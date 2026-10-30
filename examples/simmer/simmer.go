// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate -add-types -add-funcs

import (
	"image"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/cli"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/examples/baremetal"
	"cogentcore.org/lab/goal"
	"cogentcore.org/lab/goal/goalib"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/lab/labscripts"
	_ "cogentcore.org/lab/lab/labscripts"
	"cogentcore.org/lab/mvmlab"
	"cogentcore.org/lab/mvmlab/labsymbols"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensorcore"
)

// goalrun is needed for running goal commands.
var goalrun *goal.Goal

// clearErrors clears the errors of the goal commands, which are logged
// when they happen. After a command fails, goal runs no more commands
// until they are cleared, as in a script, but the operations of simmer,
// and those on each job, are independent: a command that fails, as when
// the network to the server is lost, must not stop the operations after
// it, such as when the network comes back.
func clearErrors() {
	goalrun.Errors = nil
}

var defaultJobFormat = `Name, Type
JobID, string
Status, string
Label, string
Message, string
Version, string
Args, string
Server, string
ServerJob, string
ServerStatus, string
Submit, string
Start, string
End, string
`

// Simmer manages the running and data analysis of results from simulations
// that are run on remote server(s), within a Cogent Lab browser environment,
// with the files as the left panel, and the Tabber as the right panel,
// with the terminal running the goal shell below it.
type Simmer struct {
	core.Frame
	lab.Browser

	// Project has the settings of the open simulation project,
	// which are saved in its simmer.json file.
	Project Project

	// JobsTableView is the view of the jobs table.
	JobsTableView *tensorcore.Table

	// JobsTable is the jobs Table with one row per job.
	JobsTable *table.Table

	// ResultsTableView has the results table.
	ResultsTableView *core.Table

	// ResultsList is the list of result records.
	ResultsList []*Result

	// BareMetal RPC client.
	BareMetal *baremetal.Client

	// Status info from BareMetal
	BareMetalActive      []*baremetal.Job
	BareMetalActiveTable *core.Table

	// lastConnect is the time of the last try to connect to the server
	// (see [Simmer.connectServer]).
	lastConnect time.Time
}

// startProject is the project given on the command line, if any.
var startProject string

func main() {
	startProject = projectArg()
	opts := cli.DefaultOptions("simmer", "interactive simulation running and data analysis.")
	cfg := &interpreter.Config{}
	cfg.InteractiveFunc = Interactive
	cli.Run(opts, cfg, interpreter.Run, interpreter.Build)
}

// Interactive is the cli function that gets called by default at gui startup.
// It opens the project given on the command line, or the one in the
// current directory, if any, and otherwise shows the recent projects,
// and runs the goal shell in the terminal of the window.
func Interactive(c *interpreter.Config, in *interpreter.Interpreter) error {
	b, sr := NewSimmerWindow(in)
	path := startProject
	if path == "" {
		wd, _ := os.Getwd()
		if goalib.FileExists(filepath.Join(wd, ProjectFilename)) || goalib.DirExists(filepath.Join(wd, "simdata")) {
			path = wd
		}
	}
	if path != "" {
		sr.OpenProject(core.Filename(path))
	}
	if !sr.isOpen() {
		sr.projectsTab()
	}
	mvmlab.InteractiveTerminal(c, in, sr.Terminal)
	b.RunMainWindow()
	return nil
}

// NewSimmerWindow returns a new Simmer window using given interpreter,
// without a project (see [Simmer.OpenProject]).
// do RunWindow on resulting [core.Body] to open the window.
func NewSimmerWindow(in *interpreter.Interpreter) (*core.Body, *Simmer) {
	b := core.NewBody("Simmer")
	sr := NewSimmer(b)
	sr.Interpreter = in
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		sr.Toolbar = tb
		tb.Maker(sr.MakeToolbar)
	})
	sr.Update() // makes the widgets, such as the Tabs and Terminal
	sr.InitSimmer()
	return b, sr
}

