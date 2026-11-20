// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// useRoot sets the current root and directory of the commands to a new
// directory for the duration of the test.
func useRoot(t *testing.T) *Node {
	oldRoot, oldDir, oldOut := CurRoot, CurDir, ListOutput
	root, err := NewDir("data")
	require.NoError(t, err)
	CurRoot, CurDir = root, nil
	t.Cleanup(func() { CurRoot, CurDir, ListOutput = oldRoot, oldDir, oldOut })
	return root
}

// TestCommands tests the shell-like commands that operate on the current
// directory: Mkdir, Chdir, Record, Set, SetCopy, Get and List.
func TestCommands(t *testing.T) {
	root := useRoot(t)
	sub := Mkdir("sub")
	require.NotNil(t, sub)
	assert.Same(t, sub, Mkdir("sub"))
	CurDir = nil
	assert.Nil(t, Mkdir(""))

	CurDir = nil
	require.NoError(t, Chdir("sub"))
	assert.Same(t, sub, CurDir)
	assert.Error(t, Chdir("nope"))
	assert.Same(t, sub, CurDir)

	a := tensor.NewFloat64FromValues(1, 2)
	CurDir = nil
	Record(a, "a") // recorded in root
	assert.Same(t, a, root.Value("a"))

	require.NoError(t, Chdir("sub"))
	b := tensor.NewFloat64FromValues(3)
	require.NoError(t, Set("b", b))
	assert.Same(t, b, Get("b"))
	b2 := tensor.NewFloat64FromValues(4)
	require.NoError(t, Set("b", b2)) // replaces existing value
	assert.Same(t, b2, Get("b"))

	Mkdir("deep")
	require.NoError(t, Set("deep/c", a))
	assert.Same(t, a, Get("deep/c"))
	assert.Error(t, Set("missing/c", a))
	assert.Error(t, Set("deep", a)) // a directory
	assert.Error(t, Set("", a))

	require.NoError(t, SetCopy("cp", a))
	a.SetFloat1D(10, 0)
	assert.Equal(t, 1.0, Get("cp").Float1D(0))

	assert.Nil(t, Get(""))
	assert.Nil(t, Get("missing"))
	assert.Nil(t, Get("deep"))

	var buf bytes.Buffer
	ListOutput = &buf
	require.NoError(t, List())
	assert.Equal(t, "b deep/ cp \n", buf.String())
	buf.Reset()
	require.NoError(t, List("-lr", "deep"))
	assert.Equal(t, "c [2]\n\n", buf.String())

	require.NoError(t, Chdir(""))
	assert.Same(t, root, CurDir)
	CurDir = nil
	buf.Reset()
	require.NoError(t, List("-l"))
	assert.Equal(t, "sub/\na [2]\n\n", buf.String())
}

// TestCopy tests Node.Copy of values and directories to new names, into
// directories, and onto existing values, and its errors.
func TestCopy(t *testing.T) {
	dir, _ := NewDir("root")
	a := dir.Float64("a", 2)
	a.SetFloat1D(1, 0)
	dir.Float64("b", 1)
	d := dir.Dir("d")
	d.Dir("sd").Int("x", 1)

	require.NoError(t, dir.Copy(Overwrite, "c", "a"))
	c := dir.Value("c")
	assert.Equal(t, 1.0, c.Float1D(0))
	a.SetFloat1D(2, 0)
	assert.Equal(t, 1.0, c.Float1D(0), "copy is a clone")

	require.NoError(t, dir.Copy(Overwrite, "d", "a", "b"))
	assert.Equal(t, 2.0, d.Value("a").Float1D(0))
	a.SetFloat1D(3, 0)
	require.NoError(t, dir.Copy(Preserve, "d", "a"))
	assert.Equal(t, 2.0, d.Value("a").Float1D(0), "preserved")
	require.NoError(t, dir.Copy(Overwrite, "d", "a"))
	assert.Equal(t, 3.0, d.Value("a").Float1D(0), "overwritten")

	// copy a directory, recursively
	require.NoError(t, dir.Copy(Overwrite, "d2", "d"))
	x, err := dir.NodeAtPath("d2/sd/x")
	require.NoError(t, err)
	assert.Equal(t, 1, x.Tensor.Len())
	require.NoError(t, dir.Copy(Overwrite, "d2", "d"), "directory onto directory")

	assert.Error(t, dir.Copy(Overwrite, "", "a"))
	assert.Error(t, dir.Copy(Overwrite, "e"))
	assert.Error(t, dir.Copy(Overwrite, "c", "a", "b"), "multiple onto a value")
	assert.Error(t, dir.Copy(Overwrite, "e", "missing"))
	d.Dir("c")
	assert.Error(t, dir.Copy(Overwrite, "d", "c"), "value onto directory")
	dir.Dir("e").Float64("d", 1)
	assert.Error(t, dir.Copy(Overwrite, "e", "d"), "directory onto value")
	assert.Error(t, dir.Node("a").Copy(Overwrite, "x", "a"), "not a directory")
}

