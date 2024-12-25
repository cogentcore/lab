// Copyright (c) 2024, Cogent Lab. All rights reserved.
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
	fc.SetStrings(pnms...)
}
