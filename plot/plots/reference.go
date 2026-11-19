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
	"cogentcore.org/lab/plot"
)

const (
	// HLineType is used for specifying the type name of a [Reference]
	// that draws horizontal lines at Y values.
	HLineType = "HLine"

	// VLineType is used for specifying the type name of a [Reference]
	// that draws vertical lines at X values.
	VLineType = "VLine"

	// HSpanType is used for specifying the type name of a [Reference]
	// that fills horizontal spans between Low and High Y values.
	HSpanType = "HSpan"

	// VSpanType is used for specifying the type name of a [Reference]
	// that fills vertical spans between Low and High X values.
	VSpanType = "VSpan"
)

func init() {
	plot.RegisterPlotter(HLineType, "draws a horizontal reference line across the whole plot at each Y value (e.g., at chance level, or 0). In a table, the values are those of its column, whatever its role.", []plot.Roles{}, []plot.Roles{}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return newReference(plt, data, false, false)
	})
	plot.RegisterPlotter(VLineType, "draws a vertical reference line across the whole plot at each X value (e.g., at an event time). In a table, the values are those of its column, whatever its role (other than X, which is the X axis).", []plot.Roles{}, []plot.Roles{}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return newReference(plt, data, true, false)
	})
	plot.RegisterPlotter(HSpanType, "fills a horizontal span across the whole plot between each pair of Low and High Y values.", []plot.Roles{plot.Low, plot.High}, []plot.Roles{}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return newReference(plt, data, false, true)
	})
	plot.RegisterPlotter(VSpanType, "fills a vertical span across the whole plot between each pair of Low and High X values (e.g., an epoch of time).", []plot.Roles{plot.Low, plot.High}, []plot.Roles{}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return newReference(plt, data, true, true)
	})
}

// SpanOpacity is the opacity of the default fill of a span [Reference],
// which is the line color when the Line.Fill is not set.
const SpanOpacity = 0.15

// Reference draws reference lines or spans across the whole plot,
// at fixed values on one axis (as axhline, axvline, axhspan, and
// axvspan do in matplotlib): horizontal lines at Y values, vertical
// lines at X values, or spans filled between pairs of Low and High
// values. They are included in the range of their axis.
//
// The lines, and the edges of spans if Line.On is On, are drawn with
// the Line style, which has the [colors.Scheme.OnSurface] color by
// default, as a neutral reference. The spans are filled with the
// Line.Fill if set, and otherwise the Line.Color at [SpanOpacity].
type Reference struct {
	// Low are the values of the lines, or the low edges of the spans.
	Low plot.Values

	// High are the high edges of the spans, nil for lines.
	High plot.Values

	// PX, PY are the pixel coordinates of the lines, or of the low
	// edges of the spans: the X coordinates for Vertical ones,
	// and the Y coordinates otherwise.
	PX, PY []float32

	// Style is the style for plotting.
	Style plot.Style

	// Vertical specifies that the lines or spans are vertical,
	// at X values, instead of horizontal at Y values.
	Vertical bool

	stylers plot.Stylers
}

// NewHLine adds horizontal reference lines across the whole plot,
// at each of the given Y values.
func NewHLine(plt *plot.Plot, y ...float64) *Reference {
	return newReference(plt, plot.Data{plot.Y: plot.Values(y)}, false, false)
}

// NewVLine adds vertical reference lines across the whole plot,
// at each of the given X values.
func NewVLine(plt *plot.Plot, x ...float64) *Reference {
	return newReference(plt, plot.Data{plot.X: plot.Values(x)}, true, false)
}

// NewHSpan adds a horizontal span across the whole plot,
// filled between the given low and high Y values.
func NewHSpan(plt *plot.Plot, low, high float64) *Reference {
	return newReference(plt, plot.Data{plot.Low: plot.Values{low}, plot.High: plot.Values{high}}, false, true)
}

// NewVSpan adds a vertical span across the whole plot,
// filled between the given low and high X values.
func NewVSpan(plt *plot.Plot, low, high float64) *Reference {
	return newReference(plt, plot.Data{plot.Low: plot.Values{low}, plot.High: plot.Values{high}}, true, true)
}

func newReference(plt *plot.Plot, data plot.Data, vert, span bool) *Reference {
	rf := &Reference{Vertical: vert}
	var err error
	if span {
		if _, ok := data[plot.High]; !ok {
			return nil
		}
		err = rf.SetData(data)
	} else {
		err = rf.setLines(data)
	}
	if err != nil {
		return nil
	}
	rf.Defaults()
	plt.Add(rf)
	return rf
}

func (rf *Reference) Defaults() {
	rf.Style.Defaults()
}

// lineRole returns the role of the line values.
func (rf *Reference) lineRole() plot.Roles {
	if rf.Vertical {
		return plot.X
	}
	return plot.Y
}

// setLines sets the data for lines, from the X role for Vertical lines,
// or the Y role otherwise, or else from the one role in the data,
// which is the case for table plots, where the values are those of the
// column with the line plotter.
func (rf *Reference) setLines(data plot.Data) error {
	rl := rf.lineRole()
	if _, ok := data[rl]; !ok && len(data) == 1 {
		for r := range data {
			rl = r
		}
	}
	rf.Low = plot.CopyRole(data, rl)
	if rf.Low == nil {
		return fmt.Errorf("nil %s data", rl)
	}
	rf.High = nil
	rf.stylers = plot.GetStylersFromData(data, rl)
	return nil
}