// TestValuesFuncs tests the functions that return the values and nodes of
// a directory, by names, by filter function, and in alphabetical order.
func TestValuesFuncs(t *testing.T) {
	dir, _ := NewDir("root")
	dir.Float64("z", 1)
	dir.Float64("a", 2)
	sub := dir.Dir("sub")
	sub.Int("y", 3)
	sub.Int("b", 4)

	vals, err := dir.Values()
	require.NoError(t, err)
	assert.Len(t, vals, 2)
	vals, err = dir.Values("a", "missing", "sub")
	assert.Error(t, err)
	assert.Len(t, vals, 1)
	_, err = dir.Node("a").Values()
	assert.Error(t, err)

	nds, err := dir.Nodes("a", "missing")
	assert.Error(t, err)
	assert.Len(t, nds, 1)
	_, err = dir.Node("a").Nodes()
	assert.Error(t, err)

	lens := func(ts []tensor.Tensor) []int {
		var r []int
		for _, ts := range ts {
			r = append(r, ts.Len())
		}
		return r
	}
	names := func(ns []*Node) []string {
		var r []string
		for _, n := range ns {
			r = append(r, n.Name())
		}
		return r
	}
	assert.Equal(t, []int{1, 2, 3, 4}, lens(dir.ValuesFunc(nil)))
	assert.Equal(t, []int{2, 4, 3, 1}, lens(dir.ValuesAlphaFunc(nil)))
	assert.Equal(t, []string{"a", "b", "y", "z"}, names(dir.NodesAlphaFunc(nil)))
	noSub := func(nd *Node) bool { return nd.Name() != "sub" }
	assert.Equal(t, []int{1, 2}, lens(dir.ValuesFunc(noSub)))
	assert.Equal(t, []int{2, 1}, lens(dir.ValuesAlphaFunc(noSub)))
	assert.Equal(t, []string{"a", "z"}, names(dir.NodesAlphaFunc(noSub)))
	leaf := dir.Node("a")
	assert.Nil(t, leaf.ValuesFunc(nil))
	assert.Nil(t, leaf.ValuesAlphaFunc(nil))
	assert.Nil(t, leaf.NodesAlphaFunc(nil))
	assert.Nil(t, leaf.NodesFunc(nil))

	// by modification time
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	dir.Node("z").modTime = base
	dir.Node("a").modTime = base.Add(time.Hour)
	sub.modTime = base.Add(-time.Hour)
	assert.Equal(t, []string{"sub", "z", "a"}, dir.dirNamesByTime())
}

