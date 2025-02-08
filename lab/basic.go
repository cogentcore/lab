// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"io/fs"
	"os"
	"path/filepath"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/yaegicore/coresymbols"
	"cogentcore.org/lab/goal/interpreter"
)

// Basic is a basic data browser with the files as the left panel,
// and the Tabber as the right panel.
type Basic struct {
	core.Frame
	Browser
}

// Init initializes with the data and script directories
func (br *Basic) Init() {
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
			tree.AddChildAt(w, "filetree", func(w *DataTree) {
				br.Files = w
			})
		})
		tree.AddChildAt(w, "tabs", func(w *Tabs) {
			br.Tabs = w
			Lab = br.Tabs.AsLab()
		})
	})
	br.Updater(func() {
		if br.Files != nil {
			br.Files.Tabber = br.Tabs
		}
	})
}

// NewBasicWindow returns a new Lab Browser window for given
// file system (nil for os files) and data directory.
// do RunWindow on resulting [core.Body] to open the window.
func NewBasicWindow(fsys fs.FS, dataDir string, in *interpreter.Interpreter) (*core.Body, *Basic) {
	startDir, _ := os.Getwd()
	startDir = errors.Log1(filepath.Abs(startDir))
	b := core.NewBody("Cogent Lab: " + fsx.DirAndFile(startDir))
	br := NewBasic(b)
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		br.Toolbar = tb
		tb.Maker(br.MakeToolbar)
	})
	br.Interpreter = in
	br.FS = fsys
	ddr := dataDir
	if fsys == nil {
		ddr = errors.Log1(filepath.Abs(dataDir))
	}
	br.SetDataRoot(ddr)
	if br.Interpreter == nil {
		br.InitInterp()
		in = br.Interpreter
	}
	in.Interp.Use(coresymbols.Symbols) // gui imports
	br.SetScriptsDir(filepath.Join(ddr, "labscripts"))

	LabBrowser = &br.Browser
	return b, br
}
