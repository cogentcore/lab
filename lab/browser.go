// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

//go:generate core generate

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/goal/goalib"
	"golang.org/x/exp/maps"
)

var (
	// LabBrowser is the current Lab Browser, for yaegi / Go consistent access.
	LabBrowser *Browser

	// RunScriptCode is set if labscripts is included.
	// Runs given code on browser's interpreter.
	RunScriptCode func(br *Browser, code string) error
)

// Browser holds all the elements of a data browser, for browsing data
// either on an OS filesystem or as a tensorfs virtual data filesystem.
// It supports the automatic loading of [goal] scripts as toolbar actions to
// perform pre-programmed tasks on the data, to create app-like functionality.
// Scripts are ordered alphabetically and any leading #- prefix is automatically
// removed from the label, so you can use numbers to specify a custom order.
// It is not a [core.Widget] itself, and is intended to be incorporated into
// a [core.Frame] widget, potentially along with other custom elements.
// See [Basic] for a basic implementation.
type Browser struct { //types:add -setters
	// FS is the filesystem, if browsing an FS.
	FS fs.FS

	// DataRoot is the path to the root of the data to browse.
	DataRoot string

	// StartDir is the starting directory, where the app was originally started.
	StartDir string

	// ScriptsDir is the directory containing scripts for toolbar actions.
	// It defaults to DataRoot/dbscripts
	ScriptsDir string

	// Scripts are interpreted goal scripts (via yaegi) to automate
	// routine tasks.
	Scripts map[string]string `set:"-"`

	// Interpreter is the interpreter to use for running Browser scripts.
	// is of type: *goal/interpreter.Interpreter but can't use that directly
	// to avoid importing goal unless needed. Import [labscripts] if needed.
	Interpreter any `set:"-"`

	// Files is the [DataTree] tree browser of the tensorfs or files.
	Files *DataTree

	// Tabs is the [Tabs] element managing tabs of data views.
	Tabs *Tabs

	// Toolbar is the top-level toolbar for the browser, if used.
	Toolbar *core.Toolbar

	// Splits is the overall [core.Splits] for the browser.
	Splits *core.Splits
}

// UpdateFiles Updates the files list.
func (br *Browser) UpdateFiles() { //types:add
	if br.Files == nil {
		return
	}
	files := br.Files
	if br.FS != nil {
		files.SortByModTime = true
		files.OpenPathFS(br.FS, br.DataRoot)
	} else {
		files.OpenPath(br.DataRoot)
	}
}

// UpdateScripts updates the Scripts and updates the toolbar.
func (br *Browser) UpdateScripts() { //types:add
	redo := (br.Scripts != nil)
	scr := fsx.Filenames(br.ScriptsDir, ".goal")
	br.Scripts = make(map[string]string)
	for _, s := range scr {
		snm := strings.TrimSuffix(s, ".goal")
		sc, err := os.ReadFile(filepath.Join(br.ScriptsDir, s))
		if err == nil {
			if unicode.IsLower(rune(snm[0])) {
				if !redo {
					fmt.Println("run init script:", snm)
					if RunScriptCode != nil {
						RunScriptCode(br, string(sc))
					}
				}
			} else {
				ssc := string(sc)
				br.Scripts[snm] = ssc
			}
		} else {
			slog.Error(err.Error())
		}
	}
	if br.Toolbar != nil {
		br.Toolbar.Update()
	}
}

// MakeToolbar makes a default toolbar for the browser, with update files
// and update scripts buttons, followed by MakeScriptsToolbar for the scripts.
func (br *Browser) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.UpdateFiles).SetText("").SetIcon(icons.Refresh).SetShortcut("Command+U")
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(br.UpdateScripts).SetText("").SetIcon(icons.Code)
	})
	br.MakeScriptsToolbar(p)
}

// MakeScriptsToolbar is a maker for adding buttons for each uppercase script
// to the toolbar.
func (br *Browser) MakeScriptsToolbar(p *tree.Plan) {
	scr := maps.Keys(br.Scripts)
	slices.Sort(scr)
	for _, s := range scr {
		lbl := TrimOrderPrefix(s)
		tree.AddAt(p, lbl, func(w *core.Button) {
			w.SetText(lbl).SetIcon(icons.RunCircle).
				OnClick(func(e events.Event) {
					// br.RunScript(s)
				})
			sc := br.Scripts[s]
			tt := FirstComment(sc)
			if tt == "" {
				tt = "Run Script (add a comment to top of script to provide more useful info here)"
			}
			w.SetTooltip(tt)
		})
	}
}

//////// Helpers

// FirstComment returns the first comment lines from given .goal file,
// which is used to set the tooltip for scripts.
func FirstComment(sc string) string {
	sl := goalib.SplitLines(sc)
	cmt := ""
	for _, l := range sl {
		if !strings.HasPrefix(l, "// ") {
			return cmt
		}
		cmt += strings.TrimSpace(l[3:]) + " "
	}
	return cmt
}

// TrimOrderPrefix trims any optional #- prefix from given string,
// used for ordering items by name.
func TrimOrderPrefix(s string) string {
	i := strings.Index(s, "-")
	if i < 0 {
		return s
	}
	ds := s[:i]
	if _, err := strconv.Atoi(ds); err != nil {
		return s
	}
	return s[i+1:]
}

// PromptOKCancel prompts the user for whether to do something,
// calling the given function if the user clicks OK.
func PromptOKCancel(ctx core.Widget, prompt string, fun func()) {
	d := core.NewBody(prompt)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if fun != nil {
				fun()
			}
		})
	})
	d.RunDialog(ctx)
}

// PromptString prompts the user for a string value (initial value given),
// calling the given function if the user clicks OK.
func PromptString(ctx core.Widget, str string, prompt string, fun func(s string)) {
	d := core.NewBody(prompt)
	tf := core.NewTextField(d).SetText(str)
	tf.Styler(func(s *styles.Style) {
		s.Min.X.Ch(60)
	})
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if fun != nil {
				fun(tf.Text())
			}
		})
	})
	d.RunDialog(ctx)
}

// PromptStruct prompts the user for the values in given struct (pass a pointer),
// calling the given function if the user clicks OK.
func PromptStruct(ctx core.Widget, str any, prompt string, fun func()) {
	d := core.NewBody(prompt)
	core.NewForm(d).SetStruct(str)
	d.AddBottomBar(func(bar *core.Frame) {
		d.AddCancel(bar)
		d.AddOK(bar).OnClick(func(e events.Event) {
			if fun != nil {
				fun()
			}
		})
	})
	d.RunDialog(ctx)
}
