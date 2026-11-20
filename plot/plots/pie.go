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
	"cogentcore.org/core/paint/ppath"
	"cogentcore.org/lab/plot"
)

// PieType is used for specifying the type name of a [Pie].
const PieType = "Pie"

func init() {
	plot.RegisterPlotter(PieType, "draws a pie chart of the Y values, as wedges of a circle with angles proportional to the values, with optional Label strings for the wedges, and optional Color data for their colors via the color map of the plot.", []plot.Roles{plot.Y}, []plot.Roles{plot.Label, plot.Color}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewPie(plt, data)
	})
}

// Pie draws a pie chart of the Y values: wedges of a circle with angles
// proportional to the values, as fractions of their total (as pie does
// in matplotlib). The wedges go clockwise from the StartAngle (at the
// top by default), each in a different color, or colored by the optional
// Color data via the [plot.ColorAxis] of the plot. Optional Label data
// has the labels of the wedges, drawn outside of them, and Percent draws
// the percentage of each wedge inside of it. Hole makes a donut chart.
// The edges of the wedges are drawn if Style.Line.On is On, in the
// Line.Color. Values that are NaN or not positive have no wedge.
// The pie is a circle centered in the plot area, which has no axes.
type Pie struct {
	// Values are the values of the wedges.
	Values plot.Values

	// Labels are the optional labels of the wedges.
	Labels plot.Labels

	// Color are the optional values for the colors of the wedges.
	Color plot.Values

	// StartAngle is the angle of the start of the first wedge, in degrees,
	// counter clockwise from the right: 90 is the top.
	StartAngle float32

	// CounterClockwise goes counter clockwise around the circle,
	// instead of clockwise.
	CounterClockwise bool

	// Hole is the radius of a hole in the middle, as a fraction of the
	// radius of the pie, for a donut chart.
	Hole float32

	// Percent draws the percentage of each wedge inside of it,
	// with the PercentFormat.
	Percent bool

	// PercentFormat is the format of the percentages, for fmt.Sprintf.
	PercentFormat string

	// Center, Radius are the center and radius of the pie, in pixels.
	Center math32.Vector2
	Radius float32

	// Style is the style for plotting.
	Style plot.Style

	stylers plot.Stylers
}

// NewPie adds a new [Pie] plotter to the given plot, for the given data:
// a [plot.Data] with the Y role and the optional Label and Color roles,
// or a [plot.Valuer] (e.g., a tensor) with the values. Styler functions
// are obtained from the Y metadata if present.
func NewPie(plt *plot.Plot, data any) *Pie {
	pe := &Pie{StartAngle: 90, PercentFormat: "%.0f%%"}
	if err := pe.SetData(data); err != nil {
		return nil
	}
	pe.Defaults()
	plt.Add(pe)
	return pe
}

func (pe *Pie) Defaults() {
	pe.Style.Defaults()
}

// SetData sets the plot data.
func (pe *Pie) SetData(d any) error {
	data, err := plot.DataOrValuer(d, plot.Y)
	if err != nil {
		return err
	}
	vals, err := plot.CopyValues(data[plot.Y])
	if err != nil {
		return err
	}
	n := len(vals)
	if n == 0 {
		return plot.ErrNoData
	}
	var lbls plot.Labels
	if data[plot.Label] != nil {
		if lbls = plot.CopyRoleLabels(data, plot.Label); len(lbls) != n {
			return fmt.Errorf("Pie: Label must have the length of Y, %d, not %d", n, len(lbls))
		}
	}
	var clr plot.Values
	if data[plot.Color] != nil {
		if clr = plot.CopyRole(data, plot.Color); len(clr) != n {
			return fmt.Errorf("Pie: Color must have the length of Y, %d, not %d", n, len(clr))
		}
	}
	pe.Values, pe.Labels, pe.Color = vals, lbls, clr
	pe.stylers = plot.GetStylersFromData(data, plot.Y)
	return nil
}

// Styler adds a style function to set style parameters.
func (pe *Pie) Styler(f func(s *plot.Style)) *Pie {
	pe.stylers.Add(f)
	return pe
}

func (pe *Pie) Stylers() *plot.Stylers { return &pe.stylers }

func (pe *Pie) ApplyStyle(ps *plot.PlotStyle, idx int) {
	ps.SetElementStyle(&pe.Style)
	pe.stylers.Run(&pe.Style)
}

func (pe *Pie) Data() (data plot.Data, pixX, pixY []float32) {
	data = plot.Data{plot.Y: pe.Values}
	if pe.Labels != nil {
		data[plot.Label] = pe.Labels
	}
	if pe.Color != nil {
		data[plot.Color] = pe.Color
	}
	return
}

// value returns the value of the given wedge: 0 if it is NaN or negative.
func (pe *Pie) value(i int) float64 {
	v := pe.Values[i]
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	return v
}

// Fractions returns the fraction of the total of each value.
func (pe *Pie) Fractions() []float64 {
	total := 0.0
	for i := range pe.Values {
		total += pe.value(i)
	}
	fr := make([]float64, len(pe.Values))
	if total == 0 {
		return fr
	}
	for i := range fr {
		fr[i] = pe.value(i) / total
	}
	return fr
}

// wedgeColor returns the color of the given wedge.
func (pe *Pie) wedgeColor(plt *plot.Plot, i int) image.Image {
	if pe.Color != nil {
		return plt.ColorAxis.Uniform(pe.Color[i])
	}
	return colors.Uniform(colors.Spaced(i))
}

