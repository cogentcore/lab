// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statsplot

import (
	"math"
	"math/rand"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resultsTable returns simulated results: 8 runs of 20 epochs in two
// conditions, with an error that decreases faster in condition B, and a
// skewed reaction time (RT) in three categories (Cat).
func resultsTable() *table.Table {
	rng := rand.New(rand.NewSource(1))
	dt := table.New("Results")
	run := dt.AddIntColumn("Run")
	epoch := dt.AddIntColumn("Epoch")
	cond := dt.AddStringColumn("Cond")
	cat := dt.AddStringColumn("Cat")
	errc := dt.AddFloat64Column("Err")
	rt := dt.AddFloat64Column("RT")
	cats := []string{"Easy", "Medium", "Hard"}
	for ci, cnd := range []string{"A", "B"} {
		for r := range 8 {
			for e := range 20 {
				dt.AddRows(1)
				i := dt.NumRows() - 1
				run.SetInt1D(r, i)
				epoch.SetInt1D(e, i)
				cond.SetString1D(cnd, i)
				cg := (r + e) % 3
				cat.SetString1D(cats[cg], i)
				rate := 0.1 + 0.1*float64(ci)
				errc.SetFloat1D(math.Exp(-rate*float64(e))+0.05*rng.NormFloat64(), i)
				rt.SetFloat1D(300+100*float64(cg)+50*float64(ci)+math.Exp(4+0.5*rng.NormFloat64()), i)
			}
		}
	}
	return dt
}

// plotTable plots the given table and asserts its image.
func plotTable(t *testing.T, dt *table.Table, name string) *plot.Plot {
	plt, err := plot.NewTablePlot(dt)
	require.NoError(t, err)
	imagex.Assert(t, plt.RenderImage(), name)
	return plt
}

func TestMeanError(t *testing.T) {
	dt := resultsTable()
	dir, _ := tensorfs.NewDir("Root")
	me := &MeanError{X: "Epoch", Values: []string{"Err"}, Split: "Cond"}
	me.Defaults()
	rt, err := me.Table(dir, dt)
	require.NoError(t, err)
	assert.Equal(t, []string{"Epoch", "Err A", "Err A Sem", "Err B", "Err B Sem"}, rt.Columns.Keys)
	assert.Equal(t, 20, rt.NumRows())
	// the mean at epoch 0 is about 1, and the data is in the tensorfs
	assert.InDelta(t, 1, rt.Column("Err A").Float1D(0), 0.05)
	assert.NotNil(t, dir.Node("MeanError").Node("Err B Sem"))
	plt := plotTable(t, rt, "mean-error-band")
	assert.Equal(t, 4, len(plt.Plotters))  // 2 lines, 2 bands
	_, ok := plt.Plotters[1].(*plots.Band) // the error of the first line
	assert.True(t, ok)

	// error bars, and the default value columns
	me = &MeanError{X: "Epoch", Split: "Cond", Error: CI95, ErrorPlot: Bars}
	rt, err = me.Table(dir, dt)
	require.NoError(t, err)
	assert.Equal(t, []string{"Epoch", "Err A", "Err A CI95", "Err B", "Err B CI95", "RT A", "RT A CI95", "RT B", "RT B CI95"}, rt.Columns.Keys)
	plot.Styler(rt.Column("RT A").Tensor, func(s *plot.Style) { s.On = false })
	plot.Styler(rt.Column("RT B").Tensor, func(s *plot.Style) { s.On = false })
	plot.Styler(rt.Column("RT A CI95").Tensor, func(s *plot.Style) { s.On = false })
	plot.Styler(rt.Column("RT B CI95").Tensor, func(s *plot.Style) { s.On = false })
	plotTable(t, rt, "mean-error-bars")

	// categories: bars
	me = &MeanError{X: "Cat", Values: []string{"RT"}, Split: "Cond"}
	rt, err = me.Table(dir, dt)
	require.NoError(t, err)
	assert.Equal(t, "Easy", rt.Column("Cat").String1D(0))
	plt = plotTable(t, rt, "mean-error-cat")
	_, ok = plt.Plotters[0].(*plots.Bar)
	assert.True(t, ok)

	_, err = (&MeanError{X: "Nope"}).Table(dir, dt)
	assert.Error(t, err)
}

func TestHistogram(t *testing.T) {
	dt := resultsTable()
	dir, _ := tensorfs.NewDir("Root")
	hs := &Histogram{Value: "RT", Split: "Cond"}
	hs.Defaults()
	rt, err := hs.Table(dir, dt)
	require.NoError(t, err)
	assert.Equal(t, []string{"RT", "A Count", "B Count"}, rt.Columns.Keys)
	assert.Equal(t, 20, rt.NumRows())
	total := 0.0
	for i := range 20 {
		total += rt.Column("A Count").Float1D(i)
	}
	assert.Equal(t, 160.0, total)
	plotTable(t, rt, "histogram")

	hs = &Histogram{Value: "RT", Bins: 10, Density: true}
	rt, err = hs.Table(dir, dt)
	require.NoError(t, err)
	area := 0.0
	w := rt.Column("RT").Float1D(1) - rt.Column("RT").Float1D(0)
	for i := range 10 {
		area += rt.Column("Density").Float1D(i) * w
	}
	assert.InDelta(t, 1, area, 1e-9)
}

func TestBox(t *testing.T) {
	dt := resultsTable()
	dir, _ := tensorfs.NewDir("Root")
	bx := &Box{Category: "Cat", Value: "RT", Split: "Cond"}
	bx.Defaults()
	rt, err := bx.Table(dir, dt)
	require.NoError(t, err)
	assert.Equal(t, "RT A Median", rt.Columns.Keys[2])
	assert.Equal(t, "Cat", rt.Columns.Keys[rt.NumColumns()-1])
	assert.Greater(t, rt.NumRows(), 3) // outliers
	plt := plotTable(t, rt, "box")
	ticks := plt.X.Ticker.(plot.ConstantTicks)
	assert.Equal(t, 3, len(ticks))
	assert.Equal(t, "Medium", ticks[1].Label)

	bx = &Box{Value: "RT"}
	rt, err = bx.Table(dir, dt)
	require.NoError(t, err)
	assert.Equal(t, 1, rt.NumRows()) // no outliers, min-max whiskers
}

func TestViolin(t *testing.T) {
	dt := resultsTable()
	dir, _ := tensorfs.NewDir("Root")
	vl := &Violin{Category: "Cat", Value: "RT", Split: "Cond"}
	vl.Defaults()
	rt, err := vl.Table(dir, dt)
	require.NoError(t, err)
	assert.Equal(t, 3*51, rt.NumRows())
	plt := plotTable(t, rt, "violin")
	assert.Equal(t, 4, len(plt.Plotters)) // 2 violins, 2 boxes
	vb := plt.Plotters[0].(*plots.Band)
	bx := plt.Plotters[1].(*plots.Box)
	// the violins are at the positions of the boxes
	assert.InDelta(t, bx.Style.Width.Offset, vb.Center[0], 1e-9)

	vl = &Violin{Category: "Cat", Value: "RT", Cut: 2}
	rt, err = vl.Table(dir, dt)
	require.NoError(t, err)
	plotTable(t, rt, "violin-nobox")
}
