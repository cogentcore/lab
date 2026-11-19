// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package statsplot

import (
	"fmt"
	"math"
	"reflect"

	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// Errors are the measures of the error of the mean.
type Errors int32 //enums:enum

const (
	// Sem is the standard error of the mean: Std / sqrt(n).
	Sem Errors = iota

	// Std is the standard deviation of the values.
	Std

	// CI95 is the 95% confidence interval of the mean, under a normal
	// approximation: 1.96 * Sem.
	CI95

	// NoError does not compute or plot an error.
	NoError
)

// ErrorPlots are the ways to plot the error of the mean.
type ErrorPlots int32 //enums:enum

const (
	// Band plots the error as a band around the line of the mean.
	Band ErrorPlots = iota

	// Bars plots the error as error bars.
	Bars
)

// MeanError plots the mean of the values of each of the Values columns,
// for each of the values of the X column, with the error of the mean,
// e.g., the mean of the results of several runs of a simulation at each
// Epoch, with their standard error. The means are plotted as lines, with
// an error band or error bars, or as bars with error bars, which is the
// default for an X column of strings (categories).
type MeanError struct { //types:add -setters

	// X is the column of the X axis values: the mean is computed over the
	// rows with the same X value (e.g., Epoch). If it is empty, the mean
	// is computed over all of the rows, plotted as a bar.
	X Column

	// Values are the columns to compute the mean of: all of the columns of
	// floating point numbers other than X and Split, if empty.
	Values []Column

	// Split is an optional column whose values split the rows into
	// separate series (e.g., the condition of a simulation),
	// each in their own columns.
	Split Column

	// Error is the measure of the error of the mean.
	Error Errors

	// ErrorPlot is how to plot the error, for lines.
	ErrorPlot ErrorPlots

	// Bar plots the means as bars, with error bars: the default for an X
	// column of strings, as the X values are categories.
	Bar bool
}

func (me *MeanError) Defaults() {}

// meanError returns the mean and error of the given values.
func (me *MeanError) meanError(vals []float64) (mean, err float64) {
	n := float64(len(vals))
	if n == 0 {
		return math.NaN(), math.NaN()
	}
	for _, v := range vals {
		mean += v
	}
	mean /= n
	if n < 2 {
		return mean, 0
	}
	ss := 0.0
	for _, v := range vals {
		ss += (v - mean) * (v - mean)
	}
	std := math.Sqrt(ss / (n - 1))
	switch me.Error {
	case Std:
		err = std
	case Sem:
		err = std / math.Sqrt(n)
	case CI95:
		err = 1.96 * std / math.Sqrt(n)
	}
	return
}

// valueColumns returns the Values, or the default ones.
func (me *MeanError) valueColumns(dt *table.Table) []Column {
	if len(me.Values) > 0 {
		return me.Values
	}
	var vals []Column
	for ci, cl := range dt.Columns.Values {
		nm := Column(dt.Columns.Keys[ci])
		dtp := cl.DataType()
		if nm == me.X || nm == me.Split || (dtp != reflect.Float32 && dtp != reflect.Float64) {
			continue
		}
		vals = append(vals, nm)
	}
	return vals
}

// Table computes the means and errors from the given source table into
// a MeanError directory in the given tensorfs directory, and returns
// a table of them with the plot stylers set.
func (me *MeanError) Table(dir *tensorfs.Node, dt *table.Table) (*table.Table, error) {
	splits, err := splitGroups(dt, me.Split)
	if err != nil {
		return nil, err
	}
	xs, srows, err := categories(dt, me.X, splits)
	if err != nil {
		return nil, err
	}
	valNames := me.valueColumns(dt)
	if len(valNames) == 0 {
		return nil, fmt.Errorf("statsplot.MeanError: no Values columns")
	}
	isBar := me.Bar || me.X == ""
	var xcol tensor.Values
	if me.X != "" {
		xcol, _ = column(dt, me.X)
		isBar = isBar || xcol.IsString()
	}
	b := newBuilder(dir, "MeanError", len(xs))
	xname := me.X
	if xname == "" {
		xname = "All"
	}
	if isBar || xcol.IsString() {
		b.categoryColumn(xname, xs)
	} else {
		xc := b.floats(string(xname), func(s *plot.Style) {
			s.Role = plot.X
		})
		for i, x := range xs {
			xc.SetFloat1D(x.value, i)
		}
	}
	errName := me.Error.String()
	series := 0
	for _, vn := range valNames {
		vcol, err := column(dt, vn)
		if err != nil {
			return nil, err
		}
		for si, sp := range splits {
			clr := seriesColor(series)
			series++
			name := colName(string(vn), sp.name)
			mc := b.floats(name, func(s *plot.Style) {
				s.On = true
				s.Role = plot.Y
				s.Group = name
				s.Line.Color = clr
				s.Point.Color = clr
				s.Point.Fill = clr
				if isBar {
					s.Plotter = plots.BarType
					s.Line.Fill = clr
				} else {
					s.Plotter = plots.XYType
				}
			})
			var ec *tensor.Float64
			if me.Error != NoError {
				ec = b.floats(colName(name, errName), func(s *plot.Style) {
					s.Role = plot.High
					s.Group = name
					s.NoLegend = true
					s.Line.Color = clr
					if isBar { // the error bars of the Bar
						return
					}
					s.On = true
					if me.ErrorPlot == Bars {
						s.Plotter = plots.YErrorBarsType
					} else {
						s.Plotter = plots.BandType
					}
				})
			}
			for xi, rows := range srows[si] {
				if rows == nil {
					continue
				}
				m, e := me.meanError(values(vcol, rows))
				mc.SetFloat1D(m, xi)
				if ec != nil {
					ec.SetFloat1D(e, xi)
				}
			}
		}
	}
	if len(valNames) == 1 {
		plot.Styler(b.dt.Columns.Values[0], func(s *plot.Style) {
			s.Plot.YAxisLabel = string(valNames[0])
		})
	}
	return b.dt, nil
}
