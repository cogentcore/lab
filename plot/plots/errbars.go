// Copyright (c) 2024, Cogent Core. All rights reserved.
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
	eb := &YErrorBars{}
	err := eb.SetData(data)
	if err != nil {
		return nil
	}
	eb.Defaults()
	plt.Add(eb)
	return eb
}

// SetData sets the plot data.
func (eb *YErrorBars) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.Y)
	if err != nil {
		return err
	}
	if err := dt.CheckLengths(); err != nil {
		return err
	}
	eb.X = plot.MustCopyRole(dt, plot.X)
	eb.Y = plot.MustCopyRole(dt, plot.Y)
	eb.Low = plot.CopyRole(dt, plot.Low)
	eb.High = plot.CopyRole(dt, plot.High)
	if eb.Low == nil && eb.High != nil {
		eb.Low = eb.High
	}
	if eb.X == nil || eb.Y == nil || eb.Low == nil || eb.High == nil {
		return fmt.Errorf("X or Y or Low or High is nil")
	}
	eb.stylers = plot.GetStylersFromData(dt, plot.X, plot.Low, plot.High)
	eb.ystylers = plot.GetStylersFromData(dt, plot.Y)
	return nil
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

	minX, maxX := plt.PX(plt.X.DataRange.Min), plt.PX(plt.X.DataRange.Max)
	var minY, maxY float32
	if eb.Style.RightY {
		minY, maxY = plt.PYR(plt.YR.DataRange.Max), plt.PYR(plt.YR.DataRange.Min)
	} else {
		minY, maxY = plt.PY(plt.Y.DataRange.Max), plt.PY(plt.Y.DataRange.Min)
	}

	eb.Style.Line.SetStroke(plt)
	for i, y := range eb.Y {
		x := plt.PX(eb.X.Float1D(i))
		if math32.IsNaN(x) || math.IsNaN(y) {
			continue
		}
		if x < minX || x > maxX {
			continue
		}
		ylow := plt.PY(y - math.Abs(eb.Low[i]))
		yhigh := plt.PY(y + math.Abs(eb.High[i]))
		if math32.IsNaN(ylow) || math32.IsNaN(yhigh) {
			continue
		}
		if ylow < minY || yhigh > maxY {
			continue
		}

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
func (eb *YErrorBars) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if eb.Style.RightY {
		yax = &plt.YR
	}
	plot.Range(eb.X, &plt.X.Range)
	plot.Range(eb.Y, &yax.Range)

	for i, yv := range eb.Y {
		ylow := yv - math.Abs(eb.Low[i])
		yhigh := yv + math.Abs(eb.High[i])
		if math.IsNaN(ylow) || math.IsNaN(yhigh) {
			continue
		}
		yax.Range.FitInRange(minmax.F64{ylow, yhigh})
	}

	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range
	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &eb.Style.Range)
	yax.DataRange = yax.Range
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
	eb := &XErrorBars{}
	err := eb.SetData(data)
	if err != nil {
		return nil
	}
	eb.Defaults()
	plt.Add(eb)
	return eb
}

// SetData sets the plot data.
func (eb *XErrorBars) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.Y)
	if err != nil {
		return err
	}
	if err := dt.CheckLengths(); err != nil {
		return err
	}
	eb.X = plot.MustCopyRole(dt, plot.X)
	eb.Y = plot.MustCopyRole(dt, plot.Y)
	eb.Low = plot.MustCopyRole(dt, plot.Low)
	eb.High = plot.MustCopyRole(dt, plot.High)
	eb.Low = plot.CopyRole(dt, plot.Low)
	eb.High = plot.CopyRole(dt, plot.High)
	if eb.Low == nil && eb.High != nil {
		eb.Low = eb.High
	}
	if eb.X == nil || eb.Y == nil || eb.Low == nil || eb.High == nil {
		return nil
	}
	eb.stylers = plot.GetStylersFromData(dt, plot.High)
	eb.ystylers = plot.GetStylersFromData(dt, plot.Y)
	return nil
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

	minX, maxX := plt.PX(plt.X.DataRange.Min), plt.PX(plt.X.DataRange.Max)
	var minY, maxY float32
	if eb.Style.RightY {
		minY, maxY = plt.PYR(plt.YR.DataRange.Max), plt.PYR(plt.YR.DataRange.Min)
	} else {
		minY, maxY = plt.PY(plt.Y.DataRange.Max), plt.PY(plt.Y.DataRange.Min)
	}

	eb.Style.Line.SetStroke(plt)
	for i, x := range eb.X {
		y := plt.PY(eb.Y.Float1D(i))
		if math32.IsNaN(y) || math.IsNaN(x) {
			continue
		}
		if y < minY || y > maxY {
			continue
		}
		xlow := plt.PX(x - math.Abs(eb.Low[i]))
		xhigh := plt.PX(x + math.Abs(eb.High[i]))
		if math32.IsNaN(xlow) || math32.IsNaN(xhigh) {
			continue
		}
		if xlow < minX || xhigh > maxX {
			continue
		}

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
func (eb *XErrorBars) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if eb.Style.RightY {
		yax = &plt.YR
	}
	plot.Range(eb.X, &plt.X.Range)
	plot.Range(eb.Y, &yax.Range)
	for i, xv := range eb.X {
		xlow := xv - math.Abs(eb.Low[i])
		xhigh := xv + math.Abs(eb.High[i])
		if math.IsNaN(xlow) || math.IsNaN(xhigh) {
			continue
		}
		plt.X.Range.FitInRange(minmax.F64{xlow, xhigh})
	}

	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range

	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &eb.Style.Range)
	yax.DataRange = yax.Range
}
