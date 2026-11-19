// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotcore

import (
	"image"
	"math/rand"
	"testing"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/plot/statsplot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newResultsTable returns a table of simulated results in two conditions
// (Cond, the Split), three categories (Cat), and 10 epochs (the X axis),
// with an error (Err) that is on.
func newResultsTable() *table.Table {
	rng := rand.New(rand.NewSource(1))
	dt := table.New("Results")
	epoch := dt.AddIntColumn("Epoch")
	cond := dt.AddStringColumn("Cond")
	cat := dt.AddStringColumn("Cat")
	errc := dt.AddFloat64Column("Err")
	for ci, cnd := range []string{"A", "B"} {
		for r := range 5 {
			for e := range 10 {
				dt.AddRows(1)
				i := dt.NumRows() - 1
				epoch.SetInt1D(e, i)
				cond.SetString1D(cnd, i)
				cat.SetString1D([]string{"X", "Y", "Z"}[(r+e)%3], i)
				errc.SetFloat1D(1/float64(1+e+ci)+0.05*rng.NormFloat64(), i)
			}
		}
	}
	plot.SetStyler(epoch, func(s *plot.Style) { s.Role = plot.X })
	plot.SetStyler(cond, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Split
	})
	plot.SetStyler(errc, func(s *plot.Style) { s.On = true })
	return dt
}

// TestEditorStatsPlot tests the default options of the stats plots from
// the columns of the editor, making them, the column chooser of their
// options, the Stats menu, and the dialog of a stats plot.
func TestEditorStatsPlot(t *testing.T) {
	defer core.TrackWindows()()
	b := core.NewBody()
	pl := NewEditor(b)
	pl.SetTable(newResultsTable())
	var name string
	var got *table.Table
	pl.PlotTableFunc = func(nm string, dt *table.Table) {
		name, got = nm, dt
	}
	core.RenderImage(b, core.Locked(b, func() {
		x, y, ys, split, cat := pl.statsColumns()
		assert.Equal(t, "Epoch", x)
		assert.Equal(t, "Err", y)
		assert.Equal(t, []string{"Err"}, ys)
		assert.Equal(t, "Cond", split)
		assert.Equal(t, "Cat", cat)

		me := &statsplot.MeanError{}
		pl.defaultStatsPlot(me)
		assert.Equal(t, statsplot.Column("Epoch"), me.X)
		assert.Equal(t, []statsplot.Column{"Err"}, me.Values)
		vl := &statsplot.Violin{}
		pl.defaultStatsPlot(vl)
		assert.Equal(t, statsplot.Column("Cat"), vl.Category)
		assert.Equal(t, statsplot.Column("Cond"), vl.Split)
		assert.Equal(t, 50, vl.Points)

		require.NoError(t, pl.StatsPlot("Violin", vl))
		assert.Equal(t, "Violin", name)
		require.NotNil(t, got)
		assert.Equal(t, "Err A Density", got.Columns.Keys[2])
		assert.NotNil(t, pl.StatsDir.Node("Violin"))
		assert.Error(t, pl.StatsPlot("Bad", &statsplot.Box{Value: "Nope"}))

		// the column chooser
		chooserColumns = pl.table.Columns.Keys
		col := statsplot.Column("Cat")
		cc, ok := core.NewValue(&col, "", b).(*ColumnChooser)
		require.True(t, ok)
		assert.Equal(t, 5, len(cc.Items)) // none + 4 columns
		cc.SetCurrentIndex(4)
		cc.SendChange()
		assert.Equal(t, statsplot.Column("Err"), col)

		// the Stats menu
		m := core.NewScene()
		pl.makeStatsMenu(m, image.Point{})
		assert.Equal(t, len(statsPlotTypes), len(m.Children))
		pl.statsPlotDialog("Box plot", &statsplot.Box{})
	}), func() {
		sc := windowScene(t, "Box plot")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			fm := findWidget[*core.Form](sc)
			require.NotNil(t, fm)
			bx := fm.Struct.(*statsplot.Box)
			assert.Equal(t, statsplot.Column("Err"), bx.Value)
			core.ButtonByText(sc, "Plot").Send(events.Click)
		})()
	}, core.Locked(b, func() {
		assert.Equal(t, "Err Box plot", name)
		assert.Equal(t, "Err A Low", got.Columns.Keys[0])
	}))
}

// TestEditorStatsTables tests that the tables of stats plots are plotted
// in the editor as they are by plot.NewTablePlot: the X columns of groups
// (violins, outliers) stay X, and boxes keep their fill and line colors.
func TestEditorStatsTables(t *testing.T) {
	dir, _ := tensorfs.NewDir("Stats")
	vl := &statsplot.Violin{Category: "Cat", Value: "Err", Split: "Cond"}
	vl.Defaults()
	vt, err := vl.Table(dir, newResultsTable())
	require.NoError(t, err)
	b := core.NewBody()
	pl := NewEditor(b)
	pl.SetTable(vt)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, plot.X, columnStyle(pl, "Err A Violin X").Role)
		assert.Equal(t, "Cat", pl.xAxisColumn())
		require.Equal(t, 4, len(pl.plot.Plotters))
		vb := pl.plot.Plotters[0].(*plots.Band)
		assert.NotNil(t, vb.Center)
		bst := plot.NewStyle() // the style of the box, as applied when drawn
		pl.plot.Plotters[1].(*plots.Box).Stylers().Run(bst)
		assert.Equal(t, colors.Scheme.Surface, bst.Line.Fill)
		assert.Equal(t, colors.Scheme.OnSurface, bst.Line.Color)
	}))

	bxp := &statsplot.Box{Category: "Cat", Value: "Err"}
	bxp.Defaults()
	bt, err := bxp.Table(dir, newResultsTable())
	require.NoError(t, err)
	b = core.NewBody()
	pl = NewEditor(b)
	pl.SetTable(bt)
	core.RenderImage(b, core.Locked(b, func() {
		bst := plot.NewStyle()
		pl.plot.Plotters[0].(*plots.Box).Stylers().Run(bst)
		assert.Equal(t, colors.Scheme.OnSurface, bst.Line.Color)
		assert.NotNil(t, bst.Line.Fill)
		assert.NotEqual(t, colors.Scheme.Surface, bst.Line.Fill)
		_, ok := pl.plot.Plotters[1].(*plots.XY) // the outliers
		assert.True(t, ok)
	}))
}
