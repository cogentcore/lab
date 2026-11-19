// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statsplot

import (
	"fmt"
	"math"

	"cogentcore.org/core/colors"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/stats/histogram"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensorfs"
)

// Histogram plots the histogram of the values of the Value column:
// the number of values within each of a number of bins of equal width
// across the range of the values, as a filled step outline, which can
// be compared across the values of a Split column, as the bins are the
// same for all of them.
type Histogram struct { //types:add -setters

	// Value is the column of the values.
	Value Column

	// Split is an optional column whose values split the rows into
	// separate histograms (e.g., the condition of a simulation),
	// each in their own column.
	Split Column

	// Bins is the number of bins.
	Bins int `default:"20"`

	// Min and Max are the range of values for the bins:
	// the range of the values, if they are equal.
	Min, Max float64

	// Density normalizes the counts to a probability density: the
	// fraction of values in each bin, divided by its width, so that the
	// area is 1, to compare distributions with different numbers of values.
	Density bool
}

func (hs *Histogram) Defaults() {
	hs.Bins = 20
}

// Table computes the histograms from the given source table into a
// Histogram directory in the given tensorfs directory, and returns
// a table of them with the plot stylers set.
func (hs *Histogram) Table(dir *tensorfs.Node, dt *table.Table) (*table.Table, error) {
	vcol, err := column(dt, hs.Value)
	if err != nil {
		return nil, err
	}
	splits, err := splitGroups(dt, hs.Split)
	if err != nil {
		return nil, err
	}
	nb := hs.Bins
	if nb <= 0 {
		nb = 20
	}
	mn, mx := hs.Min, hs.Max
	if mn >= mx {
		all := values(vcol, tableRows(dt))
		if len(all) == 0 {
			return nil, fmt.Errorf("statsplot.Histogram: no values in column %q", hs.Value)
		}
		mn, mx = math.Inf(1), math.Inf(-1)
		for _, v := range all {
			mn, mx = min(mn, v), max(mx, v)
		}
		if mn == mx {
			mn, mx = mn-0.5, mx+0.5
		}
	}
	width := (mx - mn) / float64(nb)
	b := newBuilder(dir, "Histogram", nb)
	bc := b.floats(string(hs.Value), func(s *plot.Style) {
		s.Role = plot.X
	})
	for i := range nb {
		bc.SetFloat1D(mn+(float64(i)+0.5)*width, i)
	}
	ylabel := "Count"
	if hs.Density {
		ylabel = "Density"
	}
	var hist []float64
	for si, sp := range splits {
		clr := seriesColor(si)
		fill := colors.Uniform(colors.ApplyOpacity(colors.ToUniform(clr), plots.BandOpacity))
		hc := b.floats(colName(sp.name, ylabel), func(s *plot.Style) {
			s.On = true
			s.Role = plot.Y
			s.Plotter = plots.XYType
			s.Line.Color = clr
			s.Line.Fill = fill
			s.Line.Step = plot.MidStep
			s.Point.On = plot.Off
			s.Range.SetMin(0)
			s.Plot.YAxisLabel = ylabel
		})
		vals := values(vcol, sp.rows)
		histogram.F64(&hist, vals, nb, mn, mx)
		norm := 1.0
		if hs.Density && len(vals) > 0 {
			norm = 1 / (float64(len(vals)) * width)
		}
		for i, c := range hist {
			hc.SetFloat1D(c*norm, i)
		}
	}
	return b.dt, nil
}
