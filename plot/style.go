// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/styles/units"
)

// Style contains the plot styling properties relevant across
// most plot types. These properties apply to individual plot elements
// while the Plot properties applies to the overall plot itself.
type Style struct { //types:add -setters

	//	Plot has overall plot-level properties, which can be set by any
	// plot element, and are updated first, before applying element-wise styles.
	Plot PlotStyle

	// On specifies whether to plot this item, for cases where it can be turned off.
	On DefaultOffOn

	// Range is the effective range of data to plot, where either end can be fixed.
	Range minmax.Range32 `display:"inline"`

	// Label provides an alternative label to use for axis, if set.
	Label string

	// NTicks sets the desired number of ticks for the axis, if > 0.
	NTicks int

	// Line has style properties for drawing lines.
	Line LineStyle

	// Point has style properties for drawing points.
	Point PointStyle

	// Text has style properties for rendering text.
	Text TextStyle

	// Width has various plot width properties.
	Width WidthStyle
}

// NewStyle returns a new Style object with defaults applied.
func NewStyle() *Style {
	st := &Style{}
	st.Defaults()
	return st
}

func (st *Style) Defaults() {
	st.Line.Defaults()
	st.Point.Defaults()
	st.Text.Defaults()
	st.Width.Defaults()
}

// WidthStyle contains various plot width properties relevant across
// different plot types.
type WidthStyle struct { //types:add -setters
	// Cap is the width of the caps drawn at the top of error bars.
	// The default is 10dp
	Cap units.Value

	// Offset for Bar plot is the offset added to each X axis value
	// relative to the Stride computed value (X = offset + index * Stride)
	// Defaults to 1.
	Offset float32

	// Stride for Bar plot is distance between bars. Defaults to 1.
	Stride float32

	// Width for Bar plot is the width of the bars, which should be less than
	// the Stride to prevent bar overlap.
	// Defaults to .8
	Width float32

	// Pad for Bar plot is additional space at start / end of data range,
	// to keep bars from overflowing ends. This amount is subtracted from Offset
	// and added to (len(Values)-1)*Stride -- no other accommodation for bar
	// width is provided, so that should be built into this value as well.
	// Defaults to 1.
	Pad float32
}

func (ws *WidthStyle) Defaults() {
	ws.Cap.Dp(10)
	ws.Offset = 1
	ws.Stride = 1
	ws.Width = .8
	ws.Pad = 1
}

// Stylers is a list of styling functions that set Style properties.
// These are called in the order added.
type Stylers []func(s *Style)

// Add Adds a styling function to the list.
func (st *Stylers) Add(f func(s *Style)) {
	*st = append(*st, f)
}

// Run runs the list of styling functions on given [Style] object.
func (st *Stylers) Run(s *Style) {
	for _, f := range *st {
		f(s)
	}
}

// NewStyle returns a new Style object with styling functions applied
// on top of Style defaults.
func (st *Stylers) NewStyle() *Style {
	s := NewStyle()
	st.Run(s)
	return s
}

// DefaultOffOn specifies whether to use the default value for a bool option,
// or to override the default and set Off or On.
type DefaultOffOn int32 //enums:enum

const (
	// Default means use the default value.
	Default DefaultOffOn = iota

	// Off means to override the default and turn Off.
	Off

	// On means to override the default and turn On.
	On
)
