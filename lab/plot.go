// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plotcore"
)

// NewPlot is a simple helper function that does [plot.New] and [plotcore.NewPlot],
// only returning the [plot.Plot] for convenient use in lab plots. See [NewPlotWidget]
// for a version that also returns the [plotcore.Plot]. See also [NewPlotFrom].
func NewPlot(parent ...tree.Node) *plot.Plot {
	plt, _ := NewPlotWidget(parent...)
	return plt
}

// NewPlotWidget is a simple helper function that does [plot.New] and [plotcore.NewPlot],
// returning both the [plot.Plot] and [plotcore.Plot] for convenient use in lab plots.
// See [NewPlot] for a version that only returns the more commonly useful [plot.Plot].
func NewPlotWidget(parent ...tree.Node) (*plot.Plot, *plotcore.Plot) {
	plt := plot.New()
	pw := plotcore.NewPlot(parent...).SetPlot(plt)
	return plt, pw
}

// NewPlotFrom is a version of [NewPlot] that copies plot data from the given starting plot.
func NewPlotFrom(from *plot.Plot, parent ...tree.Node) *plot.Plot {
	plt := NewPlot(parent...)
	*plt = *from
	return plt
}
