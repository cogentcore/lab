// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"

	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
)

const (
	// YErrorBarsType is be used for specifying the type name.
	YErrorBarsType = "YErrorBars"

	// XErrorBarsType is be used for specifying the type name.
	XErrorBarsType = "XErrorBars"
)

func init() {
	plot.RegisterPlotter(YErrorBarsType, "draws draws vertical error bars, denoting error in Y values, using either High or Low & High data roles for error deviations around X, Y coordinates.", []plot.Roles{plot.X, plot.Y, plot.High}, []plot.Roles{plot.Low}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewYErrorBars(plt, data)
	})
	plot.RegisterPlotter(XErrorBarsType, "draws draws horizontal error bars, denoting error in X values, using either High or Low & High data roles for error deviations around X, Y coordinates.", []plot.Roles{plot.X, plot.Y, plot.High}, []plot.Roles{plot.Low}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewXErrorBars(plt, data)
	})
}

// YErrorBars draws vertical error bars, denoting error in Y values,
// using ether High or Low, High data roles for error deviations
// around X, Y coordinates.
type YErrorBars struct {
	// copies of data for this line
	X, Y, Low, High plot.Values

	// PX, PY are the actual pixel plotting coordinates for each XY value.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	stylers  plot.Stylers
	ystylers plot.Stylers
}

func (eb *YErrorBars) Defaults() {
	eb.Style.Defaults()
}

// NewYErrorBars adds a new YErrorBars plotter to given plot,
// using Low, High data roles for error deviations around X, Y coordinates.
// Styler functions are obtained from the High data if present.
func NewYErrorBars(plt *plot.Plot, data plot.Data) *YErrorBars {
	if data.CheckLengths() != nil {
		return nil
	}
	eb := &YErrorBars{}
	eb.X = plot.MustCopyRole(data, plot.X)
	eb.Y = plot.MustCopyRole(data, plot.Y)
	eb.Low = plot.CopyRole(data, plot.Low)
	eb.High = plot.CopyRole(data, plot.High)
	if eb.Low == nil && eb.High != nil {
		eb.Low = eb.High
	}
	if eb.X == nil || eb.Y == nil || eb.Low == nil || eb.High == nil {
		return nil
	}
	eb.stylers = plot.GetStylersFromData(data, plot.High)
	eb.ystylers = plot.GetStylersFromData(data, plot.Y)
	eb.Defaults()
	plt.Add(eb)
	return eb
}

// Styler adds a style function to set style parameters.
func (eb *YErrorBars) Styler(f func(s *plot.Style)) *YErrorBars {
	eb.stylers.Add(f)
	return eb
}

func (eb *YErrorBars) ApplyStyle(ps *plot.PlotStyle, idx int) {
	eb.Style.Line.SpacedColor(idx)
	ps.SetElementStyle(&eb.Style)
	yst := &plot.Style{}
	eb.ystylers.Run(yst)
	eb.Style.Range = yst.Range // get range from y
	eb.stylers.Run(&eb.Style)
}

func (eb *YErrorBars) Stylers() *plot.Stylers { return &eb.stylers }

func (eb *YErrorBars) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = eb.PX
	pixY = eb.PY
	data = plot.Data{}
	data[plot.X] = eb.X
	data[plot.Y] = eb.Y
	data[plot.Low] = eb.Low
	data[plot.High] = eb.High
	return
}

func (eb *YErrorBars) Plot(plt *plot.Plot) {
	pc := plt.Painter
	uc := &pc.UnitContext

	eb.Style.Width.Cap.ToDots(uc)
	cw := 0.5 * eb.Style.Width.Cap.Dots
	nv := len(eb.X)
	eb.PX = make([]float32, nv)
	eb.PY = make([]float32, nv)
	eb.Style.Line.SetStroke(plt)
	for i, y := range eb.Y {
		x := plt.PX(eb.X.Float1D(i))
		ylow := plt.PY(y - math.Abs(eb.Low[i]))
		yhigh := plt.PY(y + math.Abs(eb.High[i]))

		eb.PX[i] = x
		eb.PY[i] = yhigh

		pc.MoveTo(x, ylow)
		pc.LineTo(x, yhigh)

		pc.MoveTo(x-cw, ylow)
		pc.LineTo(x+cw, ylow)

		pc.MoveTo(x-cw, yhigh)
		pc.LineTo(x+cw, yhigh)
		pc.Draw()
	}
}

