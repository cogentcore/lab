// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/base/sshclient"
	"github.com/mitchellh/go-homedir"
)

// Exec handles command execution for all cases, parameterized by the args.
// It executes the given command string, waiting for the command to finish,
// handling the given arguments appropriately.
// If there is any error, it adds it to the goal, and triggers CancelExecution.
//   - errOk = don't call AddError so execution will not stop on error
//   - start = calls Start on the command, which then runs asynchronously, with
//     a goroutine forked to Wait for it and close its IO
//   - output = return the output of the command as a string (otherwise return is "")
func (gl *Goal) Exec(errOk, start, output bool, cmd any, args ...any) string {
	out := ""
	if !errOk && len(gl.Errors) > 0 {
		return out
	}
	cmdIO := exec.NewCmdIO(&gl.Config)
	cmdIO.StackStart()
	if start {
		cmdIO.PushIn(nil) // no stdin for bg
	}
	cl, scmd, sargs := gl.ExecArgs(cmdIO, errOk, cmd, args...)
	if scmd == "" {
		cmdIO.PopToStart()
		return out
	}
	// The read end of the pipe of the output of this command (cmd |) goes
	// to the next command, which closes it: this command does not, as it
	// can finish before the next command has started reading.
	var pipeOut *os.File
	if len(cmdIO.PipeIn) > 0 {
		pipeOut = cmdIO.PipeIn.Pop()
		defer gl.setPipeIn(pipeOut)
	}
	if start && pipeOut != nil && cl == nil && gl.isBuiltinOrCommand(scmd) {
		gl.runBuiltinOrCommandToPipe(cmdIO, errOk, scmd, sargs...)
		return out
	}
	if scmd == "ssh" && !start && !output && len(sargs) == 1 && isHostName(sargs[0]) {
		err := gl.sshShell(cmdIO, sargs[0][1:])
		if !errOk {
			gl.AddError(err)
		}
		cmdIO.PopToStart()
		return out
	}
	if localBuiltins[scmd] { // not on the active ssh host
		cl = nil
	}
	var err error
	if cl != nil {
		switch {
		case scmd == "set":
			if len(sargs) != 2 {
				err := fmt.Errorf("expected two arguments, got %d", len(sargs))
				if !errOk {
					gl.AddError(err)
				}
				return ""
			}
			cl.SetEnv(sargs[0], sargs[1])
		case start:
			err = cl.Start(&cmdIO.StdIOState, scmd, sargs...)
		case output:
			cmdIO.PushOut(nil)
			out, err = cl.OutputContext(gl.cmdContext(), &cmdIO.StdIOState, scmd, sargs...)
		default:
			err = cl.RunContext(gl.cmdContext(), &cmdIO.StdIOState, scmd, sargs...)
		}
		if !errOk {
			gl.AddError(err)
		}
	} else {
		ran := false
		ran, out = gl.RunBuiltinOrCommand(cmdIO, errOk, start, output, scmd, sargs...)
		if !ran {
			gl.isCommand.Push(false)
			switch {
			case start:
				// fmt.Fprintf(gl.debugTrace, "start exe %s in: %#v  out: %#v  %v\n  ", scmd, cmdIO.In, cmdIO.Out, cmdIO.OutIsPipe())
				err = gl.Config.StartIO(cmdIO, scmd, sargs...)
				job := &Job{CmdIO: cmdIO, Desc: cmdIO.String()}
				jn := gl.addJob(job)
				if pipeOut == nil {
					gl.Config.StdIO.Printf("[%d]  %s\n", jn, job.Desc)
				}
				go func() {
					if cmdIO.Cmd != nil {
						cmdIO.Cmd.Wait()
					}
					cmdIO.PopToStart()
					gl.DeleteJob(job)
				}()
			case output:
				cmdIO.PushOut(nil)
				out, err = gl.Config.OutputIOContext(gl.cmdContext(), cmdIO, scmd, sargs...)
			default:
				// fmt.Fprintf(gl.debugTrace, "run exe %s in: %#v  out: %#v  %v\n  ", scmd, cmdIO.In, cmdIO.Out, cmdIO.OutIsPipe())
				err = gl.Config.RunIOContext(gl.cmdContext(), cmdIO, scmd, sargs...)
			}
			if !errOk {
				gl.AddError(err)
			}
			gl.isCommand.Pop()
		}
	}
	if !start {
		cmdIO.PopToStart()
	}
	return out
}

