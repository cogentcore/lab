// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotcore

import (
	"slices"

	"cogentcore.org/core/core"
	"cogentcore.org/lab/plot"
	_ "cogentcore.org/lab/plot/plots"
	"golang.org/x/exp/maps"
)

func init() {
	core.AddValueType[plot.PlotterName, PlotterChooser]()
	core.AddValueType[plot.ColorMapName, core.ColorMapButton]()
}

// PlotterChooser represents a [Plottername] value with a [core.Chooser]
// for selecting a plotter.
type PlotterChooser struct {
	core.Chooser
}

func (fc *PlotterChooser) Init() {
	fc.Chooser.Init()
	pnms := maps.Keys(plot.Plotters)
	slices.Sort(pnms)
	// the items are PlotterNames, as the value, so that it is found in them;
	// the first is the blank default
	items := make([]core.ChooserItem, len(pnms)+1)
	items[0] = core.ChooserItem{Value: plot.PlotterName(""), Text: "Default",
		Tooltip: "the default: the Plotter of the plot Style for numbers (XY unless set there), Labels for strings"}
	for i, nm := range pnms {
		items[i+1].Value = plot.PlotterName(nm)
		items[i+1].Tooltip = plot.Plotters[nm].Doc
	}
	fc.SetItems(items...)
}
