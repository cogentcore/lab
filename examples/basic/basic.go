// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate

import (
	"cogentcore.org/core/cli"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/tensorfs"
	"cogentcore.org/lab/yaegilab/labsymbols"
)

// important: must be run from an interactive terminal.
// Will quit immediately if not!
func main() {
	tensorfs.Mkdir("Data")
	opts := cli.DefaultOptions("basic", "basic Cogent Lab browser.")
	cfg := &interpreter.Config{}
	cfg.InteractiveFunc = Interactive
	cli.Run(opts, cfg, interpreter.Run, interpreter.Build)
}

func Interactive(c *interpreter.Config, in *interpreter.Interpreter) error {
	b, _ := lab.NewBasicWindow(tensorfs.CurRoot, "Data", in)
	in.Interp.Use(labsymbols.Symbols)
	in.Config()
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		// tb.Maker(tbv.MakeToolbar)
		tb.Maker(func(p *tree.Plan) {
			tree.Add(p, func(w *core.Button) {
				w.SetText("README").SetIcon(icons.FileMarkdown).
					SetTooltip("open README help file").OnClick(func(e events.Event) {
					core.TheApp.OpenURL("https://github.com/cogentcore/lab/blob/main/examples/basic/README.md")
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
