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
	"fmt"
	"math"

	"cogentcore.org/core/math32"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/tensor"
)

// XYType is be used for specifying the type name.
const XYType = "XY"

func init() {
	plot.RegisterPlotter(XYType, "draws lines between and / or points for X,Y data values, using optional Size data for the points, for a bubble plot.", []plot.Roles{plot.Y}, []plot.Roles{plot.X, plot.Size}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewXY(plt, data)
	})
}

// XY draws lines between and / or points for XY data values.
type XY struct {
	// copies of data for this line
	X, Y, Size plot.Values

	// PX, PY are the actual pixel plotting coordinates for each XY value.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	stylers plot.Stylers
}

// NewXY adds a new XY plotter to given plot for given data,
// which can either by a [plot.Valuer] (e.g., Tensor) with the Y values,
// or a [plot.Data] with roles, and values defined.
// Data can also include Size for the points.
// Styler functions are obtained from the Y metadata if present.
func NewXY(plt *plot.Plot, data any) *XY {
	ln := &XY{}
	err := ln.SetData(data)
	if err != nil {
		// errors.Log(err) not useful actually
		return nil
	}
	ln.Defaults()
	plt.Add(ln)
	return ln
}

// SetData sets the plot data.
func (ln *XY) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.Y)
	if err != nil {
		return err
	}
	if err := dt.CheckLengths(); err != nil {
		return err
	}
	ln.Y = plot.MustCopyRole(dt, plot.Y)
	if _, ok := dt[plot.X]; !ok {
		ln.X, err = plot.CopyValues(tensor.NewIntRange(len(ln.Y)))
		if err != nil {
			return err
		}
	} else {
		ln.X = plot.MustCopyRole(dt, plot.X)
	}
	if ln.X == nil || ln.Y == nil {
		return fmt.Errorf("X or Y is nil")
	}
	ln.stylers = plot.GetStylersFromData(dt, plot.Y)
	ln.Size = plot.CopyRole(dt, plot.Size)
	return nil
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
	return
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (ln *XY) Plot(plt *plot.Plot) {
	ln.PX = plot.PlotX(plt, ln.X)
	minX, maxX := plt.PX(plt.X.DataRange.Min), plt.PX(plt.X.DataRange.Max)
	var minY, maxY float32
	if ln.Style.RightY {
		ln.PY = plot.PlotYR(plt, ln.Y)
		// flipped due to Y flip
		minY, maxY = plt.PYR(plt.YR.DataRange.Max), plt.PYR(plt.YR.DataRange.Min)
	} else {
		ln.PY = plot.PlotY(plt, ln.Y)
		minY, maxY = plt.PY(plt.Y.DataRange.Max), plt.PY(plt.Y.DataRange.Min)
	}
	np := min(len(ln.PX), len(ln.PY))
	if np == 0 {
		return
	}
	pc := plt.Painter
	oor := plt.Style.OutOfRange

	if ln.Style.Line.HasFill() {
		botY := maxY // note: min and max are now in pixel coords, not data coords -- y flip!
		pc.Fill.Color = ln.Style.Line.Fill
		pc.Stroke.Color = nil
		prevX := ln.PX[0]
		prevY := botY
		hasPrev := false
		for i, ptx := range ln.PX {
			pty := ln.PY[i]
			if math32.IsNaN(pty) {
				continue
			}
			if ptx > maxX {
				continue // could wrap around
			}
			if ptx < minX {
				hasPrev = false
				continue
			}
			if !hasPrev && !math32.IsNaN(ptx) && !math32.IsNaN(pty) {
				pc.MoveTo(ptx, botY)
				prevX = ptx
				hasPrev = true
			}
			switch ln.Style.Line.Step {
			case plot.NoStep:
				if ptx < prevX {
					pc.LineTo(prevX, botY)
					pc.Close()
					pc.MoveTo(ptx, botY)
				}
				pc.LineTo(ptx, pty)
			case plot.PreStep:
				if i == 0 {
					continue
				}
				if ptx < prevX {
					pc.LineTo(prevX, botY)
					pc.Close()
					pc.MoveTo(ptx, botY)
				} else {
					pc.LineTo(prevX, pty)
				}
				pc.LineTo(ptx, pty)
			case plot.MidStep:
				if ptx < prevX {
					pc.LineTo(prevX, botY)
					pc.Close()
					pc.MoveTo(ptx, botY)
				} else {
					pc.LineTo(0.5*(prevX+ptx), prevY)
					pc.LineTo(0.5*(prevX+ptx), pty)
				}
				pc.LineTo(ptx, pty)
			case plot.PostStep:
				if ptx < prevX {
					pc.LineTo(prevX, botY)
					pc.Close()
					pc.MoveTo(ptx, botY)
				} else {
					pc.LineTo(ptx, prevY)
				}
				pc.LineTo(ptx, pty)
			}
			prevX, prevY = ptx, pty
		}
		pc.LineTo(prevX, botY)
		pc.Close()
		pc.Draw()
	}
	pc.Fill.Color = nil

	hasOORMarks := false
	if ln.Style.Line.SetStroke(plt) {
		if plt.HighlightPlotter == ln {
			pc.Stroke.Width.Dots *= 2
		}
		prevX, prevY := ln.PX[0], ln.PY[0]
		hasPrev := false
		if !math32.IsNaN(prevX) && !math32.IsNaN(prevY) {
			pc.MoveTo(prevX, prevY)
			hasPrev = true
		}
		for i := 1; i < np; i++ {
			ptx, pty := ln.PX[i], ln.PY[i]
			if math32.IsNaN(ptx) || math32.IsNaN(pty) {
				continue
			}
			if ptx > maxX {
				continue // could wrap around
			}
			if ptx < minX {
				hasPrev = false
				continue
			}
			if pty < minY || pty > maxY {
				pty = min(pty, maxY)
				pty = max(pty, minY)
				if oor.HasMark() {
					hasOORMarks = true
				}
				if oor.HasBreak() {
					hasPrev = false
					continue
				}
			}
			if ln.Style.Line.Step != plot.NoStep {
				if hasPrev && ptx >= prevX {
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
					hasPrev = true
				}
			}
			if !hasPrev || (!ln.Style.Line.NegativeX && ptx < prevX) {
				hasPrev = true
				pc.MoveTo(ptx, pty)
			} else {
				pc.LineTo(ptx, pty)
			}
			prevX, prevY = ptx, pty
		}
		pc.Draw()
	}

	if hasOORMarks {
		ln.Style.OutOfRangeMark.SetStroke(plt)
		for i, ptx := range ln.PX {
			pty := ln.PY[i]
			if math32.IsNaN(ptx) || math32.IsNaN(pty) {
				continue
			}
			if ptx < minX || ptx > maxX {
				continue
			}
			if pty >= minY && pty <= maxY {
				continue
			}
			pty = min(pty, maxY)
			pty = max(pty, minY)
			ln.Style.OutOfRangeMark.DrawShape(pc, math32.Vec2(ptx, pty))
		}
	}

	if ln.Style.Point.SetStroke(plt) {
		origWidth := ln.Style.Point.Width
		origSize := ln.Style.Point.Size
		for i, ptx := range ln.PX {
			pty := ln.PY[i]
			if math32.IsNaN(ptx) || math32.IsNaN(pty) {
				continue
			}
			if ptx < minX || ptx > maxX || pty < minY || pty > maxY {
				continue
			}
			pc.Stroke.Width = origWidth
			ln.Style.Point.Size = origSize
			if plt.HighlightPlotter == ln {
				if i == plt.HighlightIndex {
					pc.Stroke.Width.Dots *= 2
					ln.Style.Point.Size.Dots *= 1.5
				}
			}
			if ln.Size != nil {
				sz := ln.Size.Float1D(i)
				if math.IsNaN(sz) {
					continue
				}
				ln.Style.Point.Size.Dots = 1 + ln.Style.Point.Size.Dots*float32(plt.SizeAxis.Norm(sz))
			}
			ln.Style.Point.SetColorIndex(pc, i)
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
	pc.Fill.Color = nil
}

// UpdateRange updates the given ranges.
func (ln *XY) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if ln.Style.RightY {
		yax = &plt.YR
	}
	plot.Range(ln.X, &plt.X.Range)
	plot.Range(ln.Y, &yax.Range)

	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range
	yFits := plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &ln.Style.Range)
	yax.DataRange = yax.Range

	ln.Style.OutOfRangeMark.IsOn(plt) // does dots
	if !ln.Style.Point.IsOn(plt) && (!plt.Style.OutOfRange.HasMark() || yFits) {
		return
	}
	psz := max(ln.Style.Point.Size.Dots, ln.Style.OutOfRangeMark.Size.Dots)
	bsz := plt.DataBox()
	dy := (float64(psz) / float64(bsz.Y)) * yax.Range.Range()
	yax.Range.Min -= dy
	yax.Range.Max += dy
	dx := (float64(psz) / float64(bsz.X)) * plt.X.Range.Range()
	plt.X.Range.Min -= dx
	plt.X.Range.Max += dx
	plot.Range(ln.Size, &plt.SizeAxis.Range)
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (ln *XY) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
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
		pc.Draw()
	}

	if ln.Style.Point.SetStroke(plt) {
		midX := 0.5 * float32(ptb.Min.X+ptb.Max.X)
		ln.Style.Point.DrawShape(pc, math32.Vec2(midX, midY))
	}
	pc.Fill.Color = nil
}