// InitSimmer initializes simmer, without a project: the goal Interpreter,
// which is typically already set by this point but will be created if not,
// and the BareMetal client.
// All the simmer data of a project is contained within a "simdata"
// directory in the directory of the simulation: this dir is typically a
// symbolic link to a common collection of such simdata directories for
// all the different simulations being run (see [Simmer.NewProject]).
func (sr *Simmer) InitSimmer() {
	openRecentProjects()
	labscripts.InitInterpreter(&sr.Browser)
	in, _ := labscripts.Interpreter(&sr.Browser)
	// our Simmer is available as lab.LabBrowser
	labsymbols.Load()
	labValues := maps.Clone(labsymbols.Values["cogentcore.org/lab/lab"])
	labValues["LabBrowser"] = reflect.ValueOf(&sr)
	mvmlab.UseGUI(in, map[string]map[string]reflect.Value{"cogentcore.org/lab/lab": labValues})
	in.Config()
	lab.LabBrowser = &sr.Browser
	lab.Lab = sr.Browser.Tabs
	goalrun = in.Goal
	sr.JobsTable = table.New()

	// used for the projects that are not slurm
	sr.BareMetal = baremetal.NewClient()
	err := sr.BareMetal.Connect()
	if errors.Log(err) != nil {
		sr.BareMetal = nil
	}
}

func (sr *Simmer) Init() {
	sr.Frame.Init()
	sr.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
	})
	sr.OnShow(func(e events.Event) {
		if !sr.isOpen() {
			return
		}
		sr.UpdateFiles()
	})
	tree.AddChildAt(sr, "splits", func(w *core.Splits) {
		sr.Splits = w
		w.SetSplits(.15, .85)
		tree.AddChildAt(w, "fileframe", func(w *core.Frame) {
			w.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				s.Overflow.Set(styles.OverflowAuto)
				s.Grow.Set(1, 1)
			})
			tree.AddChildAt(w, "filetree", func(w *lab.DataTree) {
				sr.Files = w
			})
		})
		sr.AddTabsTerminal(w)
	})
	sr.Updater(func() {
		if sr.Files != nil {
			sr.Files.Tabber = sr.Tabs
		}
	})
}

// AsyncMessageSnackbar must be used for MessageSnackbar in a goroutine.
func (sr *Simmer) AsyncMessageSnackbar(message string) {
	sr.AsyncLock()
	core.MessageSnackbar(sr, message)
	sr.AsyncUnlock()
}

// IsSlurm returns true if using slurm (vs. baremetal)
func (sr *Simmer) IsSlurm() bool {
	return sr.Project.Server.Slurm
}

func (sr *Simmer) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.UpdateFiles).SetText("").SetIcon(icons.Refresh)
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Open recent").SetIcon(icons.History).SetMenu(func(m *core.Scene, pos image.Point) {
			for _, rp := range RecentProjects {
				core.NewButton(m).SetText(filepath.Dir(rp)).OnClick(func(e events.Event) {
					sr.OpenProject(core.Filename(rp))
				})
			}
			core.NewSeparator(m)
			core.NewButton(m).SetText("Clear recent projects").OnClick(func(e events.Event) {
				RecentProjects = nil
				saveRecentProjects()
			})
		})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.OpenProject).SetText("Open").SetIcon(icons.Open)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.NewProject).SetText("New project").SetIcon(icons.Add)
	})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.UpdateSims).SetText("Jobs").SetIcon(icons.ViewList).SetShortcut("Command+U")
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Bare").SetIcon(icons.Refresh).
			SetTooltip("Update BareMetal jobs").OnClick(func(e events.Event) {
			sr.UpdateBare()
		})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Queue).SetIcon(icons.List)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Status).SetIcon(icons.Sync)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Fetch).SetIcon(icons.Download)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Submit).SetIcon(icons.Add)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.ParamSearch).SetIcon(icons.Search)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Results).SetIcon(icons.Open)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Reset).SetIcon(icons.Refresh)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Diff).SetIcon(icons.Difference)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Plot).SetIcon(icons.ShowChart)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.PlotMean).SetIcon(icons.ShowChart)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Cancel).SetIcon(icons.Refresh)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Delete).SetIcon(icons.Delete)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Archive).SetIcon(icons.Archive)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.Recover).SetIcon(icons.Archive)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(sr.EditProject).SetIcon(icons.Edit)
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("README").SetIcon(icons.FileMarkdown).
			SetTooltip("open README help file").OnClick(func(e events.Event) {
			core.TheApp.OpenURL("https://github.com/cogentcore/lab/blob/main/examples/simmer/README.md")
		})
	})
}
