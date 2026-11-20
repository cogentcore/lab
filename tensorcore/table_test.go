// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorcore

import (
	"image"
	"testing"
	"time"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/keymap"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestTable returns a table with a string, a float and a 2x2 tensor
// column, with the given number of rows: the names are in reverse order
// and the values are the row indexes.
func newTestTable(rows int) *table.Table {
	dt := table.New("people")
	names := dt.AddStringColumn("Name")
	vals := dt.AddFloat64Column("Value")
	cells := dt.AddFloat64Column("Cells", 2, 2)
	dt.SetNumRows(rows)
	for i := range rows {
		names.SetString1D(string(rune('a'+rows-1-i)), i)
		vals.SetFloat1D(float64(i), i)
		cells.SetFloatRow(float64(i)/float64(rows), i, 0)
	}
	return dt
}

// valueWidget returns the value widget of the given table at the given
// visible row and column.
func valueWidget(lb *core.ListBase, row, col int) core.Widget {
	nw, idxOff := lb.This.(core.Lister).RowWidgetNs()
	return lb.ListGrid.Child(row*nw + idxOff + col).(core.Widget)
}

// headerButton returns the header button of the given column.
func headerButton(lb *core.ListBase, header *core.Frame, col int) *core.Button {
	_, idxOff := lb.This.(core.Lister).RowWidgetNs()
	return header.Child(idxOff + 2*col).(*core.Button)
}

// windowScene waits for a window with the given title to be open,
// and returns its main scene.
func windowScene(t *testing.T, title string) *core.Scene {
	t.Helper()
	var sc *core.Scene
	assert.Eventually(t, func() bool {
		if w := core.AllRenderWindows.FindName(title); w != nil {
			sc = w.MainScene()
		}
		return sc != nil
	}, 5*time.Second, 10*time.Millisecond, "the window %q", title)
	return sc
}

// TestTableSort tests sorting the table by clicking the headers of the
// columns, which toggles the order, and setting the sort by name.
func TestTableSort(t *testing.T) {
	b := core.NewBody()
	dt := newTestTable(5)
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, "", tw.SortFieldName())
		headerButton(&tw.ListBase, tw.SliceHeader(), 0).Send(events.Click)
		assert.Equal(t, 0, tw.SortIndex)
		assert.Equal(t, "Name:up", tw.SortFieldName())
		assert.Equal(t, "a", tw.Table.Column("Name").String1D(0))
		assert.Equal(t, 4.0, tw.Table.Column("Value").Float1D(0))

		headerButton(&tw.ListBase, tw.SliceHeader(), 0).Send(events.Click)
		assert.Equal(t, "Name:down", tw.SortFieldName())
		assert.Equal(t, "e", tw.Table.Column("Name").String1D(0))
		assert.Equal(t, icons.KeyboardArrowDown, headerButton(&tw.ListBase, tw.SliceHeader(), 0).Indicator)

		headerButton(&tw.ListBase, tw.SliceHeader(), 1).Send(events.Click)
		assert.Equal(t, "Value:up", tw.SortFieldName())
		assert.Equal(t, 0.0, tw.Table.Column("Value").Float1D(0))
		assert.Equal(t, icons.KeyboardArrowUp, headerButton(&tw.ListBase, tw.SliceHeader(), 1).Indicator)
		assert.Equal(t, icons.Blank, headerButton(&tw.ListBase, tw.SliceHeader(), 0).Indicator)

		tw.SortColumn(-1)
		assert.Equal(t, "", tw.SortFieldName())

		tw.SetSortFieldName("Value:down")
		assert.Equal(t, 1, tw.SortIndex)
		assert.True(t, tw.SortDescending)
		tw.SetSortFieldName("Name:up")
		assert.Equal(t, 0, tw.SortIndex)
		assert.False(t, tw.SortDescending)
		tw.SetSortFieldName("")
		assert.Equal(t, 0, tw.SortIndex)
	}))
}

