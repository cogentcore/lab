// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"fmt"
	"path/filepath"
	"strings"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/text/textcore"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plotcore"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorcore"
	"cogentcore.org/lab/tensorfs"
)

// Lab is the current Tabs, for yaegi / Go consistent access.
var Lab *Tabs

// Tabber is a [core.Tabs] based widget that has support for opening
// tabs for [plotcore.Editor] and [tensorcore.Table] editors,
// among others.
type Tabber interface {
	core.Tabber

	// AsLab returns the [lab.Tabs] widget with all the tabs methods.
	AsLab() *Tabs
}

// NewTab recycles a tab with given label, or returns the existing one
// with given type of widget within it. The existing that is returned
// is the last one in the frame, allowing for there to be a toolbar at the top.
// mkfun function is called to create and configure a new widget
// if not already existing.
func NewTab[T any](tb Tabber, label string, mkfun func(tab *core.Frame) T) T {
	tab := tb.AsLab().RecycleTab(label)
	var zv T
	if tab.HasChildren() {
		nc := tab.NumChildren()
		lc := tab.Child(nc - 1)
		if tt, ok := lc.(T); ok {
			return tt
		}
		err := fmt.Errorf("Name / Type conflict: tab %q does not have the expected type of content: is %T", label, lc)
		core.ErrorSnackbar(tb.AsLab(), err)
		return zv
	}
	w := mkfun(tab)
	return w
}

// TabAt returns widget of given type at tab of given name, nil if tab not found.
func TabAt[T any](tb Tabber, label string) T {
	var zv T
	tab := tb.AsLab().TabByName(label)
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
	core.ErrorSnackbar(tb.AsLab(), err)
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

func (ts *Tabs) AsLab() *Tabs {
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

// DirAndFileNoSlash returns [fsx.DirAndFile] with slashes replaced with spaces.
// Slashes are also used in core Widget paths, so spaces are safer.
func DirAndFileNoSlash(fpath string) string {
	return strings.ReplaceAll(fsx.DirAndFile(fpath), string(filepath.Separator), " ")
}

// GridTensorFS recycles a tab with a Grid of given [tensorfs.Node].
func (ts *Tabs) GridTensorFS(dfs *tensorfs.Node) *tensorcore.TensorGrid {
	label := DirAndFileNoSlash(dfs.Path()) + " Grid"
	if dfs.IsDir() {
		core.MessageSnackbar(ts, "Use Edit instead of Grid to view a directory")
		return nil
	}
	tsr := dfs.Tensor
	return ts.TensorGrid(label, tsr)
}

// PlotTable recycles a tab with a Plot of given table.Table.
func (ts *Tabs) PlotTable(label string, dt *table.Table) *plotcore.Editor {
	pl := NewTab(ts, label, func(tab *core.Frame) *plotcore.Editor {
		tb := core.NewToolbar(tab)
		pl := plotcore.NewEditor(tab)
		tab.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Grow.Set(1, 1)
		})
		tb.Maker(pl.MakeToolbar)
		return pl
	})
	if pl != nil {
		pl.SetTable(dt)
	}
	return pl
}

// PlotTensorFS recycles a tab with a Plot of given [tensorfs.Node].
func (ts *Tabs) PlotTensorFS(dfs *tensorfs.Node) *plotcore.Editor {
	label := DirAndFileNoSlash(dfs.Path()) + " Plot"
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

// Plot recycles a tab with given Plot using given label.
func (ts *Tabs) Plot(label string, plt *plot.Plot) *plotcore.Plot {
	pl := NewTab(ts, label, func(tab *core.Frame) *plotcore.Plot {
		pl := plotcore.NewPlot(tab)
		pl.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Grow.Set(1, 1)
		})
		pl.SetPlot(plt)
		return pl
	})
	if pl != nil {
		ts.Update()
	}
	return pl
}

// GoUpdatePlot calls GoUpdatePlot on plot at tab with given name.
// Does nothing if tab name doesn't exist (returns nil).
func (ts *Tabs) GoUpdatePlot(label string) *plotcore.Editor {
	pl := TabAt[*plotcore.Editor](ts, label)
	if pl != nil {
		pl.GoUpdatePlot()
	}
	return pl
}

// UpdatePlot calls UpdatePlot on plot at tab with given name.
// Does nothing if tab name doesn't exist (returns nil).
func (ts *Tabs) UpdatePlot(label string) *plotcore.Editor {
	pl := TabAt[*plotcore.Editor](ts, label)
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

// EditorString recycles a [textcore.Editor] tab, displaying given string.
func (ts *Tabs) EditorString(label, content string) *textcore.Editor {
	ed := NewTab(ts, label, func(tab *core.Frame) *textcore.Editor {
		ed := textcore.NewEditor(tab)
		ed.Styler(func(s *styles.Style) {
			s.Grow.Set(1, 1)
		})
		return ed
	})
	if content != "" {
		ed.Lines.SetText([]byte(content))
	}
	ts.Update()
	return ed
}

// EditorFile opens an editor tab for given file.
func (ts *Tabs) EditorFile(label, filename string) *textcore.Editor {
	ed := NewTab(ts, label, func(tab *core.Frame) *textcore.Editor {
		ed := textcore.NewEditor(tab)
		ed.Styler(func(s *styles.Style) {
			s.Grow.Set(1, 1)
		})
		return ed
	})
	ed.Lines.Open(filename)
	ts.Update()
	return ed
}
