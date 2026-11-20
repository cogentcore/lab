// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"image"
	"image/color"
	"math"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/styles/units"
)

// ColorMapName is the name of a [colormap.Map] (see [colormap.AvailableMaps]),
// which the plotcore package edits with a color map button.
type ColorMapName string

// DefaultColorMap is the default [ColorAxisStyle.ColorMap].
const DefaultColorMap = "ColdHot"

// ColorAxisStyle has the plot level style properties of the color axis,
// which maps the values of the Color data role (e.g., of XY points,
// or the cells of a Heatmap) to colors with a color map, shown in a
// color bar at the right of the plot.
type ColorAxisStyle struct { //types:add -setters

	// ColorMap is the name of the color map, ColdHot by default.
	ColorMap ColorMapName

	// Range is the range of values mapped onto the color map,
	// where either end can be fixed; otherwise it is the range of the data.
	Range minmax.Range64 `display:"inline"`

	// Scale specifies how values are scaled along the color map:
	// Linear, Log, Inverted.
	Scale AxisScales

	// Bar specifies whether to draw the color bar at the right of the plot,
	// which is drawn by default whenever any plot element uses colors.
	Bar DefaultOffOn

	// Label is the optional label of the color bar.
	Label string

	// Width is the width of the color bar, 16dp by default.
	Width units.Value
}

func (cs *ColorAxisStyle) Defaults() {
	cs.ColorMap = DefaultColorMap
	cs.Width.Dp(16)
}

// ColorAxis maps the values of the Color data role to colors, using the
// [ColorAxisStyle] of the plot. Plotters that use colors call [ColorAxis.AddRange]
// in their UpdateRange method, and [ColorAxis.Color] to get the color of a value.
// This is the "internal" data structure and should not be used for styling.
type ColorAxis struct {

	// Range has the Min, Max range of values for the axis (in raw data units).
	Range minmax.F64

	// Style has the style parameters, copied from the [PlotStyle].
	Style ColorAxisStyle

	// Scale transforms a value given in the data coordinate system
	// to the normalized coordinate system of the axis.
	Scale Normalizer

	// Map is the color map, from the Style.
	Map *colormap.Map

	// Used is set when any plotter uses colors, by [ColorAxis.AddRange].
	Used bool

	// Steps are the bounds of discrete steps of colors, in increasing order,
	// if a plotter uses them (e.g., the levels of a filled contour plot),
	// set with [ColorAxis.SetSteps]. The color bar then shows a block of the
	// color of the middle of each step, with ticks at the labeled bounds,
	// instead of a continuous range of colors.
	Steps []Tick
}

// SetSteps sets the bounds of discrete steps of colors (see [ColorAxis.Steps]),
// which plotters call in UpdateRange.
func (ca *ColorAxis) SetSteps(steps []Tick) {
	ca.Steps = steps
}

// drawConfig configures for drawing.
func (ca *ColorAxis) drawConfig() {
	ca.Scale = ca.Style.Scale.Normalizer()
	cm, ok := colormap.AvailableMaps[string(ca.Style.ColorMap)]
	if !ok {
		cm = colormap.AvailableMaps[DefaultColorMap]
	}
	ca.Map = cm
}

// AddRange marks the color axis as used, and includes the given data
// in its range, which plotters that use colors call in UpdateRange.
func (ca *ColorAxis) AddRange(data Valuer) {
	ca.Used = true
	Range(data, &ca.Range)
}

// resetRange resets the range for a new update.
func (ca *ColorAxis) resetRange() {
	ca.Used = false
	ca.Steps = nil
	ca.Range.SetInfinity()
}

// finalRange applies the fixed range of the style and makes the
// range valid, after the plotters have updated it.
func (ca *ColorAxis) finalRange() {
	if ca.Style.Range.FixMin {
		ca.Range.Min = ca.Style.Range.Min
	}
	if ca.Style.Range.FixMax {
		ca.Range.Max = ca.Style.Range.Max
	}
	ca.Range.Sanitize()
}

// Norm returns the value of x, given in the data coordinate system,
// normalized to its distance as a fraction of the range of the axis,
// clipped to the 0-1 range.
func (ca *ColorAxis) Norm(x float64) float64 {
	return math.Min(1, math.Max(0, ca.Scale.Normalize(ca.Range.Min, ca.Range.Max, x)))
}

// Color returns the color for the given value, which is transparent
// for NaN. Indexed color maps (e.g., for categories) use the value
// as the index of the color.
func (ca *ColorAxis) Color(x float64) color.RGBA {
	if math.IsNaN(x) {
		return color.RGBA{}
	}
	if ca.Map == nil {
		ca.drawConfig()
	}
	if ca.Map.Indexed {
		return ca.Map.MapIndex(int(x))
	}
	return ca.Map.Map(float32(ca.Norm(x)))
}

