// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted from github.com/gonum/plot:
// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

//go:generate core generate -add-types

import (
	"image"
	"sync"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/paint/render"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/shaped"
	"cogentcore.org/core/text/text"
)

var (
	// plotShaper is a shared text shaper.
	plotShaper shaped.Shaper

	// mutex for sharing the plotShaper.
	shaperMu sync.Mutex
)

// XAxisStyle has overall plot level styling properties for the XAxis.
type XAxisStyle struct { //types:add -setters

	// Column specifies the column to use for the common X axis,
	// for [plot.NewTablePlot] [table.Table] driven plots.
	// If empty, standard Group-based role binding is used: the last column
	// within the same group with Role=X is used.
	Column string

	// Rotation is the rotation of the X Axis labels, in degrees.
	Rotation float32

	// Label is the optional label to use for the XAxis instead of the default.
	Label string

	// Range is the effective range of XAxis data to plot, where either end can be fixed.
	Range minmax.Range64 `display:"inline"`

	// Scale specifies how values are scaled along the X axis:
	// Linear, Log, Inverted
	Scale AxisScales
}

// PlotStyle has overall plot level styling properties.
// Some properties provide defaults for individual elements, which can
// then be overwritten by element-level properties.
type PlotStyle struct { //types:add -setters

	// Title is the overall title of the plot.
	Title string

	// TitleStyle is the text styling parameters for the title.
	TitleStyle TextStyle

	// Background is the background of the plot.
	// The default is [colors.Scheme.Surface].
	Background image.Image

	// Scale multiplies the plot DPI value, to change the overall scale
	// of the rendered plot.  Larger numbers produce larger scaling.
	// Typically use larger numbers when generating plots for inclusion in
	// documents or other cases where the overall plot size will be small.
	Scale float32 `default:"1,2"`

	// Legend has the styling properties for the Legend.
	Legend LegendStyle `display:"add-fields"`

	// Axis has the styling properties for the Axis associated with this Data.
	Axis AxisStyle `display:"add-fields"`

	// XAxis has plot-level properties specific to the XAxis.
	XAxis XAxisStyle `display:"add-fields"`

	// YAxisLabel is the optional label to use for the YAxis instead of the default.
	YAxisLabel string

	// LinesOn determines whether lines are plotted by default at the overall,
	// Plot level, for elements that plot lines (e.g., plots.XY).
	LinesOn DefaultOffOn

	// LineWidth sets the default line width for data plotting lines at the
	// overall Plot level.
	LineWidth units.Value

	// PointsOn determines whether points are plotted by default at the
	// overall Plot level, for elements that plot points (e.g., plots.XY).
	PointsOn DefaultOffOn

	// PointSize sets the default point size at the overall Plot level.
	PointSize units.Value

	// LabelSize sets the default label text size at the overall Plot level.
	LabelSize units.Value

	// BarWidth for Bar plot sets the default width of the bars,
	// which should be less than the Stride (1 typically) to prevent
	// bar overlap. Defaults to .8.
	BarWidth float64

	// ShowErrors can be set to have Plot configuration errors reported.
	// This is particularly important for table-driven plots (e.g., [plotcore.Editor]),
	// but it is not on by default because often there are transitional states
	// with known errors that can lead to false alarms.
	ShowErrors bool
}

func (ps *PlotStyle) Defaults() {
	ps.TitleStyle.Defaults()
	ps.TitleStyle.Size.Dp(24)
	ps.Background = colors.Scheme.Surface
	ps.Scale = 1
	ps.Legend.Defaults()
	ps.Axis.Defaults()
	ps.LineWidth.Pt(1)
	ps.PointSize.Pt(3)
	ps.LabelSize.Dp(16)
	ps.BarWidth = .8
}

// SetElementStyle sets the properties for given element's style
// based on the global default settings in this PlotStyle.
func (ps *PlotStyle) SetElementStyle(es *Style) {
	if ps.LinesOn != Default {
		es.Line.On = ps.LinesOn
	}
	if ps.PointsOn != Default {
		es.Point.On = ps.PointsOn
	}
	es.Line.Width = ps.LineWidth
	es.Point.Size = ps.PointSize
	es.Width.Width = ps.BarWidth
	es.Text.Size = ps.LabelSize
}

