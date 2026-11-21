// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"fmt"
	"math"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
)

const (
	// BoxType is used for specifying the type name of a vertical [Box].
	BoxType = "Box"

	// XBoxType is used for specifying the type name of a horizontal [Box],
	// with the values along the X axis.
	XBoxType = "XBox"
)

func init() {
	plot.RegisterPlotter(BoxType, "draws a box plot at each position: a box from the U (first quartile) to the V (third quartile) values, a line at the Y (median) value, and optional whiskers out to the Low and High values. Positions and the X labels are as for a Bar.", []plot.Roles{plot.Y, plot.U, plot.V}, []plot.Roles{plot.Low, plot.High, plot.X}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewBox(plt, data)
	})
	plot.RegisterPlotter(XBoxType, "draws a horizontal box plot at each position, with the values along the X axis: a box from the U (first quartile) to the V (third quartile) values, a line at the Y (median) value, and optional whiskers out to the Low and High values.", []plot.Roles{plot.Y, plot.U, plot.V}, []plot.Roles{plot.Low, plot.High, plot.X}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewXBox(plt, data)
	})
}

// Box draws a box plot at each position: a box from the first quartile
// (U data role) to the third quartile (V), a line at the median (Y),
// and optional whiskers from the box out to the Low and High values
// (e.g., the min and max, or the furthest values within 1.5 times the
// interquartile range), with caps of the Width.Cap style. Outliers can
// be plotted with a separate XY scatter plot.
//
// The boxes are positioned along the category axis as for a [Bar]:
// at Width.Offset + index * Width.Stride, with a width of Width.Width,
// and the optional X data provides the labels for the category axis.
// The box is filled with the Line.Fill, and its outline, median line,
// and whiskers are drawn with the Line style; the median line is
// twice as wide.
//
// Horizontal boxes have their values along the X axis instead
// (all of the data roles are the same).
type Box struct {
	// Median, Q1, Q3 are the Y, U, and V data: the median line
	// and the first and third quartile ends of the box.
	Median, Q1, Q3 plot.Values

	// Low, High are the optional ends of the whiskers.
	Low, High plot.Values

	// XLabels are the optional labels for the category axis.
	XLabels plot.Labels

	// PX, PY are the actual pixel plotting coordinates of the
	// median at each position.
	PX, PY []float32

	// Style has the properties used to render the boxes.
	Style plot.Style

	// Horizontal specifies that the boxes are horizontal,
	// with the values along the X axis.
	Horizontal bool

	stylers plot.Stylers
}

// NewBox adds a new vertical [Box] plotter to the given plot, for the
// given [plot.Data] with the Y (median), U (first quartile), and V
// (third quartile) roles, and the optional Low, High whisker ends and
// X labels. Styler functions are obtained from the Y metadata if present.
func NewBox(plt *plot.Plot, data plot.Data) *Box {
	return newBox(plt, data, false)
}

// NewXBox adds a new horizontal [Box] plotter to the given plot,
// with the values along the X axis. See [NewBox] for the data.
func NewXBox(plt *plot.Plot, data plot.Data) *Box {
	return newBox(plt, data, true)
}

func newBox(plt *plot.Plot, data plot.Data, horiz bool) *Box {
	bx := &Box{Horizontal: horiz}
	err := bx.SetData(data)
	if err != nil {
		return nil
	}
	bx.Defaults()
	plt.Add(bx)
	return bx
}

func (bx *Box) Defaults() {
	bx.Style.Defaults()
}

// SetData sets the plot data.
func (bx *Box) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.Y)
	if err != nil {
		return err
	}
	if err := dt.CheckLengths(); err != nil {
		return err
	}
	bx.Median = plot.CopyRole(dt, plot.Y)
	bx.Q1 = plot.CopyRole(dt, plot.U)
	bx.Q3 = plot.CopyRole(dt, plot.V)
	if bx.Median == nil || bx.Q1 == nil || bx.Q3 == nil {
		return fmt.Errorf("nil Y (median), U (Q1), or V (Q3) data")
	}
	bx.Low = plot.CopyRole(dt, plot.Low)
	bx.High = plot.CopyRole(dt, plot.High)
	bx.XLabels = plot.CopyRoleLabels(dt, plot.X)
	bx.stylers = plot.GetStylersFromData(dt, plot.X, plot.Y)
	return nil
}

