// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"fmt"
	"image"
	"math"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/tensor"
)

const (
	// BandType is used for specifying the type name of a [Band]
	// that fills between Low and High Y values along the X axis.
	BandType = "Band"

	// XBandType is used for specifying the type name of a [Band]
	// that fills between Low and High X values along the Y axis.
	XBandType = "XBand"
)

func init() {
	plot.RegisterPlotter(BandType, "fills the area between Low and High Y values along the X axis: deviations around Y if given (e.g., an error band around a mean line), else absolute values (e.g., a stacked area), with Low = 0 if not given.", []plot.Roles{plot.High}, []plot.Roles{plot.Low, plot.X, plot.Y}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewBand(plt, data)
	})
	plot.RegisterPlotter(XBandType, "fills the area between Low and High X values along the Y axis: deviations around X if given (e.g., the body of a violin plot at X), else absolute values, with Low = 0 if not given.", []plot.Roles{plot.High}, []plot.Roles{plot.Low, plot.Y, plot.X}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewXBand(plt, data)
	})
}

// BandOpacity is the opacity of the default fill of a [Band],
// which is the line color when the Line.Fill is not set.
const BandOpacity = 0.3

// Band fills the area between Low and High data values along an
// independent axis (the area between two curves, as in fill_between
// in matplotlib). The band is vertical (Low, High are Y values at each
// X) unless Horizontal is set, in which case Low, High are X values
// at each Y (as in fill_betweenx), for the body of a violin plot.
//
// If the Center values (Y, or X for Horizontal) are given, Low and High
// are deviations below and above them, as in [YErrorBars] (their
// absolute values are used, and Low is High if not given), so that the
// same data can be plotted as error bars or as an error band, and a
// violin body is a Horizontal band of density values around its X.
// Otherwise, Low and High are the edges of the band, and Low is 0 if not
// given, to fill the area below the High curve (e.g., a stacked area).
//
// NaN values in any of the data break the band into separate parts.
//
// The fill is the Line.Fill style if set, and otherwise the Line.Color
// at [BandOpacity]. The edges are drawn with the Line style only if
// Line.On is On.
type Band struct {
	// Pos are the positions along the independent axis:
	// X values, or Y values for a Horizontal band.
	// They are the indexes of the values if not given.
	Pos plot.Values

	// Center are the optional values that Low and High are deviations
	// around: Y values, or X values for a Horizontal band.
	// If nil, Low and High are the edges of the band.
	Center plot.Values

	// Low, High are the low and high edges of the band at each Pos,
	// or the deviations below and above Center if it is set:
	// Y values, or X values for a Horizontal band.
	Low, High plot.Values

	// PX, PY are the actual pixel plotting coordinates of the High
	// edge at each Pos.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	// Horizontal specifies that the band runs along the Y axis,
	// so that Low, High are X values at each Y.
	Horizontal bool

	stylers plot.Stylers
}

// NewBand adds a new vertical [Band] plotter to the given plot,
// filling between Low and High Y values along the X axis.
// The data can be a [plot.Data] with the High role and optionally
// Low, X, and Y (the Center), or a [plot.Valuer] (e.g., Tensor) with the High values,
// for a fill down to 0. Styler functions are obtained from the data
// if present.
func NewBand(plt *plot.Plot, data any) *Band {
	return newBand(plt, data, false)
}

// NewXBand adds a new Horizontal [Band] plotter to the given plot,
// filling between Low and High X values along the Y axis, for the
// body of a violin plot, for example.
// The data can be a [plot.Data] with the High role and optionally
// Low, Y, and X (the Center), or a [plot.Valuer] (e.g., Tensor) with the High values,
// for a fill to 0. Styler functions are obtained from the data
// if present.
func NewXBand(plt *plot.Plot, data any) *Band {
	return newBand(plt, data, true)
}

func newBand(plt *plot.Plot, data any, horiz bool) *Band {
	bd := &Band{Horizontal: horiz}
	err := bd.SetData(data)
	if err != nil {
		return nil
	}
	bd.Defaults()
	plt.Add(bd)
	return bd
}

func (bd *Band) Defaults() {
	bd.Style.Defaults()
}

// roles returns the roles of the independent position data
// and the center data.
func (bd *Band) roles() (pos, center plot.Roles) {
	if bd.Horizontal {
		return plot.Y, plot.X
	}
	return plot.X, plot.Y
}

// SetData sets the plot data.
func (bd *Band) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.High)
	if err != nil {
		return err
	}
	if err := dt.CheckLengths(); err != nil {
		return err
	}
	bd.High = plot.CopyRole(dt, plot.High)
	if bd.High == nil {
		return fmt.Errorf("nil High data")
	}
	pr, cr := bd.roles()
	bd.Center = plot.CopyRole(dt, cr)
	bd.Low = plot.CopyRole(dt, plot.Low)
	if bd.Low == nil {
		if bd.Center != nil {
			bd.Low = bd.High
		} else {
			bd.Low = make(plot.Values, len(bd.High))
		}
	}
	bd.Pos = plot.CopyRole(dt, pr)
	if bd.Pos == nil {
		bd.Pos, err = plot.CopyValues(tensor.NewIntRange(len(bd.High)))
		if err != nil {
			return err
		}
	}
	bd.stylers = plot.GetStylersFromData(dt, pr, cr, plot.Low, plot.High)
	return nil
}

// Styler adds a style function to set style parameters.
func (bd *Band) Styler(f func(s *plot.Style)) *Band {
	bd.stylers.Add(f)
	return bd
}

