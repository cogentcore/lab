// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statsplot

import (
	"fmt"
	"image"
	"slices"

	"cogentcore.org/core/colors"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/stats/histogram"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// Box plots a box plot of the values of the Value column for each of the
// values of the Category column (e.g., a condition): a box from the first
// to the third quartile, with a line at the median, and whiskers out to
// the most extreme values within Whisker times the interquartile range,
// with the outliers beyond them as points.
type Box struct { //types:add -setters

	// Category is the column of the categories, one box for each value
	// (e.g., the condition of a simulation): one box for all of the values
	// if it is empty.
	Category Column

	// Value is the column of the values.
	Value Column

	// Split is an optional column whose values split the rows into
	// separate series of boxes next to each other, each in their own columns.
	Split Column

	// Whisker is the extent of the whiskers, in multiples of the
	// interquartile range from the quartiles, typically 1.5:
	// the min and max values if 0.
	Whisker float64 `default:"1.5"`

	// Outliers plots the values beyond the whiskers as points.
	Outliers bool `default:"true"`
}

func (bx *Box) Defaults() {
	bx.Whisker = 1.5
	bx.Outliers = true
}

// boxData has the data of a series of boxes or violins.
type boxData struct {
	splits []*group
	cats   []*group
	srows  [][][]int
	vcol   tensor.Values
}

// boxData returns the data of the boxes or violins of the given columns.
func newBoxData(dt *table.Table, category, value, split Column) (*boxData, error) {
	bd := &boxData{}
	var err error
	if bd.vcol, err = column(dt, value); err != nil {
		return nil, err
	}
	if bd.splits, err = splitGroups(dt, split); err != nil {
		return nil, err
	}
	if bd.cats, bd.srows, err = categories(dt, category, bd.splits); err != nil {
		return nil, err
	}
	return bd, nil
}

// boxStat has the box plot stats of one box, and its outliers.
type boxStat struct {
	// stats are the low whisker, Q1, median, Q3, and high whisker.
	stats [5]float64

	// outliers are the values beyond the whiskers.
	outliers []float64
}

// boxStats returns the box plot stats of each category of the given split,
// nil for a category without values. There are outliers if whisker > 0.
func (bd *boxData) boxStats(si int, whisker float64) []*boxStat {
	bss := make([]*boxStat, len(bd.cats))
	for ci, rows := range bd.srows[si] {
		vals := values(bd.vcol, rows)
		if len(vals) == 0 {
			continue
		}
		bs := &boxStat{}
		st := stats.BoxStats(tensor.NewFloat64FromValues(vals...), whisker)
		for i := range 5 {
			bs.stats[i] = st.Float1D(i)
		}
		if whisker > 0 {
			for _, v := range vals {
				if v < bs.stats[0] || v > bs.stats[4] {
					bs.outliers = append(bs.outliers, v)
				}
			}
		}
		bss[ci] = bs
	}
	return bss
}

// numOutliers returns the total number of outliers of the given boxes.
func numOutliers(bss []*boxStat) int {
	n := 0
	for _, bs := range bss {
		if bs != nil {
			n += len(bs.outliers)
		}
	}
	return n
}

// boxColumns adds the columns of the given box plots of the given split
// to the given builder, with the given column name prefix and box style.
func (bd *boxData) boxColumns(b *builder, si int, bss []*boxStat, prefix string, sty func(s *plot.Style)) {
	roles := []plot.Roles{plot.Low, plot.U, plot.Y, plot.V, plot.High}
	names := []string{"Low", "Q1", "Median", "Q3", "High"}
	cols := make([]*tensor.Float64, 5)
	for i, rl := range roles {
		cols[i] = b.floats(colName(prefix, names[i]), func(s *plot.Style) {
			s.Role = rl
			s.Group = prefix
			if rl != plot.Y {
				return
			}
			s.On = true
			s.Plotter = plots.BoxType
			s.Label = colName(bd.splits[si].name)
			sty(s)
		})
	}
	for ci, bs := range bss {
		if bs == nil {
			continue
		}
		for i := range 5 {
			cols[i].SetFloat1D(bs.stats[i], ci)
		}
	}
}

// outlierColumns adds the columns of the outliers of the given boxes of
// the given split, at the positions of the boxes, with the given color.
func (bd *boxData) outlierColumns(b *builder, si int, bss []*boxStat, prefix string, clr image.Image) {
	gp := colName(prefix, "Outliers")
	xc := b.floats(colName(gp, "X"), func(s *plot.Style) {
		s.Role = plot.X
		s.Group = gp
	})
	yc := b.floats(gp, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
		s.Group = gp
		s.Plotter = plots.XYType
		s.NoLegend = true
		s.Line.On = plot.Off
		s.Point.On = plot.On
		s.Point.Color = clr
		s.Point.Fill = clr
	})
	r := 0
	for ci, bs := range bss {
		if bs == nil {
			continue
		}
		for _, v := range bs.outliers {
			xc.SetFloat1D(plot.BarPosition(ci, si, len(bd.splits)), r)
			yc.SetFloat1D(v, r)
			r++
		}
	}
}

// axisLabels sets the axis labels of the plot of the given table,
// on its last column, which is the category column used by the boxes
// (the plot styles of columns that are not used are not applied).
func axisLabels(dt *table.Table, category, value Column) {
	plot.Styler(dt.Columns.Values[dt.NumColumns()-1], func(s *plot.Style) {
		s.Plot.XAxis.Label = string(category)
		s.Plot.YAxisLabel = string(value)
	})
}

