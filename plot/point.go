// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"image"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles/units"
)

// PointStyle has style properties for drawing points as different shapes.
type PointStyle struct { //types:add -setters
	// On indicates whether to plot points.
	On DefaultOffOn

	// Shape to draw.
	Shape Shapes

	// Color is the stroke color image specification.
	// Setting to nil turns stroke off. See also [PointStyle.ColorFunc].
	Color image.Image

	// Fill is the color to fill points.
	// Use nil to disable filling. See also [PointStyle.FillFunc].
	Fill image.Image

	// ColorFunc, if non-nil, is used instead of [PointStyle.Color].
	// The function returns the stroke color to use for a given point index.
	ColorFunc func(i int) image.Image

	// FillFunc, if non-nil, is used instead of [PointStyle.Fill].
	// The function returns the fill color to use for a given point index.
	FillFunc func(i int) image.Image

	// Width is the line width for point glyphs, with a default of 1 Pt (point).
	// Setting to 0 turns line off.
	Width units.Value

	// Size of shape to draw for each point.
	// Defaults to 3 Pt (point).
	Size units.Value
}

func (ps *PointStyle) Defaults() {
	ps.Shape = Circle
	ps.Color = colors.Scheme.OnSurface
	ps.Fill = colors.Scheme.OnSurface
	ps.Width.Pt(1)
	ps.Size.Pt(3)
}

// SpacedColor sets the Color to a default spaced color based on index,
// if it still has the initial OnSurface default.
func (ps *PointStyle) SpacedColor(idx int) {
	if ps.Color == colors.Scheme.OnSurface {
		ps.Color = colors.Uniform(colors.Spaced(idx))
	}
	if ps.Fill == colors.Scheme.OnSurface {
		ps.Fill = colors.Uniform(colors.Spaced(idx))
	}
}

// IsOn returns true if points are to be drawn.
// Also computes the dots sizes at this point.
func (ps *PointStyle) IsOn(pt *Plot) bool {
	uc := pt.UnitContext()
	ps.Width.ToDots(uc)
	ps.Size.ToDots(uc)
	if ps.On == Off || (ps.Color == nil && ps.Fill == nil && ps.ColorFunc == nil && ps.FillFunc == nil) || ps.Width.Dots == 0 || ps.Size.Dots == 0 {
		return false
	}
	return true
}

// SetStroke sets the stroke style in plot paint to current line style.
// returns false if either the Width = 0 or Color is nil
func (ps *PointStyle) SetStroke(pt *Plot) bool {
	if !ps.IsOn(pt) {
		return false
	}
	uc := pt.UnitContext()
	pc := pt.Painter
	pc.Stroke.Width = ps.Width
	pc.Stroke.Color = ps.Color
	pc.Stroke.ToDots(uc)
	if ps.Shape <= Pyramid {
		pc.Fill.Color = ps.Fill
	} else {
		pc.Fill.Color = nil
	}
	return true
}

// SetStrokeIndex sets the stroke and fill colors based on index-specific
// color functions if applicable ([PointStyle.ColorFunc] and
// [PointStyle.FillFunc]).
func (ps *PointStyle) SetStrokeIndex(pc *paint.Painter, i int) {
	if ps.ColorFunc != nil {
		pc.Stroke.Color = ps.ColorFunc(i)
	}
	if ps.FillFunc != nil && ps.Shape <= Pyramid {
		pc.Fill.Color = ps.FillFunc(i)
	}
}

// DrawShape draws the given shape
func (ps *PointStyle) DrawShape(pc *paint.Painter, pos math32.Vector2) {
	size := ps.Size.Dots
	if size == 0 {
		return
	}
	switch ps.Shape {
	case Ring:
		DrawRing(pc, pos, size)
	case Circle:
		DrawCircle(pc, pos, size)
	case Square:
		DrawSquare(pc, pos, size)
	case Box:
		DrawBox(pc, pos, size)
	case Triangle:
		DrawTriangle(pc, pos, size)
	case Pyramid:
		DrawPyramid(pc, pos, size)
	case Plus:
		DrawPlus(pc, pos, size)
	case Cross:
		DrawCross(pc, pos, size)
	}
}

func DrawRing(pc *paint.Painter, pos math32.Vector2, size float32) {
	pc.Circle(pos.X, pos.Y, size)
	pc.Draw()
}

func DrawCircle(pc *paint.Painter, pos math32.Vector2, size float32) {
	pc.Circle(pos.X, pos.Y, size)
	pc.Draw()
}

func DrawSquare(pc *paint.Painter, pos math32.Vector2, size float32) {
	x := size * 0.9
	pc.MoveTo(pos.X-x, pos.Y-x)
	pc.LineTo(pos.X+x, pos.Y-x)
	pc.LineTo(pos.X+x, pos.Y+x)
	pc.LineTo(pos.X-x, pos.Y+x)
	pc.Close()
	pc.Draw()
}

func DrawBox(pc *paint.Painter, pos math32.Vector2, size float32) {
	x := size * 0.9
	pc.MoveTo(pos.X-x, pos.Y-x)
	pc.LineTo(pos.X+x, pos.Y-x)
	pc.LineTo(pos.X+x, pos.Y+x)
	pc.LineTo(pos.X-x, pos.Y+x)
	pc.Close()
	pc.Draw()
}

func DrawTriangle(pc *paint.Painter, pos math32.Vector2, size float32) {
	x := size * 0.9
	pc.MoveTo(pos.X, pos.Y-x)
	pc.LineTo(pos.X-x, pos.Y+x)
	pc.LineTo(pos.X+x, pos.Y+x)
	pc.Close()
	pc.Draw()
}

func DrawPyramid(pc *paint.Painter, pos math32.Vector2, size float32) {
	x := size * 0.9
	pc.MoveTo(pos.X, pos.Y-x)
	pc.LineTo(pos.X-x, pos.Y+x)
	pc.LineTo(pos.X+x, pos.Y+x)
	pc.Close()
	pc.Draw()
}

func DrawPlus(pc *paint.Painter, pos math32.Vector2, size float32) {
	x := size * 1.05
	pc.MoveTo(pos.X-x, pos.Y)
	pc.LineTo(pos.X+x, pos.Y)
	pc.MoveTo(pos.X, pos.Y-x)
	pc.LineTo(pos.X, pos.Y+x)
	pc.Close()
	pc.Draw()
}

func DrawCross(pc *paint.Painter, pos math32.Vector2, size float32) {
	x := size * 0.9
	pc.MoveTo(pos.X-x, pos.Y-x)
	pc.LineTo(pos.X+x, pos.Y+x)
	pc.MoveTo(pos.X+x, pos.Y-x)
	pc.LineTo(pos.X-x, pos.Y+x)
	pc.Close()
	pc.Draw()
}

// Shapes has the options for how to draw points in the plot.
type Shapes int32 //enums:enum

const (
	// Circle is a solid circle
	Circle Shapes = iota

	// Box is a filled square
	Box

	// Pyramid is a filled triangle
	Pyramid

	// Plus is a plus sign
	Plus

	// Cross is a big X
	Cross

	// Ring is the outline of a circle
	Ring

	// Square is the outline of a square
	Square

	// Triangle is the outline of a triangle
	Triangle
)
