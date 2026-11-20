// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"errors"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestError tests the errors of Goal code, with their positions, source
// lines, and underlying errors.
func TestError(t *testing.T) {
	under := errors.New("under")
	e := &Error{Line: 2, Col: 5, Src: "x\t:= y", Msg: "bad", Err: under}
	assert.Equal(t, "2:5: bad\n    x := y\n        ^", e.Error())
	assert.ErrorIs(t, e, under)
	l, c := e.CodePos()
	assert.Equal(t, [2]int{2, 5}, [2]int{l, c})
	assert.Equal(t, "bad", e.CodeMessage())
	assert.Equal(t, "msg", (&Error{Msg: "msg"}).Error(), "no position")
	assert.Equal(t, "1:1: msg", (&Error{Line: 1, Col: 1, Msg: "msg"}).Error(), "no source")

	sl := SourceLine{Src: "abc", Cols: []token.Pos{0, 2, 3}}
	assert.Equal(t, 1, sl.SourceCol(1), "not known")
	assert.Equal(t, 3, sl.SourceCol(9), "past the end")
	assert.Equal(t, 1, (&SourceLine{}).SourceCol(5))

	sources := [][]SourceLine{{{Src: "a"}, {Src: "b"}}, {{Src: "c", Cols: []token.Pos{4}}}}
	se := SourceError(sources, 1, 1, 1, "in c", under)
	assert.Equal(t, 3, se.Line, "after the lines of the first chunk")
	assert.Equal(t, "c", se.Src)
	assert.Equal(t, 4, se.Col)
	assert.Equal(t, 0, SourceError(sources, 5, 1, 1, "x", nil).Line, "no such chunk")
}

// TestBrackets tests the errors of closing brackets without opening ones,
// and of brackets that are not closed in complete code.
func TestBrackets(t *testing.T) {
	st := NewState()
	st.TranspileCode("x := 1)\ny := 2")
	require.NotEmpty(t, st.Errors)
	assert.Contains(t, st.Errors[0].Error(), ") without a matching (")
	assert.Equal(t, 0, st.TotalDepth(), "the depths are reset")

	st = NewState()
	st.TranspileCode("if x {\n\ty := f(1,\n")
	err := st.UnclosedError()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "( is not closed: missing )")
	assert.Error(t, st.DepthError())
	assert.Contains(t, st.DepthError().Error(), "braces {}")
	st.ResetDepth()
	assert.NoError(t, st.DepthError())
	assert.Equal(t, 0, st.TotalDepth())

	st = NewState()
	st.TranspileCode("x := [1, 2]")
	assert.NoError(t, st.UnclosedError())
	st.ResetCode()
	assert.Equal(t, "", st.Code())
}

// TestGoFile tests transpiling Goal files to Go files, with and without a
// package clause.
func TestGoFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "a.goal")
	require.NoError(t, os.WriteFile(in, []byte("x := 1\nfmt.Println(x)\n"), 0o644))
	out := filepath.Join(dir, "a.go")
	st := NewState()
	require.NoError(t, st.TranspileFile(in, out))
	b, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(b), "package main")
	assert.Contains(t, string(b), "func main()")

	st = NewState()
	res, err := st.GoFile("b.goal", "b.go", "package b\n\nfunc F() int { return 1 }\n")
	require.NoError(t, err)
	assert.Contains(t, string(res), "package b")
	assert.NotContains(t, string(res), "func main()")

	st = NewState()
	_, err = st.GoFile("c.goal", "c.go", "package c\n\nfunc F( {\n")
	assert.Error(t, err)
	assert.Error(t, NewState().TranspileFile(filepath.Join(dir, "missing.goal"), out))

	// the //line directive names the goal file relative to the Go file,
	// however the paths are given (absolute, as from goalls, or relative,
	// as from goal build).
	b, err = os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(b), "\n//line a.goal:1\n")
	res, err = NewState().GoFile(filepath.Join(dir, "d.goal"), filepath.Join(dir, "d.go"), "package d\n")
	require.NoError(t, err)
	assert.Contains(t, string(res), "\n//line d.goal:1\n")
	res, err = NewState().GoFile("e.goal", "e.go", "package e\n")
	require.NoError(t, err)
	assert.Contains(t, string(res), "\n//line e.goal:1\n")
	res, err = NewState().GoFile(filepath.Join("sub", "f.goal"), filepath.Join("out", "f.go"), "package f\n")
	require.NoError(t, err)
	assert.Contains(t, string(res), "\n//line ../sub/f.goal:1\n")
}

// TestTensorfsCommands tests the tensorfs commands of math mode.
func TestTensorfsCommands(t *testing.T) {
	assert.Equal(t, []string{"cd", "ls", "mkdir"}, TensorfsCommands())
	for _, c := range []struct{ in, out string }{
		{"# cd data", `tensorfs.Chdir("data")`},
		{"# cd", `tensorfs.Chdir("")`},
		{"# mkdir data", `tensorfs.Mkdir("data")`},
		{"# ls a b", `tensorfs.List("a", "b")`},
	} {
		st := NewState()
		st.TranspileCode(c.in)
		assert.Equal(t, c.out, strings.TrimSpace(st.Code()), c.in)
	}
	st := NewState()
	st.TranspileCode("# mkdir")
	assert.NotEmpty(t, st.Errors, "mkdir needs a name")
}
