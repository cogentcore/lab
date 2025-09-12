// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/core"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plotcore"
	"cogentcore.org/lab/stats/cluster"
	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorcore"
)

//go:embed *.tsv
var tsv embed.FS

func main() {
	pats := table.New("TrainPats")
	metadata.SetDoc(pats, "Training patterns")
	// todo: meta data for grid size
	errors.Log(pats.OpenFS(tsv, "random_5x5_25.tsv", tensor.Tab))

	b := core.NewBody("grids")
	tv := core.NewTabs(b)
	nt, _ := tv.NewTab("Patterns")
	etv := tensorcore.NewTable(nt)
	tensorcore.AddGridStylerTo(pats, func(s *tensorcore.GridStyle) {
		s.TotalSize = 200
	})
	etv.SetTable(pats)
	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(etv.MakeToolbar)
	})

	lt, _ := tv.NewTab("Labels")
	gv := tensorcore.NewTensorGrid(lt)
	tsr := pats.Column("Input").RowTensor(0).Clone()
	tensorcore.AddGridStylerTo(tsr, func(s *tensorcore.GridStyle) {
		s.ColumnRotation = 45
	})
	gv.SetTensor(tsr)
	gv.RowLabels = []string{"Row 0", "Row 1,2", "", "Row 3", "Row 4"}
	gv.ColumnLabels = []string{"Col 0,1", "", "Col 2", "Col 3", "Col 4"}

	ct, _ := tv.NewTab("Cluster")
	ctb := core.NewToolbar(ct)
	plt := plotcore.NewEditor(ct)
	ctb.Maker(plt.MakeToolbar)

	dm := metric.Matrix(metric.L2Norm, pats.Column("Input"))
	cnd := cluster.Cluster("Min", dm, pats.Column("Name"))
	pdt := table.New()
	cluster.Plot(pdt, cnd, dm, pats.Column("Name"))
	stys := plotcore.BasicStylers()
	plot.Styler(pdt.Columns.At("X"), stys[0])
	plot.Styler(pdt.Columns.At("Y"), stys[1])
	plt.SetTable(pdt)

	b.RunMainWindow()
}
