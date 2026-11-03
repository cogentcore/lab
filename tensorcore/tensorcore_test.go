// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorcore

import (
	"image"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
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

// TestTableColumnWidths tests setting the widths of columns, including
// a tensor column, and dragging the handle after a header, after
// holding the mouse down.
func TestTableColumnWidths(t *testing.T) {
	b := core.NewBody()
	b.Styler(func(s *styles.Style) { s.Min.Set(units.Dp(800), units.Dp(400)) })
	pats := table.New()
	assert.NoError(t, pats.OpenCSV("testdata/random_5x5_25.tsv", tensor.Tab))
	tw := NewTable(b)
	tw.SetTable(pats)
	name, input := pats.Columns.Keys[0], pats.Columns.Keys[1]
	tw.ColumnWidths = map[string]float32{input: 200}
	changes := 0
	tw.OnChange(func(e events.Event) { changes++ })
	b.AssertRender(t, "table-column-widths", func() {
		b.AsyncLock()
		dots := tw.Styles.UnitContext.Dots(units.UnitDp)
		assert.InDelta(t, 200*dots, tw.columnDots(1), 1, "the tensor column")
		auto := tw.columnDots(0)
		hb := core.AsWidget(tw.SliceHeader().Child(1)).Geom.TotalBBox // the handle of the first
		b.AsyncUnlock()
		src := b.Scene.RenderWindow().SystemWindow.Events()
		pt := hb.Min.Add(hb.Size().Div(2))
		src.MouseButton(events.MouseDown, events.Left, pt, 0)
		time.Sleep(2 * core.TimingSettings.DragStartTime)
		for i := 1; i <= 10; i++ {
			src.MouseMove(pt.Add(image.Pt(4*i, 0)))
			time.Sleep(20 * time.Millisecond)
		}
		src.MouseButton(events.MouseUp, events.Left, pt.Add(image.Pt(40, 0)), 0)
		time.Sleep(200 * time.Millisecond)
		b.AsyncLock()
		defer b.AsyncUnlock()
		assert.InDelta(t, (auto+40)/dots, tw.ColumnWidths[name], 1)
		assert.InDelta(t, auto+40, tw.columnDots(0), 2)
		assert.Equal(t, 1, changes)
	})
}
