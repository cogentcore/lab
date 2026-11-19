// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotcore

import (
	"image"
	"reflect"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/statsplot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensorfs"
)

func init() {
	core.AddValueType[statsplot.Column, ColumnChooser]()
}

// chooserColumns are the names of the columns for new [ColumnChooser]s,
// which are those of the table of the [Editor] whose stats plot dialog
// was opened last (dialogs are made one at a time, by the user).
var chooserColumns []string

// ColumnChooser represents a [statsplot.Column] value with a [core.Chooser]
// of the columns of the table of the [Editor] (or none).
type ColumnChooser struct {
	core.Chooser
}

func (cc *ColumnChooser) Init() {
	cc.Chooser.Init()
	items := make([]core.ChooserItem, len(chooserColumns)+1)
	items[0] = core.ChooserItem{Value: statsplot.Column(""), Text: "(none)", Tooltip: "no column"}
	for i, nm := range chooserColumns {
		items[i+1] = core.ChooserItem{Value: statsplot.Column(nm), Text: nm}
	}
	cc.SetItems(items...)
}

// statsPlotTypes are the types of statistical plots in the Stats menu.
var statsPlotTypes = []struct {
	name, doc string
	icon      icons.Icon
	new       func() statsplot.Plotter
}{
	{"Mean and error", "The mean of the values of columns for each X value, with the error of the mean as a band or error bars, e.g., across runs of a simulation", icons.StackedLineChart, func() statsplot.Plotter { return &statsplot.MeanError{} }},
	{"Histogram", "The histogram of the values of a column, for each value of a Split column", icons.BarChart, func() statsplot.Plotter { return &statsplot.Histogram{} }},
	{"Box plot", "Box plots of the values of a column, for each category", icons.CandlestickChart, func() statsplot.Plotter { return &statsplot.Box{} }},
	{"Violin plot", "Violin plots of the distribution of the values of a column, for each category", icons.QueryStats, func() statsplot.Plotter { return &statsplot.Violin{} }},
}

// statsColumns returns the default columns of a stats plot from the
// current plot: the X axis, the first Y column that is on, the
// columns that are on, the Split column, and the first column of
// strings other than the Split one, as the category.
func (pl *Editor) statsColumns() (x, y string, ys []string, split, category string) {
	x = pl.xAxisColumn()
	for ci, nm := range pl.table.Columns.Keys {
		ec := pl.columns[nm]
		if ec == nil {
			continue
		}
		st := ec.style
		switch {
		case st.Role == plot.Split && st.On:
			split = nm
		case st.On && st.Role == plot.Y && !ec.isString:
			if y == "" {
				y = nm
			}
			ys = append(ys, nm)
		}
		if ec.isString && category == "" && nm != split && st.Role != plot.Split {
			category = pl.table.Columns.Keys[ci]
		}
	}
	if category == split {
		category = ""
	}
	return
}

// defaultStatsPlot sets the default options of the given stats plot,
// with the columns of the current plot.
func (pl *Editor) defaultStatsPlot(sp statsplot.Plotter) {
	sp.Defaults()
	x, y, ys, split, cat := pl.statsColumns()
	switch sp := sp.(type) {
	case *statsplot.MeanError:
		sp.X = statsplot.Column(x)
		for _, c := range ys {
			sp.Values = append(sp.Values, statsplot.Column(c))
		}
		sp.Split = statsplot.Column(split)
	case *statsplot.Histogram:
		sp.Value = statsplot.Column(y)
		sp.Split = statsplot.Column(split)
	case *statsplot.Box:
		sp.Category, sp.Value, sp.Split = statsplot.Column(cat), statsplot.Column(y), statsplot.Column(split)
	case *statsplot.Violin:
		sp.Category, sp.Value, sp.Split = statsplot.Column(cat), statsplot.Column(y), statsplot.Column(split)
	}
}

// StatsPlot computes the given stats plot of the table of the editor,
// into the StatsDir directory, and shows its table with the
// [Editor.PlotTableFunc], or in a new window if it is nil.
func (pl *Editor) StatsPlot(name string, sp statsplot.Plotter) error {
	if pl.StatsDir == nil {
		pl.StatsDir = errors.Log1(tensorfs.NewDir("StatsPlot"))
	}
	dt, err := sp.Table(pl.StatsDir, pl.table)
	if err != nil {
		return err
	}
	pl.showTable(name, dt)
	return nil
}

// showTable shows the given table of a stats plot with the given name,
// with the [Editor.PlotTableFunc], or in a new window if it is nil.
func (pl *Editor) showTable(name string, dt *table.Table) {
	if pl.PlotTableFunc != nil {
		pl.PlotTableFunc(name, dt)
		return
	}
	b := core.NewBody(name)
	sp := NewSubPlot(b)
	sp.PlotTableFunc = pl.PlotTableFunc
	sp.SetTable(dt)
	b.RunWindow()
}

// statsPlotDialog opens a dialog with the options of the given type of
// stats plot, and makes the plot when the Plot button is pressed.
func (pl *Editor) statsPlotDialog(name string, sp statsplot.Plotter) {
	chooserColumns = pl.table.Columns.Keys
	pl.defaultStatsPlot(sp)
	d := core.NewBody(name)
	core.NewForm(d).SetStruct(sp)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).SetText("Plot").OnClick(func(e events.Event) {
			title := statsPlotTitle(name, sp)
			plotIt := func() {
				if err := pl.StatsPlot(title, sp); err != nil {
					core.ErrorSnackbar(pl, err)
				}
			}
			if core.TheApp.Platform().IsMobile() {
				plotIt()
				return
			}
			// the dialog is in another window
			pl.AsyncLock()
			plotIt()
			pl.AsyncUnlock()
		})
	})
	d.RunWindowDialog(pl)
}

// statsPlotTitle returns the title of a stats plot, from its value column.
func statsPlotTitle(name string, sp statsplot.Plotter) string {
	v := reflect.ValueOf(sp).Elem().FieldByName("Value")
	if v.IsValid() && v.String() != "" {
		return v.String() + " " + name
	}
	return name
}

// makeStatsMenu makes the menu of the types of stats plots.
func (pl *Editor) makeStatsMenu(m *core.Scene, pos image.Point) {
	for _, st := range statsPlotTypes {
		bt := core.NewButton(m).SetText(st.name).SetIcon(st.icon).SetTooltip(st.doc)
		bt.OnClick(func(e events.Event) {
			pl.statsPlotDialog(st.name, st.new())
		})
	}
}
