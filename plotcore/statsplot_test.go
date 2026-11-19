// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotcore

import (
	"image"
	"math/rand"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/statsplot"
	"cogentcore.org/lab/table"
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