// PanZoom provides post-styling pan and zoom range manipulation.
type PanZoom struct {

	// XOffset adds offset to X range (pan).
	XOffset float64

	// XScale multiplies X range (zoom).
	XScale float64

	// YOffset adds offset to Y range (pan).
	YOffset float64

	// YScale multiplies Y range (zoom).
	YScale float64
}

func (pz *PanZoom) Defaults() {
	pz.XScale = 1
	pz.YScale = 1
}

// Plot is the basic type representing a plot.
// It renders into its own image.RGBA Pixels image,
// and can also save a corresponding SVG version.
type Plot struct {
	// Title of the plot
	Title Text

	// Style has the styling properties for the plot.
	// All end-user configuration should be put in here,
	// rather than modifying other fields directly on the plot.
	Style PlotStyle

	// standard text style with default options
	StandardTextStyle text.Style

	// X, Y, YR, and Z are the horizontal, vertical, right vertical, and depth axes
	// of the plot respectively. These are the actual compiled
	// state data and should not be used for styling: use Style.
	X, Y, YR, Z Axis

	// Legend is the plot's legend.
	Legend Legend

	// Plotters are drawn by calling their Plot method after the axes are drawn.
	Plotters []Plotter

	// PanZoom provides post-styling pan and zoom range factors.
	PanZoom PanZoom

	// HighlightPlotter is the Plotter to highlight. Used for mouse hovering for example.
	// It is the responsibility of the Plotter Plot function to implement highlighting.
	HighlightPlotter Plotter

	// HighlightIndex is the index of the data point to highlight, for HighlightPlotter.
	HighlightIndex int

	// TextShaper for shaping text. Can set to a shared external one,
	// or else the shared plotShaper is used under a mutex lock during Render.
	TextShaper shaped.Shaper

	// PaintBox is the bounding box for the plot within the Paint.
	// For standalone, it is the size of the image.
	PaintBox image.Rectangle

	// Current local plot bounding box in image coordinates, for computing
	// plotting coordinates.
	PlotBox math32.Box2

	// Painter is the current painter being used,
	// which is only valid during rendering, and is set by Draw function.
	// It needs to be exported for different plot types in other packages.
	Painter *paint.Painter

	// unitContext is current unit context, only valid during rendering.
	unitContext *units.Context
}

// New returns a new plot with some reasonable default settings.
func New() *Plot {
	pt := &Plot{}
	pt.Defaults()
	return pt
}

// Defaults sets defaults
func (pt *Plot) Defaults() {
	pt.SetSize(image.Point{640, 480})
	pt.Style.Defaults()
	pt.Title.Defaults()
	pt.Title.Style.Size.Dp(24)
	pt.X.Defaults(math32.X)
	pt.Y.Defaults(math32.Y)
	pt.YR.Defaults(math32.Y)
	pt.YR.RightY = true
	pt.Legend.Defaults()
	pt.PanZoom.Defaults()
	pt.StandardTextStyle.Defaults()
	pt.StandardTextStyle.WhiteSpace = text.WrapNever
}

// SetSize sets the size of the plot, typically in terms
// of actual device pixels (dots).
func (pt *Plot) SetSize(sz image.Point) {
	pt.PaintBox.Max = sz
}

// UnitContext returns the [units.Context] to use for styling.
// This includes the scaling factor.
func (pt *Plot) UnitContext() *units.Context {
	if pt.unitContext != nil {
		return pt.unitContext
	}
	uc := &units.Context{}
	*uc = pt.Painter.UnitContext
	uc.DPI *= pt.Style.Scale
	pt.unitContext = uc
	return uc
}

