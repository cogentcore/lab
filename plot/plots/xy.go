// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted from github.com/gonum/plot:
// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

//go:generate core generate

import (
	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/tensor"
)

// XYType is be used for specifying the type name.
const XYType = "XY"

func init() {
	plot.RegisterPlotter(XYType, "draws lines between and / or points for X,Y data values, using optional Size and Color data for the points, for a bubble plot.", []plot.Roles{plot.Y}, []plot.Roles{plot.X, plot.Size, plot.Color}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewXY(plt, data)
	})
}

// XY draws lines between and / or points for XY data values.
type XY struct {
	// copies of data for this line
	X, Y, Color, Size plot.Values

	// PX, PY are the actual pixel plotting coordinates for each XY value.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	stylers plot.Stylers
}

// NewXY adds a new XY plotter to given plot for given data,
// which can either by a [plot.Valuer] (e.g., Tensor) with the Y values,
// or a [plot.Data] with roles, and values defined.
// Data can also include Color and / or Size for the points.
// Styler functions are obtained from the Y metadata if present.
func NewXY(plt *plot.Plot, data any) *XY {
	dt := errors.Log1(plot.DataOrValuer(data, plot.Y))
	if dt == nil {
		return nil
	}
	if dt.CheckLengths() != nil {
		return nil
	}
	ln := &XY{}
	ln.Y = plot.MustCopyRole(dt, plot.Y)
	if _, ok := dt[plot.X]; !ok {
		ln.X = errors.Log1(plot.CopyValues(tensor.NewIntRange(len(ln.Y))))
	} else {
		ln.X = plot.MustCopyRole(dt, plot.X)
	}
	if ln.X == nil || ln.Y == nil {
		return nil
	}
	ln.stylers = plot.GetStylersFromData(dt, plot.Y)
	ln.Color = plot.CopyRole(dt, plot.Color)
	ln.Size = plot.CopyRole(dt, plot.Size)
	ln.Defaults()
	plt.Add(ln)
	return ln
}

// newXYWith is a simple helper function that creates a new XY plotter
// with lines and/or points.
func newXYWith(plt *plot.Plot, data any, line, point plot.DefaultOffOn) *XY {
	ln := NewXY(plt, data)
	if ln == nil {
		return ln
	}
	ln.Style.Line.On = line
	ln.Style.Point.On = point
	return ln
}

// NewLine adds an XY plot drawing Lines only by default, for given data
// which can either by a [plot.Valuer] (e.g., Tensor) with the Y values,
// or a [plot.Data] with roles, and values defined.
// See also [NewScatter] and [NewPointLine].
func NewLine(plt *plot.Plot, data any) *XY {
	return newXYWith(plt, data, plot.On, plot.Off)
}

// NewScatter adds an XY scatter plot drawing Points only by default, for given data
// which can either by a [plot.Valuer] (e.g., Tensor) with the Y values,
// or a [plot.Data] with roles, and values defined.
// See also [NewLine] and [NewPointLine].
func NewScatter(plt *plot.Plot, data any) *XY {
	return newXYWith(plt, data, plot.Off, plot.On)
}

// NewPointLine adds an XY plot drawing both lines and points by default, for given data
// which can either by a [plot.Valuer] (e.g., Tensor) with the Y values,
// or a [plot.Data] with roles, and values defined.
// See also [NewLine] and [NewScatter].
func NewPointLine(plt *plot.Plot, data any) *XY {
	return newXYWith(plt, data, plot.On, plot.On)
}

func (ln *XY) Defaults() {
	ln.Style.Defaults()
}

// Styler adds a style function to set style parameters.
func (ln *XY) Styler(f func(s *plot.Style)) *XY {
	ln.stylers.Add(f)
	return ln
}

func (ln *XY) Stylers() *plot.Stylers { return &ln.stylers }

func (ln *XY) ApplyStyle(ps *plot.PlotStyle, idx int) {
	ln.Style.Line.SpacedColor(idx)
	ln.Style.Point.SpacedColor(idx)
	ps.SetElementStyle(&ln.Style)
	ln.stylers.Run(&ln.Style)
}