// Uniform returns the color for the given value as a uniform image,
// for the Painter.
func (ca *ColorAxis) Uniform(x float64) image.Image {
	return colors.Uniform(ca.Color(x))
}

// barOn returns whether to draw the color bar.
func (ca *ColorAxis) barOn() bool {
	switch ca.Style.Bar {
	case Off:
		return false
	case On:
		return true
	}
	return ca.Used
}

// colorBar has the layout info for drawing the color bar.
type colorBar struct {
	ticks []Tick

	// width is the total width of the bar, ticks, labels, and padding.
	width float32

	label Text
	tick  Text
}

// colorBarConfig returns the layout of the color bar, nil if not drawn.
func (pt *Plot) colorBarConfig() *colorBar {
	ca := &pt.ColorAxis
	if !ca.barOn() {
		return nil
	}
	uc := pt.UnitContext()
	ast := &pt.Style.Axis
	ca.Style.Width.ToDots(uc)
	ast.Padding.ToDots(uc)
	ast.TickLength.ToDots(uc)
	cb := &colorBar{}
	switch {
	case len(ca.Steps) >= 2:
		cb.ticks = ca.Steps
	case !ca.Map.Indexed:
		cb.ticks = DefaultTicks{}.Ticks(ca.Range.Min, ca.Range.Max, ast.NTicks)
	}
	cb.tick.Defaults()
	cb.tick.Style = ast.TickText
	twd := float32(0)
	for _, t := range cb.ticks {
		if t.IsMinor() {
			continue
		}
		cb.tick.Text = t.Label
		cb.tick.Config(pt)
		twd = max(twd, cb.tick.Size().X)
	}
	cb.width = 2*ast.Padding.Dots + ca.Style.Width.Dots + 0.5*ast.TickLength.Dots + twd
	if ca.Style.Label != "" {
		cb.label.Defaults()
		cb.label.Style = ast.Text
		cb.label.Style.Rotation = 90
		cb.label.Text = ca.Style.Label
		cb.label.Config(pt)
		cb.width += math32.Ceil(cb.label.Size().X + cb.label.Style.Padding.Dots)
	}
	cb.width = math32.Ceil(cb.width)
	return cb
}

// drawColorBar draws the color bar starting at the given left x position,
// with the vertical extent of the plot data box.
func (pt *Plot) drawColorBar(cb *colorBar, left float32) {
	ca := &pt.ColorAxis
	pc := pt.Painter
	ast := &pt.Style.Axis
	pb := pt.PlotBox
	x0 := left + ast.Padding.Dots
	bw := ca.Style.Width.Dots
	ht := pb.Size().Y

	// the colors, from the bottom (Min) to the top (Max)
	ypos := func(v float64) float32 {
		return pb.Max.Y - ht*float32(ca.Scale.Normalize(ca.Range.Min, ca.Range.Max, v))
	}
	n := 64
	if len(ca.Steps) >= 2 {
		n = 0 // the steps instead
		for i := range len(ca.Steps) - 1 {
			lo, hi := ca.Steps[i].Value, ca.Steps[i+1].Value
			y0, y1 := ypos(hi), ypos(lo)
			pc.FillBox(math32.Vec2(x0, y0), math32.Vec2(bw, y1-y0+0.5), ca.Uniform(0.5*(lo+hi)))
		}
	}
	if ca.Map.Indexed {
		n = len(ca.Map.Colors)
	}
	dy := ht / float32(n)
	for i := range n {
		var clr color.RGBA
		if ca.Map.Indexed {
			clr = ca.Map.MapIndex(i)
		} else {
			clr = ca.Map.Map((float32(i) + 0.5) / float32(n))
		}
		y := pb.Max.Y - float32(i+1)*dy
		pc.FillBox(math32.Vec2(x0, y), math32.Vec2(bw, dy+0.5), colors.Uniform(clr))
	}
	if ast.Line.SetStroke(pt) {
		pc.Fill.Color = nil
		pc.Rectangle(x0, pb.Min.Y, bw, ht)
		pc.Draw()
	}

	// ticks and their labels
	tx := x0 + bw
	tl := 0.5 * ast.TickLength.Dots
	for _, t := range cb.ticks {
		y := ypos(t.Value)
		if y < pb.Min.Y-0.5 || y > pb.Max.Y+0.5 {
			continue
		}
		ast.TickLine.Draw(pt, math32.Vec2(tx, y), math32.Vec2(tx+tl, y))
		if t.IsMinor() {
			continue
		}
		cb.tick.Text = t.Label
		cb.tick.Config(pt)
		sz := cb.tick.Size()
		cb.tick.Draw(pt, math32.Vec2(tx+tl+ast.Padding.Dots*0.5, y-0.5*sz.Y))
	}
	if ca.Style.Label != "" {
		pos := cb.label.PosY(ht)
		pos.Y += pb.Min.Y
		pos.X = left + cb.width
		cb.label.Draw(pt, pos)
	}
}
