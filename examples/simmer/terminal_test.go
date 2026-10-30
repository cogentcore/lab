// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/base/iox/jsonx"
	"cogentcore.org/core/base/sshclient/sshtest"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/mvmlab"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/require"
)

// TestTerminalServer tests that the terminal of simmer runs the goal shell,
// with the server of the project connected as @1, in the jobs directory of
// the project on it, including interactive commands and a shell on it.
func TestTerminalServer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported")
	}
	s := sshtest.NewServer(t)
	s.Pty.Store(true) // as on a real server, for the interactive commands
	t.Setenv("HOME", t.TempDir())
	homedir.Reset()
	t.Chdir(t.TempDir())
	sim := filepath.Join(t.TempDir(), "termsim")
	require.NoError(t, os.MkdirAll(filepath.Join(sim, "simdata", "labscripts"), 0755))
	in := interpreter.NewInterpreter(interpreter.Options{})
	in.HistFile = filepath.Join(t.TempDir(), "history")
	b, sr := NewSimmerWindow(in)
	goalrun.SSH.User.KeyPath = s.KeyPath
	pr := &Project{StartDir: sim}
	pr.Defaults()
	pr.Update() // sets Server.Root
	pr.Server.SSH = s.Addr
	require.NoError(t, jsonx.Save(pr, filepath.Join(sim, ProjectFilename)))
	// the commands on the server start in its jobs directory
	s0 := filepath.Join(s.Root, pr.Server.Root, "jobs")
	require.NoError(t, os.MkdirAll(filepath.Join(s0, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(s0, "onserver.txt"), []byte("hello from server"), 0644))
	t.Cleanup(goalrun.CloseSSH)
	sr.OpenProject(core.Filename(sim))
	require.True(t, sr.isOpen())
	mvmlab.InteractiveTerminal(nil, in, sr.Terminal)
	tm := sr.Terminal
	showed := make(chan struct{})
	b.OnFinal(events.Show, func(e events.Event) { showed <- struct{}{} })
	b.RunWindow()
	<-showed
	defer func() {
		b.AsyncLock()
		b.Close()
		b.AsyncUnlock()
	}()
	text := func() string {
		b.AsyncLock()
		defer b.AsyncUnlock()
		return tm.LinesView().Lines.String()
	}
	waitFor := func(what string, cond func(s string) bool) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			if cond(text()) {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Logf("terminal text:\n%s", text())
		t.Fatalf("timeout waiting for %s", what)
	}
	run := func(cmd string) {
		b.AsyncLock()
		tm.Run(cmd)
		b.AsyncUnlock()
	}
	waitFor("the prompt", func(s string) bool { return strings.Contains(s, "> ") })
	// the input of the terminal goes to the command on the server while it
	// runs, so the prompt after it is waited for
	prompt := func(n int) func(s string) bool {
		return func(s string) bool { return strings.Count(s, "> ") >= n }
	}
	run("@1 cat onserver.txt")
	waitFor("the server output", func(s string) bool { return strings.Contains(s, "hello from server") })
	waitFor("the prompt after it", prompt(2))

	// an interactive command on the server gets the input of the terminal
	run("@1 head -1")
	time.Sleep(300 * time.Millisecond)
	run("to the server")
	waitFor("the prompt after head", prompt(3))
	run(`fmt.Println("after", "head")`)
	waitFor("the output after head", func(s string) bool { return strings.Contains(s, "after head\n") })

	// an interactive shell on the server, in the directory of @1
	run("@1 cd sub")
	waitFor("the prompt after cd", prompt(5))
	run("ssh @1")
	time.Sleep(300 * time.Millisecond)
	run("pwd")
	waitFor("the directory in the shell", func(s string) bool { return strings.Contains(s, filepath.Join(s0, "sub")+"\n") })
	run("exit")
	waitFor("the prompt after the shell", prompt(6))
	run(`fmt.Println("after", "shell")`)
	waitFor("the output after the shell", func(s string) bool { return strings.Contains(s, "after shell\n") })
	t.Log(text())
	b.AsyncLock()
	tm.SendText("\x04")
	b.AsyncUnlock()
	waitFor("the end", func(s string) bool { return strings.Contains(s, "[process exited]") })
}