// applyStyle applies all the style parameters
func (pt *Plot) applyStyle() {
	hasYright := false
	// first update the global plot style settings
	var st Style
	st.Defaults()
	st.Plot = pt.Style
	for _, plt := range pt.Plotters {
		stlr := plt.Stylers()
		stlr.Run(&st)

		var pst Style
		pst.Defaults()
		stlr.Run(&pst)
		if pst.RightY {
			hasYright = true
		}
		if pst.Label != "" {
			if pst.RightY {
				pt.YR.Label.Text = pst.Label
			} else {
				pt.Y.Label.Text = pst.Label
			}
		}
	}
	pt.Style = st.Plot
	// then apply to elements
	for i, plt := range pt.Plotters {
		plt.ApplyStyle(&pt.Style, i)
	}
	pt.Title.Style = pt.Style.TitleStyle
	if pt.Style.Title != "" {
		pt.Title.Text = pt.Style.Title
	}
	pt.Legend.Style = pt.Style.Legend
	pt.X.Style = pt.Style.Axis
	pt.X.Style.Scale = pt.Style.XAxis.Scale
	if pt.Style.XAxis.Label != "" {
		pt.X.Label.Text = pt.Style.XAxis.Label
	}
	pt.X.Label.Style = pt.Style.Axis.Text
	pt.X.TickText.Style = pt.Style.Axis.TickText
	pt.X.TickText.Style.Rotation = pt.Style.XAxis.Rotation

	pt.Y.Style = pt.Style.Axis
	pt.YR.Style = pt.Style.Axis
	pt.YR.Style.On = hasYright
	pt.Y.Label.Style = pt.Style.Axis.Text
	pt.YR.Label.Style = pt.Style.Axis.Text
	pt.Y.TickText.Style = pt.Style.Axis.TickText
	pt.YR.TickText.Style = pt.Style.Axis.TickText
	pt.Y.Label.Style.Rotation = -90
	pt.Y.Style.TickText.Align = styles.End
	pt.YR.Label.Style.Rotation = 90
	pt.YR.Style.TickText.Align = styles.Start
	pt.UpdateRange()
}

// Add adds Plotter element(s) to the plot.
// When drawing the plot, Plotters are drawn in the
// order in which they were added to the plot.
func (pt *Plot) Add(ps ...Plotter) {
	pt.Plotters = append(pt.Plotters, ps...)
}

// CurBounds returns the current render bounds from Paint
func (pt *Plot) CurBounds() image.Rectangle {
	return pt.Painter.Context().Bounds.Rect.ToRect()
}

// PushBounds returns the current render bounds from Paint
func (pt *Plot) PushBounds(tb image.Rectangle) {
	pt.Painter.PushContext(nil, render.NewBoundsRect(tb, sides.Floats{}))
}

// NominalX configures the plot to have a nominal X
// axis—an X axis with names instead of numbers.  The
// X location corresponding to each name are the integers,
// e.g., the x value 0 is centered above the first name and
// 1 is above the second name, etc.  Labels for x values
// that do not end up in range of the X axis will not have
// tick marks.
func (pt *Plot) NominalX(names ...string) {
	pt.X.Style.TickLine.Width.Pt(0)
	pt.X.Style.TickLength.Pt(0)
	pt.X.Style.Line.Width.Pt(0)
	// pt.Y.Padding.Pt(pt.X.Style.Tick.Label.Width(names[0]) / 2)
	ticks := make([]Tick, len(names))
	for i, name := range names {
		ticks[i] = Tick{float64(i), name}
	}
	pt.X.Ticker = ConstantTicks(ticks)
}

// HideX configures the X axis so that it will not be drawn.
func (pt *Plot) HideX() {
	pt.X.Style.TickLength.Pt(0)
	pt.X.Style.Line.Width.Pt(0)
	pt.X.Ticker = ConstantTicks([]Tick{})
}

// HideY configures the Y axis so that it will not be drawn.
func (pt *Plot) HideY() {
	pt.Y.Style.TickLength.Pt(0)
	pt.Y.Style.Line.Width.Pt(0)
	pt.Y.Ticker = ConstantTicks([]Tick{})
}

// HideYR configures the YR axis so that it will not be drawn.
func (pt *Plot) HideYR() {
	pt.YR.Style.TickLength.Pt(0)
	pt.YR.Style.Line.Width.Pt(0)
	pt.YR.Ticker = ConstantTicks([]Tick{})
}

