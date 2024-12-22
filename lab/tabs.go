// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"fmt"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/texteditor"
	"cogentcore.org/lab/plotcore"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorcore"
	"cogentcore.org/lab/tensorfs"
)

// CurTabber is the current Tabber. Set when one is created.
var CurTabber Tabber

// Tabber is a [core.Tabs] based widget that has support for opening
// tabs for [plotcore.PlotEditor] and [tensorcore.Table] editors,
// among others.
type Tabber interface {
	core.Tabber

	// AsDataTabs returns the underlying [databrowser.Tabs] widget.
	AsDataTabs() *Tabs

	// TensorTable recycles a tab with a [tensorcore.Table] widget
	// to view given [table.Table], using its own table.Table.
	TensorTable(label string, dt *table.Table) *tensorcore.Table

	// TensorEditor recycles a tab with a [tensorcore.TensorEditor] widget
	// to view given Tensor.
	TensorEditor(label string, tsr tensor.Tensor) *tensorcore.TensorEditor

	// TensorGrid recycles a tab with a [tensorcore.TensorGrid] widget
	// to view given Tensor.
	TensorGrid(label string, tsr tensor.Tensor) *tensorcore.TensorGrid

	// PlotTable recycles a tab with a Plot of given [table.Table].
	PlotTable(label string, dt *table.Table) *plotcore.PlotEditor

	// PlotTensorFS recycles a tab with a Plot of given [tensorfs.Node],
	// automatically using the Dir/File name of the data node for the label.
	PlotTensorFS(dfs *tensorfs.Node) *plotcore.PlotEditor

	// GoUpdatePlot calls GoUpdatePlot on plot at tab with given name.
	// Does nothing if tab name doesn't exist (returns nil).
	GoUpdatePlot(label string) *plotcore.PlotEditor

	// UpdatePlot calls UpdatePlot on plot at tab with given name.
	// Does nothing if tab name doesn't exist (returns nil).
	UpdatePlot(label string) *plotcore.PlotEditor

	// todo: PlotData of plot.Node

	// SliceTable recycles a tab with a [core.Table] widget
	// to view the given slice of structs.
	SliceTable(label string, slc any) *core.Table

	// EditorString recycles a [texteditor.Editor] tab, displaying given string.
	EditorString(label, content string) *texteditor.Editor

	// EditorFile opens an editor tab for given file.
	EditorFile(label, filename string) *texteditor.Editor
}

// NewTab recycles a tab with given label, or returns the existing one
// with given type of widget within it. The existing that is returned
// is the last one in the frame, allowing for there to be a toolbar at the top.
// mkfun function is called to create and configure a new widget
// if not already existing.
func NewTab[T any](tb Tabber, label string, mkfun func(tab *core.Frame) T) T {
	tab := tb.RecycleTab(label)
	var zv T
	if tab.HasChildren() {
		nc := tab.NumChildren()
		lc := tab.Child(nc - 1)
		if tt, ok := lc.(T); ok {
			return tt
		}
		err := fmt.Errorf("Name / Type conflict: tab %q does not have the expected type of content: is %T", label, lc)
		core.ErrorSnackbar(tb.AsDataTabs(), err)
		return zv
	}
	w := mkfun(tab)
	return w
}

// TabAt returns widget of given type at tab of given name, nil if tab not found.
func TabAt[T any](tb Tabber, label string) T {
	var zv T
	tab := tb.TabByName(label)
	if tab == nil {
		return zv
	}
	if !tab.HasChildren() { // shouldn't happen
		return zv
	}
	nc := tab.NumChildren()
	lc := tab.Child(nc - 1)
	if tt, ok := lc.(T); ok {
		return tt
	}

	err := fmt.Errorf("Name / Type conflict: tab %q does not have the expected type of content: %T", label, lc)
	core.ErrorSnackbar(tb.AsDataTabs(), err)
	return zv
}

// Tabs implements the [Tabber] interface.
type Tabs struct {
	core.Tabs
}

func (ts *Tabs) Init() {
	ts.Tabs.Init()
	ts.Type = core.FunctionalTabs
}

func (ts *Tabs) AsDataTabs() *Tabs {
	return ts
}

// TensorTable recycles a tab with a tensorcore.Table widget
// to view given table.Table, using its own table.Table as tv.Table.
// Use tv.Table.Table to get the underlying *table.Table
// Use tv.Table.Sequential to update the Indexed to view
// all of the rows when done updating the Table, and then call br.Update()
func (ts *Tabs) TensorTable(label string, dt *table.Table) *tensorcore.Table {
	tv := NewTab(ts, label, func(tab *core.Frame) *tensorcore.Table {
		tb := core.NewToolbar(tab)
		tv := tensorcore.NewTable(tab)
		tb.Maker(tv.MakeToolbar)
		return tv
	})
	tv.SetTable(dt)
	ts.Update()
	return tv
}

