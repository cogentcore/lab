// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/mvmcore/basesymbols"
	"cogentcore.org/lab/goal"
	"cogentcore.org/lab/goal/transpile"
	"cogentcore.org/lab/mvmlab/tensorsymbols"
	_ "cogentcore.org/lab/stats/metric"
	_ "cogentcore.org/lab/stats/stats"
	_ "cogentcore.org/lab/tensor/tmath"
	"cogentcore.org/lab/tensorfs"
	"github.com/cogentcore/mvm/interp"
	"github.com/cogentcore/mvm/lang/golang"
	"github.com/cogentcore/mvm/vm"
	"github.com/cogentcore/mvm/stdlib"
	_ "github.com/cogentcore/mvm/stdlib/core" // the core standard library
	"github.com/cogentcore/readline"
)

// Interpreter represents one running shell context
type Interpreter struct {
	// the goal shell
	Goal *goal.Goal

	// HistFile is the name of the history file to open / save.
	// Defaults to ~/.goal-history for the default goal shell.
	// Update this prior to running Config() to take effect.
	HistFile string

	// the mvm interpreter
	Interp *interp.Interp

	// readline is the line reader of [Interpreter.Interactive], while it
	// runs, and closed is whether [Interpreter.Close] was called;
	// readlineMu protects both.
	readline   *readline.Instance
	closed     bool
	readlineMu sync.Mutex
}

// Options are the options of a new [Interpreter]: its standard input and
// output, which are those of the process if nil.
type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// NewInterpreter returns a new [Interpreter] initialized with the given options.
// It automatically imports the standard library, the base Cogent Core and
// Cogent Lab packages, and configures necessary shell functions.
// End user app must call [Interpreter.Config] after importing any additional
// symbols with [Interpreter.Use], prior to running the interpreter.
func NewInterpreter(options Options) *Interpreter {
	in := &Interpreter{HistFile: "~/.goal-history"}
	in.Goal = goal.NewGoal()
	if options.Stdin != nil {
		in.Goal.Config.StdIO.In = options.Stdin
	}
	if options.Stdout != nil {
		in.Goal.Config.StdIO.Out = options.Stdout
	}
	if options.Stderr != nil {
		in.Goal.Config.StdIO.Err = options.Stderr
	}
	in.Goal.SaveOrigStdIO()
	in.Interp = interp.NewInterpreter(golang.GoSpec)
	in.Interp.SetIO(in.Goal.StdIOWrappers.In, in.Goal.StdIOWrappers.Out, in.Goal.StdIOWrappers.Err)
	basesymbols.Load()
	tensorsymbols.Load()
	in.Use(stdlib.Values, basesymbols.Values, tensorsymbols.Values, in.goalrunValues())
	go in.MonitorSignals()
	return in
}

// Use imports the symbols of the given packages, by import path and name,
// into the interpreter, such that they can be used without an import, by
// their package names (see [basesymbols.Import]).
func (in *Interpreter) Use(values ...map[string]map[string]reflect.Value) {
	basesymbols.Import(in.Interp, values...)
}

// Prompt returns the appropriate REPL prompt to show the user.
func (in *Interpreter) Prompt() string {
	dp := in.Goal.TrState.TotalDepth()
	pc := ">"
	dir := in.Goal.HostAndDir()
	if in.Goal.TrState.MathMode {
		pc = "#"
		dir = tensorfs.CurDir.Path()
	}
	if dp == 0 {
		return dir + " " + pc + " "
	}
	res := pc + " "
	for range dp {
		res += "    " // note: /t confuses readline
	}
	return res
}

// Eval evaluates (interprets) the given code,
// returning the value returned from the interpreter.
// HasPrint indicates whether the last line of code
// has the string print in it, which is for determining
// whether to print the result in interactive mode.
// It automatically logs any error in addition to returning it.
func (in *Interpreter) Eval(code string) (v reflect.Value, hasPrint bool, err error) {
	in.Goal.TranspileCode(code)
	source := false
	if in.Goal.SSHActive == "" {
		source = strings.HasPrefix(code, "source")
	}
	if in.Goal.TrState.TotalDepth() == 0 {
		nl := len(in.Goal.TrState.Lines)
		if nl > 0 {
			ln := in.Goal.TrState.Lines[nl-1]
			if strings.Contains(strings.ToLower(ln), "print") {
				hasPrint = true
			}
		}
		v, err = in.RunCode()
		in.Goal.Errors = nil
	}
	if source {
		v, err = in.RunCode() // run accumulated code
	}
	return
}

