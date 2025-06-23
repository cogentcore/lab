// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"cogentcore.org/core/cli"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/matrix"
	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
	"cogentcore.org/lab/yaegilab/labsymbols"
)

// important: must be run from an interactive terminal.
// Will quit immediately if not!
func main() {
	opts := cli.DefaultOptions("iris", "interactive data analysis.")
	cfg := &interpreter.Config{}
	cfg.InteractiveFunc = Interactive
	cli.Run(opts, cfg, interpreter.Run, interpreter.Build)
}

func Interactive(c *interpreter.Config, in *interpreter.Interpreter) error {
	dir := tensorfs.Mkdir("Iris")
	b, bs := lab.NewBasicWindow(tensorfs.CurRoot, "Iris", in)
	in.Interp.Use(labsymbols.Symbols)
	in.Config()
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		// tb.Maker(tbv.MakeToolbar)
		tb.Maker(func(p *tree.Plan) {
			tree.Add(p, func(w *core.Button) {
				w.SetText("README").SetIcon(icons.FileMarkdown).
					SetTooltip("open README help file").OnClick(func(e events.Event) {
					core.TheApp.OpenURL("https://github.com/cogentcore/lab/blob/main/examples/pca/README.md")
				})
			})
		})
	})
	b.OnShow(func(e events.Event) {
		go func() {
			if c.Expr != "" {
				in.Eval(c.Expr)
			}
			AnalyzeIris(dir, bs)
			in.Interactive()
		}()
	})

	b.RunWindow()
	core.Wait()
	return nil
}

func AnalyzeIris(dir *tensorfs.Node, bs *lab.Basic) {
	dt := table.New("iris")
	err := dt.OpenCSV("iris.data", tensor.Comma)
	if err != nil {
		fmt.Println(err)
		return
	}

	ddir := dir.Dir("Data")
	tensorfs.DirFromTable(ddir, dt)

	ped := bs.Tabs.PlotTable("Iris", dt)
	_ = ped

	cdt := table.New()
	cdt.AddFloat64Column("data", 4)
	cdt.AddStringColumn("class")
	err = cdt.OpenCSV("iris_nohead.data", tensor.Comma)
	if err != nil {
		fmt.Println(err)
		return
	}
	data := cdt.Column("data")
	covar := tensor.NewFloat64()
	err = metric.CovarianceMatrixOut(metric.Correlation, data, covar)
	cvg := bs.Tabs.TensorGrid("Covar", covar)
	_ = cvg

	vecs, _ := matrix.EigSym(covar)

	pcdir := dir.Dir("PCA")
	tensorfs.SetTensor(pcdir, tensor.Reslice(vecs, 3, tensor.FullAxis), "pc0")
	tensorfs.SetTensor(pcdir, tensor.Reslice(vecs, 2, tensor.FullAxis), "pc1")

	colidx := tensor.NewFloat64Scalar(3) // strongest at end
	prjn0 := tensor.NewFloat64()
	matrix.ProjectOnMatrixColumnOut(vecs, data, colidx, prjn0)

	pjdir := dir.Dir("Prjn")
	tensorfs.SetTensor(pjdir, prjn0, "pc0")
	colidx = tensor.NewFloat64Scalar(2)
	prjn1 := tensor.NewFloat64()
	matrix.ProjectOnMatrixColumnOut(vecs, data, colidx, prjn1)
	tensorfs.SetTensor(pjdir, prjn1, "pc1")

	tensorfs.SetTensor(pjdir, dt.Column("Name"), "name")
}
