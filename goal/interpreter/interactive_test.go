// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/text/terminal/pty"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setHome sets the home directory to a new temporary one for the test,
// resetting the cache of the home directory, and returns it.
func setHome(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	t.Setenv("HOME", dir)
	homedir.Reset()
	t.Cleanup(homedir.Reset)
	return dir
}

// answerTerminal copies the output of the given pseudo-terminal to the
// given writer, answering the queries of the cursor position of the line
// reader, as a terminal does.
func answerTerminal(p *pty.Pty, w io.Writer) {
	buf := make([]byte, 4096)
	for {
		n, err := p.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			if bytes.Contains(buf[:n], []byte("\x1b[6n")) {
				p.Write([]byte("\x1b[1;1R"))
			}
		}
		if err != nil {
			return
		}
	}
}

// TestInteractive tests the interactive shell in a pseudo-terminal (as
// in a terminal widget): running lines, printing their values, history
// commands, and the end of the input, which saves the history.
func TestInteractive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the console of the pseudo-terminal is attached to the process on Windows")
	}
	home := setHome(t)
	in := NewInterpreter(Options{})
	in.HistFile = filepath.Join(home, "hist")
	require.NoError(t, os.WriteFile(in.HistFile, []byte("old := 1"), 0o644))
	p, err := pty.Run(func(tin, tout *os.File) error {
		in.SetIO(tin, tout, tout)
		return in.Interactive()
	}, 80, 24)
	require.NoError(t, err)
	var out syncBuffer
	go answerTerminal(p, &out)
	waitFor := func(s string) {
		t.Helper()
		require.Eventually(t, func() bool { return strings.Contains(out.String(), s) }, 10*time.Second, 10*time.Millisecond, "%q in %q", s, out.String())
	}

	waitFor(">") // the prompt
	p.Write([]byte("x := 6 * 7\r"))
	p.Write([]byte("x\r"))
	waitFor("42")
	p.Write([]byte("!1\r")) // the history of x := 6 * 7 (after old := 1)
	waitFor("h:1")
	p.Write([]byte("!99\r"))
	waitFor("not in range")
	p.Write([]byte("!a\r"))
	waitFor("not a number")
	in.SizeChanged()
	p.Write([]byte{4}) // Control+D: the end of the input
	require.NoError(t, p.Wait())

	b, err := os.ReadFile(in.HistFile)
	require.NoError(t, err)
	assert.Contains(t, string(b), "x := 6 * 7", "the history is saved")
	in.Close()
	in.Close() // more than once
}

// TestInteractiveClosed tests that the interactive shell returns at once
// when the interpreter is closed.
func TestInteractiveClosed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the console of the pseudo-terminal is attached to the process on Windows")
	}
	setHome(t)
	in := NewInterpreter(Options{})
	in.Close()
	p, err := pty.Run(func(tin, tout *os.File) error {
		in.SetIO(tin, tout, tout)
		return in.Interactive()
	}, 80, 24)
	require.NoError(t, err)
	go answerTerminal(p, io.Discard)
	require.NoError(t, p.Wait())
}

// TestPrompt tests the prompt of the shell, in the current directory, in
// math mode, and within a block.
func TestPrompt(t *testing.T) {
	home := setHome(t)
	in := NewInterpreter(Options{})
	in.Goal.Config.Dir = filepath.Join(home, "proj")
	assert.Equal(t, filepath.Join("~", "proj")+string(filepath.Separator)+" > ", in.Prompt())
	in.Goal.TrState.MathMode = true
	assert.True(t, strings.HasSuffix(in.Prompt(), " # "), in.Prompt())
	in.Goal.TrState.MathMode = false
	in.Goal.TranspileCode("if true {")
	assert.Equal(t, ">     ", in.Prompt(), "indented within the block")
}

// TestRunConfig tests running the startup config file in the home
// directory, and interrupting the interpreter.
func TestRunConfig(t *testing.T) {
	home := setHome(t)
	var out bytes.Buffer
	in := NewInterpreter(Options{Stdout: &out, Stderr: &out})
	in.Config() // no config file
	require.NoError(t, os.WriteFile(filepath.Join(home, ".goal"), []byte(`fmt.Println("configured")`+"\n"), 0o644))
	require.NoError(t, in.RunConfig())
	assert.Contains(t, out.String(), "configured")
	in.Interrupt()
}

// TestRunFile tests running a goal file with an expression and arguments,
// as the run command does.
func TestRunFile(t *testing.T) {
	setHome(t)
	dir := t.TempDir()
	fn := filepath.Join(dir, "a.goal")
	require.NoError(t, os.WriteFile(fn, []byte("y := 5\n"), 0o644))
	var ran bool
	c := &Config{Input: fn, Expr: "z := y * 2", Args: []string{"a"}, Dir: dir,
		InteractiveFunc: func(c *Config, in *Interpreter) error { ran = true; return nil }}
	t.Chdir(dir)
	require.NoError(t, Run(c))
	assert.False(t, ran, "not interactive")
	c.Interactive = true
	require.NoError(t, Run(c))
	assert.True(t, ran, "interactive after the file")

	ran = false
	require.NoError(t, Run(&Config{InteractiveFunc: func(c *Config, in *Interpreter) error { ran = true; return nil }}))
	assert.True(t, ran, "interactive without an input")

	assert.Error(t, Run(&Config{Input: filepath.Join(dir, "bad.goal"), Expr: "undefinedFunc()"}))
}
