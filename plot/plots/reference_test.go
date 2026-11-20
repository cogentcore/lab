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
)

func TestReference(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Reference lines and spans"
	vs := NewVSpan(plt, 40, 60)
	assert.NotNil(t, vs)
	hs := NewHSpan(plt, 20, 30)
	assert.NotNil(t, hs)
	hs.Styler(func(s *plot.Style) {
		s.Line.Color = colors.Uniform(colors.Green)
		s.Line.On = plot.On
	})
	ln := NewLine(plt, sinDataXY())
	assert.NotNil(t, ln)
	hl := NewHLine(plt, 50)
	assert.NotNil(t, hl)
	hl.Styler(func(s *plot.Style) {
		s.Line.Dashes = []float32{4, 2}
	})
	vl := NewVLine(plt, 25, 75)
	assert.NotNil(t, vl)
	vl.Styler(func(s *plot.Style) {
		s.Line.Color = colors.Uniform(colors.Red)
	})
	plt.Legend.Add("Sine", ln)
	plt.Legend.Add("Chance", hl)
	plt.Legend.Add("Epoch", vs)

	assert.True(t, vs.IsSpan())
	assert.False(t, hl.IsSpan())
	assert.True(t, vl.Vertical)
	pd, _, _ := vl.Data()
	assert.Equal(t, plot.Values{25, 75}, pd[plot.X])
	pd, _, _ = hs.Data()
	assert.Equal(t, plot.Values{30}, pd[plot.High])

	imagex.Assert(t, plt.RenderImage(), "reference")
	_, px, _ := vl.Data()
	assert.Less(t, px[0], px[1])
}

func TestReferenceRange(t *testing.T) {
	// reference values are included in the range of their axis only
	plt := plot.New()
	NewLine(plt, sinDataXY())
	NewHLine(plt, 150)
	NewVSpan(plt, -50, -10)
	plt.RenderImage()
	assert.Equal(t, 150.0, plt.Y.Range.Max)
	assert.Equal(t, -50.0, plt.X.Range.Min)
	assert.InDelta(t, 10, plt.Y.Range.Min, 0.1)

	assert.Nil(t, newReference(plot.New(), plot.Data{plot.X: plot.Values{1}, plot.Low: plot.Values{1}}, false, false))
	// the one role in the data, as for tables
	assert.NotNil(t, newReference(plot.New(), plot.Data{plot.X: plot.Values{1}}, false, false))
	assert.Nil(t, newReference(plot.New(), plot.Data{plot.Low: plot.Values{1}}, false, true))
}

func TestGrid(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Y grid"
	// plot styles can be set by any plotter
	NewBar(plt, plot.Data{plot.Y: plot.Values{3, 5, 2, 4}}).Styler(func(s *plot.Style) {
		s.Plot.Grid.Y = true
	})
	imagex.Assert(t, plt.RenderImage(), "grid-y")

	plt = plot.New()
	plt.Title.Text = "X and Y grid"
	NewLine(plt, sinDataXY()).Styler(func(s *plot.Style) {
		s.Plot.Grid.X = true
		s.Plot.Grid.Y = true
		s.Plot.Grid.Line.Dashes = []float32{2, 2}
	})
	imagex.Assert(t, plt.RenderImage(), "grid-xy")
}

func TestTableReference(t *testing.T) {
	n := 20
	tx, ty := tensor.NewFloat64(n), tensor.NewFloat64(n)
	for i := range n {
		tx.SetFloat1D(float64(i), i)
		ty.SetFloat1D(math.Sin(float64(i)/3), i)
	}
	// event times, with NaN for no event, in their own column
	ev := tensor.NewFloat64(n)
	for i := range n {
		ev.SetFloat1D(math.NaN(), i)
	}
	ev.SetFloat1D(5, 5)
	ev.SetFloat1D(12, 12)
	plot.SetStyler(tx, func(s *plot.Style) {
		s.Role = plot.X
		s.Plot.Title = "Table VLine events, Y grid"
		s.Plot.Grid.Y = true
	})
	plot.SetStyler(ty, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
	})
	plot.SetStyler(ev, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y // any role but X, which is the X axis
		s.Group = "Events"
		s.Plotter = VLineType
		s.Line.Dashes = []float32{4, 2}
	})
	dt := table.New("Reference")
	dt.AddColumn("X", tx)
	dt.AddColumn("Y", ty)
	dt.AddColumn("Events", ev)
	plt, err := plot.NewTablePlot(dt)
	assert.NoError(t, err)
	imagex.Assert(t, plt.RenderImage(), "table-reference")
}
