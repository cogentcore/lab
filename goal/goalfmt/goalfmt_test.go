// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goalfmt

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourcePackage(t *testing.T) {
	// the Goal code (tensor indexing) is indented, the Go code is
	// formatted, the unused import is removed, and the used one added
	src := `package foo

import (
	"fmt"
	"os"
)

func F(ni, di, X int) string {
x := 1
		Neurons[ni, di, X] = 2
	return  fmt.Sprint(x)+strings.ToUpper("a")
}
`
	want := `package foo

import (
	"fmt"
	"strings"
)

func F(ni, di, X int) string {
	x := 1
	Neurons[ni, di, X] = 2
	return fmt.Sprint(x) + strings.ToUpper("a")
}
`
	fname := filepath.Join(t.TempDir(), "test.goal")
	out, err := Source(fname, []byte(src), nil)
	assert.NoError(t, err)
	assert.Equal(t, want, string(out))
	out, err = Source(fname, []byte(want), nil)
	assert.NoError(t, err)
	assert.Equal(t, want, string(out), "formatted code is unchanged")

	// without imports, the imports are not fixed
	out, err = Source(fname, []byte(src), &Options{FormatOnly: true})
	assert.NoError(t, err)
	assert.Contains(t, string(out), "\t\"os\"\n")
	assert.NotContains(t, string(out), "\t\"strings\"\n")
}

func TestSourceScript(t *testing.T) {
	// a script has no imports, and its code is not indented by the
	// main function that it is in, in its Go code
	src := "x := 10\n  for i := range x {\nfmt.Println(i)\n     }\nls -la\n##\na := zeros(3)\n##\n"
	want := "x := 10\nfor i := range x {\n\tfmt.Println(i)\n}\nls -la\n##\na := zeros(3)\n##\n"
	out, err := Source(filepath.Join(t.TempDir(), "test.goal"), []byte(src), nil)
	assert.NoError(t, err)
	assert.Equal(t, want, string(out))
}

func TestSourceErrors(t *testing.T) {
	// the syntax errors of the Go code are at their positions in the Goal code
	src := "package foo\n\nfunc F() {\n\tx := \n}\n"
	_, err := Source("test.goal", []byte(src), nil)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "test.goal:5:1: expected operand")
	}
}
