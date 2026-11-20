// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// peaksGrid returns the values of two gaussian peaks of opposite signs
// (as in the matplotlib contour demo) on a grid of n by n points over
// [-3, 3], with the coordinates of the columns and rows.
func peaksGrid(n int) (z, x, y *tensor.Float64) {
	x = tensor.NewFloat64SpacedLinear(tensor.NewFloat64Scalar(-3), tensor.NewFloat64Scalar(3), n, tensor.IncludeEnd)
	y = tensor.NewFloat64SpacedLinear(tensor.NewFloat64Scalar(-2), tensor.NewFloat64Scalar(2), n, tensor.IncludeEnd)
	z = tensor.NewFloat64(n, n)
	for r := range n {
		for c := range n {
			xv, yv := x.Float1D(c), y.Float1D(r)
			z.Set(math.Exp(-xv*xv-yv*yv)-math.Exp(-(xv-1)*(xv-1)-(yv-1)*(yv-1)), r, c)
		}
	}
	return
}

// axisGrid returns the coordinates 0..n-1.
func axisGrid(n int) []float64 {
	v := make([]float64, n)
	for i := range n {
		v[i] = float64(i)
	}
	return v
}

// polyArea returns the signed area of the given polygon,
// positive for counter clockwise.
func polyArea(pts []math32.Vector2) float64 {
	a := 0.0
	for i, p := range pts {
		q := pts[(i+1)%len(pts)]
		a += float64(p.X*q.Y - q.X*p.Y)
	}
	return a / 2
}

func TestContourGridLines(t *testing.T) {
	// a cone, whose contours are circles around the center
	n := 21
	z := make([]float64, n*n)
	for r := range n {
		for c := range n {
			z[r*n+c] = math.Hypot(float64(c-10), float64(r-10))
		}
	}
	g := newContourGrid(z, n, n, axisGrid(n), axisGrid(n))
	assert.Equal(t, 4*20*20, len(g.tris))
	lines := g.lines(5)
	require.Equal(t, 1, len(lines))
	ln := lines[0]
	assert.Equal(t, ln[0], ln[len(ln)-1], "closed")
	for _, p := range ln {
		assert.InDelta(t, 5, math.Hypot(float64(p.X-10), float64(p.Y-10)), 0.1)
	}
	// the region within the circle is cut out of the one at or above 5:
	// the area of the region is the square minus the circle
	area := 0.0
	for _, p := range g.region(5) {
		a := polyArea(p)
		assert.GreaterOrEqual(t, a, 0.0, "counter clockwise") // 0 for a vertex at the level
		area += a
	}
	assert.InDelta(t, 400-math.Pi*25, area, 1)

	// a line that ends at the boundary
	lines = g.lines(12)
	assert.Equal(t, 4, len(lines)) // at the four corners
	for _, ln := range lines {
		assert.NotEqual(t, ln[0], ln[len(ln)-1], "open")
	}
	assert.Empty(t, g.lines(100))
}

func TestContourGridRegion(t *testing.T) {
	// a plane, z = x, whose region at or above a level has an exact area
	n := 5
	z := make([]float64, n*n)
	for i := range z {
		z[i] = float64(i % n)
	}
	g := newContourGrid(z, n, n, axisGrid(n), axisGrid(n))
	area := 0.0
	for _, p := range g.region(1.5) {
		area += polyArea(p)
	}
	assert.InDelta(t, 2.5*4, area, 1e-5)

	// NaN values leave out their cells
	z[2*n+2] = math.NaN()
	g = newContourGrid(z, n, n, axisGrid(n), axisGrid(n))
	assert.Equal(t, 4*(16-4), len(g.tris))
}

func TestContourLevels(t *testing.T) {
	assert.Equal(t, []float64{0.25, 0.5, 0.75, 1}, contourLevels(0, 1, 5, nil))
	assert.Equal(t, []float64{0.5, 0.7}, contourLevels(0, 1, 5, []float64{0.7, 2, 0.5, 0.5}))
}

func TestContour(t *testing.T) {
	z, x, y := peaksGrid(41)
	plt := plot.New()
	plt.Title.Text = "Contour"
	ct := NewContour(plt, plot.Data{plot.Color: z, plot.X: x, plot.Y: y})
	require.NotNil(t, ct)
	imagex.Assert(t, plt.RenderImage(), "contour")
	assert.Equal(t, -3.0, plt.X.Range.Min)
	assert.Equal(t, 2.0, plt.Y.Range.Max)
	assert.True(t, plt.ColorAxis.Used)
	assert.Greater(t, ct.PY[0], ct.PY[41], "row 0 at the bottom")
	pd, _, _ := ct.Data()
	assert.Equal(t, 41, pd[plot.X].Len())

	plt = plot.New()
	plt.Title.Text = "Filled contour"
	ct = NewFilledContour(plt, plot.Data{plot.Color: z, plot.X: x, plot.Y: y})
	ct.Styler(func(s *plot.Style) {
		s.Line.On = plot.On
		s.Line.Width.Pt(0.5)
		s.Plot.ColorAxis.Label = "Z"
	})
	ct.Levels = 8
	imagex.Assert(t, plt.RenderImage(), "contour-filled")

	// explicit levels, the grid indexes, and NaN holes
	z.Set(math.NaN(), 20, 20)
	plt = plot.New()
	ct = NewContour(plt, z)
	ct.LevelValues = []float64{-0.5, 0, 0.5}
	ct.Styler(func(s *plot.Style) { s.Line.Width.Pt(2) })
	imagex.Assert(t, plt.RenderImage(), "contour-levels")
	assert.Equal(t, 40.0, plt.X.Range.Max)
	assert.Nil(t, NewContour(plot.New(), plot.Data{plot.Color: plot.Values{1, 2}}))
}

func TestContourTable(t *testing.T) {
	// a table column of grids: the last row is drawn by default, and the
	// X column of the table is not used as the coordinates of the columns
	nr := 3
	grids := tensor.NewFloat64(nr, 21, 21)
	for r := range nr {
		z, _, _ := peaksGrid(21)
		for i := range z.Len() {
			grids.SetFloat1D(float64(r+1)*z.Float1D(i), r*21*21+i)
		}
	}
	plot.SetStyler(grids, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Color
		s.Plotter = FilledContourType
		s.Plot.Title = "Filled contour of the last grid in a table column"
	})
	dt := table.New("Grids")
	dt.AddColumn("X", tensor.NewIntRange(nr))
	dt.AddColumn("Grids", grids)
	plt, err := plot.NewTablePlot(dt)
	require.NoError(t, err)
	imagex.Assert(t, plt.RenderImage(), "contour-table")
	ct := plt.Plotters[0].(*Contour)
	assert.True(t, ct.Filled)
	assert.Equal(t, 21, ct.Columns)
	assert.Equal(t, 20.0, ct.X[20])
	assert.Equal(t, grids.Float1D(2*21*21+1), ct.Values[1])
}
