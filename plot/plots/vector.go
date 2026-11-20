// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/paint"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
)

// VectorType is used for specifying the type name of a [Vector].
const VectorType = "Vector"

func init() {
	plot.RegisterPlotter(VectorType, "draws arrows at X, Y positions with the vector components U, V (a quiver plot, e.g., of a flow field or gradient), with lengths scaled so that the longest arrow is about the spacing of the arrows. Optional Color data colors the arrows via the color map of the plot.", []plot.Roles{plot.X, plot.Y, plot.U, plot.V}, []plot.Roles{plot.Color}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewVector(plt, data)
	})
}

// Pivots are the points of an arrow of a [Vector] plot that are at the
// positions of the data.
type Pivots int32 //enums:enum

const (
	// PivotTail puts the tail of the arrow at the position.
	PivotTail Pivots = iota

	// PivotMiddle puts the middle of the arrow at the position.
	PivotMiddle

	// PivotTip puts the tip of the arrow at the position.
	PivotTip
)

// Vector draws arrows at X, Y positions with the vector components U, V
// (as quiver does in matplotlib), e.g., for a flow field or a gradient.
// The arrows point in the direction of U, V on the screen (U to the right
// and V up), independent of the scales of the axes, and their lengths are
// proportional to the magnitudes of the vectors, scaled so that the
// longest arrow is about the spacing of the arrows, times Scale.
// Optional Color data colors each arrow via the [plot.ColorAxis] of the
// plot; otherwise the arrows are in the Line.Color.
//
// The data are X, Y, U, V (and Color) values of the same length, e.g.,
// columns of a table, or a grid: X and Y with the coordinates of the
// columns and rows, and U, V (and Color) with the values of the grid,
// row by row (e.g., 2D tensors of rows by columns, as for a [Contour]).
// Arrows with NaN values are not drawn.
type Vector struct {
	// X, Y are the positions of the arrows.
	X, Y plot.Values

	// U, V are the components of the vectors.
	U, V plot.Values

	// Color are the optional values for the colors of the arrows.
	Color plot.Values

	// Scale multiplies the automatic lengths of the arrows.
	Scale float32

	// Pivot is the point of each arrow at its position.
	Pivot Pivots

	// Key draws a key arrow at the top right of the plot, for the scale
	// of the arrows (as quiverkey does in matplotlib): a horizontal arrow
	// with the length of a vector of magnitude KeyValue, with its KeyLabel.
	Key bool

	// KeyValue is the magnitude of the vector of the key arrow: a round
	// value about the largest magnitude of the vectors if 0.
	KeyValue float64

	// KeyLabel is the label of the key arrow: the KeyValue if empty.
	KeyLabel string

	// PX, PY are the pixel coordinates of the positions of the arrows.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	stylers plot.Stylers
}

// NewVector adds a new [Vector] plotter to the given plot, for the given
// data: a [plot.Data] with the X, Y, U, V roles and the optional Color
// role (see [Vector]). Styler functions are obtained from the Y metadata
// if present (the column that is On in a table).
func NewVector(plt *plot.Plot, data plot.Data) *Vector {
	vc := &Vector{Scale: 1}
	if err := vc.SetData(data); err != nil {
		return nil
	}
	vc.Defaults()
	plt.Add(vc)
	return vc
}

func (vc *Vector) Defaults() {
	vc.Style.Defaults()
}

// SetData sets the plot data, a [plot.Data] as for [NewVector].
func (vc *Vector) SetData(d any) error {
	data, err := plot.DataOrValuer(d, plot.U)
	if err != nil {
		return err
	}
	for _, rl := range []plot.Roles{plot.X, plot.Y, plot.U, plot.V} {
		if data[rl] == nil || data[rl].Len() == 0 {
			return fmt.Errorf("Vector: %s data is missing", rl)
		}
	}
	u, v := plot.CopyRole(data, plot.U), plot.CopyRole(data, plot.V)
	x, y := plot.CopyRole(data, plot.X), plot.CopyRole(data, plot.Y)
	n := len(u)
	if len(v) != n {
		return fmt.Errorf("Vector: U and V must have the same length, not %d and %d", n, len(v))
	}
	var clr plot.Values
	if data[plot.Color] != nil {
		if clr = plot.CopyRole(data, plot.Color); len(clr) != n {
			return fmt.Errorf("Vector: Color must have the length of U and V, %d, not %d", n, len(clr))
		}
	}
	switch {
	case len(x) == n && len(y) == n:
	case len(x)*len(y) == n: // a grid
		gx, gy := make(plot.Values, n), make(plot.Values, n)
		for i := range n {
			gx[i], gy[i] = x[i%len(x)], y[i/len(x)]
		}
		x, y = gx, gy
	default:
		return fmt.Errorf("Vector: X and Y must have the length of U and V, %d, or be the coordinates of the columns and rows of a grid of them, not %d and %d", n, len(x), len(y))
	}
	vc.X, vc.Y, vc.U, vc.V, vc.Color = x, y, u, v, clr
	vc.stylers = plot.GetStylersFromData(data, plot.Y)
	return nil
}