// Styler adds a style function to set style parameters.
func (bx *Box) Styler(f func(s *plot.Style)) *Box {
	bx.stylers.Add(f)
	return bx
}

func (bx *Box) Stylers() *plot.Stylers { return &bx.stylers }

func (bx *Box) ApplyStyle(ps *plot.PlotStyle, idx int) {
	bx.Style.Line.SpacedFill(idx)
	ps.SetElementStyle(&bx.Style)
	bx.stylers.Run(&bx.Style)
}

func (bx *Box) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = bx.PX
	pixY = bx.PY
	data = plot.Data{}
	data[plot.Y] = bx.Median
	data[plot.U] = bx.Q1
	data[plot.V] = bx.Q3
	if bx.Low != nil {
		data[plot.Low] = bx.Low
	}
	if bx.High != nil {
		data[plot.High] = bx.High
	}
	if bx.XLabels != nil {
		data[plot.X] = bx.XLabels
	}
	return
}

// numValid returns the number of boxes up to the last one with a median,
// excluding NaN padding at the end (e.g., of a table with other data).
func (bx *Box) numValid() int {
	n := len(bx.Median)
	for n > 0 && math.IsNaN(bx.Median[n-1]) {
		n--
	}
	return n
}

// pos returns the position of the box at index i on the category axis.
func (bx *Box) pos(i int) float64 {
	return bx.Style.Width.Offset + float64(i)*bx.Style.Width.Stride
}

