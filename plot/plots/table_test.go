// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"math/rand"
	"slices"
	"strconv"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/colors"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

// todo: move into statplot and test everything

func TestTable(t *testing.T) {
	rand.Seed(1)
	n := 21
	tx, ty := tensor.NewFloat64(n), tensor.NewFloat64(n)
	tl, th := tensor.NewFloat64(n), tensor.NewFloat64(n)
	ts := tensor.NewFloat64(n)
	lbls := tensor.NewString(n)
	for i := range n {
		tx.SetFloat1D(float64(i*5), i)
		ty.SetFloat1D(50.0+40*math.Sin((float64(i)/8)*math.Pi), i)
		tl.SetFloat1D(5*rand.Float64(), i)
		th.SetFloat1D(5*rand.Float64(), i)
		ts.SetFloat1D(1+5*rand.Float64(), i)
		lbls.SetString1D(strconv.Itoa(i), i)
	}
	ptyps := maps.Keys(plot.Plotters)
	slices.Sort(ptyps)
	for _, ttyp := range ptyps {
		// attach stylers to the Y axis data: that is where plotter looks for it
		genst := func(s *plot.Style) {
			s.Plot.Title = "Test " + ttyp
			s.Plot.XAxis.Label = "X Axis"
			s.Plot.YAxisLabel = "Y Axis"
			s.Plotter = plot.PlotterName(ttyp)
			s.Plot.SetLinesOn(plot.On).SetPointsOn(plot.On)
			s.Line.Color = colors.Uniform(colors.Red)
			s.Point.Color = colors.Uniform(colors.Blue)
			s.Range.SetMin(0).SetMax(100)
		}
		plot.SetStyler(ty, genst, func(s *plot.Style) {
			s.On = true
			s.Role = plot.Y
			s.Group = "Y"
		})
		// others get basic styling
		plot.SetStyler(tx, func(s *plot.Style) {
			s.Role = plot.X
			s.Group = "Y"
		})
		plot.SetStyler(tl, func(s *plot.Style) {
			s.Role = plot.Low
			s.Group = "Y"
		})
		plot.SetStyler(th, genst, func(s *plot.Style) {
			s.On = true
			s.Role = plot.High
			s.Group = "Y"
		})
		plot.SetStyler(ts, func(s *plot.Style) {
			s.Role = plot.Size
			s.Group = "Y"
		})
		plot.SetStyler(lbls, genst, func(s *plot.Style) {
			s.On = true
			s.Role = plot.Label
			s.Group = "Y"
		})
		dt := table.New("Test Table") // todo: use Name by default for plot.
		dt.AddColumn("X", tx)
		dt.AddColumn("Y", ty)
		dt.AddColumn("Low", tl)
		dt.AddColumn("High", th)
		dt.AddColumn("Size", ts)
		dt.AddColumn("Labels", lbls)

		plt, err := plot.NewTablePlot(dt)
		assert.NoError(t, err)
		fnm := "table_" + ttyp + ".png"
		imagex.Assert(t, plt.RenderImage(), fnm)
	}
}

func TestTableSplit(t *testing.T) {
	ngps := 3
	nitms := 4
	n := ngps * nitms
	lbls := tensor.NewString(n)
	tx, ty := tensor.NewFloat64(n), tensor.NewFloat64(n)
	groups := []string{"Aaa", "Bbb", "Ccc"}
	for i := range n {
		tx.SetFloat1D(float64(i*5), i)
		ty.SetFloat1D(50.0+40*math.Sin((float64(i)/8)*math.Pi), i)
		lbls.SetString1D(groups[i/nitms], i)
	}
	// attach stylers to the Y axis data: that is where plotter looks for it
	genst := func(s *plot.Style) {
		s.Plot.Title = "Test Split"
		s.Plot.XAxis.Label = "X Axis"
		s.Plot.YAxisLabel = "Y Axis"
		s.Plotter = plot.PlotterName("XY")
		s.Plot.SetLinesOn(plot.On).SetPointsOn(plot.On)
		s.Line.Color = colors.Uniform(colors.Red)
		s.Point.Color = colors.Uniform(colors.Blue)
		s.Range.SetMin(0).SetMax(100)
	}
	plot.SetStyler(ty, genst, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
		s.Group = "Y"
	})
	plot.SetStyler(tx, func(s *plot.Style) {
		s.Role = plot.X
		s.Group = "Y"
	})
	plot.SetStyler(lbls, genst, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Split
		s.Group = "Y"
	})
	dt := table.New("Test Table") // todo: use Name by default for plot.
	dt.AddColumn("X", tx)
	dt.AddColumn("Y", ty)
	dt.AddColumn("Labels", lbls)

	plt, err := plot.NewTablePlot(dt)
	assert.NoError(t, err)
	fnm := "table_split.png"
	imagex.Assert(t, plt.RenderImage(), fnm)
}
