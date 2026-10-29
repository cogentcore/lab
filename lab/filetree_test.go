// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/lab/tensorcore"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
)

// TestFileNodeDoubleClick tests that double clicking a tensorfs directory
// with only data edits it, as do one with only subdirectories that have
// only single values (as a matrix) and a group of GroupStats, and others
// open and close.
func TestFileNodeDoubleClick(t *testing.T) {
	root, _ := tensorfs.NewDir("root")
	sims := root.Dir("Sims")
	tensorfs.Value[float64](sims.Dir("Leaf"), "X", 3)
	tensorfs.Value[float64](sims.Dir("Parent").Dir("Child"), "Y", 3)
	tensorfs.Value[float64](sims.Dir("Parent"), "Z", 3)
	tensorfs.Scalar[float64](sims.Dir("Describe").Dir("A"), "Mean").SetFloat1D(1.5, 0)
	tensorfs.Scalar[float64](sims.Dir("Describe").Dir("B"), "Mean").SetFloat1D(2.5, 0)
	for _, vn := range []string{"A", "B"} { // as from GroupStats
		vd := sims.Dir("method").Dir(vn)
		tensorfs.Value[string](vd, "method", 2).SetString1D("x", 1)
		tensorfs.Value[float64](vd, "Mean", 2).SetFloat1D(4.5, 1)
	}
	tensorfs.Value[float64](sims.Dir("Vectors").Dir("A"), "Mean", 3) // neither
	b, br := NewBasicWindow(root, "Sims")
	b.RunWindow()
	defer func() {
		b.AsyncLock()
		b.Close()
		b.AsyncUnlock()
	}()
	node := func(name string) *FileNode {
		var fn *FileNode
		deadline := time.Now().Add(5 * time.Second)
		for fn == nil && time.Now().Before(deadline) {
			b.AsyncLock()
			br.Files.WidgetWalkDown(func(cw core.Widget, cwb *core.WidgetBase) bool {
				if n, ok := cw.(*FileNode); ok && n.Name == name {
					fn = n
				}
				return fn == nil
			})
			b.AsyncUnlock()
			time.Sleep(20 * time.Millisecond)
		}
		if fn == nil {
			t.Fatalf("no file node %s", name)
		}
		return fn
	}
	doubleClick := func(fn *FileNode) {
		b.AsyncLock()
		fn.Parts.Send(events.DoubleClick)
		b.AsyncUnlock()
	}

	leaf := node("Leaf")
	doubleClick(leaf)
	b.AsyncLock()
	assert.True(t, br.Tabs.TabByName("Sims/Leaf") != nil, "the edit tab")
	assert.True(t, leaf.Closed, "not opened")
	b.AsyncUnlock()

	desc := node("Describe")
	doubleClick(desc)
	b.AsyncLock()
	if tab := br.Tabs.TabByName("Sims/Describe"); assert.NotNil(t, tab, "the matrix tab") {
		tv := tab.AsTree().Child(1).(*tensorcore.Table)
		assert.Equal(t, []string{"Name", "A", "B"}, tv.Table.Columns.Keys)
		assert.Equal(t, 2.5, tv.Table.Column("B").Float1D(0))
	}
	b.AsyncUnlock()

	// plotting uses the same table
	b.AsyncLock()
	desc.PlotFile()
	assert.NotNil(t, br.Tabs.TabByName("Sims Describe Plot"), "the plot tab")
	assert.Equal(t, []string{"Name", "A", "B"}, dirTable(TensorFS(desc.AsNode())).Columns.Keys)
	b.AsyncUnlock()

	doubleClick(node("method"))
	b.AsyncLock()
	if tab := br.Tabs.TabByName("Sims/method"); assert.NotNil(t, tab, "the group stats tab") {
		tv := tab.AsTree().Child(1).(*tensorcore.Table)
		assert.Equal(t, []string{"method", "A/Mean", "B/Mean"}, tv.Table.Columns.Keys)
		assert.Equal(t, "x", tv.Table.Column("method").String1D(1))
		assert.Equal(t, 4.5, tv.Table.Column("B/Mean").Float1D(1))
	}
	b.AsyncUnlock()

	vectors := node("Vectors")
	b.AsyncLock()
	closed := vectors.Closed
	b.AsyncUnlock()
	doubleClick(vectors)
	b.AsyncLock()
	assert.NotEqual(t, closed, vectors.Closed, "vectors toggled")
	assert.Nil(t, br.Tabs.TabByName("Sims/Vectors"), "no vectors tab")
	b.AsyncUnlock()

	parent := node("Parent")
	b.AsyncLock()
	closed = parent.Closed
	b.AsyncUnlock()
	doubleClick(parent)
	b.AsyncLock()
	assert.NotEqual(t, closed, parent.Closed, "toggled")
	assert.Nil(t, br.Tabs.TabByName("Sims/Parent"), "no edit tab")
	b.AsyncUnlock()
}