func (ln *XY) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = ln.PX
	pixY = ln.PY
	data = plot.Data{}
	data[plot.X] = ln.X
	data[plot.Y] = ln.Y
	if ln.Size != nil {
		data[plot.Size] = ln.Size
	}
	if ln.Color != nil {
		data[plot.Color] = ln.Color
	}
	return
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (ln *XY) Plot(plt *plot.Plot) {
	ln.PX = plot.PlotX(plt, ln.X)
	ln.PY = plot.PlotY(plt, ln.Y)
	np := len(ln.PX)
	if np == 0 || len(ln.PY) == 0 {
		return
	}
	pc := plt.Paint
	if ln.Style.Line.HasFill() {
		pc.FillStyle.Color = ln.Style.Line.Fill
		minY := plt.PY(plt.Y.Range.Min)
		prevX := ln.PX[0]
		prevY := minY
		pc.MoveTo(prevX, prevY)
		for i, ptx := range ln.PX {
			pty := ln.PY[i]
			switch ln.Style.Line.Step {
			case plot.NoStep:
				if ptx < prevX {
					pc.LineTo(prevX, minY)
					pc.ClosePath()
					pc.MoveTo(ptx, minY)
				}
				pc.LineTo(ptx, pty)
			case plot.PreStep:
				if i == 0 {
					continue
				}
				if ptx < prevX {
					pc.LineTo(prevX, minY)
					pc.ClosePath()
					pc.MoveTo(ptx, minY)
				} else {
					pc.LineTo(prevX, pty)
				}
				pc.LineTo(ptx, pty)
			case plot.MidStep:
				if ptx < prevX {
					pc.LineTo(prevX, minY)
					pc.ClosePath()
					pc.MoveTo(ptx, minY)
				} else {
					pc.LineTo(0.5*(prevX+ptx), prevY)
					pc.LineTo(0.5*(prevX+ptx), pty)
				}
				pc.LineTo(ptx, pty)
			case plot.PostStep:
				if ptx < prevX {
					pc.LineTo(prevX, minY)
					pc.ClosePath()
					pc.MoveTo(ptx, minY)
				} else {
					pc.LineTo(ptx, prevY)
				}
				pc.LineTo(ptx, pty)
			}
			prevX, prevY = ptx, pty
		}
		pc.LineTo(prevX, minY)
		pc.ClosePath()
		pc.Fill()
	}
	pc.FillStyle.Color = nil

	if ln.Style.Line.SetStroke(plt) {
		if plt.HighlightPlotter == ln {
			pc.StrokeStyle.Width.Dots *= 2
		}
		prevX, prevY := ln.PX[0], ln.PY[0]
		pc.MoveTo(prevX, prevY)
		for i := 1; i < np; i++ {
			ptx, pty := ln.PX[i], ln.PY[i]
			if ln.Style.Line.Step != plot.NoStep {
				if ptx >= prevX {
					switch ln.Style.Line.Step {
					case plot.PreStep:
						pc.LineTo(prevX, pty)
					case plot.MidStep:
						pc.LineTo(0.5*(prevX+ptx), prevY)
						pc.LineTo(0.5*(prevX+ptx), pty)
					case plot.PostStep:
						pc.LineTo(ptx, prevY)
					}
				} else {
					pc.MoveTo(ptx, pty)
				}
			}
			if !ln.Style.Line.NegativeX && ptx < prevX {
				pc.MoveTo(ptx, pty)
			} else {
				pc.LineTo(ptx, pty)
			}
			prevX, prevY = ptx, pty
		}
		pc.Stroke()
	}
	if ln.Style.Point.SetStroke(plt) {
		origWidth := ln.Style.Point.Width
		origSize := ln.Style.Point.Size
		for i, ptx := range ln.PX {
			pty := ln.PY[i]
			if plt.HighlightPlotter == ln {
				if i == plt.HighlightIndex {
					pc.StrokeStyle.Width.Dots *= 2
					ln.Style.Point.Size.Dots *= 1.5
				} else {
					pc.StrokeStyle.Width = origWidth
					ln.Style.Point.Size = origSize
				}
			}
			ln.Style.Point.DrawShape(pc, math32.Vec2(ptx, pty))
		}
		ln.Style.Point.Size = origSize
	} else if plt.HighlightPlotter == ln {
		op := ln.Style.Point.On
		origSize := ln.Style.Point.Size
		ln.Style.Point.On = plot.On
		ln.Style.Point.Width.Pt(2)
		ln.Style.Point.Size.Pt(4.5)
		ln.Style.Point.SetStroke(plt)
		ptx := ln.PX[plt.HighlightIndex]
		pty := ln.PY[plt.HighlightIndex]
		ln.Style.Point.DrawShape(pc, math32.Vec2(ptx, pty))
		ln.Style.Point.On = op
		ln.Style.Point.Size = origSize
	}
	pc.FillStyle.Color = nil
}

// UpdateRange updates the given ranges.
func (ln *XY) UpdateRange(plt *plot.Plot, xr, yr, zr *minmax.F64) {
	// todo: include point sizes!
	plot.Range(ln.X, xr)
	plot.RangeClamp(ln.Y, yr, &ln.Style.Range)
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (ln *XY) Thumbnail(plt *plot.Plot) {
	pc := plt.Paint
	ptb := pc.Bounds
	midY := 0.5 * float32(ptb.Min.Y+ptb.Max.Y)

	if ln.Style.Line.Fill != nil {
		tb := ptb
		if ln.Style.Line.Width.Value > 0 {
			tb.Min.Y = int(midY)
		}
		pc.FillBox(math32.FromPoint(tb.Min), math32.FromPoint(tb.Size()), ln.Style.Line.Fill)
	}

	if ln.Style.Line.SetStroke(plt) {
		pc.MoveTo(float32(ptb.Min.X), midY)
		pc.LineTo(float32(ptb.Max.X), midY)
		pc.Stroke()
	}

	if ln.Style.Point.SetStroke(plt) {
		midX := 0.5 * float32(ptb.Min.X+ptb.Max.X)
		ln.Style.Point.DrawShape(pc, math32.Vec2(midX, midY))
	}
	pc.FillStyle.Color = nil
}
