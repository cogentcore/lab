// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted from gonum/plot:
// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"image"
	"image/color"
	"math"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
)

// Plot is the basic type representing a plot.
// It renders into its own image.RGBA Pixels image,
// and can also save a corresponding SVG version.
type Plot struct {
	// Title of the plot
	Title Text

	// Background is the background color of the plot.
	// The default is White.
	Background color.Color

	// X and Y are the horizontal and vertical axes
	// of the plot respectively.
	X, Y Axis

	// Legend is the plot's legend.
	Legend Legend

	// plotters are drawn by calling their Plot method
	// after the axes are drawn.
	Plotters []Plotter

	// size is the target size of the image to render to
	Size image.Point

	// painter for rendering
	Paint *paint.Context

	// pixels that we render into
	Pixels *image.RGBA `copier:"-" json:"-" xml:"-" edit:"-"`

	// standard text style with default params
	StdTextStyle styles.Text
}

// Defaults sets defaults
func (pt *Plot) Defaults() {
	pt.Title.Defaults()
	pt.Title.Style.Size.Pt(16)
	pt.Title.Style.Align = styles.Center
	pt.Background = color.White
	pt.X.Defaults(math32.X)
	pt.Y.Defaults(math32.Y)
	pt.Legend.Defaults()
	pt.Size = image.Point{1280, 1024}
	pt.StdTextStyle.Defaults()
	pt.StdTextStyle.WhiteSpace = styles.WhiteSpaceNowrap
}

// New returns a new plot with some reasonable default settings.
func New() *Plot {
	pt := &Plot{}
	pt.Defaults()
	return pt
}

// Add adds a Plotters to the plot.
//
// If the plotters implements DataRanger then the
// minimum and maximum values of the X and Y
// axes are changed if necessary to fit the range of
// the data.
//
// When drawing the plot, Plotters are drawn in the
// order in which they were added to the plot.
func (pt *Plot) Add(ps ...Plotter) {
	for _, d := range ps {
		if x, ok := d.(DataRanger); ok {
			xmin, xmax, ymin, ymax := x.DataRange()
			pt.X.Min = math.Min(pt.X.Min, xmin)
			pt.X.Max = math.Max(pt.X.Max, xmax)
			pt.Y.Min = math.Min(pt.Y.Min, ymin)
			pt.Y.Max = math.Max(pt.Y.Max, ymax)
		}
	}

	pt.Plotters = append(pt.Plotters, ps...)
}

// Resize sets the size of the output image to given size.
// Does nothing if already the right size.
func (pt *Plot) Resize(sz image.Point) {
	if pt.Pixels != nil {
		ib := pt.Pixels.Bounds().Size()
		if ib == sz {
			pt.Size = sz
			return // already good
		}
	}
	pt.Pixels = image.NewRGBA(image.Rectangle{Max: sz})
	pt.Paint = paint.NewContextFromImage(pt.Pixels)
	pt.Size = sz
}

func (pt *Plot) SaveImage(filename string) error {
	return imagex.Save(pt.Pixels, filename)
}

// NominalX configures the plot to have a nominal X
// axis—an X axis with names instead of numbers.  The
// X location corresponding to each name are the integers,
// e.g., the x value 0 is centered above the first name and
// 1 is above the second name, etc.  Labels for x values
// that do not end up in range of the X axis will not have
// tick marks.
func (pt *Plot) NominalX(names ...string) {
	pt.X.TickLine.Width.Pt(0)
	pt.X.TickLength.Pt(0)
	pt.X.Line.Width.Pt(0)
	// pt.Y.Padding.Pt(pt.X.Tick.Label.Width(names[0]) / 2)
	ticks := make([]Tick, len(names))
	for i, name := range names {
		ticks[i] = Tick{float64(i), name}
	}
	pt.X.Ticker = ConstantTicks(ticks)
}

// HideX configures the X axis so that it will not be drawn.
func (pt *Plot) HideX() {
	pt.X.TickLength.Pt(0)
	pt.X.Line.Width.Pt(0)
	pt.X.Ticker = ConstantTicks([]Tick{})
}

// HideY configures the Y axis so that it will not be drawn.
func (pt *Plot) HideY() {
	pt.Y.TickLength.Pt(0)
	pt.Y.Line.Width.Pt(0)
	pt.Y.Ticker = ConstantTicks([]Tick{})
}

// HideAxes hides the X and Y axes.
func (pt *Plot) HideAxes() {
	pt.HideX()
	pt.HideY()
}

// NominalY is like NominalX, but for the Y axis.
func (pt *Plot) NominalY(names ...string) {
	pt.Y.TickLine.Width.Pt(0)
	pt.Y.TickLength.Pt(0)
	pt.Y.Line.Width.Pt(0)
	// pt.X.Padding = pt.Y.Tick.Label.Height(names[0]) / 2
	ticks := make([]Tick, len(names))
	for i, name := range names {
		ticks[i] = Tick{float64(i), name}
	}
	pt.Y.Ticker = ConstantTicks(ticks)
}