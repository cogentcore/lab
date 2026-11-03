// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorcore

import (
	"fmt"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
)

// The column widths of [Table] and [TensorEditor], whose ColumnWidths have
// the widths of the columns, in Dp, by column name, for those whose widths
// are set, as the user does by dragging the handles after the headers of
// the columns, which sends a change event. The other columns have automatic
// widths. The widths are applied in the layout of the grid, for which both
// are a [core.ColumnWidther].

// headerName returns the name of the header of the column at the given
// index with the given name in the plan of the header, which includes the
// index, as it is in the closures, so that the header is remade when its
// column has another index.
func headerName(col int, name string) string {
	return fmt.Sprintf("head-%d-%s", col, name)
}

// makeColumnHandle adds the handle after the header of the column at the
// given index with the given name of the given list, with the given column
// widths, to the given plan, which sets its width by dragging, and makes it
// automatic again by double clicking.
func makeColumnHandle(p *tree.Plan, lb *core.ListBase, widths *map[string]float32, col int, name string) {
	tree.AddAt(p, fmt.Sprintf("width-%d-%s", col, name), func(w *core.Handle) { // see headerName
		w.SetTooltip("Drag to set the width of the column, or double click for the automatic width")
		w.Styler(func(s *styles.Style) {
			s.Direction = styles.Row
			s.SetAbilities(false, abilities.Focusable)
			s.SetAbilities(true, abilities.DoubleClickable)
		})
		w.FinalStyler(func(s *styles.Style) {
			s.Min.X.Dp(3)
			s.Min.Y.Em(1.5)
			s.Margin.SetHorizontal(units.Dp(2))
		})
		var start float32 // the width of the column when the drag starts, in dots
		w.On(events.SlideStart, func(e events.Event) {
			start = columnDots(lb, col)
		})
		w.On(events.SlideMove, func(e events.Event) {
			e.SetHandled()
			if *widths == nil {
				*widths = map[string]float32{}
			}
			dots := start + float32(e.Pos().X-e.StartPos().X)
			(*widths)[name] = max(dots/lb.Styles.UnitContext.Dots(units.UnitDp), 16)
			lb.Restyle()
			lb.NeedsLayout()
		})
		w.On(events.SlideStop, func(e events.Event) {
			lb.SendChange(e)
		})
		w.OnDoubleClick(func(e events.Event) {
			delete(*widths, name)
			lb.Restyle()
			lb.NeedsLayout()
			lb.SendChange(e)
		})
	})
}

// columnWidthDots returns the given width of a column in Dp in dots,
// for the given list.
func columnWidthDots(lb *core.ListBase, width float32) float32 {
	return width * lb.Styles.UnitContext.Dots(units.UnitDp)
}

// columnDots returns the width of the column at the given index
// of the given list, in dots, without the gap after it.
func columnDots(lb *core.ListBase, col int) float32 {
	_, idxOff := lb.This.(core.Lister).RowWidgetNs()
	if lb.ListGrid == nil || idxOff+col >= lb.ListGrid.NumChildren() {
		return 0
	}
	return core.AsWidget(lb.ListGrid.Child(idxOff + col)).Geom.Size.Alloc.Total.X
}

// sizeHeader sets the widths of the headers of the columns of the given
// list, with the given number of columns, to those of their columns: each
// header, with the handle after it, has the width of its column, with the
// gap after it, as the header has no gaps.
func sizeHeader(lb *core.ListBase, header *core.Frame, nCols int) {
	lg := lb.ListGrid
	_, idxOff := lb.This.(core.Lister).RowWidgetNs()
	if lg == nil || header.NumChildren() != idxOff+2*nCols || lg.NumChildren() < idxOff+nCols {
		return
	}
	gap := lg.Styles.Gap.X.Dots
	setWidth := func(w *core.WidgetBase, width float32) {
		ksz := &w.Geom.Size
		d := width - ksz.Actual.Total.X
		ksz.Actual.Total.X += d
		ksz.Actual.Content.X += d
		ksz.Alloc.Total.X = ksz.Actual.Total.X
		ksz.Alloc.Content.X = ksz.Actual.Content.X
	}
	colWidth := func(i int) float32 {
		return core.AsWidget(lg.Child(i)).Geom.Size.Alloc.Total.X + gap
	}
	if idxOff > 0 && colWidth(0) > gap {
		setWidth(core.AsWidget(header.Child(0)), colWidth(0))
	}
	for col := range nCols {
		if colWidth(idxOff+col) <= gap { // not laid out yet
			continue
		}
		hi := idxOff + 2*col
		handle := core.AsWidget(header.Child(hi + 1)).Geom.Size.Actual.Total.X
		setWidth(core.AsWidget(header.Child(hi)), colWidth(idxOff+col)-handle)
	}
}