// TestTableEdit tests editing the values of a table by typing into its
// string and number fields, which sets the values of the table at the
// rows of the view, and sends a change event.
func TestTableEdit(t *testing.T) {
	b := core.NewBody()
	dt := newTestTable(5)
	tw := NewTable(b)
	tw.SetTable(dt)
	changes := 0
	tw.OnChange(func(e events.Event) { changes++ })
	core.RenderImage(b, core.Locked(b, func() {
		tw.SortColumn(0) // a view in another order: row 0 is the last one
		tf := valueWidget(&tw.ListBase, 0, 0).(*core.TextField)
		assert.Equal(t, "a", tf.Text())
		tf.SetText("a much longer name")
		tf.SendChange()
		assert.Equal(t, "a much longer name", dt.Column("Name").String1D(4))
		assert.Equal(t, len("a much longer name"), tw.colMaxWidths[0])
		assert.Equal(t, 1, changes)

		sp := valueWidget(&tw.ListBase, 1, 1).(*core.Spinner)
		assert.Equal(t, float32(3), sp.Value)
		sp.SetValue(42)
		sp.SendChange()
		assert.Equal(t, 42.0, dt.Column("Value").Float1D(3))
		assert.Equal(t, 2, changes)

		tw.SetReadOnly(true)
		tw.Update()
		assert.True(t, valueWidget(&tw.ListBase, 0, 0).AsWidget().IsReadOnly())
	}))
}

// TestTableRowsBlank tests that the rows after the end of a short table,
// including the tensor cells, are invisible blanks, and the index
// column that is shown with ShowIndexes.
func TestTableRowsBlank(t *testing.T) {
	b := core.NewBody()
	b.Styler(func(s *styles.Style) { s.Min.Set(units.Dp(600), units.Dp(600)) })
	dt := newTestTable(2)
	tw := NewTable(b)
	tw.ShowIndexes = true
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		require.Greater(t, tw.VisibleRows, 2)
		nw, idxOff := tw.RowWidgetNs()
		assert.Equal(t, 4, nw)
		assert.Equal(t, 1, idxOff)
		blank := valueWidget(&tw.ListBase, 2, 2).(*TensorGrid)
		assert.True(t, blank.StateIs(0) || !blank.IsVisible())
		assert.Same(t, tw.blankCells[2], blank.Tensor)
		assert.Equal(t, []int{2, 2}, blank.Tensor.ShapeSizes())
		assert.Equal(t, 0.0, blank.Tensor.Float1D(0))
		cell := valueWidget(&tw.ListBase, 1, 2).(*TensorGrid)
		assert.Equal(t, 0.5, cell.Tensor.Float1D(0))

		w, ok := tw.RowFirstVisWidget(0)
		assert.True(t, ok)
		assert.NotNil(t, w)
		_, ok = tw.RowFirstVisWidget(-1)
		assert.False(t, ok)

		fw := tw.RowGrabFocus(1)
		require.NotNil(t, fw)
		assert.Same(t, fw, tw.RowGrabFocus(1), "already has the focus")
		assert.Nil(t, tw.RowGrabFocus(100))
	}))
}

// TestTableNewDeleteRows tests inserting and deleting rows in the table.
func TestTableNewDeleteRows(t *testing.T) {
	b := core.NewBody()
	dt := newTestTable(3)
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		tw.NewAt(1)
		assert.Equal(t, 4, dt.NumRows())
		assert.Equal(t, 4, tw.SliceSize)
		assert.Equal(t, "", tw.Table.Column("Name").String1D(1))
		assert.Equal(t, 1, tw.SelectedIndex)

		tw.DeleteAt(0) // from the view only
		assert.Equal(t, 3, tw.Table.NumRows())
		assert.Equal(t, 3, tw.SliceSize)
		assert.Equal(t, 4, dt.NumRows())
		assert.Equal(t, "", tw.Table.Column("Name").String1D(0))
		tw.DeleteAt(10) // out of range
		assert.Equal(t, 3, tw.Table.NumRows())
	}))
}

