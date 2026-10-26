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