// Table computes the box plots from the given source table into a Box
// directory in the given tensorfs directory, and returns a table of them
// with the plot stylers set.
func (bx *Box) Table(dir *tensorfs.Node, dt *table.Table) (*table.Table, error) {
	bd, err := newBoxData(dt, bx.Category, bx.Value, bx.Split)
	if err != nil {
		return nil, err
	}
	bss := make([][]*boxStat, len(bd.splits))
	nrows := len(bd.cats) // the outliers can be more rows than the boxes
	for si := range bd.splits {
		bss[si] = bd.boxStats(si, bx.Whisker)
		if bx.Outliers {
			nrows = max(nrows, numOutliers(bss[si]))
		}
	}
	b := newBuilder(dir, "Box", nrows)
	for si, sp := range bd.splits {
		clr := seriesColor(si)
		prefix := colName(string(bx.Value), sp.name)
		bd.boxColumns(b, si, bss[si], prefix, func(s *plot.Style) {
			s.Line.Fill = clr
		})
		if bx.Outliers && bx.Whisker > 0 {
			bd.outlierColumns(b, si, bss[si], prefix, clr)
		}
	}
	// the category labels are the last X column, used by the boxes
	b.categoryColumn(bx.Category, bd.cats)
	axisLabels(b.dt, bx.Category, bx.Value)
	return b.dt, nil
}

// Violin plots a violin plot of the values of the Value column for each of
// the values of the Category column (e.g., a condition): the distribution
// of the values as a gaussian kernel density estimate, mirrored on both
// sides of the position of each category, with a box plot inside it.
type Violin struct { //types:add -setters

	// Category is the column of the categories, one violin for each value
	// (e.g., the condition of a simulation): one violin for all of the
	// values if it is empty.
	Category Column

	// Value is the column of the values.
	Value Column

	// Split is an optional column whose values split the rows into
	// separate series of violins next to each other, each in their own columns.
	Split Column

	// Points is the number of points of the density of each violin.
	Points int `default:"50"`

	// Bandwidth is the bandwidth of the kernel density estimate:
	// the standard deviation of its gaussian kernel. If 0, it is computed
	// for each violin with Silverman's rule of thumb
	// (see [histogram.KDEBandwidth]).
	Bandwidth float64

	// Cut extends the density beyond the range of the values, in multiples
	// of the bandwidth: 0 is the range of the values (as matplotlib does),
	// and 2 shows the tails of the density (as seaborn does).
	Cut float64

	// Box plots a box plot inside each violin.
	Box bool `default:"true"`

	// Whisker is the extent of the whiskers of the box plot, in multiples
	// of the interquartile range from the quartiles: the min and max
	// values if 0.
	Whisker float64 `default:"1.5"`
}

func (vl *Violin) Defaults() {
	vl.Points = 50
	vl.Box = true
	vl.Whisker = 1.5
}

// Table computes the violin plots from the given source table into a
// Violin directory in the given tensorfs directory, and returns a table
// of them with the plot stylers set.
func (vl *Violin) Table(dir *tensorfs.Node, dt *table.Table) (*table.Table, error) {
	bd, err := newBoxData(dt, vl.Category, vl.Value, vl.Split)
	if err != nil {
		return nil, err
	}
	np := vl.Points
	if np < 2 {
		np = 50
	}
	nsp := len(bd.splits)
	ncat := len(bd.cats)
	b := newBuilder(dir, "Violin", ncat*(np+1)) // NaN rows between violins
	halfWidth := 0.45 * plot.BarSeriesWidth(nsp)
	for si, sp := range bd.splits {
		clr := seriesColor(si)
		prefix := colName(string(vl.Value), sp.name)
		gp := colName(prefix, "Violin")
		vx := b.floats(colName(gp, "X"), func(s *plot.Style) {
			s.Role = plot.X
			s.Group = gp
		})
		vy := b.floats(colName(gp, "Y"), func(s *plot.Style) {
			s.Role = plot.Y
			s.Group = gp
		})
		vd := b.floats(colName(prefix, "Density"), func(s *plot.Style) {
			s.On = true
			s.Role = plot.High
			s.Group = gp
			s.Plotter = plots.XBandType
			s.Label = colName(sp.name)
			s.Line.Color = clr
		})
		for ci, rows := range bd.srows[si] {
			vals := values(bd.vcol, rows)
			if len(vals) == 0 {
				continue
			}
			vt := tensor.NewFloat64FromValues(vals...)
			grid := histogram.KDEGrid(vt, np, vl.Bandwidth, vl.Cut)
			dens := histogram.KDE(vt, grid, vl.Bandwidth)
			dmax := slices.Max(tensor.AsFloat64(dens).Values)
			if dmax <= 0 {
				continue
			}
			pos := plot.BarPosition(ci, si, nsp)
			for i := range np {
				r := ci*(np+1) + i
				vx.SetFloat1D(pos, r)
				vy.SetFloat1D(grid.Float1D(i), r)
				vd.SetFloat1D(halfWidth*dens.Float1D(i)/dmax, r)
			}
		}
		// the box plot also provides the category labels
		bd.boxColumns(b, si, bd.boxStats(si, vl.Whisker), prefix, func(s *plot.Style) {
			s.NoLegend = true // the violins are in the legend
			if !vl.Box {      // only the labels
				s.Line.On = plot.Off
				s.Line.Fill = nil
				return
			}
			s.Width.Width = 0.15
			s.Width.Cap.Dp(0)
			s.Line.Fill = colors.Scheme.Surface
		})
	}
	b.categoryColumn(vl.Category, bd.cats)
	axisLabels(b.dt, vl.Category, vl.Value)
	if ncat == 0 {
		return nil, fmt.Errorf("statsplot.Violin: no values in column %q", vl.Value)
	}
	return b.dt, nil
}
