// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
)

const (
	// ContourType is used for specifying the type name of a [Contour]
	// with contour lines.
	ContourType = "Contour"

	// FilledContourType is used for specifying the type name of a
	// [Contour] with filled regions between the levels.
	FilledContourType = "FilledContour"
)

func init() {
	plot.RegisterPlotter(ContourType, "draws the contour lines of the values of a 2D tensor in the Color data role (rows by columns, as for a Heatmap), at levels colored via the color map of the plot, with optional X and Y coordinates of the columns and rows.", []plot.Roles{plot.Color}, []plot.Roles{plot.X, plot.Y}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewContour(plt, data)
	})
	plot.RegisterPlotter(FilledContourType, "fills the regions between the contour levels of the values of a 2D tensor in the Color data role (rows by columns, as for a Heatmap), colored via the color map of the plot, with optional X and Y coordinates of the columns and rows.", []plot.Roles{plot.Color}, []plot.Roles{plot.X, plot.Y}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewFilledContour(plt, data)
	})
}

// Contour draws contour lines of the values of a 2D grid (rows by
// columns), the lines along which the values are equal to each of a
// number of levels, or fills the regions between the levels if Filled
// (as contour and contourf do in matplotlib). The levels are colored via
// the [plot.ColorAxis] of the plot, which is shown in a color bar.
// The values are interpolated linearly between the grid points, and NaN
// values leave holes.
//
// The data is a tensor in the Color role, as for a [Heatmap]. Column j is
// at X = j, and row i at Y = i, unless the X and Y data have the
// coordinates of the columns and rows (e.g., from [tensor.NewFloat64SpacedLinear]),
// with one value for each.
type Contour struct {
	// Values are the values of the grid, row by row.
	Values plot.Values

	// Rows, Columns are the number of rows and columns.
	Rows, Columns int

	// X, Y are the coordinates of the columns and rows.
	X, Y plot.Values

	// Levels is the number of levels, about which round values are
	// chosen across the range of the values, as for axis ticks,
	// if LevelValues is empty.
	Levels int

	// LevelValues are the values of the levels, if set.
	LevelValues []float64

	// Filled fills the regions between the levels, colored by the
	// middle value of each, instead of drawing contour lines.
	// The lines between the regions are also drawn if Style.Line.On is On,
	// with the Style.Line.Color.
	Filled bool

	// Row is the row of the outer dimension of a higher-dimensional
	// tensor that is drawn; -1 (the default) is the last row.
	// It is used by SetData.
	Row int

	// PX, PY are the pixel coordinates of the grid points,
	// in the order of the Values.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	data    plot.Data
	stylers plot.Stylers
}

// NewContour adds a new [Contour] plotter with contour lines to the given
// plot, for the given data: a [plot.Data] with the Color role and the
// optional X and Y coordinates, or a [tensor.Tensor] with the values
// (see [Contour]). Styler functions are obtained from the Color metadata
// if present.
func NewContour(plt *plot.Plot, data any) *Contour {
	ct := &Contour{Levels: 10, Row: -1}
	if err := ct.SetData(data); err != nil {
		return nil
	}
	ct.Defaults()
	plt.Add(ct)
	return ct
}

// NewFilledContour adds a new [Contour] plotter with filled regions between
// the levels to the given plot, for the given data (see [NewContour]).
func NewFilledContour(plt *plot.Plot, data any) *Contour {
	ct := NewContour(plt, data)
	if ct != nil {
		ct.Filled = true
	}
	return ct
}

func (ct *Contour) Defaults() {
	ct.Style.Defaults()
}

// SetData sets the plot data.
func (ct *Contour) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.Color)
	if err != nil {
		return err
	}
	ct.data = dt
	ct.Values, ct.Rows, ct.Columns, err = gridValues(ContourType, dt, ct.Row)
	if err != nil {
		return err
	}
	ct.X = ct.coords(plot.X, ct.Columns)
	ct.Y = ct.coords(plot.Y, ct.Rows)
	ct.stylers = plot.GetStylersFromData(dt, plot.Color)
	return nil
}

// coords returns the coordinates for the given role, if it has n numeric
// values, and otherwise the indexes 0..n-1.
func (ct *Contour) coords(role plot.Roles, n int) plot.Values {
	if v := plot.CopyRole(ct.data, role); len(v) == n {
		return v
	}
	v := make(plot.Values, n)
	for i := range n {
		v[i] = float64(i)
	}
	return v
}

// Styler adds a style function to set style parameters.
func (ct *Contour) Styler(f func(s *plot.Style)) *Contour {
	ct.stylers.Add(f)
	return ct
}

