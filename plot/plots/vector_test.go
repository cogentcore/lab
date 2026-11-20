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
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// swirlGrid returns a flow field rotating around the origin, on a grid
// of n by n points over [-2, 2], with the coordinates of the columns and
// rows, and the magnitude of the vectors.
func swirlGrid(n int) (x, y, u, v, mag *tensor.Float64) {
	x = tensor.NewFloat64SpacedLinear(tensor.NewFloat64Scalar(-2), tensor.NewFloat64Scalar(2), n, tensor.IncludeEnd)
	y = x.Clone().(*tensor.Float64)
	u, v, mag = tensor.NewFloat64(n, n), tensor.NewFloat64(n, n), tensor.NewFloat64(n, n)
	for r := range n {
		for c := range n {
			xv, yv := x.Float1D(c), y.Float1D(r)
			d := math.Exp(-0.3 * (xv*xv + yv*yv))
			u.Set(-yv*d, r, c)
			v.Set(xv*d, r, c)
			mag.Set(math.Hypot(xv, yv)*d, r, c)
		}
	}
	return
}

func TestVector(t *testing.T) {
	x, y, u, v, mag := swirlGrid(13)
	plt := plot.New()
	plt.Title.Text = "Vector field"
	vc := NewVector(plt, plot.Data{plot.X: x, plot.Y: y, plot.U: u, plot.V: v, plot.Color: mag})
	require.NotNil(t, vc)
	vc.Styler(func(s *plot.Style) { s.Plot.ColorAxis.Label = "Speed" })
	vc.Key = true
	assert.Equal(t, 13*13, len(vc.X), "the grid is expanded")
	assert.Equal(t, x.Float1D(3), vc.X[13+3])
	assert.Equal(t, y.Float1D(1), vc.Y[13+3])
	imagex.Assert(t, plt.RenderImage(), "vector")
	assert.True(t, plt.ColorAxis.Used)
	assert.InDelta(t, -2-0.9*4.0/12, plt.X.Range.Min, 1e-9, "a margin of the spacing of the grid")
	pd, _, _ := vc.Data()
	assert.Equal(t, 13*13, pd[plot.Color].Len())

	// the middle of the arrows at the positions, in the line color, scaled
	u.Set(math.NaN(), 6, 6)
	plt = plot.New()
	vc = NewVector(plt, plot.Data{plot.X: x, plot.Y: y, plot.U: u, plot.V: v})
	vc.Pivot = PivotMiddle
	vc.Scale = 0.7
	vc.Key = true
	assert.True(t, vc.KeyPosition.Top, "top right by default")
	assert.False(t, vc.KeyPosition.Left)
	vc.KeyPosition.Top, vc.KeyPosition.Left = false, true
	vc.KeyPosition.XOffs.Dp(20)
	vc.KeyValue = 0.5
	vc.KeyLabel = "0.5 m/s"
	imagex.Assert(t, plt.RenderImage(), "vector-middle")
	assert.False(t, plt.ColorAxis.Used)
	assert.False(t, vc.valid(6*13+6))

	// errors
	assert.Nil(t, NewVector(plot.New(), plot.Data{plot.X: x, plot.Y: y, plot.U: u}))
	assert.Nil(t, NewVector(plot.New(), plot.Data{plot.X: plot.Values{1, 2}, plot.Y: plot.Values{1, 2, 3}, plot.U: u, plot.V: v}))
	assert.Nil(t, NewVector(plot.New(), plot.Data{plot.X: x, plot.Y: y, plot.U: u, plot.V: plot.Values{1}}))
	assert.Nil(t, NewVector(plot.New(), plot.Data{plot.X: x, plot.Y: y, plot.U: u, plot.V: v, plot.Color: plot.Values{1}}))
}

func TestVectorTable(t *testing.T) {
	// arrows at scattered positions, from the columns of a table
	rng := rand.New(rand.NewSource(1))
	n := 40
	dt := table.New("Arrows")
	xc, yc := dt.AddFloat64Column("X"), dt.AddFloat64Column("Y")
	uc, vc := dt.AddFloat64Column("U"), dt.AddFloat64Column("V")
	dt.SetNumRows(n)
	for i := range n {
		x, y := 10*rng.Float64(), 5*rng.Float64()
		xc.SetFloat1D(x, i)
		yc.SetFloat1D(y, i)
		uc.SetFloat1D(1+0.2*y, i) // a shear flow
		vc.SetFloat1D(0.2*math.Sin(x), i)
	}
	plot.SetStyler(xc, func(s *plot.Style) { s.Role = plot.X })
	// the Y column is On, with the U and V columns in its group
	plot.SetStyler(uc, func(s *plot.Style) { s.Role = plot.U; s.Group = "A" })
	plot.SetStyler(vc, func(s *plot.Style) { s.Role = plot.V; s.Group = "A" })
	plot.SetStyler(yc, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
		s.Group = "A"
		s.Plotter = VectorType
		s.Plot.Title = "Vectors from a table"
	})
	plt, err := plot.NewTablePlot(dt)
	require.NoError(t, err)
	imagex.Assert(t, plt.RenderImage(), "vector-table")
	vp := plt.Plotters[0].(*Vector)
	assert.Equal(t, n, len(vp.X))
	assert.Equal(t, yc.Float1D(5), vp.Y[5])
}

func TestVectorKeyValue(t *testing.T) {
	vc := &Vector{}
	assert.Equal(t, 0.5, vc.keyValue(0.78))
	assert.Equal(t, 2.0, vc.keyValue(4.9))
	assert.Equal(t, 10.0, vc.keyValue(19))
	assert.Equal(t, 5.0, vc.keyValue(5))
	vc.KeyValue = 3
	assert.Equal(t, 3.0, vc.keyValue(19))
}
