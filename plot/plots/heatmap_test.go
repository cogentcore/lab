// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"math/rand"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

// waveGrid returns a rows x cols grid of a 2D wave, offset by off.
func waveGrid(rows, cols int, off float64) *tensor.Float64 {
	g := tensor.NewFloat64(rows, cols)
	for r := range rows {
		for c := range cols {
			g.Set(math.Sin(float64(c)/2+off)*math.Cos(float64(r)/3), r, c)
		}
	}
	return g
}

func TestHeatmap(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Heatmap"
	g := waveGrid(12, 16, 0)
	g.Set(math.NaN(), 0, 0)
	hm := NewHeatmap(plt, g)
	assert.NotNil(t, hm)
	assert.Equal(t, 12, hm.Rows)
	assert.Equal(t, 16, hm.Columns)
	imagex.Assert(t, plt.RenderImage(), "heatmap")
	assert.Equal(t, -0.5, plt.X.Range.Min)
	assert.Equal(t, 15.5, plt.X.Range.Max)
	assert.True(t, plt.ColorAxis.Used)
	// row 0 is at the top
	assert.Less(t, hm.PY[0], hm.PY[16])
	pd, _, _ := hm.Data()
	assert.Equal(t, 12*16, pd[plot.Color].Len())

	// row 0 at the bottom, 1D and errors
	plt = plot.New()
	hm = NewHeatmap(plt, g)
	hm.TopZero = false
	plt.RenderImage()
	assert.Greater(t, hm.PY[0], hm.PY[16])
	hm = NewHeatmap(plot.New(), tensor.NewFloat64FromValues(1, 2, 3))
	assert.Equal(t, 1, hm.Rows)
	assert.Nil(t, NewHeatmap(plot.New(), plot.Data{plot.Color: plot.Values{1, 2}}))
}

func TestHeatmapCorrelation(t *testing.T) {
	// correlation matrix of 5 variables, with labels
	rng := rand.New(rand.NewSource(1))
	n := 100
	vars := tensor.NewFloat64(n, 5)
	for i := range n {
		a, b := rng.NormFloat64(), rng.NormFloat64()
		vars.Set(a, i, 0)
		vars.Set(a+0.5*rng.NormFloat64(), i, 1)
		vars.Set(-a+rng.NormFloat64(), i, 2)
		vars.Set(b, i, 3)
		vars.Set(b+a+rng.NormFloat64(), i, 4)
	}
	corr := metric.CovarianceMatrix(metric.Correlation, vars)
	names := tensor.NewStringFromValues("A", "B", "C", "D", "E")
	plt := plot.New()
	plt.Title.Text = "Correlation matrix"
	hm := NewHeatmap(plt, plot.Data{plot.Color: corr, plot.X: names, plot.Y: names})
	assert.NotNil(t, hm)
	hm.Styler(func(s *plot.Style) {
		s.Line.On = plot.On
		s.Plot.ColorAxis.Range.SetMin(-1).SetMax(1)
		s.Plot.ColorAxis.Label = "Correlation"
	})
	imagex.Assert(t, plt.RenderImage(), "heatmap-corr")
	ticks := plt.Y.Ticker.(plot.ConstantTicks)
	assert.Equal(t, "A", ticks[0].Label)
	assert.Equal(t, 4.0, ticks[0].Value) // at the top
	assert.Equal(t, -1.0, plt.ColorAxis.Range.Min)

	// labels of the wrong length are not used
	hm = NewHeatmap(plot.New(), plot.Data{plot.Color: corr, plot.X: plot.Values{1, 2}})
	assert.Nil(t, hm.XLabels)
}

func TestHeatmapRows(t *testing.T) {
	// a table column of matrices: the last row is drawn by default
	nr := 3
	mats := tensor.NewFloat64(nr, 6, 8)
	for r := range nr {
		g := waveGrid(6, 8, float64(r))
		for i := range g.Len() {
			mats.SetFloat1D(g.Float1D(i), r*48+i)
		}
	}
	tx := tensor.NewIntRange(nr)
	plot.SetStyler(mats, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Color
		s.Plotter = HeatmapType
		s.Plot.Title = "Last matrix in a table column"
	})
	dt := table.New("Mats")
	dt.AddColumn("X", tx)
	dt.AddColumn("Mats", mats)
	plt, err := plot.NewTablePlot(dt)
	assert.NoError(t, err)
	imagex.Assert(t, plt.RenderImage(), "heatmap-table")
	hm := plt.Plotters[0].(*Heatmap)
	assert.Equal(t, 6, hm.Rows)
	assert.Equal(t, 8, hm.Columns)
	assert.Equal(t, mats.Float1D(2*48+1), hm.Values[1])

	// another row
	hm = &Heatmap{Row: 0}
	assert.NoError(t, hm.SetData(plot.Data{plot.Color: mats}))
	assert.Equal(t, mats.Float1D(1), hm.Values[1])
}
