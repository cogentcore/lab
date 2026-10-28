// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotcore

import (
	"cogentcore.org/lab/stats/cluster"
	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/table"
)

// ClusterPlot creates a cluster plot in given [Editor] plot
// using data from given data table, in column dataColumn,
// and labels from labelColumn, with given distance metric
// and cluster metric functions. It is in this package, not in
// [cluster], so that cluster does not depend on the GUI.
func ClusterPlot(plt *Editor, dt *table.Table, distMetric metric.Metrics, clustMetric cluster.Metrics, dataColumn, labelColumn string) {
	pt := table.New()
	cluster.PlotFromTableToTable(pt, dt, distMetric, clustMetric, dataColumn, labelColumn)
	plt.SetTable(pt)
}
