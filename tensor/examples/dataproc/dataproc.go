// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"
	"fmt"
	"math"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/tensor/stats/agg"
	"cogentcore.org/core/tensor/stats/split"
	"cogentcore.org/core/tensor/table"
	"cogentcore.org/core/tensor/tensorview"
)

// Planets is raw data
var Planets *table.Table

// PlanetsDesc are descriptive stats of all (non-Null) data
var PlanetsDesc *table.Table

// PlanetsNNDesc are descriptive stats of planets where entire row is non-null
var PlanetsNNDesc *table.Table

// GpMethodOrbit shows the median of orbital period as a function of method
var GpMethodOrbit *table.Table

// GpMethodYear shows all stats of year described by orbit
var GpMethodYear *table.Table

// GpMethodDecade shows number of planets found in each decade by given method
var GpMethodDecade *table.Table

// GpDecade shows number of planets found in each decade
var GpDecade *table.Table

//go:embed *.csv
var csv embed.FS

// AnalyzePlanets analyzes planets.csv data following some of the examples
// given here, using pandas:
//
//	https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html
func AnalyzePlanets() {
	Planets = table.NewTable(0, "planets")
	Planets.OpenFS(csv, "planets.csv", table.Comma)

	PlanetsAll := table.NewIndexView(Planets) // full original data

	NonNull := table.NewIndexView(Planets)
	NonNull.Filter(table.FilterNull) // filter out all rows with Null values

	PlanetsDesc = agg.DescAll(PlanetsAll) // individually excludes Null values in each col, but not row-wise
	PlanetsNNDesc = agg.DescAll(NonNull)  // standard descriptive stats for row-wise non-nulls

	byMethod := split.GroupBy(PlanetsAll, []string{"method"})
	split.Agg(byMethod, "orbital_period", agg.AggMedian)
	GpMethodOrbit = byMethod.AggsToTable(table.AddAggName)

	byMethod.DeleteAggs()
	split.Desc(byMethod, "year") // full desc stats of year

	byMethod.Filter(func(idx int) bool {
		ag := byMethod.AggByColumnName("year:Std")
		return ag.Aggs[idx][0] > 0 // exclude results with 0 std
	})

	GpMethodYear = byMethod.AggsToTable(table.AddAggName)

	byMethodDecade := split.GroupByFunc(PlanetsAll, func(row int) []string {
		meth := Planets.StringValue("method", row)
		yr := Planets.Float("year", row)
		decade := math.Floor(yr/10) * 10
		return []string{meth, fmt.Sprintf("%gs", decade)}
	})
	byMethodDecade.SetLevels("method", "decade")

	split.Agg(byMethodDecade, "number", agg.AggSum)

	// uncomment this to switch to decade first, then method
	// byMethodDecade.ReorderLevels([]int{1, 0})
	// byMethodDecade.SortLevels()

	decadeOnly, _ := byMethodDecade.ExtractLevels([]int{1})
	split.Agg(decadeOnly, "number", agg.AggSum)
	GpDecade = decadeOnly.AggsToTable(table.AddAggName)

	GpMethodDecade = byMethodDecade.AggsToTable(table.AddAggName) // here to ensure that decadeOnly didn't mess up..

	// todo: need unstack -- should be specific to the splits data because we already have the cols and
	// groups etc -- the ExtractLevels method provides key starting point.

	// todo: pivot table -- neeeds unstack function.

	// todo: could have a generic unstack-like method that takes a column for the data to turn into columns
	// and another that has the data to put in the cells.
}

func main() {
	AnalyzePlanets()

	b := core.NewBody("dataproc")
	tv := core.NewTabs(b)

	nt := tv.NewTab("Planets Data")
	tbv := tensorview.NewTableView(nt).SetTable(Planets)
	b.AddAppBar(tbv.ConfigToolbar)
	b.AddAppBar(func(tb *core.Toolbar) {
		core.NewButton(tb).SetText("README").SetIcon(icons.FileMarkdown).
			SetTooltip("open README help file").OnClick(func(e events.Event) {
			core.TheApp.OpenURL("https://github.com/emer/table/blob/master/examples/dataproc/README.md")
		})
	})

	nt = tv.NewTab("Non-Null Rows Desc")
	tensorview.NewTableView(nt).SetTable(PlanetsNNDesc)
	nt = tv.NewTab("All Desc")
	tensorview.NewTableView(nt).SetTable(PlanetsDesc)
	nt = tv.NewTab("By Method Orbit")
	tensorview.NewTableView(nt).SetTable(GpMethodOrbit)
	nt = tv.NewTab("By Method Year")
	tensorview.NewTableView(nt).SetTable(GpMethodYear)
	nt = tv.NewTab("By Method Decade")
	tensorview.NewTableView(nt).SetTable(GpMethodDecade)
	nt = tv.NewTab("By Decade")
	tensorview.NewTableView(nt).SetTable(GpDecade)

	tv.SelectTabIndex(0)

	b.RunMainWindow()
}
