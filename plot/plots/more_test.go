// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"testing"

	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestXYData tests XY data setting, defaults and errors.
func TestXYData(t *testing.T) {
	plt := plot.New()
	ln := NewPointLine(plt, tensor.NewFloat64FromValues(3, 1, 2))
	require.NotNil(t, ln)
	assert.Equal(t, plot.On, ln.Style.Line.On)
	assert.Equal(t, plot.On, ln.Style.Point.On)
	assert.Equal(t, plot.Values{0, 1, 2}, ln.X)
	d, _, _ := ln.Data()
	assert.Nil(t, d[plot.Size])
	assert.Same(t, ln, ln.Styler(func(s *plot.Style) {}))
	assert.Len(t, *ln.Stylers(), 1)

	sz := plot.Values{1, 2, 3}
	require.NoError(t, ln.SetData(plot.Data{plot.Y: plot.Values{1, 2, 3}, plot.Size: sz}))
	d, _, _ = ln.Data()
	assert.Equal(t, sz, d[plot.Size])

	assert.Error(t, ln.SetData(3))
	assert.Error(t, ln.SetData(plot.Data{plot.X: plot.Values{1}, plot.Y: plot.Values{1, 2}}))
	assert.Error(t, ln.SetData(plot.Data{plot.X: plot.Values{1}}))
	assert.Nil(t, NewXY(plt, plot.Data{plot.X: plot.Values{1}}))
	assert.Nil(t, NewLine(plt, 3))
	assert.Len(t, plt.Plotters, 1)
}

// TestXYRightYHighlight tests an XY plot on the right Y axis, along with
// the highlighting of a point, and out of range values that are marked.
func TestXYRightYHighlight(t *testing.T) {
	plt := plot.New()
	left := NewLine(plt, plot.Values{0, 10, 20, 30})
	right := NewLine(plt, plot.Data{plot.Y: plot.Values{1, 2, 3, 50}, plot.Size: plot.Values{1, math.NaN(), 2, 3}})
	right.Styler(func(s *plot.Style) {
		s.RightY = true
		s.Range.SetMax(4)
		s.Point.On = plot.On
	})
	plt.Style.OutOfRange = plot.BreakMark
	plt.HighlightPlotter = right
	plt.HighlightIndex = 1
	plt.RenderImage()
	assert.Equal(t, minmax.F64{Min: 0, Max: 30}, plt.Y.Range)
	assert.InDelta(t, 4.0, plt.YR.Range.Max, 0.1) // padded for the points
	for i, y := range right.Y[:3] {
		assert.InDelta(t, plt.PYR(y), right.PY[i], 1.0e-3)
	}
	for i, y := range left.Y {
		assert.InDelta(t, plt.PY(y), left.PY[i], 1.0e-3)
	}

	// highlight with points off
	plt.HighlightPlotter = left
	plt.HighlightIndex = 2
	plt.RenderImage()
}

// TestXYSteps tests that step lines with out of range breaks render, and
// that the pixel positions are within the plot box.
func TestXYSteps(t *testing.T) {
	for _, step := range plot.StepKindValues() {
		plt := plot.New()
		ln := NewLine(plt, plot.Data{plot.X: plot.Values{0, 1, 2, 3, 2, 4}, plot.Y: plot.Values{1, 5, math.NaN(), 2, 3, 4}})
		ln.Styler(func(s *plot.Style) {
			s.Line.Step = step
			s.Range.SetMax(4)
		})
		plt.Style.OutOfRange = plot.Break
		plt.Style.XAxis.Range.SetMax(3.5)
		plt.RenderImage()
		assert.Equal(t, 3.5, plt.X.Range.Max, step.String())
		for i, px := range ln.PX {
			if ln.X[i] <= 3.5 {
				assert.GreaterOrEqual(t, px, plt.PlotBox.Min.X-1e-3)
				assert.LessOrEqual(t, px, plt.PlotBox.Max.X+1e-3)
			}
		}
	}
}