// EvalCode evaluates (interprets) the given complete code, such as that of
// a file or a code block, as [Interpreter.Eval]: unlike Eval, for which
// a bracket that is not closed continues the code on the next line,
// as in an interactive shell, that is an error.
func (in *Interpreter) EvalCode(code string) (v reflect.Value, hasPrint bool, err error) {
	v, hasPrint, err = in.Eval(code)
	st := &in.Goal.TrState
	if err == nil && st.TotalDepth() != 0 {
		err = st.UnclosedError()
		if err == nil {
			err = st.DepthError()
		}
		st.ResetCode()
		st.ResetDepth()
		st.Errors = nil
	}
	return
}

// RunCode runs the accumulated set of code lines
// and clears the stack of code lines.
// It automatically logs any error in addition to returning it.
func (in *Interpreter) RunCode() (reflect.Value, error) {
	st := &in.Goal.TrState
	if len(st.Errors) > 0 { // errors of transpiling the code
		errs := st.Errors
		st.Errors = nil
		st.ResetCode()
		return reflect.Value{}, errors.Join(errs...)
	}
	if len(in.Goal.Errors) > 0 {
		return reflect.Value{}, errors.Join(in.Goal.Errors...)
	}
	st.AddChunk()
	code, sources := st.Chunks, st.ChunkSources
	st.ResetCode()
	// The errors that the code logs stop it, as errors of it.
	prevHook := errors.SetLogHook(logHook)
	defer errors.SetLogHook(prevHook)
	var v reflect.Value
	var err error
	for ci, ch := range code {
		// mvm takes empty source as the name of a package to run.
		if strings.TrimSpace(ch) == "" {
			continue
		}
		in.Goal.StartContext()
		v, err = in.Interp.Eval("goal", ch)
		in.Goal.EndContext()
		if err != nil {
			err = in.sourceError(err, sources, ci)
			cancelled := errors.Is(err, context.Canceled)
			// fmt.Println("cancelled:", cancelled)
			in.Goal.DeleteAllJobs()
			in.Goal.RestoreOrigStdIO()
			in.Goal.TrState.ResetDepth()
			if !cancelled {
				in.Goal.AddError(err)
			} else {
				in.Goal.Errors = nil
			}
			break
		}
		// the errors of the commands that ran stop the code
		if len(in.Goal.Errors) > 0 {
			err = errors.Join(in.Goal.Errors...)
			break
		}
	}
	return v, err
}

// sourceError returns the given error of running the Go code of the
// given chunk as an error at its position in the Goal source code
// (see [transpile.SourceError]), if it has a position.
func (in *Interpreter) sourceError(err error, sources [][]transpile.SourceLine, chunk int) error {
	d, ok := in.Interp.Diagnose(err)
	if !ok || d.Source != "goal" {
		return err
	}
	var le *LogError
	switch {
	case errors.As(err, &le): // logged by the Go code that the code called
		d.Msg = le.Error()
	case errors.Is(err, vm.ErrInterrupted): // by Ctrl-C
		d.Msg = "interrupted"
	}
	return transpile.SourceError(sources, chunk, d.Line, d.Col, d.Msg, err)
}

// RunConfig runs the .goal startup config file in the user's
// home directory if it exists.
func (in *Interpreter) RunConfig() error {
	err := in.Goal.TranspileConfig()
	if err != nil {
		errors.Log(err)
	}
	_, err = in.RunCode()
	return err
}

// MonitorSignals monitors the operating system signals to appropriately
// stop the interpreter and prevent the shell from closing on Control+C.
// It is called automatically in another goroutine in [NewInterpreter].
func (in *Interpreter) MonitorSignals() {
	c := make(chan os.Signal, 1)
	// todo: syscall.SIGSEGV not defined on web
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	for {
		<-c
		in.Goal.CancelExecution()
		in.Interp.Interrupt() // the Go code that is running
	}
}