func (bd *Band) Stylers() *plot.Stylers { return &bd.stylers }

func (bd *Band) ApplyStyle(ps *plot.PlotStyle, idx int) {
	bd.Style.Line.SpacedColor(idx)
	ps.SetElementStyle(&bd.Style)
	bd.stylers.Run(&bd.Style)
}

func (bd *Band) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = bd.PX
	pixY = bd.PY
	data = plot.Data{}
	pr, cr := bd.roles()
	data[pr] = bd.Pos
	if bd.Center != nil {
		data[cr] = bd.Center
	}
	data[plot.Low] = bd.Low
	data[plot.High] = bd.High
	return
}

// fill returns the fill color for the band: the Line.Fill if set,
// and otherwise the Line.Color at BandOpacity (nil if neither).
func (bd *Band) fill() image.Image {
	ls := &bd.Style.Line
	if ls.Fill != nil {
		if !ls.HasFill() {
			return nil
		}
		return colors.Uniform(colors.ToUniform(ls.Fill))
	}
	if ls.Color == nil {
		return nil
	}
	return colors.Uniform(colors.ApplyOpacity(colors.ToUniform(ls.Color), BandOpacity))
}

// edges returns the low and high edges of the band at index i.
func (bd *Band) edges(i int) (lo, hi float64) {
	if bd.Center == nil {
		return bd.Low[i], bd.High[i]
	}
	c := bd.Center[i]
	return c - math.Abs(bd.Low[i]), c + math.Abs(bd.High[i])
}

// segments returns the [start, end) index ranges of the parts of the
// band without any NaN values.
func (bd *Band) segments() [][2]int {
	var segs [][2]int
	st := -1
	for i, p := range bd.Pos {
		lo, hi := bd.edges(i)
		ok := !math.IsNaN(p) && !math.IsNaN(lo) && !math.IsNaN(hi)
		switch {
		case ok && st < 0:
			st = i
		case !ok && st >= 0:
			segs = append(segs, [2]int{st, i})
			st = -1
		}
	}
	if st >= 0 {
		segs = append(segs, [2]int{st, len(bd.Pos)})
	}
	return segs
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (bd *Band) Plot(plt *plot.Plot) {
	pc := plt.Painter
	n := len(bd.Pos)
	bd.PX = make([]float32, n)
	bd.PY = make([]float32, n)
	pxl := make([]float32, n)
	pyl := make([]float32, n)
	py := plt.PY
	if bd.Style.RightY {
		py = plt.PYR
	}
	for i, p := range bd.Pos {
		lo, hi := bd.edges(i)
		if bd.Horizontal {
			bd.PX[i] = plt.PX(hi)
			pxl[i] = plt.PX(lo)
			bd.PY[i] = py(p)
			pyl[i] = bd.PY[i]
		} else {
			bd.PX[i] = plt.PX(p)
			pxl[i] = bd.PX[i]
			bd.PY[i] = py(hi)
			pyl[i] = py(lo)
		}
	}
	segs := bd.segments()

	// fill: the High edge forward, then the Low edge backward.
	if fill := bd.fill(); fill != nil {
		pc.Fill.Color = fill
		pc.Stroke.Color = nil
		for _, sg := range segs {
			pc.MoveTo(bd.PX[sg[0]], bd.PY[sg[0]])
			for i := sg[0] + 1; i < sg[1]; i++ {
				pc.LineTo(bd.PX[i], bd.PY[i])
			}
			for i := sg[1] - 1; i >= sg[0]; i-- {
				pc.LineTo(pxl[i], pyl[i])
			}
			pc.Close()
		}
		pc.Draw()
		pc.Fill.Color = nil
	}

	// edges: only if lines are explicitly on.
	if bd.Style.Line.On == plot.On && bd.Style.Line.SetStroke(plt) {
		if plt.HighlightPlotter == bd {
			pc.Stroke.Width.Dots *= 2
		}
		for _, sg := range segs {
			pc.MoveTo(bd.PX[sg[0]], bd.PY[sg[0]])
			for i := sg[0] + 1; i < sg[1]; i++ {
				pc.LineTo(bd.PX[i], bd.PY[i])
			}
			pc.MoveTo(pxl[sg[0]], pyl[sg[0]])
			for i := sg[0] + 1; i < sg[1]; i++ {
				pc.LineTo(pxl[i], pyl[i])
			}
		}
		pc.Draw()
	}
}

// UpdateRange updates the given ranges.
func (bd *Band) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if bd.Style.RightY {
		yax = &plt.YR
	}
	posRange, valRange := &yax.Range, &plt.X.Range
	if !bd.Horizontal {
		posRange, valRange = valRange, posRange
	}
	plot.Range(bd.Pos, posRange)
	for i := range bd.Pos {
		lo, hi := bd.edges(i)
		if math.IsNaN(lo) || math.IsNaN(hi) {
			continue
		}
		valRange.FitInRange(minmax.F64{Min: min(lo, hi), Max: max(lo, hi)})
	}

	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range
	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &bd.Style.Range)
	yax.DataRange = yax.Range
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (bd *Band) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
	if fill := bd.fill(); fill != nil {
		pc.FillBox(math32.FromPoint(ptb.Min), math32.FromPoint(ptb.Size()), fill)
	}
	if bd.Style.Line.On == plot.On && bd.Style.Line.SetStroke(plt) {
		pc.Rectangle(float32(ptb.Min.X), float32(ptb.Min.Y), float32(ptb.Dx()), float32(ptb.Dy()))
		pc.Draw()
	}
	pc.Fill.Color = nil
}