// whisker returns the value of the given whisker data at index i,
// NaN if there is none.
func whisker(w plot.Values, i int) float64 {
	if i >= len(w) {
		return math.NaN()
	}
	return w[i]
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (bx *Box) Plot(plt *plot.Plot) {
	pc := plt.Painter
	n := len(bx.Median)
	bx.PX = make([]float32, n)
	bx.PY = make([]float32, n)
	py := plt.PY
	if bx.Style.RightY {
		py = plt.PYR
	}
	// pt returns the pixel point for the given category position and value.
	pt := func(cat, val float64) math32.Vector2 {
		if bx.Horizontal {
			return math32.Vec2(plt.PX(val), py(cat))
		}
		return math32.Vec2(plt.PX(cat), py(val))
	}
	line := func(cat0, val0, cat1, val1 float64) {
		p0, p1 := pt(cat0, val0), pt(cat1, val1)
		pc.MoveTo(p0.X, p0.Y)
		pc.LineTo(p1.X, p1.Y)
	}

	hw := 0.5 * bx.Style.Width.Width
	stroke := bx.Style.Line.SetStroke(plt)
	if stroke && plt.HighlightPlotter == bx {
		pc.Stroke.Width.Dots *= 2
	}
	width := pc.Stroke.Width.Dots
	bx.Style.Width.Cap.ToDots(&pc.UnitContext)
	cw := 0.5 * bx.Style.Width.Cap.Dots
	for i, med := range bx.Median {
		q1, q3 := bx.Q1[i], bx.Q3[i]
		if math.IsNaN(med) || math.IsNaN(q1) || math.IsNaN(q3) {
			continue
		}
		cat := bx.pos(i)
		mp := pt(cat, med)
		bx.PX[i], bx.PY[i] = mp.X, mp.Y

		// the box
		p0, p1 := pt(cat-hw, q1), pt(cat+hw, q3)
		box := math32.Box2{Min: p0.Min(p1), Max: p0.Max(p1)}
		pc.Fill.Color = bx.Style.Line.Fill
		if !stroke {
			pc.Stroke.Color = nil
		}
		pc.Rectangle(box.Min.X, box.Min.Y, box.Size().X, box.Size().Y)
		pc.Draw()
		pc.Fill.Color = nil
		if !stroke {
			continue
		}

		// whiskers, with caps
		for _, wv := range []struct{ end, edge float64 }{{whisker(bx.Low, i), q1}, {whisker(bx.High, i), q3}} {
			if math.IsNaN(wv.end) {
				continue
			}
			line(cat, wv.edge, cat, wv.end)
			ep := pt(cat, wv.end)
			if bx.Horizontal {
				pc.MoveTo(ep.X, ep.Y-cw)
				pc.LineTo(ep.X, ep.Y+cw)
			} else {
				pc.MoveTo(ep.X-cw, ep.Y)
				pc.LineTo(ep.X+cw, ep.Y)
			}
		}
		pc.Draw()

		// the median line, twice as wide
		pc.Stroke.Width.Dots = 2 * width
		line(cat-hw, med, cat+hw, med)
		pc.Draw()
		pc.Stroke.Width.Dots = width
	}
}

// numCategories returns the number of categories, up to the last one with
// a median or a label: the categories are shared by series of boxes next
// to each other (e.g., from a Split column), which can have no values
// for some of them.
func (bx *Box) numCategories() int {
	n := len(bx.Median)
	for n > 0 && math.IsNaN(bx.Median[n-1]) && bx.label(n-1) == "" {
		n--
	}
	return n
}

// label returns the label of the given category, if any.
func (bx *Box) label(i int) string {
	if i < len(bx.XLabels) {
		return bx.XLabels[i]
	}
	return ""
}

// UpdateRange updates the given ranges.
func (bx *Box) UpdateRange(plt *plot.Plot) {
	bw := bx.Style.Width
	n := bx.numCategories()
	catMin := bw.Offset - bw.Pad
	catMax := bw.Offset + float64(n-1)*bw.Stride + bw.Pad
	yax := &plt.Y
	if bx.Style.RightY {
		yax = &plt.YR
	}
	catRange, valRange := &plt.X.Range, &yax.Range
	if bx.Horizontal {
		catRange, valRange = valRange, &plt.X.Range
	}
	var ticks plot.ConstantTicks
	for i := range n {
		for _, v := range []float64{bx.Median[i], bx.Q1[i], bx.Q3[i], whisker(bx.Low, i), whisker(bx.High, i)} {
			if !math.IsNaN(v) {
				valRange.FitValInRange(v)
			}
		}
		if bx.XLabels == nil {
			continue
		}
		lb := bx.label(i)
		if lb == "" {
			if math.IsNaN(bx.Median[i]) {
				continue
			}
			lb = fmt.Sprintf("%d", i)
		}
		ticks = append(ticks, plot.Tick{Value: bx.pos(i) + bw.TickOffset, Label: lb})
	}
	catRange.FitInRange(minmax.F64{Min: catMin, Max: catMax})
	if ticks != nil {
		if bx.Horizontal {
			yax.Ticker = ticks
		} else {
			plt.X.Ticker = ticks
		}
	}

	if bx.Horizontal {
		plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &bx.Style.Range)
		plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &plt.Style.XAxis.Range)
	} else {
		plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
		plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &bx.Style.Range)
	}
	plt.X.DataRange = plt.X.Range
	yax.DataRange = yax.Range
}

// Thumbnail fulfills the plot.Thumbnailer interface.
func (bx *Box) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
	if !bx.Style.Line.SetStroke(plt) {
		pc.Stroke.Color = nil
	}
	pc.Fill.Color = bx.Style.Line.Fill
	pc.Rectangle(float32(ptb.Min.X), float32(ptb.Min.Y), float32(ptb.Size().X), float32(ptb.Size().Y))
	pc.Draw()
	pc.Fill.Color = nil
	if pc.Stroke.Color != nil {
		midY := 0.5 * float32(ptb.Min.Y+ptb.Max.Y)
		pc.MoveTo(float32(ptb.Min.X), midY)
		pc.LineTo(float32(ptb.Max.X), midY)
		pc.Draw()
	}
}
