// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate -add-types -add-funcs

import (
	"os"
	"path/filepath"
	"reflect"

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
	"cogentcore.org/lab/yaegilab/gui"
	"github.com/cogentcore/yaegi/interp"
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

// SimRun manages the running and data analysis of results from simulations
// that are run on remote server(s), within a Cogent Lab browser environment,
// with the files as the left panel, and the Tabber as the right panel.
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

	// ResultsList is the list of result records.
	ResultsList []*Result
}

// important: must be run from an interactive terminal.
// Will quit immediately if not!
func main() {
	opts := cli.DefaultOptions("simrun", "interactive simulation running and data analysis.")
	cfg := &interpreter.Config{}
	cfg.InteractiveFunc = Interactive
	cli.Run(opts, cfg, interpreter.Run, interpreter.Build)
}

// Interactive is the cli function that gets called by default at gui startup.
func Interactive(c *interpreter.Config, in *interpreter.Interpreter) error {
	b, _ := NewSimRunWindow(in)
	b.OnShow(func(e events.Event) {
		// go func() {
		// 	if c.Expr != "" {
		// 		in.Eval(c.Expr)
		// 	}
		// 	in.Interactive()
		// }()
	})
	b.RunWindow()
	core.Wait()
	return nil
}

// NewSimRunWindow returns a new SimRun window using given interpreter.
// do RunWindow on resulting [core.Body] to open the window.
func NewSimRunWindow(in *interpreter.Interpreter) (*core.Body, *SimRun) {
	startDir, _ := os.Getwd()
	startDir = errors.Log1(filepath.Abs(startDir))
	b := core.NewBody("SimRun: " + fsx.DirAndFile(startDir))
	br := NewSimRun(b)
	br.Interpreter = in
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		br.Toolbar = tb
		tb.Maker(br.MakeToolbar)
	})
	br.InitSimRun(startDir)
	return b, br
}

// InitSimRun initializes the simrun configuration and data
// for given starting directory, which should be the main github
// current working directory for the simulation being run.
// All the simrun data is contained within a "simdata" directory
// under the startDir: this dir is typically a symbolic link
// to a common collection of such simdata directories for all
// the different simulations being run.
// The goal Interpreter is typically already set by this point
// but will be created if not.
func (br *SimRun) InitSimRun(startDir string) {
	br.StartDir = startDir
	ddr := errors.Log1(filepath.Abs("simdata"))
	br.SetDataRoot(ddr)
	if br.Interpreter == nil {
		br.InitInterp()
	}
	in := br.Interpreter
	in.Interp.Use(coresymbols.Symbols) // gui imports
	in.Interp.Use(gui.Symbols)         // gui imports
	in.Interp.Use(interp.Exports{
		"cogentcore.org/lab/simrun/simrun": map[string]reflect.Value{
			"SimRun": reflect.ValueOf(br), // our SimRun is available as simrun.SimRun
		},
	})
	in.Config()
	br.SetScriptsDir(filepath.Join(br.DataRoot, "dbscripts"))
	lab.TheBrowser = &br.Browser
	lab.CurTabber = br.Browser.Tabs
	goalrun = in.Goal
	br.Config.StartDir = br.StartDir
	br.Config.DataRoot = br.DataRoot
	br.Config.Defaults()
	br.JobsTable = table.New()
	br.UpdateScripts() // automatically runs lowercase init scripts
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
		w.SetFunc(br.UpdateFiles).SetText("").SetIcon(icons.Refresh)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.UpdateSims).SetText("Jobs").SetIcon(icons.ViewList).SetShortcut("Command+U")
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Queue).SetIcon(icons.List)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Status).SetIcon(icons.Sync)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Finalize).SetIcon(icons.Sync)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Fetch).SetIcon(icons.Download)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Submit).SetIcon(icons.Add)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Results).SetIcon(icons.Refresh)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Reset).SetIcon(icons.Refresh)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Diff).SetIcon(icons.Difference)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Plot).SetIcon(icons.ShowChart)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Cancel).SetIcon(icons.Refresh)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Delete).SetIcon(icons.Delete)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.Archive).SetIcon(icons.Archive)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.EditConfig).SetIcon(icons.Edit)
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("README").SetIcon(icons.FileMarkdown).
			SetTooltip("open README help file").OnClick(func(e events.Event) {
			core.TheApp.OpenURL("https://github.com/cogentcore/core/blob/main/tensor/examples/planets/README.md")
		})
	})
}
