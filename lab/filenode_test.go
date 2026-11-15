// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/text/textcore"
	"cogentcore.org/lab/plotcore"
	"cogentcore.org/lab/tensorcore"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// labWindow runs a window of a basic browser of the given file system and
// data directory, and returns it, with a function that waits for a file
// node with the given path, and closes the window at the end of the test.
func labWindow(t *testing.T, fsys *tensorfs.Node, dataDir string) (*core.Body, *Basic, func(p string) *FileNode) {
	t.Helper()
	var b *core.Body
	var br *Basic
	if fsys == nil {
		b, br = NewBasicWindow(nil, dataDir)
	} else {
		b, br = NewBasicWindow(fsys, dataDir)
	}
	b.RunWindow()
	t.Cleanup(func() {
		b.AsyncLock()
		b.Close()
		b.AsyncUnlock()
	})
	node := func(p string) *FileNode {
		t.Helper()
		var fn *FileNode
		deadline := time.Now().Add(5 * time.Second)
		for fn == nil && time.Now().Before(deadline) {
			b.AsyncLock()
			if br.Files != nil {
				if n, ok := br.Files.FindFile(p); ok && string(n.Filepath) == p {
					fn, _ = n.This.(*FileNode)
				}
			}
			b.AsyncUnlock()
			if fn == nil {
				time.Sleep(20 * time.Millisecond)
			}
		}
		if fn == nil {
			t.Fatalf("no file node %s", p)
		}
		return fn
	}
	return b, br, node
}

// locked calls the given function with the lock of the given body held.
func locked(b *core.Body, fun func()) {
	b.AsyncLock()
	defer b.AsyncUnlock()
	fun()
}

// waitFor waits until the given condition is true, with the lock of the
// given body held.
func waitFor(t *testing.T, b *core.Body, what string, cond func() bool) {
	t.Helper()
	assert.Eventually(t, func() bool {
		b.AsyncLock()
		defer b.AsyncUnlock()
		return cond()
	}, 5*time.Second, 10*time.Millisecond, what)
}

// TestFileNodeValues tests opening the tensorfs values: a number and a
// string, which are edited in dialogs, an empty value, and a tensor, and
// their icons and tooltips.
func TestFileNodeValues(t *testing.T) {
	root, _ := tensorfs.NewDir("root")
	sims := root.Dir("Sims")
	tensorfs.Scalar[float64](sims, "N").SetFloat1D(2.5, 0)
	tensorfs.Scalar[string](sims, "S").SetString1D("hi", 0)
	tensorfs.Value[float64](sims, "E", 0)
	tensorfs.Value[float64](sims, "T", 3, 3).SetFloat1D(1, 4)
	b, br, node := labWindow(t, root, "Sims")

	n := node("Sims/N")
	locked(b, func() {
		assert.Equal(t, fileinfo.Number, n.Info.Known)
		assert.Equal(t, fileinfo.Data, n.Info.Cat)
		tt, _ := n.WidgetTooltip(image.Point{})
		assert.Contains(t, tt, "N")
		n.OpenFile()
	})
	waitFor(t, b, "the number dialog", func() bool { return findSpinner(core.TopScene(b)) != nil })
	locked(b, func() {
		sc := core.TopScene(b)
		findSpinner(sc).SetValue(7)
		core.ButtonByText(sc, "OK").Send(events.Click)
	})
	waitFor(t, b, "the new number", func() bool { return sims.Node("N").Tensor.Float1D(0) == 7 })

	s := node("Sims/S")
	locked(b, func() {
		assert.Equal(t, fileinfo.String, s.Info.Known)
		s.EditFile() // the same as opening data
	})
	waitFor(t, b, "the string dialog", func() bool { return findText(core.TopScene(b)) != nil })
	locked(b, func() {
		sc := core.TopScene(b)
		findText(sc).SetText("there")
		core.ButtonByText(sc, "OK").Send(events.Click)
	})
	waitFor(t, b, "the new string", func() bool { return sims.Node("S").Tensor.String1D(0) == "there" })

	e := node("Sims/E")
	locked(b, func() {
		e.OpenFile() // a snackbar: no data
		assert.Nil(t, findSpinner(core.TopScene(b)))
	})

	tn := node("Sims/T")
	locked(b, func() {
		assert.Equal(t, fileinfo.Tensor, tn.Info.Known)
		tn.OpenFile()
		assert.NotNil(t, TabAt[*tensorcore.TensorEditor](br.Tabs, "Sims/T"))
		tn.GridFile()
		assert.NotNil(t, TabAt[*tensorcore.TensorGrid](br.Tabs, "Sims T Grid"))
		tn.PlotFile()
		assert.NotNil(t, TabAt[*plotcore.Editor](br.Tabs, "Sims T Plot"))
	})
}

// findSpinner returns the first spinner in the given widget.
func findSpinner(w core.Widget) *core.Spinner {
	if w == nil {
		return nil
	}
	for cw := range w.AsWidget().AllWidgets() {
		if sp, ok := cw.(*core.Spinner); ok {
			return sp
		}
	}
	return nil
}