// isBuiltinOrCommand returns whether cmd is a builtin or a command function.
func (gl *Goal) isBuiltinOrCommand(cmd string) bool {
	_, hasCmd := gl.Commands[cmd]
	_, hasBlt := gl.Builtins[cmd]
	return hasCmd || hasBlt
}

// runBuiltinOrCommandToPipe runs the builtin or command function cmd, whose
// output is piped to the next command (cmd |). It runs here, not in the
// background, as its execs use the IO of the goal, which is set to its own
// only while it runs, with its output captured, which is then written to
// the pipe in the background, so that it does not block until the next
// command reads it, and the pipe is then closed.
func (gl *Goal) runBuiltinOrCommandToPipe(cmdIO *exec.CmdIO, errOk bool, cmd string, args ...string) {
	pipe := cmdIO.Out
	var buf bytes.Buffer
	cmdIO.PushOut(&buf)
	gl.RunBuiltinOrCommand(cmdIO, errOk, false, false, cmd, args...)
	cmdIO.PopOut()
	go func() {
		pipe.Write(buf.Bytes())
		cmdIO.PopToStart()
	}()
}

// sshShell runs an interactive shell on the ssh host of the given name,
// for ssh @name, in its current directory, until it ends (see
// [sshclient.Client.Shell]). Other uses of ssh run the ssh command.
func (gl *Goal) sshShell(cmdIO *exec.CmdIO, host string) error {
	cl, err := gl.SSHByHost(host)
	if err != nil {
		return fmt.Errorf("ssh: %w", err)
	}
	return cl.Shell(&cmdIO.StdIOState)
}

// localBuiltins are the builtins that run on the local host even when the
// commands run on an ssh host (see [Goal.SSHActive]), as they act on the
// shell itself: its ssh connections, files on several hosts, jobs, and
// the Goal code that it runs.
var localBuiltins = map[string]bool{"gossh": true, "scp": true, "exit": true, "jobs": true, "source": true}

// RunBuiltinOrCommand runs a builtin or a command, returning true if it ran,
// and the output string if running in output mode.
func (gl *Goal) RunBuiltinOrCommand(cmdIO *exec.CmdIO, errOk, start, output bool, cmd string, args ...string) (bool, string) {
	out := ""
	cmdFun, hasCmd := gl.Commands[cmd]
	bltFun, hasBlt := gl.Builtins[cmd]

	if !hasCmd && !hasBlt {
		return false, out
	}

	if hasCmd {
		gl.commandArgs.Push(args)
		gl.isCommand.Push(true)
	}

	// note: we need to set both os. and wrapper versions, so it works the same
	// in compiled vs. interpreted mode
	var oldsh, oldwrap, oldstd *exec.StdIO
	save := func() {
		oldsh = gl.Config.StdIO.Set(&cmdIO.StdIO)
		oldwrap = gl.StdIOWrappers.SetWrappers(&cmdIO.StdIO)
		oldstd = cmdIO.SetToOS()
	}

	done := func() {
		if hasCmd {
			gl.isCommand.Pop()
			gl.commandArgs.Pop()
		}
		// fmt.Fprintf(gl.debugTrace, "%s restore %#v\n", cmd, oldstd.In)
		oldstd.SetToOS()
		gl.StdIOWrappers.SetWrappers(oldwrap)
		gl.Config.StdIO = *oldsh
	}

	switch {
	case start:
		wg := sync.WaitGroup{}
		wg.Add(1)
		job := &Job{CmdIO: cmdIO, Desc: cmd}
		gl.Config.StdIO.Printf("[%d]  %s\n", gl.addJob(job), cmd)
		go func() {
			if hasCmd {
				oldwrap = gl.StdIOWrappers.SetWrappers(&cmdIO.StdIO)
				// oldstd = cmdIO.SetToOS()
				// fmt.Fprintf(gl.debugTrace, "%s oldstd in: %#v  out: %#v\n", cmd, oldstd.In, oldstd.Out)
				cmdFun(args...)
				// oldstd.SetToOS()
				gl.StdIOWrappers.SetWrappers(oldwrap)
				gl.isCommand.Pop()
				gl.commandArgs.Pop()
			} else {
				gl.AddError(bltFun(cmdIO, args...))
			}
			time.Sleep(time.Millisecond)
			wg.Done()
		}()
		go func() {
			wg.Wait()
			cmdIO.PopToStart()
			gl.DeleteJob(job)
		}()
	case output:
		save()
		obuf := &bytes.Buffer{}
		// os.Stdout = obuf // needs a file
		gl.Config.StdIO.Out = obuf
		gl.StdIOWrappers.SetWrappedOut(obuf)
		cmdIO.PushOut(obuf)
		if hasCmd {
			cmdFun(args...)
		} else {
			gl.AddError(bltFun(cmdIO, args...))
		}
		out = strings.TrimSuffix(obuf.String(), "\n")
		done()
	default:
		save()
		if hasCmd {
			cmdFun(args...)
		} else {
			gl.AddError(bltFun(cmdIO, args...))
		}
		done()
	}
	return true, out
}