// TestMetadataCalc tests SetMetaItems and CalcAll.
func TestMetadataCalc(t *testing.T) {
	dir, _ := NewDir("root")
	a := dir.Float64("a", 1)
	b := dir.Dir("sub").Float64("b", 1)
	require.NoError(t, dir.SetMetaItems("Color", "red", "a"))
	v, _ := metadata.GetFromData[string](*a.Metadata(), "Color")
	assert.Equal(t, "red", v)
	assert.Error(t, dir.SetMetaItems("Color", "red", "missing"))

	err := dir.CalcAll()
	assert.Error(t, err, "no calc funcs set")

	tensor.SetCalcFunc(a, func() error {
		a.SetFloat1D(3, 0)
		return nil
	})
	tensor.SetCalcFunc(b, func() error {
		b.SetFloat1D(4, 0)
		return nil
	})
	require.NoError(t, dir.CalcAll())
	assert.Equal(t, 3.0, a.Float1D(0))
}

// TestNodeValues tests the typed value constructors, ValueType, NewValues,
// Set, and the FileInfo and DirEntry methods of a value node.
func TestNodeValues(t *testing.T) {
	dir, _ := NewDir("root")
	f32 := dir.Float32("f32", 2)
	assert.Same(t, f32, dir.Float32("f32"))
	assert.Equal(t, 2, dir.Int("i", 2).Len())
	s := dir.StringValue("s", 1)
	s.SetString1D("hi", 0)
	assert.Equal(t, 1, Scalar[float64](dir, "sc").Len())

	vt := ValueType(dir, "vt", reflect.Float32, 3)
	assert.Equal(t, reflect.Float32, vt.DataType())
	assert.Same(t, vt, ValueType(dir, "vt", reflect.Int))

	NewValues[int](dir, []int{2, 2}, "n1", "n2")
	assert.Equal(t, []int{2, 2}, dir.Value("n2").ShapeSizes())

	tsr := tensor.NewIntFromValues(1, 2, 3)
	nd := dir.Set("set", tsr)
	assert.Same(t, tsr, nd.Tensor)
	tsr2 := tensor.NewIntFromValues(4)
	assert.Same(t, nd, dir.Set("set", tsr2))
	assert.Same(t, tsr2, nd.Tensor)

	// FileInfo, DirEntry
	assert.Same(t, tsr2, nd.Sys())
	assert.NotNil(t, dir.Sys())
	assert.Equal(t, fs.FileMode(0444), nd.Mode())
	assert.Equal(t, fs.ModeDir, dir.Type())
	info, err := nd.Info()
	require.NoError(t, err)
	assert.Equal(t, "set", info.Name())
	assert.Equal(t, int64(0), dir.Size())
	assert.Equal(t, fileinfo.Unknown, dir.KnownFileInfo())
	assert.Equal(t, fileinfo.Tensor, dir.Node("vt").KnownFileInfo())
	assert.Equal(t, fileinfo.String, dir.Node("s").KnownFileInfo())
	assert.Equal(t, fileinfo.Number, nd.KnownFileInfo())

	b, err := dir.ReadFile("set")
	require.NoError(t, err)
	assert.Equal(t, tsr2.Bytes(), b)
	_, err = dir.ReadFile("missing")
	assert.Error(t, err)
	sub := dir.Dir("sub")
	_, err = dir.ReadFile("sub")
	assert.Error(t, err)

	// the short recursive listing
	sub.Int("x", 1)
	ls := dir.List(Short, Recursive)
	assert.True(t, strings.HasPrefix(ls, "f32 i s sc vt n1 n2 set \n"), ls)
	assert.Contains(t, ls, "x ")
	assert.Contains(t, dir.String(), "sub/ ")
	assert.Equal(t, "set [1]", nd.String())
}

// TestSub tests the relative path resolution of Sub and DirAtPath,
// including going up with "..", and its errors.
func TestSub(t *testing.T) {
	root, _ := NewDir("root")
	a := root.Dir("a/b")
	root.Float64("v", 1)

	d, err := a.DirAtPath("../..")
	require.NoError(t, err)
	assert.Same(t, root, d)
	d, err = root.DirAtPath("./a")
	require.NoError(t, err)
	assert.Equal(t, "a", d.Name())
	_, err = root.DirAtPath("..")
	assert.Error(t, err, "already at root")
	_, err = root.DirAtPath("missing")
	assert.Error(t, err)
	_, err = root.DirAtPath("v")
	assert.Error(t, err, "not a directory")
	_, err = root.Node("v").Sub("x")
	assert.Error(t, err)
	_, err = root.NodeAtPath("../x")
	assert.Error(t, err, "invalid path")
	_, err = root.Node("v").NodeAtPath("x")
	assert.Error(t, err)
	_, err = root.NodeAtPath("missing/x")
	assert.Error(t, err)
	_, err = root.ReadDir("missing")
	assert.Error(t, err)
	nd, err := root.NodeAtPath(".")
	require.NoError(t, err)
	assert.Same(t, root, nd)
}

