// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorcore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
)

func TestTable(t *testing.T) {
	b := core.NewBody()

	pats := table.New()
	err := pats.OpenCSV("testdata/random_5x5_25.tsv", tensor.Tab)
	assert.NoError(t, err)

	tw := NewTable(b)
	tw.SetTable(pats)
	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(tw.MakeToolbar)
	})
	b.AssertRender(t, "table")
}

func TestGrid(t *testing.T) {
	b := core.NewBody()

	pats := table.New()
	err := pats.OpenCSV("testdata/random_5x5_25.tsv", tensor.Tab)
	assert.NoError(t, err)

	gv := NewTensorGrid(b)
	tsr := pats.Column("Input").RowTensor(0).Clone()
	AddGridStylerTo(tsr, func(s *GridStyle) {
		s.ColumnRotation = 45
	})
	gv.SetTensor(tsr)
	gv.RowLabels = []string{"Row 0", "Row 1,2", "", "Row 3", "Row 4"}
	gv.ColumnLabels = []string{"Col 0,1", "", "Col 2", "Col 3", "Col 4"}
	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(gv.MakeToolbar)
	})
	b.AssertRender(t, "grid")
}

func TestTensor(t *testing.T) {
	b := core.NewBody()

	pats := table.New()
	err := pats.OpenCSV("testdata/random_5x5_25.tsv", tensor.Tab)
	assert.NoError(t, err)

	te := NewTensorEditor(b)
	tsr := pats.Column("Input").RowTensor(0).Clone()
	te.SetTensor(tsr)
	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(te.MakeToolbar)
	})
	b.AssertRender(t, "tensor")
}

// TestTableKeepScroll tests that setting the table that is already viewed,
// as when its data are updated, keeps the scroll position and selection,
// and that another table starts at the top.
func TestTableKeepScroll(t *testing.T) {
	dt := table.New()
	dt.AddIntColumn("N")
	dt.SetNumRows(100)
	b := core.NewBody()
	tw := NewTable(b)
	tw.SetReadOnly(true)
	tw.SetTable(dt)
	showed := make(chan struct{})
	b.OnFinal(events.Show, func(e events.Event) { showed <- struct{}{} })
	b.RunWindow()
	<-showed
	b.AsyncLock()
	defer func() {
		b.Close()
		b.AsyncUnlock()
	}()

	tw.ScrollToIndex(80)
	tw.SelectIndex(80)
	start := tw.StartIndex
	assert.Greater(t, start, 0)

	dt.SetNumRows(101) // a new row
	tw.SetTable(dt)
	assert.Equal(t, start, tw.StartIndex)
	assert.Contains(t, tw.SelectedIndexes, 80)

	other := table.New()
	other.AddIntColumn("N")
	other.SetNumRows(100)
	tw.SetTable(other)
	assert.Equal(t, 0, tw.StartIndex)
}