// Styler adds a style function to set style parameters.
func (vc *Vector) Styler(f func(s *plot.Style)) *Vector {
	vc.stylers.Add(f)
	return vc
}

func (vc *Vector) Stylers() *plot.Stylers { return &vc.stylers }

func (vc *Vector) ApplyStyle(ps *plot.PlotStyle, idx int) {
	ps.SetElementStyle(&vc.Style)
	vc.stylers.Run(&vc.Style)
}

func (vc *Vector) Data() (data plot.Data, pixX, pixY []float32) {
	data = plot.Data{plot.X: vc.X, plot.Y: vc.Y, plot.U: vc.U, plot.V: vc.V}
	if vc.Color != nil {
		data[plot.Color] = vc.Color
	}
	return data, vc.PX, vc.PY
}

// valid returns whether the arrow at the given index has values.
func (vc *Vector) valid(i int) bool {
	return !math.IsNaN(vc.X[i]) && !math.IsNaN(vc.Y[i]) && !math.IsNaN(vc.U[i]) && !math.IsNaN(vc.V[i])
}

// spacing returns the typical distance between the positions of the
// arrows, in pixels: the area of the plot per arrow, or the distance
// between the columns and rows of a grid, the smaller of the two.
func (vc *Vector) spacing(plt *plot.Plot) float32 {
	nx, ny := distinct(vc.PX), distinct(vc.PY)
	sz := plt.PlotBox.Size()
	sp := math32.Sqrt(sz.X * sz.Y / float32(max(len(vc.PX), 1)))
	if nx > 1 && ny > 1 && nx*ny == len(vc.PX) { // a grid
		sp = min(sz.X/float32(nx), sz.Y/float32(ny))
	}
	return sp
}

// dataSpacing returns the typical distance between the positions of the
// arrows in data units along X and Y, for the given ranges of the
// positions: the distance between the columns and rows of a grid, or
// else the square root of the area per arrow.
func (vc *Vector) dataSpacing(xr, yr minmax.F64) (sx, sy float64) {
	n := len(vc.X)
	nx, ny := distinct(vc.X), distinct(vc.Y)
	if nx > 1 && ny > 1 && nx*ny == n {
		return xr.Range() / float64(nx-1), yr.Range() / float64(ny-1)
	}
	w, h := xr.Range(), yr.Range()
	switch {
	case w > 0 && h > 0:
		s := math.Sqrt(w * h / float64(n))
		return s, s
	case w > 0:
		return w / float64(n), w / float64(n)
	default:
		return h / float64(max(n, 1)), h / float64(max(n, 1))
	}
}

