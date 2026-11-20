// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/colors"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPie(t *testing.T) {
	names := tensor.NewStringFromValues("Apples", "Bananas", "Cherries", "Dates", "Elderberries")
	vals := plot.Values{35, 25, 20, 12, 8}
	plt := plot.New()
	plt.Title.Text = "Pie"
	pe := NewPie(plt, plot.Data{plot.Y: vals, plot.Label: names})
	require.NotNil(t, pe)
	pe.Percent = true
	imagex.Assert(t, plt.RenderImage(), "pie")
	assert.InDelta(t, 0.35, pe.Fractions()[0], 1e-9)
	assert.False(t, plt.X.Style.On, "no axes")
	assert.Greater(t, pe.Radius, float32(50))

	// a donut, counter clockwise, with edges
	plt = plot.New()
	plt.Title.Text = "Donut"
	pe = NewPie(plt, plot.Data{plot.Y: vals, plot.Label: names})
	pe.Hole = 0.5
	pe.CounterClockwise = true
	pe.Percent = true
	pe.Styler(func(s *plot.Style) {
		s.Line.On = plot.On
		s.Line.Width.Pt(2)
		s.Line.Color = colors.Uniform(colors.White)
	})
	imagex.Assert(t, plt.RenderImage(), "pie-donut")

	// one value is the whole circle; NaN and negative values have no wedge;
	// colors from Color data
	plt = plot.New()
	pe = NewPie(plt, plot.Values{3, math.NaN(), -1})
	imagex.Assert(t, plt.RenderImage(), "pie-one")
	assert.Equal(t, []float64{1, 0, 0}, pe.Fractions())
	plt = plot.New()
	NewPie(plt, plot.Data{plot.Y: vals, plot.Color: vals})
	imagex.Assert(t, plt.RenderImage(), "pie-color")
	assert.True(t, plt.ColorAxis.Used)

	// errors
	assert.Nil(t, NewPie(plot.New(), plot.Data{plot.Y: vals, plot.Label: tensor.NewStringFromValues("A")}))
	assert.Nil(t, NewPie(plot.New(), plot.Data{plot.Y: vals, plot.Color: plot.Values{1}}))
	assert.Nil(t, NewPie(plot.New(), plot.Values{}))
}

func TestPieTable(t *testing.T) {
	dt := table.New("Fruit")
	names := dt.AddStringColumn("Fruit")
	counts := dt.AddFloat64Column("Count")
	for i, nm := range []string{"Apples", "Bananas", "Cherries"} {
		dt.AddRows(1)
		names.SetString1D(nm, i)
		counts.SetFloat1D(float64(3-i), i)
	}
	plot.SetStyler(names, func(s *plot.Style) { s.Role = plot.Label })
	plot.SetStyler(counts, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
		s.Plotter = PieType
		s.Plot.Title = "Pie from a table"
	})
	plt, err := plot.NewTablePlot(dt)
	require.NoError(t, err)
	imagex.Assert(t, plt.RenderImage(), "pie-table")
	pe := plt.Plotters[0].(*Pie)
	assert.Equal(t, "Bananas", pe.Labels[1])
}
