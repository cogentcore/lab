// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate -add-types -add-funcs

import (
	"io/fs"
	"os"
	"path/filepath"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/cli"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/yaegicore/coresymbols"
	"cogentcore.org/lab/goal"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensorcore"
	"cogentcore.org/lab/tensorfs"
)

// goalrun is needed for running goal commands.
var goalrun *goal.Goal

var defaultJobFormat = `Name, Type
JobID, string
Version, string
Status, string
Args, string
Message, string
Label, string
Server, string
ServerJob, string
ServerStatus, string
Submit, string
Start, string
End, string
`

// SimRun is a data browser with the files as the left panel,
// and the Tabber as the right panel.
type SimRun struct {
	core.Frame
	lab.Browser

	// Config holds all the configuration settings.
	Config Configuration

	// JobsTableView is the view of the jobs table.
	JobsTableView *tensorcore.Table

	// JobsTable is the jobs Table with one row per job.
	JobsTable *table.Table

	// ResultsTableView has the results table.
	ResultsTableView *core.Table

	// Results is the list of result records.
	Results []*Result
}

// important: must be run from an interactive terminal.
// Will quit immediately if not!
func main() {
	opts := cli.DefaultOptions("simrun", "interactive simulation running and data analysis.")
	cfg := &interpreter.Config{}
	cfg.InteractiveFunc = Interactive
	cli.Run(opts, cfg, interpreter.Run, interpreter.Build)
}

func Interactive(c *interpreter.Config, in *interpreter.Interpreter) error {
	in.Interp.Use(coresymbols.Symbols) // gui imports
	in.Config()
	b, br := NewSimRunWindow(tensorfs.CurRoot, "SimRun")
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		tb.Maker(br.MakeToolbar)
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

// NewSimRunWindow returns a new data Browser window for given
// file system (nil for os files) and data directory.
// do RunWindow on resulting [core.Body] to open the window.
func NewSimRunWindow(fsys fs.FS, dataDir string) (*core.Body, *SimRun) {
	startDir, _ := os.Getwd()
	startDir = errors.Log1(filepath.Abs(startDir))
	b := core.NewBody("SimRun: " + fsx.DirAndFile(startDir))
	br := NewSimRun(b)
	br.FS = fsys
	ddr := dataDir
	if fsys == nil {
		ddr = errors.Log1(filepath.Abs(dataDir))
	}
	br.SetDataRoot(ddr)
	br.InitSimRun()
	return b, br
}

func (br *SimRun) InitSimRun() {
	br.InitInterp()
	br.Interpreter.Interp.Use(coresymbols.Symbols) // gui imports
	ddr := br.DataRoot
	br.SetScriptsDir(filepath.Join(ddr, "simscripts"))
	lab.TheBrowser = &br.Browser
	lab.CurTabber = br.Browser.Tabs
	goalrun = br.Interpreter.Goal
	br.Interpreter.Eval("br := databrowser.TheBrowser") // grab it
	br.Config.StartDir = br.StartDir
	br.Config.DataRoot = br.DataRoot
	br.Config.Defaults()
	br.JobsTable = table.New()
	br.UpdateScripts()
}

func (br *SimRun) Init() {
	br.Frame.Init()
	br.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})
	br.OnShow(func(e events.Event) {
		br.UpdateFiles()
	})
	tree.AddChildAt(br, "splits", func(w *core.Splits) {
		br.Splits = w
		w.SetSplits(.15, .85)
		tree.AddChildAt(w, "fileframe", func(w *core.Frame) {
			w.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				s.Overflow.Set(styles.OverflowAuto)
				s.Grow.Set(1, 1)
			})
			tree.AddChildAt(w, "filetree", func(w *lab.DataTree) {
				br.Files = w
			})
		})
		tree.AddChildAt(w, "tabs", func(w *lab.Tabs) {
			br.Tabs = w
		})
	})
	br.Updater(func() {
		if br.Files != nil {
			br.Files.Tabber = br.Tabs
		}
	})
}

func (br *SimRun) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.UpdateFiles).SetText("").SetIcon(icons.Refresh).SetShortcut("Command+U")
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Jobs).SetText("Jobs").SetIcon(icons.Refresh).SetShortcut("Command+U")
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("README").SetIcon(icons.FileMarkdown).
			SetTooltip("open README help file").OnClick(func(e events.Event) {
			core.TheApp.OpenURL("https://github.com/cogentcore/core/blob/main/tensor/examples/planets/README.md")
		})
	})
}