// TestAdd tests adding nodes to a directory, including the error for an
// existing name and for a non-directory.
func TestAdd(t *testing.T) {
	root, _ := NewDir("root")
	root.Float64("v", 1)
	nd, _ := newNode(nil, "v")
	assert.ErrorIs(t, root.Add(nd), fs.ErrExist)
	assert.Error(t, root.Node("v").Add(nd))
	ex, err := newNode(root, "v")
	assert.ErrorIs(t, err, fs.ErrExist)
	assert.Same(t, root.Node("v"), ex)
	_, err = newNode(root.Node("v"), "x")
	assert.Error(t, err)
	_, err = NewDir("", root)
	require.NoError(t, err)
	assert.NotNil(t, root.Node("root"))
}

// TestTarOptions tests Tar without compression and with an include
// filter, Untar of a large compressed value, and Untar errors.
func TestTarOptions(t *testing.T) {
	dir, _ := NewDir("root")
	dir.Float64("keep", 2).SetFloat1D(1.5, 1)
	dir.Float64("skip", 2)
	dir.Dir("sub").Int("x", 3)

	var b bytes.Buffer
	require.NoError(t, Tar(&b, dir, NoGzip, func(nd *Node) bool { return nd.Name() != "skip" }))
	nd, _ := NewDir("root")
	require.NoError(t, Untar(&b, nd, NoGzip))
	assert.Equal(t, "keep [2]\nsub/\n\tx [3]\n", nd.ListAll())
	assert.Equal(t, 1.5, nd.Value("keep").Float1D(1))

	// a value larger than a single read of the decompressor
	big := dir.Float64("big", 200000)
	for i := range big.Len() {
		big.SetFloat1D(float64(i), i)
	}
	b.Reset()
	require.NoError(t, Tar(&b, dir, Gzip, nil))
	nd, _ = NewDir("root")
	require.NoError(t, Untar(&b, nd, Gzip))
	nb := nd.Value("big")
	require.Equal(t, big.Len(), nb.Len())
	assert.Equal(t, float64(big.Len()-1), nb.Float1D(big.Len()-1))

	assert.Error(t, Untar(strings.NewReader("not gzip"), nd, Gzip))
	err := Untar(strings.NewReader(strings.Repeat("x", 1024)), nd, NoGzip)
	assert.Error(t, err)
}

// TestAllFiles tests AllFiles on a directory of real files, with an
// exclusion pattern.
func TestAllFiles(t *testing.T) {
	td := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(td, "sub"), 0o755))
	for _, f := range []string{"a.txt", "b.dat", "sub/c.txt"} {
		require.NoError(t, os.WriteFile(filepath.Join(td, f), []byte("x"), 0o644))
	}
	files, err := AllFiles(td, filepath.Join(td, "*.dat"))
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(td, "a.txt"), filepath.Join(td, "sub", "c.txt")}, files)
	_, err = AllFiles(filepath.Join(td, "missing"))
	assert.True(t, errors.Is(err, fs.ErrNotExist))
}

// TestWriteErrors tests the path errors of the write operations for
// paths through a value or a missing directory.
func TestWriteErrors(t *testing.T) {
	root, _ := NewDir("root")
	root.Float64("v", 1)
	assert.Error(t, root.Rename("v/x", "y"), "not a directory")
	assert.ErrorIs(t, root.Rename("missing/x", "y"), fs.ErrNotExist)
	assert.Error(t, root.RemoveAll("."), "cannot change the root")
}