// TensorEditor recycles a tab with a tensorcore.TensorEditor widget
// to view given Tensor.
func (ts *Tabs) TensorEditor(label string, tsr tensor.Tensor) *tensorcore.TensorEditor {
	tv := NewTab(ts, label, func(tab *core.Frame) *tensorcore.TensorEditor {
		tb := core.NewToolbar(tab)
		tv := tensorcore.NewTensorEditor(tab)
		tb.Maker(tv.MakeToolbar)
		return tv
	})
	tv.SetTensor(tsr)
	ts.Update()
	return tv
}

// TensorGrid recycles a tab with a tensorcore.TensorGrid widget
// to view given Tensor.
func (ts *Tabs) TensorGrid(label string, tsr tensor.Tensor) *tensorcore.TensorGrid {
	tv := NewTab(ts, label, func(tab *core.Frame) *tensorcore.TensorGrid {
		// tb := core.NewToolbar(tab)
		tv := tensorcore.NewTensorGrid(tab)
		// tb.Maker(tv.MakeToolbar)
		return tv
	})
	tv.SetTensor(tsr)
	ts.Update()
	return tv
}

// PlotTable recycles a tab with a Plot of given table.Table.
func (ts *Tabs) PlotTable(label string, dt *table.Table) *plotcore.PlotEditor {
	pl := NewTab(ts, label, func(tab *core.Frame) *plotcore.PlotEditor {
		tb := core.NewToolbar(tab)
		pl := plotcore.NewPlotEditor(tab)
		tab.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Grow.Set(1, 1)
		})
		tb.Maker(pl.MakeToolbar)
		return pl
	})
	if pl != nil {
		pl.SetTable(dt)
		ts.Update()
	}
	return pl
}

// PlotTensorFS recycles a tab with a Plot of given [tensorfs.Node].
func (ts *Tabs) PlotTensorFS(dfs *tensorfs.Node) *plotcore.PlotEditor {
	label := fsx.DirAndFile(dfs.Path()) + " Plot"
	if dfs.IsDir() {
		return ts.PlotTable(label, tensorfs.DirTable(dfs, nil))
	}
	tsr := dfs.Tensor
	dt := table.New(label)
	dt.Columns.Rows = tsr.DimSize(0)
	if ix, ok := tsr.(*tensor.Rows); ok {
		dt.Indexes = ix.Indexes
	}
	rc := dt.AddIntColumn("Row")
	for r := range dt.Columns.Rows {
		rc.Values[r] = r
	}
	dt.AddColumn(dfs.Name(), tsr.AsValues())
	return ts.PlotTable(label, dt)
}

// GoUpdatePlot calls GoUpdatePlot on plot at tab with given name.
// Does nothing if tab name doesn't exist (returns nil).
func (ts *Tabs) GoUpdatePlot(label string) *plotcore.PlotEditor {
	pl := TabAt[*plotcore.PlotEditor](ts, label)
	if pl != nil {
		pl.GoUpdatePlot()
	}
	return pl
}

// UpdatePlot calls UpdatePlot on plot at tab with given name.
// Does nothing if tab name doesn't exist (returns nil).
func (ts *Tabs) UpdatePlot(label string) *plotcore.PlotEditor {
	pl := TabAt[*plotcore.PlotEditor](ts, label)
	if pl != nil {
		pl.UpdatePlot()
	}
	return pl
}

// SliceTable recycles a tab with a core.Table widget
// to view the given slice of structs.
func (ts *Tabs) SliceTable(label string, slc any) *core.Table {
	tv := NewTab(ts, label, func(tab *core.Frame) *core.Table {
		return core.NewTable(tab)
	})
	tv.SetSlice(slc)
	ts.Update()
	return tv
}

// EditorString recycles a [texteditor.Editor] tab, displaying given string.
func (ts *Tabs) EditorString(label, content string) *texteditor.Editor {
	ed := NewTab(ts, label, func(tab *core.Frame) *texteditor.Editor {
		ed := texteditor.NewEditor(tab)
		ed.Styler(func(s *styles.Style) {
			s.Grow.Set(1, 1)
		})
		return ed
	})
	if content != "" {
		ed.Buffer.SetText([]byte(content))
	}
	ts.Update()
	return ed
}

// EditorFile opens an editor tab for given file.
func (ts *Tabs) EditorFile(label, filename string) *texteditor.Editor {
	ed := NewTab(ts, label, func(tab *core.Frame) *texteditor.Editor {
		ed := texteditor.NewEditor(tab)
		ed.Styler(func(s *styles.Style) {
			s.Grow.Set(1, 1)
		})
		return ed
	})
	ed.Buffer.Open(core.Filename(filename))
	ts.Update()
	return ed
}