func (ct *Contour) Stylers() *plot.Stylers { return &ct.stylers }

func (ct *Contour) ApplyStyle(ps *plot.PlotStyle, idx int) {
	ps.SetElementStyle(&ct.Style)
	ct.stylers.Run(&ct.Style)
}

func (ct *Contour) Data() (data plot.Data, pixX, pixY []float32) {
	return plot.Data{plot.Color: ct.Values, plot.X: ct.X, plot.Y: ct.Y}, ct.PX, ct.PY
}

// valueRange returns the range of the values, ignoring NaNs.
func (ct *Contour) valueRange() minmax.F64 {
	var rng minmax.F64
	rng.SetInfinity()
	for _, v := range ct.Values {
		if !math.IsNaN(v) {
			rng.FitValInRange(v)
		}
	}
	return rng
}

// levels returns the levels of the contours (see [contourLevels]).
func (ct *Contour) levels() []float64 {
	rng := ct.valueRange()
	return contourLevels(rng.Min, rng.Max, ct.Levels, ct.LevelValues)
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (ct *Contour) Plot(plt *plot.Plot) {
	pc := plt.Painter
	py := plt.PY
	if ct.Style.RightY {
		py = plt.PYR
	}
	n := len(ct.Values)
	ct.PX = make([]float32, n)
	ct.PY = make([]float32, n)
	for i := range n {
		ct.PX[i], ct.PY[i] = plt.PX(ct.X[i%ct.Columns]), py(ct.Y[i/ct.Columns])
	}
	rng := ct.valueRange()
	if rng.Min > rng.Max {
		return
	}
	g := newContourGrid(ct.Values, ct.Rows, ct.Columns, ct.X, ct.Y)
	levels := ct.levels()
	// path adds the given points, in data coordinates, to the path
	path := func(pts []math32.Vector2) {
		for i, p := range pts {
			x, y := plt.PX(float64(p.X)), py(float64(p.Y))
			if i == 0 {
				pc.MoveTo(x, y)
			} else {
				pc.LineTo(x, y)
			}
		}
	}
	if ct.Filled {
		// each region at or above a level is filled over the one below
		bounds := append([]float64{rng.Min}, levels...)
		for i, lo := range bounds {
			hi := rng.Max
			if i+1 < len(bounds) {
				hi = bounds[i+1]
			}
			for _, poly := range g.region(lo) {
				path(poly)
				pc.Close()
			}
			pc.Fill.Color = plt.ColorAxis.Uniform(0.5 * (lo + hi))
			pc.Stroke.Color = nil
			pc.Draw()
		}
		pc.Fill.Color = nil
		if ct.Style.Line.On != plot.On || !ct.Style.Line.SetStroke(plt) {
			return
		}
		pc.Fill.Color = nil
		for _, lv := range levels {
			for _, ln := range g.lines(lv) {
				path(ln)
			}
		}
		pc.Draw()
		return
	}
	if ct.Style.Line.On == plot.Off || !ct.Style.Line.SetStroke(plt) {
		return
	}
	pc.Fill.Color = nil // the lines are closed for loops
	for _, lv := range levels {
		for _, ln := range g.lines(lv) {
			path(ln)
		}
		pc.Stroke.Color = plt.ColorAxis.Uniform(lv)
		pc.Draw()
	}
}

// UpdateRange updates the given ranges.
func (ct *Contour) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if ct.Style.RightY {
		yax = &plt.YR
	}
	plot.Range(ct.X, &plt.X.Range)
	plot.Range(ct.Y, &yax.Range)
	plt.ColorAxis.AddRange(ct.Values)
	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range
	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &ct.Style.Range)
	yax.DataRange = yax.Range
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (ct *Contour) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
	c := math32.FromPoint(ptb.Min).Add(math32.FromPoint(ptb.Size()).MulScalar(0.5))
	sz := math32.FromPoint(ptb.Size()).MulScalar(0.5)
	n := 3
	for i := range n {
		f := float32(n-i) / float32(n)
		clr := colors.Uniform(plt.ColorAxis.Map.Map((float32(i) + 0.5) / float32(n)))
		pc.Ellipse(c.X, c.Y, f*sz.X, f*sz.Y)
		if ct.Filled {
			pc.Fill.Color = clr
			pc.Stroke.Color = nil
		} else {
			pc.Fill.Color = nil
			pc.Stroke.Color = clr
		}
		pc.Draw()
	}
	pc.Fill.Color = nil
}
