// Copyright (c) 2020, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package random plots histograms of random distributions.
package main

//go:generate core generate

import (
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/base/randx"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/plotcore"
	"cogentcore.org/lab/stats/histogram"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
)

// Random is the random distribution plotter widget.
type Random struct {
	core.Frame
	Data
}

// Data contains the random distribution plotter data and options.
type Data struct { //types:add
	// random params
	Dist randx.RandParams `display:"add-fields"`

	// number of samples
	NumSamples int

	// number of bins in the histogram
	NumBins int

	// range for histogram
	Range minmax.F64

	// table for raw data
	Table *table.Table `display:"no-inline"`

	// histogram of data
	Histogram *table.Table `display:"no-inline"`

	// the plot
	plot *plotcore.PlotEditor `display:"-"`
}

// logPrec is the precision for saving float values in logs.
const logPrec = 4

func (rd *Random) Init() {
	rd.Frame.Init()

	rd.Dist.Defaults()
	rd.Dist.Dist = randx.Gaussian
	rd.Dist.Mean = 0.5
	rd.Dist.Var = 0.15
	rd.NumSamples = 1000000
	rd.NumBins = 100
	rd.Range.Set(0, 1)
	rd.Table = table.New()
	rd.Histogram = table.New()
	rd.ConfigTable(rd.Table)
	rd.Plot()

	rd.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})
	tree.AddChild(rd, func(w *core.Splits) {
		w.SetSplits(0.3, 0.7)
		tree.AddChild(w, func(w *core.Form) {
			w.SetStruct(&rd.Data)
			w.OnChange(func(e events.Event) {
				rd.Plot()
			})
		})
		tree.AddChild(w, func(w *plotcore.PlotEditor) {
			w.SetTable(rd.Histogram)
			rd.plot = w
		})
	})
}

// Plot generates the data and plots a histogram of results.
func (rd *Random) Plot() { //types:add
	dt := rd.Table

	dt.SetNumRows(rd.NumSamples)
	for vi := 0; vi < rd.NumSamples; vi++ {
		vl := rd.Dist.Gen()
		dt.Column("Value").SetFloat(float64(vl), vi)
	}

	histogram.F64Table(rd.Histogram, dt.Columns.Values[0].(*tensor.Float64).Values, rd.NumBins, rd.Range.Min, rd.Range.Max)
	if rd.plot != nil {
		rd.plot.UpdatePlot()
	}
}

func (rd *Random) ConfigTable(dt *table.Table) {
	metadata.SetName(dt, "Data")
	// metadata.SetReadOnly(dt, true)
	tensor.SetPrecision(dt, logPrec)
	val := dt.AddFloat64Column("Value")
	plot.SetStylersTo(val, func(s *plot.Style) {
		s.Role = plot.X
		s.Plotter = plots.BarType
		s.Plot.XAxis.Rotation = 45
		s.Plot.Title = "Random distribution histogram"
	})
}

func (rd *Random) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(rd.Plot).SetIcon(icons.ScatterPlot)
	})
	tree.Add(p, func(w *core.Separator) {})
	if rd.plot != nil {
		rd.plot.MakeToolbar(p)
	}
}

func main() {
	b := core.NewBody("Random numbers")
	rd := NewRandom(b)
	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(rd.MakeToolbar)
	})
	b.RunMainWindow()
}
