// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"testing"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
)

// TestFileNodeDoubleClick tests that double clicking a tensorfs directory
// with only data edits it, and one with subdirectories opens and closes.
func TestFileNodeDoubleClick(t *testing.T) {
	root, _ := tensorfs.NewDir("root")
	sims := root.Dir("Sims")
	tensorfs.Value[float64](sims.Dir("Leaf"), "X", 3)
	tensorfs.Value[float64](sims.Dir("Parent").Dir("Child"), "Y", 3)
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

	parent := node("Parent")
	b.AsyncLock()
	closed := parent.Closed
	b.AsyncUnlock()
	doubleClick(parent)
	b.AsyncLock()
	assert.NotEqual(t, closed, parent.Closed, "toggled")
	assert.Nil(t, br.Tabs.TabByName("Sims/Parent"), "no edit tab")
	b.AsyncUnlock()
}
