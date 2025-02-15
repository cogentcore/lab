// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"errors"
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"strings"

	"cogentcore.org/core/base/logx"
	"cogentcore.org/core/base/num"
	"cogentcore.org/core/base/stringsx"
	"golang.org/x/tools/imports"
)

// State holds the transpiling state
type State struct {
	// FuncToVar translates function definitions into variable definitions,
	// which is the default for interactive use of random code fragments
	// without the complete go formatting.
	// For pure transpiling of a complete codebase with full proper Go formatting
	// this should be turned off.
	FuncToVar bool

	// MathMode is on when math mode is turned on.
	MathMode bool

	// MathRecord is state of auto-recording of data into current directory
	// in math mode.
	MathRecord bool

	// depth of delim at the end of the current line. if 0, was complete.
	ParenDepth, BraceDepth, BrackDepth, TypeDepth, DeclDepth int

	// Chunks of code lines that are accumulated during Transpile,
	// each of which should be evaluated separately, to avoid
	// issues with contextual effects from import, package etc.
	Chunks []string

	// current stack of transpiled lines, that are accumulated into
	// code Chunks.
	Lines []string

	// stack of runtime errors.
	Errors []error

	// if this is non-empty, it is the name of the last command defined.
	// triggers insertion of the AddCommand call to add to list of defined commands.
	lastCommand string
}

// NewState returns a new transpiling state; mostly for testing
func NewState() *State {
	st := &State{FuncToVar: true}
	return st
}

// TranspileCode processes each line of given code,
// adding the results to the LineStack
func (st *State) TranspileCode(code string) {
	lns := strings.Split(code, "\n")
	n := len(lns)
	if n == 0 {
		return
	}
	for _, ln := range lns {
		hasDecl := st.DeclDepth > 0
		tl := st.TranspileLine(ln)
		st.AddLine(tl)
		if st.BraceDepth == 0 && st.BrackDepth == 0 && st.ParenDepth == 1 && st.lastCommand != "" {
			st.lastCommand = ""
			nl := len(st.Lines)
			st.Lines[nl-1] = st.Lines[nl-1] + ")"
			st.ParenDepth--
		}
		if hasDecl && st.DeclDepth == 0 { // break at decl
			st.AddChunk()
		}
	}
}

// TranspileFile transpiles the given input goal file to the
// given output Go file. If no existing package declaration
// is found, then package main and func main declarations are
// added. This also affects how functions are interpreted.
func (st *State) TranspileFile(in string, out string) error {
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	code := string(b)
	lns := stringsx.SplitLines(code)
	hasPackage := false
	for _, ln := range lns {
		if strings.HasPrefix(ln, "package ") {
			hasPackage = true
			break
		}
	}
	if hasPackage {
		st.FuncToVar = false // use raw functions
	}
	st.TranspileCode(code)
	st.FuncToVar = true
	if err != nil {
		return err
	}

	hdr := `package main
import (
	"cogentcore.org/lab/goal"
	"cogentcore.org/lab/goal/goalib"
	"cogentcore.org/lab/tensor"
	_ "cogentcore.org/lab/tensor/tmath"
	_ "cogentcore.org/lab/stats/stats"
	_ "cogentcore.org/lab/stats/metric"
)

func main() {
	goalrun := goal.NewGoal()
	_ = goalrun
`

	src := st.Code()
	res := []byte(src)
	bsrc := res
	gen := fmt.Sprintf("// Code generated by \"goal build\"; DO NOT EDIT.\n//line %s:1\n", in)
	if hasPackage {
		bsrc = []byte(gen + src)
		res, err = format.Source(bsrc)
	} else {
		bsrc = []byte(gen + hdr + src + "\n}")
		res, err = imports.Process(out, bsrc, nil)
	}
	if err != nil {
		res = bsrc
		fmt.Println(err.Error())
	} else {
		err = st.DepthError()
	}
	werr := os.WriteFile(out, res, 0666)
	return errors.Join(err, werr)
}

// TotalDepth returns the sum of any unresolved paren, brace, or bracket depths.
func (st *State) TotalDepth() int {
	return num.Abs(st.ParenDepth) + num.Abs(st.BraceDepth) + num.Abs(st.BrackDepth)
}

// ResetCode resets the stack of transpiled code
func (st *State) ResetCode() {
	st.Chunks = nil
	st.Lines = nil
}

// ResetDepth resets the current depths to 0
func (st *State) ResetDepth() {
	st.ParenDepth, st.BraceDepth, st.BrackDepth, st.TypeDepth, st.DeclDepth = 0, 0, 0, 0, 0
}

// DepthError reports an error if any of the parsing depths are not zero,
// to be called at the end of transpiling a complete block of code.
func (st *State) DepthError() error {
	if st.TotalDepth() == 0 {
		return nil
	}
	str := ""
	if st.ParenDepth != 0 {
		str += fmt.Sprintf("Incomplete parentheses (), remaining depth: %d\n", st.ParenDepth)
	}
	if st.BraceDepth != 0 {
		str += fmt.Sprintf("Incomplete braces [], remaining depth: %d\n", st.BraceDepth)
	}
	if st.BrackDepth != 0 {
		str += fmt.Sprintf("Incomplete brackets {}, remaining depth: %d\n", st.BrackDepth)
	}
	if str != "" {
		slog.Error(str)
		return errors.New(str)
	}
	return nil
}

// AddLine adds line on the stack
func (st *State) AddLine(ln string) {
	st.Lines = append(st.Lines, ln)
}

// Code returns the current transpiled lines,
// split into chunks that should be compiled separately.
func (st *State) Code() string {
	st.AddChunk()
	if len(st.Chunks) == 0 {
		return ""
	}
	return strings.Join(st.Chunks, "\n")
}

// AddChunk adds current lines into a chunk of code
// that should be compiled separately.
func (st *State) AddChunk() {
	if len(st.Lines) == 0 {
		return
	}
	st.Chunks = append(st.Chunks, strings.Join(st.Lines, "\n"))
	st.Lines = nil
}

// AddError adds the given error to the error stack if it is non-nil,
// and calls the Cancel function if set, to stop execution.
// This is the main way that goal errors are handled.
// It also prints the error.
func (st *State) AddError(err error) error {
	if err == nil {
		return nil
	}
	st.Errors = append(st.Errors, err)
	logx.PrintlnError(err)
	return err
}
