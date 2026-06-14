// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"fmt"
	"image"
	"math"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
)

// LabelsType is be used for specifying the type name.
const LabelsType = "Labels"

func init() {
	plot.RegisterPlotter(LabelsType, "draws text labels at specified X, Y points.", []plot.Roles{plot.X, plot.Y, plot.Label}, []plot.Roles{}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewLabels(plt, data)
	})
}

// Labels draws text labels at specified X, Y points.
type Labels struct {
	// copies of data for this line
	X, Y   plot.Values
	Labels plot.Labels

	// PX, PY are the actual pixel plotting coordinates for each XY value.
	PX, PY []float32

	// Style is the style of the label text.
	Style plot.Style

	// plot size and number of TextStyle when styles last generated -- don't regen
	styleSize image.Point
	stylers   plot.Stylers
	ystylers  plot.Stylers
}

// NewLabels adds a new Labels to given plot for given data,
// which must specify X, Y and Label roles.
// Styler functions are obtained from the Label metadata if present.
func NewLabels(plt *plot.Plot, data plot.Data) *Labels {
	lb := &Labels{}
	err := lb.SetData(data)
	if err != nil {
		return nil
	}
	lb.Defaults()
	plt.Add(lb)
	return lb
}

// SetData sets the plot data.
func (lb *Labels) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.Y)
	if err != nil {
		return err
	}
	if err := dt.CheckLengths(); err != nil {
		return err
	}
	lb.X = plot.MustCopyRole(dt, plot.X)
	lb.Y = plot.MustCopyRole(dt, plot.Y)
	ld := dt[plot.Label]
	if ld == nil || lb.X == nil || lb.Y == nil {
		return fmt.Errorf("Label or X or Y is nil")
	}
	lb.Labels = make(plot.Labels, lb.X.Len())
	for i := range ld.Len() {
		lb.Labels[i] = ld.String1D(i)
	}

	lb.stylers = plot.GetStylersFromData(dt, plot.Label)
	lb.ystylers = plot.GetStylersFromData(dt, plot.Y)
	return nil
}

func (lb *Labels) Defaults() {
	lb.Style.Defaults()
}

// Styler adds a style function to set style parameters.
func (lb *Labels) Styler(f func(s *plot.Style)) *Labels {
	lb.stylers.Add(f)
	return lb
}

func (lb *Labels) ApplyStyle(ps *plot.PlotStyle, idx int) {
	lb.Style.Line.SpacedColor(idx)
	ps.SetElementStyle(&lb.Style)
	yst := &plot.Style{}
	lb.ystylers.Run(yst)
	lb.Style.Range = yst.Range // get range from y
	lb.stylers.Run(&lb.Style)  // can still override here
}

func (lb *Labels) Stylers() *plot.Stylers { return &lb.stylers }

func (lb *Labels) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = lb.PX
	pixY = lb.PY
	data = plot.Data{}
	data[plot.X] = lb.X
	data[plot.Y] = lb.Y
	data[plot.Label] = lb.Labels
	return
}

// Plot implements the Plotter interface, drawing labels.
func (lb *Labels) Plot(plt *plot.Plot) {
	pc := plt.Painter
	uc := &pc.UnitContext
	lb.PX = plot.PlotX(plt, lb.X)
	minX, maxX := plt.PX(plt.X.DataRange.Min), plt.PX(plt.X.DataRange.Max)
	var minY, maxY float32
	if lb.Style.RightY {
		lb.PY = plot.PlotYR(plt, lb.Y)
		// flipped due to Y flip
		minY, maxY = plt.PYR(plt.YR.DataRange.Max), plt.PYR(plt.YR.DataRange.Min)
	} else {
		lb.PY = plot.PlotY(plt, lb.Y)
		minY, maxY = plt.PY(plt.Y.DataRange.Max), plt.PY(plt.Y.DataRange.Min)
	}
	st := &lb.Style.Text
	st.Offset.ToDots(uc)
	var ltxt plot.Text
	ltxt.Defaults()
	ltxt.Style = *st
	ltxt.ToDots(uc)
	nskip := lb.Style.LabelSkip
	skip := nskip // start with label
	for i, label := range lb.Labels {
		if label == "" {
			continue
		}
		if skip != nskip {
			skip++
			continue
		}
		skip = 0
		ptx, pty := lb.PX[i], lb.PY[i]
		if math32.IsNaN(ptx) || math32.IsNaN(pty) {
			continue
		}
		if ptx < minX || ptx > maxX || pty < minY || pty > maxY {
			continue
		}
		ltxt.Text = label
		ltxt.Config(plt)
		tht := ltxt.Size().Y
		ltxt.Draw(plt, math32.Vec2(ptx+st.Offset.X.Dots, pty+st.Offset.Y.Dots-tht))
	}
}

// UpdateRange updates the given ranges.
func (lb *Labels) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if lb.Style.RightY {
		yax = &plt.YR
	}
	plot.Range(lb.X, &plt.X.Range)
	plot.Range(lb.Y, &yax.Range)

	var pxToData math32.Vector2
	bsz := plt.DataBox()
	pxToData.X = float32(plt.X.Range.Range()) / float32(bsz.X)
	pxToData.Y = float32(yax.Range.Range()) / float32(bsz.Y)
	st := &lb.Style.Text
	var ltxt plot.Text
	ltxt.Style = *st
	for i, label := range lb.Labels {
		if label == "" {
			continue
		}
		xv := lb.X[i]
		yv := lb.Y[i]
		if math.IsNaN(xv) || math.IsNaN(yv) {
			continue
		}
		ltxt.Text = label
		ltxt.Config(plt)
		tht := pxToData.Y * ltxt.Size().Y
		twd := 1.1 * pxToData.X * ltxt.Size().X
		maxx := xv + float64(pxToData.X*st.Offset.X.Dots+twd)
		maxy := yv + float64(pxToData.Y*st.Offset.Y.Dots+tht) // y is up here
		plt.X.Range.FitInRange(minmax.F64{xv, maxx})
		yax.Range.FitInRange(minmax.F64{yv, maxy})
	}

	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range

	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &lb.Style.Range)
	yax.DataRange = yax.Range
}
