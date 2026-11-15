// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotosl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMoveLines tests moving a region of lines up to an earlier line.
func TestMoveLines(t *testing.T) {
	lines := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
	MoveLines(&lines, 1, 3, 5)
	var s []string
	for _, l := range lines {
		s = append(s, string(l))
	}
	assert.Equal(t, []string{"a", "d", "e", "b", "c"}, s)
}

// TestFileNames tests the names of files with other extensions.
func TestFileNames(t *testing.T) {
	assert.Equal(t, "dir/a.wgsl", wgslFile("dir/a.go"))
	assert.Equal(t, "dir/a", bareFile("dir/a.go"))
}

// TestCopyFiles tests copying files, and the files of packages into the
// imports directory, once.
func TestCopyFiles(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.wgsl")
	require.NoError(t, os.WriteFile(src, []byte("line 1\nline 2\n"), 0o644))
	dst := filepath.Join(dir, "b.wgsl")
	lines, err := CopyFile(src, dst)
	require.NoError(t, err)
	assert.Equal(t, "line 1", string(lines[0]))
	b, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Contains(t, string(b), "line 2")
	_, err = CopyFile(filepath.Join(dir, "missing"), dst)
	assert.Error(t, err)

	st := &State{ImportsDir: dir}
	require.NoError(t, st.CopyPackageFile("slrand.wgsl", "cogentcore.org/lab/gosl/slrand"))
	require.Len(t, st.SLImportFiles, 1)
	assert.FileExists(t, filepath.Join(dir, "slrand.wgsl"))
	require.NoError(t, st.CopyPackageFile("slrand.wgsl", "cogentcore.org/lab/gosl/slrand"))
	assert.Len(t, st.SLImportFiles, 1, "not again")
	assert.Error(t, st.CopyPackageFile("x.wgsl", "cogentcore.org/lab/nonexistent/pkg"))
}

// TestPrintFuncGraph tests printing the graph of the functions that call
// other functions.
func TestPrintFuncGraph(t *testing.T) {
	st := &State{FuncGraph: map[string]*Function{}}
	f := NewFunction("F")
	f.Funcs["G"] = NewFunction("G")
	st.FuncGraph["F"] = f
	st.FuncGraph["G"] = f.Funcs["G"]
	st.PrintFuncGraph()
}
