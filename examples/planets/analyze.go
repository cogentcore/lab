// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"
	"math"

	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

//go:embed *.csv
var csv embed.FS

// AnalyzePlanets analyzes planets.csv data following some of the examples
// in pandas from:
// https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html
func AnalyzePlanets(br *lab.Basic, dir *tensorfs.Node) {
	Planets := table.New("planets")
	Planets.OpenFS(csv, "planets.csv", tensor.Comma)

	tensorfs.DirFromTable(dir.Dir("Data"), Planets)

	noNA := table.NewView(Planets).FilterNA(table.AnyNA).New()
	tensorfs.DirFromTable(dir.Dir("DataNoNA"), noNA)

	// describe the main data values:
	cols := stats.DescribeTableAll(dir, noNA)

	decade := Planets.AddFloat64Column("decade")
	year := Planets.Column("year")
	for row := range Planets.NumRows() {
		yr := year.FloatRow(row, 0)
		dec := math.Floor(yr/10) * 10
		decade.SetFloatRow(dec, row, 0)
	}

	stats.TableGroups(dir, Planets, "method", "decade")
	stats.TableGroupDescribe(dir, Planets, cols...)

	methTable := stats.GroupStatsDirTable(dir.Dir("Stats").Dir("method"))
	for i, tsr := range methTable.Columns.Values {
		plot.SetFirstStyler(tsr, func(s *plot.Style) {
			s.Plotter = "Bar"
			if i == 0 {
				s.Role = plot.X
				s.Plot.XAxis.Rotation = -45
			}
		})
	}
	br.Tabs.PlotTable("Method Stats", methTable)

	decTable := stats.GroupStatsDirTable(dir.Dir("Stats").Dir("decade"))
	for i, tsr := range decTable.Columns.Values {
		plot.SetFirstStyler(tsr, func(s *plot.Style) {
			s.Plotter = "Bar"
			if i == 0 {
				s.Role = plot.X
				s.Plot.XAxis.Rotation = -45
			}
		})
	}
	br.Tabs.PlotTable("Decade Stats", decTable)

	// byMethod := split.GroupBy(PlanetsAll, "method")
	// split.AggColumn(byMethod, "orbital_period", stats.Median)
	// GpMethodOrbit = byMethod.AggsToTable(table.AddAggName)

	// byMethod.DeleteAggs()
	// split.DescColumn(byMethod, "year") // full desc stats of year

	// byMethod.Filter(func(idx int) bool {
	// 	ag := errors.Log1(byMethod.AggByColumnName("year:Std"))
	// 	return ag.Aggs[idx][0] > 0 // exclude results with 0 std
	// })

	// GpMethodYear = byMethod.AggsToTable(table.AddAggName)

	// split.AggColumn(byMethodDecade, "number", stats.Sum)

	// uncomment this to switch to decade first, then method
	// byMethodDecade.ReorderLevels([]int{1, 0})
	// byMethodDecade.SortLevels()

	// decadeOnly := errors.Log1(byMethodDecade.ExtractLevels([]int{1}))
	// split.AggColumn(decadeOnly, "number", stats.Sum)
	// GpDecade = decadeOnly.AggsToTable(table.AddAggName)
	//
	// GpMethodDecade = byMethodDecade.AggsToTable(table.AddAggName) // here to ensure that decadeOnly didn't mess up..

	// todo: need unstack -- should be specific to the splits data because we already have the cols and
	// groups etc -- the ExtractLevels method provides key starting point.

	// todo: pivot table -- neeeds unstack function.

	// todo: could have a generic unstack-like method that takes a column for the data to turn into columns
	// and another that has the data to put in the cells.
}