// point returns the point on the circle of the given radius at the given
// angle, in radians, counter clockwise from the right.
func (pe *Pie) point(r, angle float32) math32.Vector2 {
	sin, cos := math32.Sincos(angle)
	return pe.Center.Add(math32.Vec2(r*cos, -r*sin))
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (pe *Pie) Plot(plt *plot.Plot) {
	pc := plt.Painter
	pb := plt.PlotBox
	sz := pb.Size()
	pe.Center = pb.Min.Add(sz.MulScalar(0.5))
	frac := 0.9
	if pe.Labels != nil {
		frac = 0.7 // room for the labels
	}
	pe.Radius = 0.5 * min(sz.X, sz.Y) * float32(frac)
	r, hole := pe.Radius, pe.Hole*pe.Radius
	dir := float32(-1) // clockwise
	sweep, back := ppath.Clockwise, ppath.CounterClockwise
	if pe.CounterClockwise {
		dir = 1
		sweep, back = back, sweep
	}
	edges := pe.Style.Line.On == plot.On
	a := math32.DegToRad(pe.StartAngle)
	fracs := pe.Fractions()
	wedges := make([]pieWedge, len(fracs))
	for i, f := range fracs {
		span := float32(2*math.Pi*f) * dir
		wedges[i] = pieWedge{a + 0.5*span, float32(f)}
		if f == 0 {
			continue
		}
		large := ppath.SmallArc
		if f > 0.5 {
			large = ppath.LargeArc
		}
		// arc adds an arc of the given radius from the start of the wedge
		// to its end, in two halves for the whole circle
		arc := func(rr, from, to float32, sw ppath.ArcSweep) {
			if f > 0.999999 {
				mid := pe.point(rr, 0.5*(from+to))
				pc.ArcTo(rr, rr, 0, ppath.SmallArc, sw, mid.X, mid.Y)
			}
			end := pe.point(rr, to)
			pc.ArcTo(rr, rr, 0, large, sw, end.X, end.Y)
		}
		start := pe.point(r, a)
		if hole > 0 {
			pc.MoveTo(start.X, start.Y)
			arc(r, a, a+span, sweep)
			in := pe.point(hole, a+span)
			pc.LineTo(in.X, in.Y)
			arc(hole, a+span, a, back)
		} else {
			pc.MoveTo(pe.Center.X, pe.Center.Y)
			pc.LineTo(start.X, start.Y)
			arc(r, a, a+span, sweep)
		}
		pc.Close()
		pc.Fill.Color = pe.wedgeColor(plt, i)
		pc.Stroke.Color = nil
		if edges && pe.Style.Line.SetStroke(plt) {
			pc.Fill.Color = pe.wedgeColor(plt, i)
		}
		pc.Draw()
		a += span
	}
	pc.Fill.Color = nil
	pe.drawText(plt, wedges, r, hole)
}

// pieWedge has the middle angle of a wedge, in radians, and its fraction.
type pieWedge struct{ mid, frac float32 }

// drawText draws the labels and percentages of the given wedges,
// with their middle angles, for the given radii of the pie and hole.
func (pe *Pie) drawText(plt *plot.Plot, wedges []pieWedge, r, hole float32) {
	if pe.Labels == nil && !pe.Percent {
		return
	}
	pc := plt.Painter
	text := func(s string, pos math32.Vector2, align float32) {
		tx := &plot.Text{Text: s}
		tx.Defaults()
		tx.Style = plt.Style.Axis.TickText
		tx.Config(plt)
		tsz := tx.PaintText.Bounds.Size()
		pc.DrawText(tx.PaintText, pos.Sub(math32.Vec2(align*tsz.X, 0.5*tsz.Y)))
	}
	for i, w := range wedges {
		if w.frac == 0 {
			continue
		}
		if pe.Labels != nil {
			pos := pe.point(1.08*r, w.mid)
			// left of the center aligns the end of the label, and the top
			// and bottom center it
			cos := math32.Cos(w.mid)
			align := 0.5 - 0.5*max(-1, min(1, 3*cos))
			text(pe.Labels[i], pos, align)
		}
		if pe.Percent {
			text(fmt.Sprintf(pe.PercentFormat, 100*w.frac), pe.point(0.5*(r+hole)+0.1*(r-hole), w.mid), 0.5)
		}
	}
}

// UpdateRange updates the given ranges, and turns off the axes,
// as the pie is centered in the plot area.
func (pe *Pie) UpdateRange(plt *plot.Plot) {
	plt.X.Range.FitInRange(minmax.F64{Min: -1, Max: 1})
	plt.Y.Range.FitInRange(minmax.F64{Min: -1, Max: 1})
	plt.X.Style.On = false
	plt.Y.Style.On = false
	if pe.Color != nil {
		plt.ColorAxis.AddRange(pe.Color)
	}
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (pe *Pie) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
	c := math32.FromPoint(ptb.Min).Add(math32.FromPoint(ptb.Size()).MulScalar(0.5))
	r := 0.5 * float32(min(ptb.Dx(), ptb.Dy()))
	a := float32(math.Pi / 2)
	for i, f := range []float32{0.5, 0.3, 0.2} {
		span := -2 * math32.Pi * f
		s := c.Add(math32.Vec2(r*math32.Cos(a), -r*math32.Sin(a)))
		e := c.Add(math32.Vec2(r*math32.Cos(a+span), -r*math32.Sin(a+span)))
		large := ppath.SmallArc
		if f > 0.5 {
			large = ppath.LargeArc
		}
		pc.MoveTo(c.X, c.Y)
		pc.LineTo(s.X, s.Y)
		pc.ArcTo(r, r, 0, large, ppath.Clockwise, e.X, e.Y)
		pc.Close()
		pc.Fill.Color = colors.Uniform(colors.Spaced(i))
		pc.Stroke.Color = nil
		pc.Draw()
		a += span
	}
	pc.Fill.Color = nil
}