// distinct returns the number of distinct values.
func distinct[T comparable](vals []T) int {
	m := map[T]bool{}
	for _, v := range vals {
		m[v] = true
	}
	return len(m)
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (vc *Vector) Plot(plt *plot.Plot) {
	pc := plt.Painter
	py := plt.PY
	if vc.Style.RightY {
		py = plt.PYR
	}
	n := len(vc.U)
	vc.PX = make([]float32, n)
	vc.PY = make([]float32, n)
	maxMag := 0.0
	for i := range n {
		vc.PX[i], vc.PY[i] = plt.PX(vc.X[i]), py(vc.Y[i])
		if vc.valid(i) {
			maxMag = max(maxMag, math.Hypot(vc.U[i], vc.V[i]))
		}
	}
	if maxMag == 0 || !vc.Style.Line.SetStroke(plt) {
		return
	}
	scale := vc.Scale
	if scale == 0 {
		scale = 1
	}
	scale *= 0.9 * vc.spacing(plt) / float32(maxMag)
	lineColor := pc.Stroke.Color
	width := pc.Stroke.Width.Dots
	for i := range n {
		if !vc.valid(i) {
			continue
		}
		d := math32.Vec2(float32(vc.U[i]), -float32(vc.V[i])).MulScalar(scale)
		ln := d.Length()
		if ln == 0 {
			continue
		}
		p := math32.Vec2(vc.PX[i], vc.PY[i])
		switch vc.Pivot {
		case PivotMiddle:
			p = p.Sub(d.MulScalar(0.5))
		case PivotTip:
			p = p.Sub(d)
		}
		clr := lineColor
		if vc.Color != nil {
			if math.IsNaN(vc.Color[i]) {
				continue
			}
			clr = plt.ColorAxis.Uniform(vc.Color[i])
		}
		drawArrow(pc, p, d, clr, width)
	}
	pc.Fill.Color = nil
	if vc.Key {
		vc.drawKey(plt, scale, maxMag, lineColor, width)
	}
}

// drawArrow draws an arrow from the given point along the given vector,
// in pixels, with the given color and line width.
func drawArrow(pc *paint.Painter, p, d math32.Vector2, clr image.Image, width float32) {
	ln := d.Length()
	if ln == 0 {
		return
	}
	// the head: a triangle at the tip, smaller for short arrows
	dir := d.DivScalar(ln)
	hl := min(0.4*ln, 4*width+3)
	hw := 0.6 * hl
	tip := p.Add(d)
	base := tip.Sub(dir.MulScalar(hl))
	perp := math32.Vec2(-dir.Y, dir.X).MulScalar(hw)
	pc.Fill.Color = nil
	pc.Stroke.Color = clr
	pc.MoveTo(p.X, p.Y)
	pc.LineTo(base.X, base.Y)
	pc.Draw()
	pc.Stroke.Color = nil
	pc.Fill.Color = clr
	pc.MoveTo(tip.X, tip.Y)
	pc.LineTo(base.X+perp.X, base.Y+perp.Y)
	pc.LineTo(base.X-perp.X, base.Y-perp.Y)
	pc.Close()
	pc.Draw()
	pc.Fill.Color = nil
}

// keyValue returns the magnitude of the vector of the key arrow: the
// KeyValue, or else the largest round value (1, 2, or 5 times a power
// of 10) up to the given largest magnitude of the vectors.
func (vc *Vector) keyValue(maxMag float64) float64 {
	if vc.KeyValue > 0 {
		return vc.KeyValue
	}
	p := math.Pow(10, math.Floor(math.Log10(maxMag)))
	for _, f := range []float64{5, 2} {
		if f*p <= maxMag*(1+1e-9) {
			return f * p
		}
	}
	return p
}

// drawKey draws the key arrow at the top right of the plot, with its label
// to its left, on a translucent background, for the given scale of the
// arrows, in pixels per unit of magnitude.
func (vc *Vector) drawKey(plt *plot.Plot, scale float32, maxMag float64, clr image.Image, width float32) {
	pc := plt.Painter
	val := vc.keyValue(maxMag)
	lbl := vc.KeyLabel
	if lbl == "" {
		lbl = strconv.FormatFloat(val, 'g', 4, 64)
	}
	tx := &plot.Text{Text: lbl}
	tx.Defaults()
	tx.Style = plt.Style.Axis.TickText
	tx.Config(plt)
	tsz := tx.PaintText.Bounds.Size()
	pad := 0.5 * tsz.Y
	ln := float32(val) * scale
	pb := plt.PlotBox
	right := pb.Max.X - pad
	y := pb.Min.Y + pad + 0.5*tsz.Y
	left := right - ln - pad - tsz.X
	bg := colors.Uniform(colors.ApplyOpacity(colors.ToUniform(colors.Scheme.Surface), 0.85))
	pc.FillBox(math32.Vec2(left-0.5*pad, y-0.5*tsz.Y-0.5*pad), math32.Vec2(right-left+pad, tsz.Y+pad), bg)
	pc.DrawText(tx.PaintText, math32.Vec2(left, y-0.5*tsz.Y))
	drawArrow(pc, math32.Vec2(right-ln, y), math32.Vec2(ln, 0), clr, width)
}

// UpdateRange updates the given ranges.
func (vc *Vector) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if vc.Style.RightY {
		yax = &plt.YR
	}
	plot.Range(vc.X, &plt.X.Range)
	plot.Range(vc.Y, &yax.Range)
	// a margin for the arrows, whose lengths are about their spacing
	var xr, yr minmax.F64
	xr.SetInfinity()
	yr.SetInfinity()
	plot.Range(vc.X, &xr)
	plot.Range(vc.Y, &yr)
	sx, sy := vc.dataSpacing(xr, yr)
	f := 0.9 * float64(vc.Scale)
	if vc.Scale == 0 {
		f = 0.9
	}
	if vc.Pivot == PivotMiddle {
		f *= 0.5
	}
	plt.X.Range.FitInRange(minmax.F64{Min: xr.Min - f*sx, Max: xr.Max + f*sx})
	yax.Range.FitInRange(minmax.F64{Min: yr.Min - f*sy, Max: yr.Max + f*sy})
	if vc.Color != nil {
		plt.ColorAxis.AddRange(vc.Color)
	}
	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range
	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &vc.Style.Range)
	yax.DataRange = yax.Range
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (vc *Vector) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
	if !vc.Style.Line.SetStroke(plt) {
		return
	}
	clr := pc.Stroke.Color
	y := float32(ptb.Min.Y) + 0.5*float32(ptb.Dy())
	x0, x1 := float32(ptb.Min.X), float32(ptb.Max.X)
	hl := 0.4 * (x1 - x0)
	pc.Fill.Color = nil
	pc.MoveTo(x0, y)
	pc.LineTo(x1-hl, y)
	pc.Draw()
	pc.Stroke.Color = nil
	pc.Fill.Color = clr
	pc.MoveTo(x1, y)
	pc.LineTo(x1-hl, y-0.6*hl)
	pc.LineTo(x1-hl, y+0.6*hl)
	pc.Close()
	pc.Draw()
	pc.Fill.Color = nil
}