// UpdateRange updates the given ranges.
func (eb *YErrorBars) UpdateRange(plt *plot.Plot, x, y, yr, z *minmax.F64) {
	if eb.Style.RightY {
		y = yr
	}
	plot.Range(eb.X, x)
	plot.RangeClamp(eb.Y, y, &eb.Style.Range)
	for i, yv := range eb.Y {
		ylow := yv - math.Abs(eb.Low[i])
		yhigh := yv + math.Abs(eb.High[i])
		y.FitInRange(minmax.F64{ylow, yhigh})
	}
	return
}

//////// XErrorBars

// XErrorBars draws horizontal error bars, denoting error in X values,
// using ether High or Low, High data roles for error deviations
// around X, Y coordinates.
type XErrorBars struct {
	// copies of data for this line
	X, Y, Low, High plot.Values

	// PX, PY are the actual pixel plotting coordinates for each XY value.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	stylers  plot.Stylers
	ystylers plot.Stylers
	yrange   minmax.Range64
}

func (eb *XErrorBars) Defaults() {
	eb.Style.Defaults()
}

// NewXErrorBars adds a new XErrorBars plotter to given plot,
// using Low, High data roles for error deviations around X, Y coordinates.
func NewXErrorBars(plt *plot.Plot, data plot.Data) *XErrorBars {
	if data.CheckLengths() != nil {
		return nil
	}
	eb := &XErrorBars{}
	eb.X = plot.MustCopyRole(data, plot.X)
	eb.Y = plot.MustCopyRole(data, plot.Y)
	eb.Low = plot.MustCopyRole(data, plot.Low)
	eb.High = plot.MustCopyRole(data, plot.High)
	eb.Low = plot.CopyRole(data, plot.Low)
	eb.High = plot.CopyRole(data, plot.High)
	if eb.Low == nil && eb.High != nil {
		eb.Low = eb.High
	}
	if eb.X == nil || eb.Y == nil || eb.Low == nil || eb.High == nil {
		return nil
	}
	eb.stylers = plot.GetStylersFromData(data, plot.High)
	eb.ystylers = plot.GetStylersFromData(data, plot.Y)
	eb.Defaults()
	plt.Add(eb)
	return eb
}

// Styler adds a style function to set style parameters.
func (eb *XErrorBars) Styler(f func(s *plot.Style)) *XErrorBars {
	eb.stylers.Add(f)
	return eb
}

func (eb *XErrorBars) ApplyStyle(ps *plot.PlotStyle, idx int) {
	eb.Style.Line.SpacedColor(idx)
	ps.SetElementStyle(&eb.Style)
	yst := &plot.Style{}
	eb.ystylers.Run(yst)
	eb.yrange = yst.Range // get range from y
	eb.stylers.Run(&eb.Style)
}

func (eb *XErrorBars) Stylers() *plot.Stylers { return &eb.stylers }

func (eb *XErrorBars) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = eb.PX
	pixY = eb.PY
	data = plot.Data{}
	data[plot.X] = eb.X
	data[plot.Y] = eb.Y
	data[plot.Low] = eb.Low
	data[plot.High] = eb.High
	return
}

func (eb *XErrorBars) Plot(plt *plot.Plot) {
	pc := plt.Painter
	uc := &pc.UnitContext

	eb.Style.Width.Cap.ToDots(uc)
	cw := 0.5 * eb.Style.Width.Cap.Dots
	nv := len(eb.X)
	eb.PX = make([]float32, nv)
	eb.PY = make([]float32, nv)
	eb.Style.Line.SetStroke(plt)
	for i, x := range eb.X {
		y := plt.PY(eb.Y.Float1D(i))
		xlow := plt.PX(x - math.Abs(eb.Low[i]))
		xhigh := plt.PX(x + math.Abs(eb.High[i]))

		eb.PX[i] = xhigh
		eb.PY[i] = y

		pc.MoveTo(xlow, y)
		pc.LineTo(xhigh, y)

		pc.MoveTo(xlow, y-cw)
		pc.LineTo(xlow, y+cw)

		pc.MoveTo(xhigh, y-cw)
		pc.LineTo(xhigh, y+cw)
		pc.Draw()
	}
}

// UpdateRange updates the given ranges.
func (eb *XErrorBars) UpdateRange(plt *plot.Plot, x, y, yr, z *minmax.F64) {
	if eb.Style.RightY {
		y = yr
	}
	plot.RangeClamp(eb.X, x, &eb.Style.Range)
	plot.RangeClamp(eb.Y, y, &eb.yrange)
	for i, xv := range eb.X {
		xlow := xv - math.Abs(eb.Low[i])
		xhigh := xv + math.Abs(eb.High[i])
		x.FitInRange(minmax.F64{xlow, xhigh})
	}
	return
}