func (gl *Goal) HandleArgErr(errok bool, err error) error {
	if err == nil {
		return err
	}
	if errok {
		gl.Config.StdIO.ErrPrintln(err.Error())
	} else {
		gl.AddError(err)
	}
	return err
}

// ExecArgs processes the args to given exec command,
// handling all of the input / output redirection and
// file globbing, homedir expansion, etc.
func (gl *Goal) ExecArgs(cmdIO *exec.CmdIO, errOk bool, cmd any, args ...any) (*sshclient.Client, string, []string) {
	// the output of the previous command, if piped, is the input
	if r := gl.takePipeIn(); r != nil {
		cmdIO.PushIn(r) // closed by PopToStart
	}
	scmd := reflectx.ToString(cmd)
	cl := gl.ActiveSSH()
	// isCmd := gl.isCommand.Peek()
	sargs := make([]string, 0, len(args))
	var err error

	homeDir := func(s string) string {
		if cl == nil {
			s, err = homedir.Expand(s)
			gl.HandleArgErr(errOk, err)
			// note: handling globbing in a later pass, to not clutter..
		} else {
			if s[0] == '~' {
				s = "$HOME/" + s[1:]
			}
		}
		return s
	}

	for _, a := range args {
		if sa, ok := a.([]string); ok {
			for _, s := range sa {
				if s == "" {
					continue
				}
				s = homeDir(s)
				sargs = append(sargs, s)
			}
			continue
		}
		if sa, ok := a.([]any); ok {
			for _, aa := range sa {
				s := reflectx.ToString(aa)
				if s == "" {
					continue
				}
				s = homeDir(s)
				sargs = append(sargs, s)
			}
			continue
		}
		s := reflectx.ToString(a)
		if s == "" {
			continue
		}
		s = homeDir(s)
		sargs = append(sargs, s)
	}
	if scmd[0] == '@' {
		newHost := ""
		if scmd == "@0" { // local
			cl = nil
		} else {
			hnm := scmd[1:]
			if scl, ok := gl.SSHClients[hnm]; ok {
				newHost = hnm
				cl = scl
			} else {
				gl.HandleArgErr(errOk, fmt.Errorf("goal: ssh connection named: %q not found", hnm))
			}
		}
		if len(sargs) > 0 {
			scmd = sargs[0]
			sargs = sargs[1:]
		} else { // just a ssh switch
			gl.SSHActive = newHost
			return nil, "", nil
		}
	}
	for i := 0; i < len(sargs); i++ { // we modify so no range
		s := sargs[i]
		n := len(sargs)
		switch {
		case s[0] == '>':
			sargs = gl.OutToFile(cl, cmdIO, errOk, sargs, i)
		case s[0] == '<':
			sargs = gl.InFromFile(cl, cmdIO, errOk, sargs, i)
		case s[0] == '|':
			sargs = gl.OutToPipe(cl, cmdIO, errOk, sargs, i)
		case cl == nil && strings.HasPrefix(s, "args"):
			sargs = gl.CmdArgs(errOk, sargs, i)
			i-- // back up because we consume this one
			continue
		}
		if len(sargs) < n { // redirection args consumed: check the next one here
			i--
		}
	}
	// do globbing late here so we don't have to wade through everything.
	// only for local.
	if cl == nil {
		gargs := make([]string, 0, len(sargs))
		for _, s := range sargs {
			g, err := filepath.Glob(s)
			if err != nil || len(g) == 0 { // not valid
				gargs = append(gargs, s)
			} else {
				gargs = append(gargs, g...)
			}
		}
		sargs = gargs
	}
	return cl, scmd, sargs
}