// HideAxes hides the X and Y axes.
func (pt *Plot) HideAxes() {
	pt.HideX()
	pt.HideY()
	pt.HideYR()
}

// NominalY is like NominalX, but for the Y axis.
func (pt *Plot) NominalY(names ...string) {
	pt.Y.Style.TickLine.Width.Pt(0)
	pt.Y.Style.TickLength.Pt(0)
	pt.Y.Style.Line.Width.Pt(0)
	// pt.X.Padding = pt.Y.Tick.Label.Height(names[0]) / 2
	ticks := make([]Tick, len(names))
	for i, name := range names {
		ticks[i] = Tick{float64(i), name}
	}
	pt.Y.Ticker = ConstantTicks(ticks)
}

// UpdateRange updates the axis range values based on current Plot values.
// This first resets the range so any fixed additional range values should
// be set after this point.
func (pt *Plot) UpdateRange() {
	pt.X.Range.SetInfinity()
	pt.Y.Range.SetInfinity()
	pt.YR.Range.SetInfinity()
	pt.Z.Range.SetInfinity()
	if pt.Style.XAxis.Range.FixMin {
		pt.X.Range.Min = pt.Style.XAxis.Range.Min
	}
	if pt.Style.XAxis.Range.FixMax {
		pt.X.Range.Max = pt.Style.XAxis.Range.Max
	}
	for _, pl := range pt.Plotters {
		pl.UpdateRange(pt, &pt.X.Range, &pt.Y.Range, &pt.YR.Range, &pt.Z.Range)
	}
	pt.X.Range.Sanitize()
	pt.Y.Range.Sanitize()
	pt.YR.Range.Sanitize()
	pt.Z.Range.Sanitize()

	pt.X.Range.Min *= pt.PanZoom.XScale
	pt.X.Range.Max *= pt.PanZoom.XScale
	pt.X.Range.Min += pt.PanZoom.XOffset
	pt.X.Range.Max += pt.PanZoom.XOffset

	pt.Y.Range.Min *= pt.PanZoom.YScale
	pt.Y.Range.Max *= pt.PanZoom.YScale
	pt.Y.Range.Min += pt.PanZoom.YOffset
	pt.Y.Range.Max += pt.PanZoom.YOffset

	pt.YR.Range.Min *= pt.PanZoom.YScale
	pt.YR.Range.Max *= pt.PanZoom.YScale
	pt.YR.Range.Min += pt.PanZoom.YOffset
	pt.YR.Range.Max += pt.PanZoom.YOffset
}

// PX returns the X-axis plotting coordinate for given raw data value
// using the current plot bounding region
func (pt *Plot) PX(v float64) float32 {
	return pt.PlotBox.ProjectX(float32(pt.X.Norm(v)))
}

// PY returns the Y-axis plotting coordinate for given raw data value
func (pt *Plot) PY(v float64) float32 {
	return pt.PlotBox.ProjectY(float32(1 - pt.Y.Norm(v)))
}

// PYR returns the Y-axis plotting coordinate for given raw data value
func (pt *Plot) PYR(v float64) float32 {
	return pt.PlotBox.ProjectY(float32(1 - pt.YR.Norm(v)))
}

// ClosestDataToPixel returns the Plotter data point closest to given pixel point,
// in the Pixels image.
func (pt *Plot) ClosestDataToPixel(px, py int) (plt Plotter, plotterIndex, pointIndex int, dist float32, pixel math32.Vector2, data Data, legend string) {
	tp := math32.Vec2(float32(px), float32(py))
	dist = float32(math32.MaxFloat32)
	for pi, pl := range pt.Plotters {
		dts, pxX, pxY := pl.Data()
		if len(pxY) != len(pxX) {
			continue
		}
		for i, ptx := range pxX {
			pty := pxY[i]
			pxy := math32.Vec2(ptx, pty)
			d := pxy.DistanceTo(tp)
			if d < dist {
				dist = d
				pixel = pxy
				plt = pl
				plotterIndex = pi
				pointIndex = i
				data = dts
				legend = pt.Legend.LegendForPlotter(pl)
			}
		}
	}
	return
}
