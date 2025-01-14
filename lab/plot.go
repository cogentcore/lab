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
// for a version that also returns the [plotcore.Plot].
func NewPlot(parent ...tree.Node) *plot.Plot {
	plt := plot.New()
	plotcore.NewPlot(parent...).SetPlot(plt)
	return plt
}
