// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

// colorData returns sine data with Color values that follow X,
// and a missing color.
func colorData() plot.Data {
	data := sinDataXY()
	n := len(data[plot.X].(plot.Values))
	cv := make(plot.Values, n)
	for i := range cv {
		cv[i] = float64(i) / 2
	}
	cv[3] = math.NaN()
	data[plot.Color] = cv
	return data
}

func TestXYColor(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Points colored by value"
	sc := NewScatter(plt, colorData())
	assert.NotNil(t, sc)
	sc.Styler(func(s *plot.Style) {
		s.Point.Size.Pt(6)
		s.Plot.ColorAxis.Label = "Color value"
	})
	pd, _, _ := sc.Data()
	assert.NotNil(t, pd[plot.Color])
	imagex.Assert(t, plt.RenderImage(), "xy-color")
	assert.True(t, plt.ColorAxis.Used)
	assert.Equal(t, 0.0, plt.ColorAxis.Range.Min)
	assert.Equal(t, 10.0, plt.ColorAxis.Range.Max)
	assert.Equal(t, 1.0, plt.ColorAxis.Norm(20))
	assert.Equal(t, uint8(0), plt.ColorAxis.Color(math.NaN()).A)
	assert.Equal(t, plt.ColorAxis.Map.Map(0.5), plt.ColorAxis.Color(5))

	// fixed range, another color map, rings, and no color bar
	plt = plot.New()
	plt.Title.Text = "Fixed color range, no color bar"
	sc = NewScatter(plt, colorData())
	sc.Styler(func(s *plot.Style) {
		s.Point.Size.Pt(6)
		s.Point.Shape = plot.Ring
		s.Plot.ColorAxis.ColorMap = "Viridis"
		s.Plot.ColorAxis.Range.SetMin(2).SetMax(4)
		s.Plot.ColorAxis.Bar = plot.Off
	})
	imagex.Assert(t, plt.RenderImage(), "xy-color-fixed")
	assert.Equal(t, 2.0, plt.ColorAxis.Range.Min)
	assert.Equal(t, "Viridis", plt.ColorAxis.Map.Name)

	// not used
	plt = plot.New()
	NewLine(plt, sinDataXY())
	plt.RenderImage()
	assert.False(t, plt.ColorAxis.Used)
}

func TestTableColor(t *testing.T) {
	data := colorData()
	tx := tensor.NewFloat64FromValues(data[plot.X].(plot.Values)...)
	ty := tensor.NewFloat64FromValues(data[plot.Y].(plot.Values)...)
	tc := tensor.NewFloat64FromValues(data[plot.Color].(plot.Values)...)
	plot.SetStyler(tx, func(s *plot.Style) {
		s.Role = plot.X
	})
	plot.SetStyler(ty, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
		s.Plot.Title = "Table Color role"
		s.Plot.PointsOn = plot.On
	})
	plot.SetStyler(tc, func(s *plot.Style) {
		s.Role = plot.Color
	})
	dt := table.New("Color")
	dt.AddColumn("X", tx)
	dt.AddColumn("Y", ty)
	dt.AddColumn("C", tc)
	plt, err := plot.NewTablePlot(dt)
	assert.NoError(t, err)
	imagex.Assert(t, plt.RenderImage(), "table-color")
	assert.True(t, plt.ColorAxis.Used)
}
