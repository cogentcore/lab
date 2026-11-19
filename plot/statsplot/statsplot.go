// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package statsplot computes the statistics of table data for common
// statistical plots: the mean with error bars or bands ([MeanError]),
// histograms ([Histogram]), box plots ([Box]), and violin plots ([Violin]).
// Each one computes its data from the columns of a source [table.Table]
// into a [tensorfs] directory, and returns a [table.Table] of those
// columns with the [plot.Style] stylers set to plot them, which can be
// plotted directly with [plot.NewTablePlot], or in a plotcore.Editor
// where the plot can be further customized.
//
// The rows can be grouped by the values of a Split column, which are
// plotted as separate series, each in its own columns of the result
// (named with the split value), so that they can be turned on and off
// separately, with a different color for each.
package statsplot

//go:generate core generate

import (
	"fmt"
	"image"
	"math"
	"slices"
	"strconv"
	"strings"

	"cogentcore.org/core/colors"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// Plotter is the interface of the types of statistical plots,
// which have their options as fields.
type Plotter interface {
	// Defaults sets the default options.
	Defaults()

	// Table computes the plot data from the given source table into
	// the given tensorfs directory, and returns a table of the result
	// columns, with the plot stylers set.
	Table(dir *tensorfs.Node, dt *table.Table) (*table.Table, error)
}

// group is a group of rows with the same value of a column.
type group struct {
	// name is the value as a string.
	name string

	// value is the value, for numbers.
	value float64

	// rows are the indexes of the rows in the raw column data.
	rows []int
}

// tableRows returns the indexes of the rows of the given table in its raw
// column data: its Indexes, or all of the rows.
func tableRows(dt *table.Table) []int {
	if dt.Indexes != nil {
		return dt.Indexes
	}
	rows := make([]int, dt.Columns.Rows)
	for i := range rows {
		rows[i] = i
	}
	return rows
}

// column returns the raw data of the given column of the table,
// with an error if it is not found.
func column(dt *table.Table, name string) (tensor.Values, error) {
	cl := dt.Columns.At(name)
	if cl == nil {
		return nil, fmt.Errorf("statsplot: column %q not found", name)
	}
	return cl, nil
}

// groupRows returns the groups of the given rows of the given column by
// its values: numbers sorted by value, and strings (categories) in the
// order in which they first appear, or one group with all of the rows
// if the column is nil. Rows with NaN numbers are left out.
func groupRows(cl tensor.Values, rows []int) []*group {
	if cl == nil {
		return []*group{{rows: rows}}
	}
	gps := map[string]*group{}
	var list []*group // in the order of first appearance
	isStr := cl.IsString()
	for _, r := range rows {
		var nm string
		var v float64
		if isStr {
			nm = cl.StringRow(r, 0)
		} else {
			v = cl.FloatRow(r, 0)
			if math.IsNaN(v) {
				continue
			}
			nm = strconv.FormatFloat(v, 'g', -1, 64)
		}
		gp := gps[nm]
		if gp == nil {
			gp = &group{name: nm, value: v}
			gps[nm] = gp
			list = append(list, gp)
		}
		gp.rows = append(gp.rows, r)
	}
	if isStr {
		return list
	}
	slices.SortFunc(list, func(a, b *group) int {
		switch {
		case a.value < b.value:
			return -1
		case a.value > b.value:
			return 1
		}
		return 0
	})
	return list
}

// splitGroups returns the groups of the rows of the table by the
// given split column, one group with all of the rows if it is "".
func splitGroups(dt *table.Table, split string) ([]*group, error) {
	var cl tensor.Values
	if split != "" {
		var err error
		if cl, err = column(dt, split); err != nil {
			return nil, err
		}
	}
	return groupRows(cl, tableRows(dt)), nil
}

// categories returns the groups of all of the rows of the table by the
// given category column (one group if ""), and for each split group,
// the rows of each category, in the same order (nil if none).
func categories(dt *table.Table, category string, splits []*group) ([]*group, [][][]int, error) {
	var cl tensor.Values
	if category != "" {
		var err error
		if cl, err = column(dt, category); err != nil {
			return nil, nil, err
		}
	}
	cats := groupRows(cl, tableRows(dt))
	catIndex := map[string]int{}
	for i, c := range cats {
		catIndex[c.name] = i
	}
	srows := make([][][]int, len(splits))
	for si, sp := range splits {
		srows[si] = make([][]int, len(cats))
		for _, c := range groupRows(cl, sp.rows) {
			srows[si][catIndex[c.name]] = c.rows
		}
	}
	return cats, srows, nil
}

// values returns the non-NaN values of the given rows of the given column.
func values(cl tensor.Values, rows []int) []float64 {
	vals := make([]float64, 0, len(rows))
	for _, r := range rows {
		v := cl.FloatRow(r, 0)
		if !math.IsNaN(v) {
			vals = append(vals, v)
		}
	}
	return vals
}

// colName returns the name of a result column from the given parts,
// leaving out empty ones.
func colName(parts ...string) string {
	var nm []string
	for _, p := range parts {
		if p != "" {
			nm = append(nm, p)
		}
	}
	return strings.Join(nm, " ")
}

// seriesColor returns the color of the given series index.
func seriesColor(i int) image.Image {
	return colors.Uniform(colors.Spaced(i))
}

var nan = math.NaN()

// builder makes the columns of a result table in a tensorfs directory.
type builder struct {
	dir  *tensorfs.Node
	dt   *table.Table
	rows int
}

// newBuilder returns a builder of a table with the given number of rows,
// with its columns in a new subdirectory of the given name in dir.
func newBuilder(dir *tensorfs.Node, name string, rows int) *builder {
	sd := dir.Dir(name)
	dt := table.New(name)
	dt.SetNumRows(rows)
	return &builder{dir: sd, dt: dt, rows: rows}
}

// floats adds a column of numbers, all NaN, with the given stylers.
func (b *builder) floats(name string, st ...func(s *plot.Style)) *tensor.Float64 {
	tsr := tensorfs.Value[float64](b.dir, name, b.rows)
	tsr.SetShapeSizes(b.rows) // if it exists already
	for i := range b.rows {
		tsr.SetFloat1D(nan, i)
	}
	plot.SetStyler(tsr, st...)
	b.dt.AddColumn(name, tsr)
	return tsr.(*tensor.Float64)
}

// strings adds a column of strings, with the given stylers.
func (b *builder) strings(name string, st ...func(s *plot.Style)) *tensor.String {
	tsr := tensorfs.Value[string](b.dir, name, b.rows)
	tsr.SetShapeSizes(b.rows)
	for i := range b.rows {
		tsr.SetString1D("", i)
	}
	plot.SetStyler(tsr, st...)
	b.dt.AddColumn(name, tsr)
	return tsr.(*tensor.String)
}

// categoryColumn adds the column of the category labels,
// as the X axis labels of bar-like plots.
func (b *builder) categoryColumn(name string, cats []*group) {
	if name == "" {
		name = "Category"
	}
	cl := b.strings(name, func(s *plot.Style) {
		s.Role = plot.X
	})
	for i, c := range cats {
		cl.SetString1D(c.name, i)
	}
}
