// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"
	"math"

	"cogentcore.org/core/cli"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/yaegicore/coresymbols"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/lab"
	_ "cogentcore.org/lab/lab/labscripts"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
	"cogentcore.org/lab/yaegilab/labsymbols"
)

//go:embed *.csv
var csv embed.FS

// AnalyzePlanets analyzes planets.csv data following some of the examples
// in pandas from:
// https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html
func AnalyzePlanets(dir *tensorfs.Node) {
	Planets := table.New("planets")
	Planets.OpenFS(csv, "planets.csv", tensor.Comma)

	vals := []string{"number", "orbital_period", "mass", "distance", "year"}
	stats.DescribeTable(dir, Planets, vals...)

	decade := Planets.AddFloat64Column("decade")
	year := Planets.Column("year")
	for row := range Planets.NumRows() {
		yr := year.FloatRow(row, 0)
		dec := math.Floor(yr/10) * 10
		decade.SetFloatRow(dec, row, 0)
	}

	stats.TableGroups(dir, Planets, "method", "decade")
	stats.TableGroupDescribe(dir, Planets, vals...)

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

// important: must be run from an interactive terminal.
// Will quit immediately if not!
func main() {
	dir := tensorfs.Mkdir("Planets")
	AnalyzePlanets(dir)

	opts := cli.DefaultOptions("planets", "interactive data analysis.")
	cfg := &interpreter.Config{}
	cfg.InteractiveFunc = Interactive
	cli.Run(opts, cfg, interpreter.Run, interpreter.Build)
}

func Interactive(c *interpreter.Config, in *interpreter.Interpreter) error {
	b, _ := lab.NewBasicWindow(tensorfs.CurRoot, "Planets")
	in.Interp.Use(coresymbols.Symbols)
	in.Interp.Use(labsymbols.Symbols)
	in.Config()
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		// tb.Maker(tbv.MakeToolbar)
		tb.Maker(func(p *tree.Plan) {
			tree.Add(p, func(w *core.Button) {
				w.SetText("README").SetIcon(icons.FileMarkdown).
					SetTooltip("open README help file").OnClick(func(e events.Event) {
					core.TheApp.OpenURL("https://github.com/cogentcore/lab/blob/main/examples/planets/README.md")
				})
			})
		})
	})
	b.OnShow(func(e events.Event) {
		go func() {
			if c.Expr != "" {
				in.Eval(c.Expr)
			}
			in.Interactive()
		}()
	})
	b.RunWindow()
	core.Wait()
	return nil
}
