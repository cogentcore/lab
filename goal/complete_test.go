// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cogentcore.org/core/base/sshclient"
	"cogentcore.org/core/base/sshclient/sshtest"
	"cogentcore.org/lab/tensorfs"
)

// TestComplete checks the completions of file names with spaces and other
// chars that the shell gives a meaning, unquoted, escaped and quoted, and
// of command names, and that the completed word is read as the name.
func TestComplete(t *testing.T) {
	dir := t.TempDir()
	for _, d := range []string{"My Documents", "plain"} {
		if err := os.Mkdir(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"My File.txt", "a{b}.txt", ".hidden", "My Documents/inner file.txt", "plain/x.go"} {
		if err := os.WriteFile(filepath.Join(dir, f), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gl := NewGoal()
	gl.Config.Dir = dir

	tests := []struct {
		line    string
		adds    []string
		baseLen int
	}{
		{`ls My`, []string{`\ Documents/`, `\ File.txt `}, 2},
		{`ls My\ D`, []string{`ocuments/`}, 5},
		{`ls "My D`, []string{`ocuments/`}, 5},
		{`ls "My F`, []string{`ile.txt" `}, 5},
		{`ls 'My F`, []string{`ile.txt' `}, 5},
		{`ls My\ Documents/`, []string{`inner\ file.txt `}, 0},
		{`ls "My Documents/in`, []string{`ner file.txt" `}, 2},
		{`cat < My\ Documents/inner\ `, []string{`file.txt `}, 7},
		{`ls a`, []string{`\{b\}.txt `}, 1},
		{`ls .h`, []string{`idden `}, 2},
		{`ls pl`, []string{`ain/`}, 2},
		{`jo`, []string{`bs `}, 2},
		{`echo hi; jo`, []string{`bs `}, 2},
		{`ls jo`, nil, 2},
	}
	for _, test := range tests {
		w, comps := gl.completions(test.line)
		var adds []string
		for _, c := range comps {
			adds = append(adds, c.add)
		}
		if !slices.Equal(adds, test.adds) || w.baseLen != test.baseLen {
			t.Errorf("%s: completions %q, base %d, want %q, %d", test.line, adds, w.baseLen, test.adds, test.baseLen)
			continue
		}
		// the completed word is read as the path of the name
		for _, c := range comps {
			got := lastShellWord(strings.TrimSuffix(test.line+c.add, " ")).value
			want := filepath.Join(filepath.Dir(w.value+"x"), c.name)
			if strings.HasSuffix(c.add, "/") {
				want += "/"
			}
			if strings.HasPrefix(want, "./") {
				want = want[2:]
			}
			if c.icon != "terminal" && got != want {
				t.Errorf("%s + %s: read as %q, want %q", test.line, c.add, got, want)
			}
		}
	}

	// readline gets the same
	rc := &ReadlineCompleter{Goal: gl}
	line := []rune(`ls "My F`)
	newLine, n := rc.Do(line, len(line))
	if len(newLine) != 1 || string(newLine[0]) != `ile.txt" ` || n != 5 {
		t.Errorf("readline: %q %d", newLine, n)
	}
}

// TestCompleteMath checks the completions of math mode, which are of the
// tensorfs nodes and commands.
func TestCompleteMath(t *testing.T) {
	tensorfs.Chdir("")
	defer tensorfs.Chdir("")
	dir := tensorfs.Mkdir("CompData")
	tensorfs.Value[float64](dir, "Answer", 1)
	tensorfs.Value[float64](dir, `A"b`, 1)
	tensorfs.Value[float64](dir.Dir("Sub Dir"), "Inner", 1)
	tensorfs.Chdir("CompData")
	gl := NewGoal()

	tests := []struct {
		line    string
		adds    []string
		baseLen int
	}{
		{`l`, []string{`s `}, 1},
		{`cd S`, []string{`ub\ Dir/`}, 1},
		{`ls `, []string{`Sub\ Dir/`}, 0},
		{`cd "Sub Dir/`, nil, 0}, // no directories in it
		{`x := get("A`, []string{`\"b"`, `nswer"`}, 1},
		{`x := get("Sub Dir/I`, []string{`nner"`}, 1},
		{"x := get(`A", []string{`"b` + "`", "nswer`"}, 1},
		{`x := get("a\"b", "A`, []string{`\"b"`, `nswer"`}, 1},
		{`x := 'A'; ls ../Comp`, []string{`Data/`}, 4},
		{`x := An`, nil, 2},
	}
	gl.TrState.MathMode = true
	for _, test := range tests {
		w, comps := gl.completions(test.line)
		var adds []string
		for _, c := range comps {
			adds = append(adds, c.add)
		}
		if !slices.Equal(adds, test.adds) || w.baseLen != test.baseLen {
			t.Errorf("%s: completions %q, base %d, want %q, %d", test.line, adds, w.baseLen, test.adds, test.baseLen)
		}
	}

	// a line of math mode, and the shell
	gl.TrState.MathMode = false
	for line, want := range map[string]string{`# cd S`: `ub\ Dir/`, ` ## cd S`: `ub\ Dir/`} {
		if _, comps := gl.completions(line); len(comps) != 1 || comps[0].add != want {
			t.Errorf("%s: completions %v, want %q", line, comps, want)
		}
	}
	gl.Config.Dir = t.TempDir()
	if _, comps := gl.completions(`cd S`); len(comps) != 0 {
		t.Errorf("shell: completions %v, want none", comps)
	}
}

// TestCompleteRemote checks the completions of the files on an ssh host,
// for the commands that run on it, and the names of the hosts.
func TestCompleteRemote(t *testing.T) {
	s := sshtest.NewServer(t)
	for _, d := range []string{"remote dir", "sub"} {
		if err := os.Mkdir(filepath.Join(s.Root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"remote.txt", ".hidden", "sub/inner.txt"} {
		if err := os.WriteFile(filepath.Join(s.Root, f), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(s.Root, "sub"), filepath.Join(s.Root, "link")); err != nil {
		t.Fatal(err)
	}
	local := t.TempDir()
	if err := os.WriteFile(filepath.Join(local, "local.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	gl := NewGoal()
	gl.Config.Dir = local
	gl.SSH.User.KeyPath = s.KeyPath
	cl := sshclient.NewClient(gl.SSH)
	if err := cl.Connect(s.Addr); err != nil {
		t.Fatal(err)
	}
	defer cl.Close()
	gl.SSHClients["h"] = cl

	tests := []struct {
		active  string // the active host
		line    string
		adds    []string
		baseLen int
	}{
		{"", `@h ls rem`, []string{`ote\ dir/`, `ote.txt `}, 3},
		{"", `@h cat sub/in`, []string{`ner.txt `}, 2},
		{"", `@h ls li`, []string{`nk/`}, 2},
		{"", `@h ls .h`, []string{`idden `}, 2},
		{"", `echo hi | @h cat rem`, []string{`ote\ dir/`, `ote.txt `}, 3},
		{"", `ls rem`, nil, 3},
		{"", `ls lo`, []string{`cal.txt `}, 2},
		{"", `scp @h:rem`, []string{`ote\ dir/`, `ote.txt `}, 3},
		{"", `scp local.txt @h:sub/`, []string{`inner.txt `}, 0},
		{"", `@`, []string{`h `, `0 `}, 1}, // the local host last
		{"", `@h`, []string{` `}, 2},
		{"", `@x ls rem`, nil, 3},
		{"h", `ls rem`, []string{`ote\ dir/`, `ote.txt `}, 3},
		{"h", `cat ~/sub/`, []string{`inner.txt `}, 0},
		{"h", `scp lo`, []string{`cal.txt `}, 2},
		{"h", `@0 ls lo`, []string{`cal.txt `}, 2},
	}
	for _, test := range tests {
		gl.SSHActive = test.active
		w, comps := gl.completions(test.line)
		var adds []string
		for _, c := range comps {
			adds = append(adds, c.add)
		}
		if !slices.Equal(adds, test.adds) || w.baseLen != test.baseLen {
			t.Errorf("%s (active %q): completions %q, base %d, want %q, %d", test.line, test.active, adds, w.baseLen, test.adds, test.baseLen)
		}
	}
}
