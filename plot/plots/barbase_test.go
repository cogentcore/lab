// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestBarErrLow(t *testing.T) {
	// asymmetric error bars: Low below and High above the top
	plt := plot.New()
	plt.Title.Text = "Bar with Low and High errors"
	bc := NewBar(plt, plot.Data{plot.Y: plot.Values{3, 5, 2}, plot.Low: plot.Values{1, 0.5, math.NaN()}, plot.High: plot.Values{0.5, 1, 0.3}})
	assert.NotNil(t, bc)
	pd, _, _ := bc.Data()
	assert.Equal(t, plot.Values{0.5, 1, 0.3}, pd[plot.High])
	assert.NotNil(t, pd[plot.Low])
	imagex.Assert(t, plt.RenderImage(), "bar-err-low")
	assert.Equal(t, 6.0, plt.Y.Range.Max)
}

func TestBarBase(t *testing.T) {
	// a waterfall: each bar starts at the running total
	steps := plot.Values{5, 3, -2, 4, -3}
	low := make(plot.Values, len(steps))
	ht := make(plot.Values, len(steps))
	tot := 0.0
	for i, s := range steps {
		low[i] = min(tot, tot+s)
		ht[i] = math.Abs(s)
		tot += s
	}
	plt := plot.New()
	plt.Title.Text = "Waterfall (floating bars)"
	bc := NewBar(plt, plot.Data{plot.Y: ht, plot.X: tensor.NewStringFromValues("a", "b", "c", "d", "e")})
	assert.NotNil(t, bc)
	bc.Base = low
	assert.Equal(t, 8.0, bc.BarHeight(1))
	assert.Equal(t, 6.0, bc.Bottom(2))
	imagex.Assert(t, plt.RenderImage(), "bar-base")
	assert.Equal(t, 0.0, plt.Y.Range.Min)
	assert.Equal(t, 10.0, plt.Y.Range.Max)

	// stacked on a floating bar, and missing bottoms
	plt = plot.New()
	plt.Title.Text = "Gantt (horizontal floating bars, stacked)"
	b1 := NewBar(plt, plot.Data{plot.Y: plot.Values{2, 3, 1}})
	b1.Base = plot.Values{0, 1, math.NaN()}
	b1.Horizontal = true
	b2 := NewBar(plt, plot.Data{plot.Y: plot.Values{1, 1, 2}})
	b2.Horizontal = true
	b2.StackOn(b1)
	assert.Equal(t, 5.0, b2.BarHeight(1))
	assert.True(t, math.IsNaN(b2.Bottom(2)))
	imagex.Assert(t, plt.RenderImage(), "bar-base-horiz")
	assert.Equal(t, 5.0, plt.X.Range.Max)
}

func TestBarRightY(t *testing.T) {
	// bars on the right Y axis are drawn with its scale
	plt := plot.New()
	NewLine(plt, plot.Values{0, 1})
	bc := NewBar(plt, plot.Values{30, 50})
	bc.Styler(func(s *plot.Style) {
		s.RightY = true
	})
	plt.RenderImage()
	assert.Equal(t, plt.PYR(50), bc.PY[1])
	assert.NotEqual(t, plt.PY(50), bc.PY[1])
}