// OutToFile processes the > arg that sends output to a file
func (gl *Goal) OutToFile(cl *sshclient.Client, cmdIO *exec.CmdIO, errOk bool, sargs []string, i int) []string {
	n := len(sargs)
	s := sargs[i]
	sn := len(s)
	fn := ""
	narg := 1
	if i < n-1 {
		fn = sargs[i+1]
		narg = 2
	}
	appn := false
	errf := false
	switch {
	case sn > 1 && s[1] == '>':
		appn = true
		if sn > 2 && s[2] == '&' {
			errf = true
		}
	case sn > 1 && s[1] == '&':
		errf = true
	case sn > 1:
		fn = s[1:]
		narg = 1
	}
	if fn == "" {
		gl.HandleArgErr(errOk, fmt.Errorf("missing file name after %s", s))
		return sargs
	}
	// The file is on the host of the command, unless it is @name:file, on
	// the host of that ssh connection, or @0:file, on the local host.
	var fcl *sshclient.Client // the host of the file, if not local
	if host, p, ok := strings.Cut(fn, ":"); ok && strings.HasPrefix(host, "@") {
		fn = p
		if host != "@0" {
			hcl, err := gl.SSHByHost(host[1:])
			if err != nil {
				gl.HandleArgErr(errOk, err)
				return sargs
			}
			fcl = hcl
		}
	} else if cl != nil {
		fcl = cl
	}
	if fcl != nil && fcl == cl { // the shell of the host redirects it
		if narg == 2 {
			sargs[i+1] = fn
		} else {
			sargs[i] = ">" + fn // as >file
		}
		return sargs
	}
	sargs = slices.Delete(sargs, i, i+narg)
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if appn {
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	}
	var f io.WriteCloser
	var err error
	if fcl != nil {
		f, err = fcl.OpenFile(fn, flag)
	} else {
		f, err = os.OpenFile(fn, flag, 0644)
	}
	if err == nil {
		cmdIO.PushOut(f)
		if errf {
			cmdIO.PushErr(f)
		}
	} else {
		gl.HandleArgErr(errOk, err)
	}
	return sargs
}

// InFromFile processes the < arg that reads the input from a file,
// which is closed when the command is done.
func (gl *Goal) InFromFile(cl *sshclient.Client, cmdIO *exec.CmdIO, errOk bool, sargs []string, i int) []string {
	s := sargs[i]
	fn := ""
	narg := 1
	switch {
	case len(s) > 1:
		fn = s[1:]
	case i < len(sargs)-1:
		fn = sargs[i+1]
		narg = 2
	}
	if fn == "" {
		gl.HandleArgErr(errOk, fmt.Errorf("missing file name after <"))
		return sargs
	}
	if cl != nil {
		if !strings.HasPrefix(fn, "@0:") {
			return sargs
		}
		fn = fn[3:]
	}
	sargs = slices.Delete(sargs, i, i+narg)
	f, err := os.Open(fn)
	if err != nil {
		gl.HandleArgErr(errOk, err)
		return sargs
	}
	cmdIO.PushIn(f)
	return sargs
}

// OutToPipe processes the | arg that sends output to a pipe
func (gl *Goal) OutToPipe(cl *sshclient.Client, cmdIO *exec.CmdIO, errOk bool, sargs []string, i int) []string {
	s := sargs[i]
	sn := len(s)
	errf := false
	if sn > 1 && s[1] == '&' {
		errf = true
	}
	sargs = slices.Delete(sargs, i, i+1)
	cmdIO.PushOutPipe()
	if errf {
		cmdIO.PushErr(cmdIO.Out)
	}
	return sargs
}

// CmdArgs processes expressions involving "args" for commands
func (gl *Goal) CmdArgs(errOk bool, sargs []string, i int) []string {
	// n := len(sargs)
	// s := sargs[i]
	// sn := len(s)
	args := gl.commandArgs.Peek()

	// fmt.Println("command args:", args)

	switch {
	case sargs[i] == "args...":
		sargs = slices.Delete(sargs, i, i+1)
		sargs = slices.Insert(sargs, i, args...)
	}

	return sargs
}

// cmdContext returns the context with which the commands are run, which
// interrupts them when the execution is canceled while they run, as for
// Control+C (see [Goal.CancelExecution]). It is nil when there is none,
// or it is already done, as it is after an error ([Goal.AddError]), for
// the commands that run after it, which do not stop on errors.
func (gl *Goal) cmdContext() context.Context {
	if gl.Ctx == nil || gl.Ctx.Err() != nil {
		return nil
	}
	return gl.Ctx
}

// CancelExecution calls the Cancel() function if set.
func (gl *Goal) CancelExecution() {
	if gl.Cancel != nil {
		gl.Cancel()
	}
}