// TestTableCopyPaste tests copying the selected rows as CSV data, and
// pasting them into the table, assigning to and inserting rows,
// with the selected strings of a column.
func TestTableCopyPaste(t *testing.T) {
	b := core.NewBody()
	dt := newTestTable(4)
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Nil(t, tw.CopySelectToMime(), "no selection")
		assert.Nil(t, tw.SelectedColumnStrings("Name"))
		assert.Equal(t, fileinfo.DataCSV, tw.MimeDataType())

		tw.SortColumn(0) // a, b, c, d = rows 3, 2, 1, 0
		tw.SelectIndexEvent(0, events.SelectOne)
		tw.SelectIndexEvent(1, events.ExtendContinuous)
		assert.Equal(t, []string{"a", "b"}, tw.SelectedColumnStrings("Name"))
		md := tw.CopySelectToMime()
		require.Len(t, md, 1)
		assert.Equal(t, fileinfo.DataCSV, md[0].Type)
		recs := tw.FromMimeData(md)
		require.Len(t, recs, 3, "the header and the rows")
		assert.Equal(t, "a", recs[1][0])
		assert.Equal(t, "3", recs[1][1])
		assert.Equal(t, "b", recs[2][0])

		tw.PasteAssign(md, 3) // row d = row 0
		assert.Equal(t, "a", dt.Column("Name").String1D(0))
		assert.Equal(t, 3.0, dt.Column("Value").Float1D(0))

		tw.PasteAtIndex(md, 1)
		assert.Equal(t, 6, dt.NumRows())
		assert.Equal(t, 0, tw.SelectedIndex-1)

		tw.PasteAssign(nil, 0)
		tw.PasteAtIndex(nil, 0)
		assert.Equal(t, 6, dt.NumRows())
	}))
}

// TestTableSearch tests searching the values of the table, and
// highlighting and selecting the matches, as the Find of a window does,
// in a view of the table in another order.
func TestTableSearch(t *testing.T) {
	b := core.NewBody()
	dt := newTestTable(5)
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Nil(t, tw.TextRunes())
		tw.SortColumn(0) // a = row 4 is first
		ms, handled := tw.Search("a", core.IgnoreCase)
		assert.True(t, handled)
		require.Len(t, ms, 1)
		assert.Equal(t, 0, ms[0].Region.Start.Line/tw.nCols, "the row of the view")
		assert.Equal(t, 0, ms[0].Region.Start.Line%tw.nCols, "the Name column")
		ms, _ = tw.Search("4", core.IgnoreCase) // also 0.4 in the cells of the row of c
		require.Len(t, ms, 2)
		assert.Equal(t, 2, ms[1].Region.Start.Line/tw.nCols, "the row of the view")
		assert.Equal(t, 2, ms[1].Region.Start.Line%tw.nCols, "the Cells column")
		assert.Equal(t, 0, ms[0].Region.Start.Line/tw.nCols, "the row of the view")
		assert.Equal(t, 1, ms[0].Region.Start.Line%tw.nCols, "the Value column")

		assert.True(t, tw.HighlightMatches(ms))
		tw.SelectMatch(ms, 0, core.SelectScroll, core.SelectSet)
		assert.True(t, tw.HighlightMatches(nil))
	}))
}

// TestTableSearchScroll tests selecting a search match in a row that is
// not visible, which scrolls to it.
func TestTableSearchScroll(t *testing.T) {
	b := core.NewBody()
	dt := table.New()
	ns := dt.AddIntColumn("N")
	dt.SetNumRows(200)
	for i := range 200 {
		ns.SetInt1D(i, i)
	}
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		ms, _ := tw.Search("150", core.IgnoreCase)
		require.Len(t, ms, 1)
		assert.Equal(t, 0, tw.StartIndex)
		tw.SelectMatch(ms, 0, core.SelectNoScroll, core.SelectSet) // not scrolled
		assert.Equal(t, 0, tw.StartIndex)
		tw.SelectMatch(ms, 0, core.SelectScroll, core.SelectSet)
		assert.LessOrEqual(t, tw.StartIndex, 150)
		assert.Greater(t, tw.StartIndex+tw.VisibleRows, 150)
	}))
}

// TestTableToolbar tests the actions of the toolbar of the table:
// adding rows, and filtering and unfiltering it.
func TestTableToolbar(t *testing.T) {
	b := core.NewBody()
	dt := newTestTable(5)
	tw := NewTable(b)
	tw.SetTable(dt)
	var tb *core.Toolbar
	b.AddTopBar(func(bar *core.Frame) {
		tb = core.NewToolbar(bar)
		tb.Maker(tw.MakeToolbar)
	})
	core.RenderImage(b, core.Locked(b, func() {
		assert.NotNil(t, core.ButtonByText(tb, "Sort"))
		assert.NotNil(t, core.ButtonByText(tb, "Filter"))
		require.NoError(t, tw.Table.FilterString("Name", "b", tensor.StringMatch{}))
		tw.Update()
		assert.Equal(t, 1, tw.SliceSize)
		core.ButtonByText(tb, "Unfilter").Send(events.Click)
		assert.Equal(t, 5, tw.SliceSize)
	}))

	empty := NewTable(core.NewBody())
	assert.NotPanics(t, func() { core.NewToolbar(core.NewBody()).Maker(empty.MakeToolbar) })
}