// TestBarData tests bar data setting, stacking and errors.
func TestBarData(t *testing.T) {
	plt := plot.New()
	b1 := NewBar(plt, plot.Data{plot.Y: plot.Values{1, 2, 3}, plot.High: plot.Values{0.5, 0.5, 0.5}})
	require.NotNil(t, b1)
	b2 := NewBar(plt, plot.Values{1, 1, math.NaN()})
	require.NotNil(t, b2)
	b1.Style.Width.Offset = 2
	b2.StackOn(b1)
	assert.Equal(t, 2.0, b2.Style.Width.Offset)
	assert.Equal(t, 2.0, b2.BarHeight(0))
	assert.Equal(t, 3.0, b2.BarHeight(1))
	assert.Equal(t, 0.0, b2.BarHeight(5)) // out of range
	var nb *Bar
	assert.Equal(t, 0.0, nb.BarHeight(0))
	assert.Same(t, b1, b1.Styler(func(s *plot.Style) {}))

	d, _, _ := b1.Data()
	assert.Equal(t, plot.Values{0.5, 0.5, 0.5}, d[plot.High])
	d, _, _ = b2.Data()
	assert.Nil(t, d[plot.High])

	assert.Error(t, b1.SetData(3))
	assert.Error(t, b1.SetData(plot.Data{plot.Y: plot.Values{1}, plot.High: plot.Values{1, 2}}))
	assert.Error(t, b1.SetData(plot.Data{plot.X: plot.Values{1}}))
	assert.Nil(t, NewBar(plt, 3))
}

// TestBarRanges tests the axis ranges for vertical, horizontal,
// stacked, labeled and right-axis bar plots.
func TestBarRanges(t *testing.T) {
	plt := plot.New()
	b1 := NewBar(plt, plot.Data{plot.Y: plot.Values{1, 2, 3}, plot.High: plot.Values{0.5, 0.5, 0.5}})
	b2 := NewBar(plt, plot.Data{plot.Y: plot.Values{1, 1, 1}, plot.X: plot.Labels{"a", "b", "c"}})
	b2.XLabels = b2.XLabels[:2] // fewer labels than bars
	b2.StackOn(b1)
	plt.RenderImage()
	assert.Equal(t, 0.0, plt.Y.Range.Min)
	assert.Equal(t, 4.0, plt.Y.Range.Max) // stacked 3 + 1
	assert.Equal(t, minmax.F64{Min: 0, Max: 4}, plt.X.Range)
	ticks := plt.X.Ticker.Ticks(0, 1, 5)
	require.Len(t, ticks, 3)
	assert.Equal(t, "a", ticks[0].Label)
	assert.Equal(t, 1.0, ticks[0].Value)
	assert.Equal(t, "2", ticks[2].Label) // no label for the third bar
	assert.Equal(t, b1.X, b2.X)
	assert.Equal(t, plot.Values{2, 3, 4}, b2.Yp)

	plt = plot.New()
	hb := NewBar(plt, plot.Values{2, -1, 5})
	hb.Horizontal = true
	plt.RenderImage()
	assert.Equal(t, minmax.F64{Min: -1, Max: 5}, plt.X.Range)
	assert.Equal(t, minmax.F64{Min: 0, Max: 4}, plt.Y.Range)

	plt = plot.New()
	rb := NewBar(plt, plot.Values{2, 7})
	rb.Styler(func(s *plot.Style) {
		s.RightY = true
		s.Range.SetMax(5)
	})
	plt.Style.OutOfRange = plot.Mark
	plt.RenderImage()
	assert.InDelta(t, 0, plt.YR.Range.Min, 0.2) // padded for the marks
	assert.InDelta(t, 5, plt.YR.Range.Max, 0.2)
	assert.True(t, plt.YR.Style.On)
}

