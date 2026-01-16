// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/base/tolassert"
	"cogentcore.org/lab/plot"
	_ "cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

var clusters = `
0: 
	9.181170003996987: 
		5.534356399283666: 
			4.859933131085473: 
				3.4641016151377544: Mark_sad Mark_happy 
				3.4641016151377544: Zane_sad Zane_happy 
			3.4641016151377544: Alberto_sad Alberto_happy 
		5.111664626761644: 
			4.640135790634417: 
				4: Lisa_sad Lisa_happy 
				3.4641016151377544: Betty_sad Betty_happy 
			3.605551275463989: Wendy_sad Wendy_happy `

func TestClust(t *testing.T) {
	dt := table.New()
	err := dt.OpenCSV("testdata/faces.dat", tensor.Tab)
	assert.NoError(t, err)
	in := dt.Column("Input")
	out := metric.Matrix(metric.L2Norm, in)

	cl := Cluster(Avg, out, dt.Column("Name"))

	var dists []float64

	var gather func(n *Node)
	gather = func(n *Node) {
		dists = append(dists, n.Dist)
		for _, kn := range n.Kids {
			gather(kn)
		}
	}
	gather(cl)

	exdists := []float64{0, 9.181170119179619, 5.534356355667114, 4.859933137893677, 3.464101552963257, 0, 0, 3.464101552963257, 0, 0, 3.464101552963257, 0, 0, 5.111664593219757, 4.640135824680328, 4, 0, 0, 3.464101552963257, 0, 0, 3.605551242828369, 0, 0}

	tolassert.EqualTolSlice(t, exdists, dists, 1.0e-7)
}

func TestTableCluster(t *testing.T) {
	dt := table.New()
	err := dt.OpenCSV("testdata/faces.dat", tensor.Tab)
	assert.NoError(t, err)

	pt := table.New()
	PlotFromTableToTable(pt, dt, metric.MetricL2Norm, Min, "Input", "Name")
	plt, err := plot.NewTablePlot(pt)
	assert.NoError(t, err)
	fnm := "table_cluster.png"
	imagex.Assert(t, plt.RenderImage(), fnm)
}