// SetData sets the plot data: Low and High for spans,
// or the values of lines in the X or Y role (or a [plot.Valuer]).
func (rf *Reference) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, rf.lineRole())
	if err != nil {
		return err
	}
	if _, ok := dt[plot.High]; !ok {
		return rf.setLines(dt)
	}
	if err := dt.CheckLengths(); err != nil {
		return err
	}
	rf.Low = plot.CopyRole(dt, plot.Low)
	rf.High = plot.CopyRole(dt, plot.High)
	if rf.Low == nil || rf.High == nil {
		return fmt.Errorf("nil Low or High data")
	}
	rf.stylers = plot.GetStylersFromData(dt, plot.Low, plot.High)
	return nil
}

// IsSpan returns whether these are spans, not lines.
func (rf *Reference) IsSpan() bool {
	return rf.High != nil
}

// Styler adds a style function to set style parameters.
func (rf *Reference) Styler(f func(s *plot.Style)) *Reference {
	rf.stylers.Add(f)
	return rf
}

func (rf *Reference) Stylers() *plot.Stylers { return &rf.stylers }

func (rf *Reference) ApplyStyle(ps *plot.PlotStyle, idx int) {
	ps.SetElementStyle(&rf.Style)
	rf.stylers.Run(&rf.Style)
}

func (rf *Reference) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = rf.PX
	pixY = rf.PY
	data = plot.Data{}
	if rf.IsSpan() {
		data[plot.Low] = rf.Low
		data[plot.High] = rf.High
	} else {
		data[rf.lineRole()] = rf.Low
	}
	return
}

// fill returns the fill color for spans: the Line.Fill if set,
// and otherwise the Line.Color at SpanOpacity (nil if neither).
func (rf *Reference) fill() image.Image {
	ls := &rf.Style.Line
	if ls.Fill != nil {
		if !ls.HasFill() {
			return nil
		}
		return ls.Fill
	}
	if ls.Color == nil {
		return nil
	}
	return colors.Uniform(colors.ApplyOpacity(colors.ToUniform(ls.Color), SpanOpacity))
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (rf *Reference) Plot(plt *plot.Plot) {
	pc := plt.Painter
	pb := plt.PlotBox
	pv := plt.PY
	if rf.Vertical {
		pv = plt.PX
	} else if rf.Style.RightY {
		pv = plt.PYR
	}
	n := len(rf.Low)
	rf.PX = make([]float32, n)
	rf.PY = make([]float32, n)
	// line draws a line across the plot at pixel position p.
	line := func(p float32) {
		if rf.Vertical {
			pc.MoveTo(p, pb.Min.Y)
			pc.LineTo(p, pb.Max.Y)
		} else {
			pc.MoveTo(pb.Min.X, p)
			pc.LineTo(pb.Max.X, p)
		}
	}
	for i, v := range rf.Low {
		p := pv(v)
		if rf.Vertical {
			rf.PX[i], rf.PY[i] = p, pb.Min.Y
		} else {
			rf.PX[i], rf.PY[i] = pb.Min.X, p
		}
	}

	if !rf.IsSpan() {
		if !rf.Style.Line.SetStroke(plt) {
			return
		}
		if plt.HighlightPlotter == rf {
			pc.Stroke.Width.Dots *= 2
		}
		for _, v := range rf.Low {
			if !math.IsNaN(v) {
				line(pv(v))
			}
		}
		pc.Draw()
		return
	}

	if fill := rf.fill(); fill != nil {
		pc.Fill.Color = fill
		pc.Stroke.Color = nil
		for i, lo := range rf.Low {
			hi := rf.High[i]
			if math.IsNaN(lo) || math.IsNaN(hi) {
				continue
			}
			pl, ph := pv(lo), pv(hi)
			if rf.Vertical {
				pc.Rectangle(min(pl, ph), pb.Min.Y, math32.Abs(ph-pl), pb.Size().Y)
			} else {
				pc.Rectangle(pb.Min.X, min(pl, ph), pb.Size().X, math32.Abs(ph-pl))
			}
		}
		pc.Draw()
		pc.Fill.Color = nil
	}
	if rf.Style.Line.On == plot.On && rf.Style.Line.SetStroke(plt) {
		for i, lo := range rf.Low {
			hi := rf.High[i]
			if math.IsNaN(lo) || math.IsNaN(hi) {
				continue
			}
			line(pv(lo))
			line(pv(hi))
		}
		pc.Draw()
	}
}

// UpdateRange updates the given ranges.
func (rf *Reference) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if rf.Style.RightY {
		yax = &plt.YR
	}
	vr := &yax.Range
	if rf.Vertical {
		vr = &plt.X.Range
	}
	plot.Range(rf.Low, vr)
	if rf.High != nil {
		plot.Range(rf.High, vr)
	}
	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range
	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &rf.Style.Range)
	yax.DataRange = yax.Range
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (rf *Reference) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
	if rf.IsSpan() {
		if fill := rf.fill(); fill != nil {
			pc.FillBox(math32.FromPoint(ptb.Min), math32.FromPoint(ptb.Size()), fill)
		}
		return
	}
	if rf.Style.Line.SetStroke(plt) {
		midY := 0.5 * float32(ptb.Min.Y+ptb.Max.Y)
		pc.MoveTo(float32(ptb.Min.X), midY)
		pc.LineTo(float32(ptb.Max.X), midY)
		pc.Draw()
	}
}