// TestErrorBars tests Y and X error bar data, ranges and pixel positions,
// including error bars on the right Y axis.
func TestErrorBars(t *testing.T) {
	plt := plot.New()
	NewLine(plt, plot.Values{0, 100})
	eb := NewYErrorBars(plt, plot.Data{plot.X: plot.Values{0, 1}, plot.Y: plot.Values{1, 2}, plot.High: plot.Values{0.5, 1}})
	require.NotNil(t, eb)
	assert.Equal(t, eb.High, eb.Low) // low defaults to high
	assert.Same(t, eb, eb.Styler(func(s *plot.Style) { s.RightY = true }))
	d, _, _ := eb.Data()
	assert.Equal(t, plot.Values{0.5, 1}, d[plot.Low])
	plt.RenderImage()
	assert.Equal(t, minmax.F64{Min: 0, Max: 100}, plt.Y.Range)
	assert.Equal(t, minmax.F64{Min: 0.5, Max: 3}, plt.YR.Range)
	assert.InDelta(t, plt.PYR(1.5), eb.PY[0], 1.0e-3)
	assert.InDelta(t, plt.PYR(3), eb.PY[1], 1.0e-3)

	plt = plot.New()
	NewLine(plt, plot.Values{0, 100})
	xe := NewXErrorBars(plt, plot.Data{plot.X: plot.Values{1, 2}, plot.Y: plot.Values{1, 2}, plot.Low: plot.Values{0.5, 0.5}, plot.High: plot.Values{1, 2}})
	require.NotNil(t, xe)
	assert.Same(t, xe, xe.Styler(func(s *plot.Style) { s.RightY = true }))
	d, _, _ = xe.Data()
	assert.Equal(t, plot.Values{1, 2}, d[plot.High])
	plt.RenderImage()
	assert.Equal(t, minmax.F64{Min: 0, Max: 4}, plt.X.Range)
	assert.InDelta(t, plt.PYR(1), xe.PY[0], 1.0e-3)
	assert.InDelta(t, plt.PX(4), xe.PX[1], 1.0e-3)

	for _, bad := range []plot.Data{
		{plot.X: plot.Values{1}, plot.Y: plot.Values{1}},
		{plot.X: plot.Values{1}, plot.Y: plot.Values{1, 2}, plot.High: plot.Values{1}},
	} {
		assert.Nil(t, NewYErrorBars(plt, bad))
		assert.Nil(t, NewXErrorBars(plt, bad))
	}
	assert.Error(t, eb.SetData(3))
	assert.Error(t, xe.SetData(3))
}

// TestLabelsData tests label data, skipping and the right Y axis.
func TestLabelsData(t *testing.T) {
	plt := plot.New()
	NewLine(plt, plot.Values{0, 100})
	lb := NewLabels(plt, plot.Data{plot.X: plot.Values{0, 0.5, 1, math.NaN()}, plot.Y: plot.Values{1, 2, 3, 4}, plot.Label: plot.Labels{"a", "", "c", "d"}})
	require.NotNil(t, lb)
	assert.Same(t, lb, lb.Styler(func(s *plot.Style) {
		s.RightY = true
		s.LabelSkip = 1
	}))
	d, _, _ := lb.Data()
	assert.Equal(t, plot.Labels{"a", "", "c", "d"}, d[plot.Label])
	plt.RenderImage()
	assert.Equal(t, 1.0, plt.YR.Range.Min)
	assert.Greater(t, plt.YR.Range.Max, 3.0) // room for the label text
	assert.Equal(t, minmax.F64{Min: 0, Max: 100}, plt.Y.Range)
	assert.InDelta(t, plt.PYR(3), lb.PY[2], 1.0e-3)

	assert.Nil(t, NewLabels(plt, plot.Data{plot.X: plot.Values{1}, plot.Y: plot.Values{1}}))
	assert.Nil(t, NewLabels(plt, plot.Data{plot.X: plot.Values{1}, plot.Y: plot.Values{1, 2}, plot.Label: plot.Labels{"a"}}))
	assert.Error(t, lb.SetData(3))
}
