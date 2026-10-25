// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer is a bytes.Buffer that is safe for concurrent writes, as the
// commands of a pipe make.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// shellTest runs the given lines of Goal code in a new interpreter, in a
// new current directory, each as the interactive shell runs a line, with
// the given standard input, and returns what they print, and the errors of
// the lines. It checks that the standard IO of the process and of the
// interpreter are restored after each line.
func shellTest(t *testing.T, stdin string, lines ...string) (string, []error) {
	t.Helper()
	t.Chdir(t.TempDir())
	osIn, osOut, osErr := os.Stdin, os.Stdout, os.Stderr
	var out syncBuffer
	in := NewInterpreter(Options{Stdout: &out, Stderr: &out, Stdin: strings.NewReader(stdin)})
	var errs []error
	for _, line := range lines {
		if _, _, err := in.Eval(line); err != nil {
			errs = append(errs, err)
		}
		if os.Stdin != osIn || os.Stdout != osOut || os.Stderr != osErr {
			t.Errorf("%q: the standard IO of the process is not restored", line)
		}
		cur, orig := in.Goal.Config.StdIO, in.Goal.OrigStdIO
		if cur.Out != orig.Out || cur.Err != orig.Err || cur.In != orig.In {
			t.Errorf("%q: the standard IO of the interpreter is not restored", line)
		}
	}
	return out.String(), errs
}

// TestShell runs shell commands of Goal, with their redirection of input and
// output, pipes, and commands defined in Goal (see docs/content/shell.md),
// as the interactive shell runs them, checking what they print.
func TestShell(t *testing.T) {
	hello := "command hello {\n\techo hello args...\n}"
	upper := "command upper {\n\ttr a-z A-Z\n}"
	goPrint := "command gp {\n\tfmt.Println(\"from go\", args)\n}"
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{"echo", []string{"echo hi"}, "hi\n"},
		{"go-expr-arg", []string{"echo {1+2}"}, "3\n"},
		{"multiple", []string{"echo a; echo b"}, "a\nb\n"},
		{"go", []string{"fmt.Println(1 + 2)"}, "3\n"},
		{"stdin", []string{"cat"}, "from stdin\n"},
		{"cd", []string{"mkdir sub", "cd sub", "fmt.Println(filepath.Base($pwd$))"}, "sub\n"},
		{"set-env", []string{"set FOO bar", "printenv FOO"}, "bar\n"},
		{"optional", []string{"[false]", "echo ok"}, "ok\n"},

		{"out-file", []string{"echo a b > f.txt", "cat f.txt"}, "a b\n"},
		{"append-file", []string{"echo a > f.txt", "echo b >> f.txt", "cat f.txt"}, "a\nb\n"},
		{"err-file", []string{"[ls nonexist >& f.txt]", "grep -c nonexist < f.txt"}, "1\n"},
		{"in-file", []string{"echo abc > f.txt", "cat < f.txt"}, "abc\n"},
		{"in-out-file", []string{"echo abc > f.txt", "cat < f.txt > g.txt", "cat g.txt"}, "abc\n"},

		{"pipe", []string{"echo abc | cat"}, "abc\n"},
		{"pipe-3", []string{"echo abc | tr a-z A-Z | cat"}, "ABC\n"},
		{"pipe-err", []string{"[ls nonexist |& grep -c nonexist]"}, "1\n"},
		{"pipe-in-file", []string{"echo abc > f.txt", "cat < f.txt | tr a-z A-Z"}, "ABC\n"},
		{"pipe-out-file", []string{"echo abc | tr a-z A-Z > f.txt", "cat f.txt"}, "ABC\n"},
		{"pipe-repeat", []string{"echo a | cat", "echo b | cat", "echo c | cat"}, "a\nb\nc\n"},

		{"capture", []string{"x := $echo hi$", "fmt.Println(x + \"!\")"}, "hi!\n"},
		{"capture-pipe", []string{"x := $echo abc | tr a-z A-Z$", "fmt.Println(x + \"!\")"}, "ABC!\n"},

		{"command", []string{hello, "hello world"}, "hello world\n"},
		{"command-file", []string{hello, "hello world > f.txt", "cat f.txt"}, "hello world\n"},
		{"command-capture", []string{hello, "y := $hello world$", "fmt.Println(y + \"!\")"}, "hello world!\n"},
		{"command-pipe-out", []string{hello, "hello world | tr a-z A-Z"}, "HELLO WORLD\n"},
		{"command-pipe-in", []string{upper, "echo abc | upper"}, "ABC\n"},
		{"command-pipe-middle", []string{upper, "echo abc | upper | cat"}, "ABC\n"},
		{"command-in-file", []string{upper, "echo abc > f.txt", "upper < f.txt"}, "ABC\n"},
		{"go-command-file", []string{goPrint, "gp a > f.txt", "cat f.txt"}, "from go [a]\n"},
		{"go-command-pipe", []string{goPrint, "gp a | tr a-z A-Z"}, "FROM GO [A]\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, errs := shellTest(t, "from stdin\n", test.lines...)
			if len(errs) > 0 {
				t.Errorf("errors: %v", errs)
			}
			if got != test.want {
				t.Errorf("output %q, want %q", got, test.want)
			}
		})
	}
}

// TestShellErrors checks that a failing command is an error of its line,
// unless it is optional ([ ]), and that it stops the commands after it in
// the same code, as in a script.
func TestShellErrors(t *testing.T) {
	got, errs := shellTest(t, "", "false", "echo after", "[false]", "ls nonexist > f.txt", "cat < nonexist")
	if len(errs) != 3 {
		t.Errorf("%d errors, want 3: %v", len(errs), errs)
	}
	if !strings.HasPrefix(got, "after\n") {
		t.Errorf("output %q", got)
	}

	got, errs = shellTest(t, "", "false\necho skipped")
	if len(errs) != 1 || strings.Contains(got, "skipped") {
		t.Errorf("errors %v, output %q, want an error and no output", errs, got)
	}
	got, errs = shellTest(t, "", "[false]\necho ran")
	if len(errs) != 0 || got != "ran\n" {
		t.Errorf("errors %v, output %q, want no error and ran", errs, got)
	}
}

// TestShellBackground checks a command run in the background (&), which
// jobs lists while it runs, and which the next commands do not wait for.
func TestShellBackground(t *testing.T) {
	t.Chdir(t.TempDir())
	var out syncBuffer
	in := NewInterpreter(Options{Stdout: &out, Stderr: &out})
	for _, line := range []string{"sleep 0.5 &", "jobs", "echo next"} {
		if _, _, err := in.Eval(line); err != nil {
			t.Fatalf("%s: %v", line, err)
		}
	}
	// the job is listed when it starts, and by jobs: [1]  <pid>  <path>
	job := regexp.MustCompile(`^\[1\]  \d+ .*sleep 0\.5\n`)
	got := out.String()
	for range 2 {
		loc := job.FindStringIndex(got)
		if loc == nil {
			t.Fatalf("output %q: no job line", out.String())
		}
		got = got[loc[1]:]
	}
	if got != "next\n" {
		t.Errorf("output %q: no next", out.String())
	}
	for range 100 {
		if in.Goal.NumJobs() == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("the job did not finish")
}