// TestTableAsyncUpdate tests updating the table from another goroutine
// after rows are added, which shows the last row.
func TestTableAsyncUpdate(t *testing.T) {
	b := core.NewBody()
	dt := table.New()
	dt.AddIntColumn("N")
	dt.SetNumRows(5)
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, func() {
		b.AsyncLock()
		dt.SetNumRows(200)
		b.AsyncUnlock()
		tw.AsyncUpdateTable()
		b.AsyncLock()
		defer b.AsyncUnlock()
		assert.Equal(t, 200, tw.SliceSize)
		assert.Greater(t, tw.StartIndex, 0)
	})
}

// TestTableSetSlice tests viewing a slice of structs as a table.
func TestTableSetSlice(t *testing.T) {
	type city struct {
		Area       float32
		Population float64
	}
	b := core.NewBody()
	tw := NewTable(b)
	tw.SetSlice([]city{{500, 62000}, {800, 85000}})
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, 2, tw.SliceSize)
		assert.Equal(t, []string{"Area", "Population"}, tw.Table.Columns.Keys)
		assert.Equal(t, float32(85000), valueWidget(&tw.ListBase, 1, 1).(*core.Spinner).Value)
	}))
}

// TestTableGridStyle tests the grid styles of the tensor columns: the
// stylers in their metadata, and the dialogs of the context menu of the
// header of a tensor column, which edit the style of the column and the
// global one.
func TestTableGridStyle(t *testing.T) {
	defer core.TrackWindows()()
	b := core.NewBody()
	dt := newTestTable(3)
	AddGridStylerTo(dt.Column("Cells").Tensor, func(s *GridStyle) { s.GridFill = 0.5 })
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		gs := tw.GetColumnGridStyle(2)
		assert.Equal(t, float32(0.5), gs.GridFill)
		assert.Equal(t, float32(0.9), tw.GetColumnGridStyle(1).GridFill)
		cell := valueWidget(&tw.ListBase, 0, 2).(*TensorGrid)
		assert.Equal(t, float32(0.5), cell.GridStyle.GridFill)

		hdr := headerButton(&tw.ListBase, tw.SliceHeader(), 2)
		require.Len(t, hdr.ContextMenus, 1, "only for a tensor column")
		assert.Empty(t, headerButton(&tw.ListBase, tw.SliceHeader(), 1).ContextMenus)
		m := core.NewScene()
		hdr.ContextMenus[0](m, image.Point{})
		core.ButtonByText(m, "Edit grid style").Send(events.Click)
	}), func() {
		sc := windowScene(t, "Tensor grid style")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			fm := findWidget[*core.Form](sc)
			require.NotNil(t, fm)
			fm.SendChange()
			core.ButtonByText(sc, "Edit global style").Send(events.Click)
		})()
	}, core.Locked(b, func() {
		assert.NotNil(t, tw.ColumnGridStyle[2], "set by the change")
	}))
}

// findWidget returns the first widget of the given type in the given
// widget and its descendants.
func findWidget[T core.Widget](w core.Widget) T {
	var zero T
	for cw := range w.AsWidget().AllWidgets() {
		if x, ok := cw.(T); ok {
			return x
		}
	}
	return zero
}

// TestValueButtons tests the buttons of the tensor and table values, whose
// text is the name of the table, and that open their editors.
func TestValueButtons(t *testing.T) {
	defer core.TrackWindows()()
	dt := newTestTable(2)
	metadata.SetName(dt, "people")
	unnamed := table.New()
	tsr := tensor.NewFloat64(2, 2)
	b := core.NewBody()
	tbn := core.Bind(&dt, NewTableButton(b))
	tbu := core.Bind(&unnamed, NewTableButton(b))
	var nilTable *table.Table
	tbz := core.Bind(&nilTable, NewTableButton(b))
	var tt tensor.Tensor = tsr
	tsb := core.Bind(&tt, NewTensorButton(b))
	var nt tensor.Tensor
	tsz := core.Bind(&nt, NewTensorButton(b))
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, "people", tbn.Text)
		assert.Equal(t, "Table", tbu.Text)
		assert.Equal(t, "None", tbz.Text)
		assert.Equal(t, "Tensor", tsb.Text)
		assert.Equal(t, "None", tsz.Text)
		assert.Same(t, dt, *tbn.WidgetValue().(**table.Table))
		tbn.Send(events.Click)
		tsb.Send(events.Click)
	}))
}

