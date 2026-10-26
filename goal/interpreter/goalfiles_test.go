// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
			_, _, err = in.Eval(string(src))
			assert.NoError(t, err, out.String())
			want, err := os.ReadFile(strings.TrimSuffix(f, ".goal") + ".want")
			if err != nil {
				return
			}
			assert.Equal(t, string(want), out.String())
		})
	}
}
