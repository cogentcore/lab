// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mvmlab

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/terminal"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
)

// TestInteractiveTerminal tests the goal shell in a terminal, with the
// data of the app in tensorfs, completing its names, interrupting code
// and commands, and ending and restarting the shell.
func TestInteractiveTerminal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported")
	}
	tensorfs.Chdir("") // the root
	dir := tensorfs.Mkdir("TermData")
	tensorfs.Value[float64](dir, "Answer", 1).SetFloat1D(42, 0)

	in := interpreter.NewInterpreter(interpreter.Options{})
	in.HistFile = filepath.Join(t.TempDir(), "history")
	UseGUI(in)
	b := core.NewBody()
	tm := terminal.NewTerminal(b)
	tm.Styler(func(s *styles.Style) {
		s.Min.Set(units.Ch(80), units.Em(20))
	})
	InteractiveTerminal(&interpreter.Config{Expr: `startup := "from expr"`}, in, tm)
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
		return tm.LinesView().Lines.String()
	}
	waitFor := func(what string, cond func(s string) bool) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			b.AsyncLock()
			s := text()
			b.AsyncUnlock()
			if cond(s) {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		b.AsyncLock()
		t.Logf("terminal text:\n%s", text())
		b.AsyncUnlock()
		t.Fatalf("timeout waiting for %s", what)
	}
	run := func(cmd string) {
		b.AsyncLock()
		tm.Run(cmd)
		b.AsyncUnlock()
	}
	waitFor("the prompt", func(s string) bool { return strings.Contains(s, "> ") })

	// the data of the app, in math mode
	run("##")
	waitFor("math mode", func(s string) bool { return strings.Contains(s, "data # ") })
	b.AsyncLock()
	tm.SendText("cd TermD\t") // completes the tensorfs directory
	b.AsyncUnlock()
	waitFor("the completion", func(s string) bool { return strings.Contains(s, "cd TermData/") })
	run("")
	run("ls")
	waitFor("the listing", func(s string) bool { return strings.Contains(s, "Answer") })
	run("##")

	// Go code, with the variable from the expression of the config
	run(`fmt.Println(startup, tensorfs.Get("Answer").Float1D(0))`)
	waitFor("the output", func(s string) bool { return strings.Contains(s, "from expr 42\n") })

	// Control+C interrupts code that is running
	run("n := 0; for { n++ }")
	time.Sleep(200 * time.Millisecond)
	b.AsyncLock()
	tm.SendText("\x03")
	b.AsyncUnlock()
	run(`fmt.Println("after", "interrupt")`)
	waitFor("the interrupt", func(s string) bool { return strings.Contains(s, "after interrupt\n") })

	// and commands that are running
	start := time.Now()
	run("sleep 10")
	time.Sleep(300 * time.Millisecond)
	b.AsyncLock()
	tm.SendText("\x03")
	b.AsyncUnlock()
	run(`fmt.Println("after", "sleep")`)
	waitFor("the interrupt of the command", func(s string) bool { return strings.Contains(s, "after sleep\n") })
	assert.Less(t, time.Since(start), 5*time.Second)

	// the shell ends with Control+D, and runs again when restarted
	b.AsyncLock()
	tm.SendText("\x04")
	b.AsyncUnlock()
	waitFor("the end", func(s string) bool { return strings.Contains(s, "[process exited]") })
	b.AsyncLock()
	assert.False(t, tm.IsRunning())
	tm.Restart()
	b.AsyncUnlock()
	run(`fmt.Println("again", startup)`)
	waitFor("the restart", func(s string) bool { return strings.Contains(s, "again from expr\n") })

	// it ends before the window closes, so that it restores the standard
	// input and output of the process before the next test
	b.AsyncLock()
	tm.SendText("\x04")
	b.AsyncUnlock()
	waitFor("the second end", func(s string) bool { return strings.Count(s, "[process exited]") == 2 })
}