// TestTableKeys tests copying the selected rows with the copy key,
// as CSV data in the clipboard.
func TestTableKeys(t *testing.T) {
	b := core.NewBody()
	dt := newTestTable(3)
	tw := NewTable(b)
	tw.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		tw.SelectIndexEvent(2, events.SelectOne)
		core.SendKeyFunction(tw, keymap.Copy)
		md := tw.Clipboard().Read([]string{fileinfo.DataCSV})
		recs := tw.FromMimeData(md)
		require.Len(t, recs, 2)
		assert.Equal(t, "a", recs[1][0])
	}))
}

// tableNames returns the names of the rows of the given table, in the
// order of its view, as a [Table] shows them (its edits change the
// indexes of its view of the table).
func tableNames(dt *table.Table) []string {
	var names []string
	for i := range dt.NumRows() {
		names = append(names, dt.Column("Name").String1D(i))
	}
	return names
}

// dragTableRow drags the given row of the given table by its index to
// the other given row of the other given table (which can be the same),
// and inserts it after that row with the menu of the drop.
func dragTableRow(t *testing.T, b *core.Body, from *Table, fromRow int, to *Table, toRow int) {
	t.Helper()
	var fp, tp image.Point
	core.Locked(b, func() {
		index := func(tw *Table, row int) image.Point {
			nw, _ := tw.RowWidgetNs()
			bb := tw.ListGrid.Child(row * nw).(core.Widget).AsWidget().Geom.TotalBBox
			return bb.Min.Add(bb.Size().Div(2))
		}
		fp, tp = index(from, fromRow), index(to, toRow)
	})()
	core.DragMouse(b, fp, tp)
	var after *core.Button
	require.Eventually(t, func() bool {
		core.Locked(b, func() {
			if pops := core.Popups(b); len(pops) > 0 {
				after = core.ButtonByText(pops[len(pops)-1], "Insert after")
			}
		})()
		return after != nil
	}, 5*time.Second, 20*time.Millisecond, "the drop menu")
	core.Locked(b, func() { after.Send(events.Click) })()
	time.Sleep(100 * time.Millisecond)
}

// TestTableDragMoves tests dragging the rows of tables by their indexes:
// with DragMoves, a drag moves the row that is dragged, also to another
// table, deleting it from its table, and without it, a drag copies it.
func TestTableDragMoves(t *testing.T) {
	b := core.NewBody()
	b.Styler(func(s *styles.Style) { s.Min.Set(units.Dp(600), units.Dp(600)) })
	dt := newTestTable(4) // d c b a
	tw := NewTable(b)
	tw.ShowIndexes = true
	tw.DragMoves = true
	tw.SetTable(dt)
	dt2 := newTestTable(2) // b a
	tw2 := NewTable(b)
	tw2.ShowIndexes = true
	tw2.SetTable(dt2)
	core.RenderImage(b, func() {
		dragTableRow(t, b, tw, 0, tw, 2)
		core.Locked(b, func() {
			assert.Equal(t, []string{"c", "b", "d", "a"}, tableNames(tw.Table), "d moved after b")
		})()

		dragTableRow(t, b, tw, 3, tw2, 0)
		core.Locked(b, func() {
			assert.Equal(t, []string{"b", "a", "a"}, tableNames(tw2.Table), "a moved after b in the other table")
			assert.Equal(t, []string{"c", "b", "d"}, tableNames(tw.Table), "a deleted from its table")
		})()

		dragTableRow(t, b, tw2, 0, tw2, 1)
		core.Locked(b, func() {
			assert.Equal(t, []string{"b", "a", "b", "a"}, tableNames(tw2.Table), "b copied after a, without DragMoves")
		})()
	})
}
