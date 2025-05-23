// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted initially from gonum/plot:
// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"math"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
)

// AxisScales are the scaling options for how values are distributed
// along an axis: Linear, Log, etc.
type AxisScales int32 //enums:enum

const (
	// Linear is a linear axis scale.
	Linear AxisScales = iota

	// Log is a Logarithmic axis scale.
	Log

	// InverseLinear is an inverted linear axis scale.
	InverseLinear

	// InverseLog is an inverted log axis scale.
	InverseLog
)

// AxisStyle has style properties for the axis.
type AxisStyle struct { //types:add -setters

	// On determines whether the axis is rendered.
	On bool

	// Text has the text style parameters for the text label.
	Text TextStyle

	// Line has styling properties for the axis line.
	Line LineStyle

	// Padding between the axis line and the data.  Having
	// non-zero padding ensures that the data is never drawn
	// on the axis, thus making it easier to see.
	Padding units.Value

	// NTicks is the desired number of ticks (actual likely
	// will be different). If < 2 then the axis will not be drawn.
	NTicks int

	// Scale specifies how values are scaled along the axis:
	// Linear, Log, Inverted
	Scale AxisScales

	// TickText has the text style for rendering tick labels,
	// and is shared for actual rendering.
	TickText TextStyle

	// TickLine has line style for drawing tick lines.
	TickLine LineStyle

	// TickLength is the length of tick lines.
	TickLength units.Value
}

func (ax *AxisStyle) Defaults() {
	ax.On = true
	ax.Line.Defaults()
	ax.Text.Defaults()
	ax.Text.Size.Dp(20)
	ax.Padding.Pt(5)
	ax.NTicks = 5
	ax.TickText.Defaults()
	ax.TickText.Size.Dp(16)
	ax.TickText.Padding.Dp(2)
	ax.TickLine.Defaults()
	ax.TickLength.Pt(8)
}

// Axis represents either a horizontal or vertical axis of a plot.
// This is the "internal" data structure and should not be used for styling.
type Axis struct {

	// Range has the Min, Max range of values for the axis (in raw data units.)
	Range minmax.F64

	// specifies which axis this is: X, Y or Z.
	Axis math32.Dims

	// For a Y axis, this puts the axis on the right (i.e., the second Y axis).
	RightY bool

	// Label for the axis.
	Label Text

	// Style has the style parameters for the Axis.
	Style AxisStyle

	// TickText is used for rendering the tick text labels.
	TickText Text

	// Ticker generates the tick marks. Any tick marks
	// returned by the Marker function that are not in
	// range of the axis are not drawn.
	Ticker Ticker

	// Scale transforms a value given in the data coordinate system
	// to the normalized coordinate system of the axis—its distance
	// along the axis as a fraction of the axis range.
	Scale Normalizer

	// AutoRescale enables an axis to automatically adapt its minimum
	// and maximum boundaries, according to its underlying Ticker.
	AutoRescale bool

	// cached list of ticks, set in size
	ticks []Tick
}

// Sets Defaults, range is (∞, ­∞), and thus any finite
// value is less than Min and greater than Max.
func (ax *Axis) Defaults(dim math32.Dims) {
	ax.Style.Defaults()
	ax.Axis = dim
	if dim == math32.Y {
		ax.Label.Style.Rotation = -90
		if ax.RightY {
			ax.Style.TickText.Align = styles.Start
		} else {
			ax.Style.TickText.Align = styles.End
		}
	}
	ax.Scale = LinearScale{}
	ax.Ticker = DefaultTicks{}
}

// drawConfig configures for drawing.
func (ax *Axis) drawConfig() {
	switch ax.Style.Scale {
	case Linear:
		ax.Scale = LinearScale{}
	case Log:
		ax.Scale = LogScale{}
	case InverseLinear:
		ax.Scale = InvertedScale{LinearScale{}}
	case InverseLog:
		ax.Scale = InvertedScale{LogScale{}}
	}
}

// SanitizeRange ensures that the range of the axis makes sense.
func (ax *Axis) SanitizeRange() {
	ax.Range.Sanitize()
	if ax.AutoRescale {
		marks := ax.Ticker.Ticks(ax.Range.Min, ax.Range.Max, ax.Style.NTicks)
		for _, t := range marks {
			ax.Range.FitValInRange(t.Value)
		}
	}
}

// Normalizer rescales values from the data coordinate system to the
// normalized coordinate system.
type Normalizer interface {
	// Normalize transforms a value x in the data coordinate system to
	// the normalized coordinate system.
	Normalize(min, max, x float64) float64
}

// LinearScale an be used as the value of an Axis.Scale function to
// set the axis to a standard linear scale.
type LinearScale struct{}

var _ Normalizer = LinearScale{}

// Normalize returns the fractional distance of x between min and max.
func (LinearScale) Normalize(min, max, x float64) float64 {
	return (x - min) / (max - min)
}

// LogScale can be used as the value of an Axis.Scale function to
// set the axis to a log scale.
type LogScale struct{}

var _ Normalizer = LogScale{}

// Normalize returns the fractional logarithmic distance of
// x between min and max.
func (LogScale) Normalize(min, max, x float64) float64 {
	if min <= 0 || max <= 0 || x <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}
	logMin := math.Log(min)
	return (math.Log(x) - logMin) / (math.Log(max) - logMin)
}

// InvertedScale can be used as the value of an Axis.Scale function to
// invert the axis using any Normalizer.
type InvertedScale struct{ Normalizer }

var _ Normalizer = InvertedScale{}

// Normalize returns a normalized [0, 1] value for the position of x.
func (is InvertedScale) Normalize(min, max, x float64) float64 {
	return is.Normalizer.Normalize(max, min, x)
}

// Norm returns the value of x, given in the data coordinate
// system, normalized to its distance as a fraction of the
// range of this axis.  For example, if x is a.Min then the return
// value is 0, and if x is a.Max then the return value is 1.
func (ax *Axis) Norm(x float64) float64 {
	return ax.Scale.Normalize(ax.Range.Min, ax.Range.Max, x)
}
