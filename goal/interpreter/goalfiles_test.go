// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
)

// TestGoalFiles runs each of the Goal files of the transpile tests,
// ../transpile/testdata/*.goal, checking that its output is that of the
// .want file of the same name, if it exists. The transpile tests build
// them as Go programs.
func TestGoalFiles(t *testing.T) {
	files, err := filepath.Glob("../transpile/testdata/*.goal")
	assert.NoError(t, err)
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".goal")
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(f)
			assert.NoError(t, err)
			var out syncBuffer
			in := NewInterpreter(Options{Stdout: &out, Stderr: &out})
			_, _, err = in.EvalCode(string(src))
			assert.NoError(t, err, out.String())
			want, err := os.ReadFile(strings.TrimSuffix(f, ".goal") + ".want")
			if err != nil {
				return
			}
			assert.Equal(t, string(want), out.String())
		})
	}
}

// goalFence matches the Goal code blocks of a markdown page: its
// indentation and code.
var goalFence = regexp.MustCompile("(?m)^(\\s*)```Goal\\s*\n((?:.*\n)*?)\\s*```")

// TestMathDocs runs the Goal code blocks of the math mode page of the docs,
// in order in one interpreter, as the docs do, recording the variables of
// math mode in tensorfs, and checks that their output is that of
// testdata/mathdocs.want. The docs test that the compiled code of the
// blocks prints the same.
func TestMathDocs(t *testing.T) {
	page, err := os.ReadFile("../../docs/content/math.md")
	assert.NoError(t, err)
	var out syncBuffer
	in := NewInterpreter(Options{Stdout: &out, Stderr: &out})
	tensorfs.ListOutput = &out
	defer func() { tensorfs.ListOutput = nil }()
	for i, m := range goalFence.FindAllStringSubmatch(string(page), -1) {
		var lines []string
		for _, l := range strings.Split(strings.TrimSuffix(m[2], "\n"), "\n") {
			lines = append(lines, strings.TrimPrefix(l, m[1]))
		}
		fmt.Fprintf(&out, "######## block %d\n", i)
		in.Goal.TrState.MathRecord = true
		_, _, err := in.EvalCode(strings.Join(lines, "\n") + "\n")
		assert.NoError(t, err, "block %d", i)
	}
	want, err := os.ReadFile("testdata/mathdocs.want")
	assert.NoError(t, err)
	assert.Equal(t, string(want), out.String())
}