// Config performs final configuration after all the imports have been Use'd:
// it runs the user's .goal startup config file.
func (in *Interpreter) Config() {
	in.RunConfig()
}

// OpenHistory opens history from the current HistFile
// and loads it into the readline history for given rl instance
func (in *Interpreter) OpenHistory(rl *readline.Instance) error {
	err := in.Goal.OpenHistory(in.HistFile)
	if err == nil {
		for _, h := range in.Goal.Hist {
			rl.SaveToHistory(h)
		}
	}
	return err
}

// SaveHistory saves last 500 (or HISTFILESIZE env value) lines of history,
// to the current HistFile.
func (in *Interpreter) SaveHistory() error {
	n := 500
	if hfs := os.Getenv("HISTFILESIZE"); hfs != "" {
		en, err := strconv.Atoi(hfs)
		if err != nil {
			in.Goal.Config.StdIO.ErrPrintf("SaveHistory: environment variable HISTFILESIZE: %q not a number: %s", hfs, err.Error())
		} else {
			n = en
		}
	}
	return in.Goal.SaveHistory(n, in.HistFile)
}

// setReadline sets the line reader of [Interpreter.Interactive], returning
// false if the interpreter is already closed.
func (in *Interpreter) setReadline(rl *readline.Instance) bool {
	in.readlineMu.Lock()
	defer in.readlineMu.Unlock()
	if in.closed {
		return false
	}
	in.readline = rl
	return true
}

// isClosed returns whether [Interpreter.Close] was called.
func (in *Interpreter) isClosed() bool {
	in.readlineMu.Lock()
	defer in.readlineMu.Unlock()
	return in.closed
}

// Close stops [Interpreter.Interactive], restoring the terminal from the
// raw mode in which it reads a line, which is otherwise left on if the
// program exits while it reads one, as a GUI app does when its window is
// closed (see cogentcore.org/lab/mvmlab.UseGUI, which calls Close when the
// app quits). It can be called from any goroutine, more than once.
func (in *Interpreter) Close() {
	in.readlineMu.Lock()
	in.closed = true
	rl := in.readline
	in.readline = nil
	in.readlineMu.Unlock()
	if rl != nil {
		rl.Close()
	}
}

// Interactive runs an interactive shell that allows the user to input goal.
// Must have done in.Config() prior to calling.
func (in *Interpreter) Interactive() error {
	in.Goal.TrState.MathRecord = true
	rl, err := readline.NewFromConfig(&readline.Config{
		AutoComplete: &goal.ReadlineCompleter{Goal: in.Goal},
		Undo:         true,
	})
	if err != nil {
		return err
	}
	if !in.setReadline(rl) {
		rl.Close()
		return nil
	}
	in.OpenHistory(rl)
	defer in.Close()
	log.SetOutput(rl.Stderr()) // redraw the prompt correctly after log output
	defer log.SetOutput(os.Stderr)

	for {
		rl.SetPrompt(in.Prompt())
		line, err := rl.ReadLine()
		if in.isClosed() {
			in.SaveHistory()
			return nil
		}
		if errors.Is(err, readline.ErrInterrupt) {
			continue
		}
		if errors.Is(err, io.EOF) {
			in.SaveHistory()
			os.Exit(0)
		}
		if err != nil {
			in.SaveHistory()
			return err
		}
		if len(line) > 0 && line[0] == '!' { // history command
			hl, err := strconv.Atoi(line[1:])
			nh := len(in.Goal.Hist)
			if err != nil {
				in.Goal.Config.StdIO.ErrPrintf("history number: %q not a number: %s", line[1:], err.Error())
				line = ""
			} else if hl >= nh {
				in.Goal.Config.StdIO.ErrPrintf("history number: %d not in range: [0:%d]", hl, nh)
				line = ""
			} else {
				line = in.Goal.Hist[hl]
				fmt.Printf("h:%d\t%s\n", hl, line)
			}
		} else if line != "" && !strings.HasPrefix(line, "history") && line != "h" {
			in.Goal.AddHistory(line)
		}
		in.Goal.Errors = nil
		v, hasPrint, err := in.Eval(line)
		if err == nil && !hasPrint && v.IsValid() && !v.IsZero() && v.Kind() != reflect.Func {
			in.Goal.Config.StdIO.Println(v.Interface())
		}
	}
}