// TestFileNodeSelected tests the actions on the selected nodes of the
// context menu: Edit, Plot, and Grid, its disabled items, and Diff Dirs,
// which needs two directories.
func TestFileNodeSelected(t *testing.T) {
	root, _ := tensorfs.NewDir("root")
	sims := root.Dir("Sims")
	tensorfs.Value[float64](sims.Dir("Run"), "X", 3)
	tensorfs.Value[float64](sims, "A", 3)
	tensorfs.Value[float64](sims, "B", 3)
	b, br, node := labWindow(t, root, "Sims")
	a, bn, run := node("Sims/A"), node("Sims/B"), node("Sims/Run")
	menu := func(fn *FileNode) *core.Scene {
		m := core.NewScene()
		fn.ContextMenu(m, image.Point{})
		return m
	}
	disabled := func(m *core.Scene, text string) bool {
		bt := core.ButtonByText(m, text)
		require.NotNil(t, bt, text)
		bt.Style()
		return bt.StateIs(states.Disabled)
	}
	locked(b, func() {
		m := menu(a)
		assert.True(t, disabled(m, "Edit"), "no selection")
		assert.True(t, disabled(m, "Diff Dirs"))

		a.SelectEvent(events.SelectOne)
		bn.SelectEvent(events.ExtendOne)
		m = menu(a)
		assert.False(t, disabled(m, "Edit"))
		assert.False(t, disabled(m, "Plot"))
		assert.False(t, disabled(m, "Grid"))
		assert.True(t, disabled(m, "Diff Dirs"), "not a directory")

		core.ButtonByText(m, "Edit").Send(events.Click)
		assert.NotNil(t, TabAt[*tensorcore.TensorEditor](br.Tabs, "Sims/A"))
		assert.NotNil(t, TabAt[*tensorcore.TensorEditor](br.Tabs, "Sims/B"))
		core.ButtonByText(m, "Plot").Send(events.Click)
		assert.NotNil(t, TabAt[*plotcore.Editor](br.Tabs, "Sims A Plot"))
		assert.NotNil(t, TabAt[*plotcore.Editor](br.Tabs, "Sims B Plot"))
		core.ButtonByText(m, "Grid").Send(events.Click)
		assert.NotNil(t, TabAt[*tensorcore.TensorGrid](br.Tabs, "Sims B Grid"))

		run.SelectEvent(events.SelectOne)
		m = menu(run)
		assert.False(t, disabled(m, "Diff Dirs"))
		core.ButtonByText(m, "Diff Dirs").Send(events.Click) // a snackbar: only one
		run.GridFile()                                       // a snackbar: a directory
		assert.Nil(t, br.Tabs.TabByName("Sims Run Grid"))
		run.PlotFile()
		assert.NotNil(t, TabAt[*plotcore.Editor](br.Tabs, "Sims Run Plot"))
	})
}

// TestFileNodeOSFiles tests opening the files of the operating system:
// a table file, which is edited and plotted as a table, an invalid one,
// and a text file, which is edited in a text editor, and diffing two
// directories.
func TestFileNodeOSFiles(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir()) // as the paths of the tree
	require.NoError(t, err)
	t.Chdir(dir)
	data := filepath.Join(dir, "data")
	write := func(name, content string) string {
		fn := filepath.Join(data, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(fn), 0o755))
		require.NoError(t, os.WriteFile(fn, []byte(content), 0o644))
		return fn
	}
	tsv := write("a.tsv", "X\tY\n1\t2\n3\t4\n")
	txt := write("notes.txt", "hello\n")
	write("d1/f.txt", "one\n")
	write("d2/f.txt", "two\n")
	b, br, node := labWindow(t, nil, "data")
	defer core.TrackWindows()() // after the lab window, which the cleanup closes
	at, nt := node(tsv), node(txt)
	locked(b, func() {
		assert.Same(t, &br.Browser, LabBrowser)
		assert.Equal(t, fileinfo.Data, at.Info.Cat)
		at.OpenFile()
		tv := TabAt[*tensorcore.Table](br.Tabs, "data/a.tsv")
		require.NotNil(t, tv)
		assert.Equal(t, 2, tv.SliceSize)
		assert.Equal(t, []string{"X", "Y"}, tv.Table.Columns.Keys)
		at.PlotFile()
		assert.NotNil(t, TabAt[*plotcore.Editor](br.Tabs, "data/a.tsv Plot"))
		at.GridFile() // not tensorfs data: nothing
		assert.Nil(t, br.Tabs.TabByName("data a.tsv Grid"))

		nt.EditFile()
		ed := TabAt[*textcore.Editor](br.Tabs, "data/notes.txt")
		require.NotNil(t, ed)
		assert.Equal(t, "hello\n", ed.Lines.String())
		nt.PlotFile() // not data: nothing
		assert.Nil(t, br.Tabs.TabByName("data/notes.txt Plot"))
		nt.OpenFile()

		d1, _ := br.Files.FindFile(filepath.Join(data, "d1"))
		d2, _ := br.Files.FindFile(filepath.Join(data, "d2"))
		require.NotNil(t, d1)
		require.NotNil(t, d2)
		d1.SelectEvent(events.SelectOne)
		d2.SelectEvent(events.ExtendOne)
		d1.This.(*FileNode).DiffDirs()
	})
	assert.Eventually(t, func() bool { return core.AllRenderWindows.Len() >= 2 }, 5*time.Second, 10*time.Millisecond, "the diff browser")
}
