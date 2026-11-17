// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"testing"

	"cogentcore.org/core/base/iox/jsonx"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/types"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/table"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlotMean tests PlotMean with the group columns that it is given,
// which are saved in the project, and prompted for by its button.
func TestPlotMean(t *testing.T) {
	defer core.TrackWindows()()
	// In a test, the main package is named by its import path, so the type
	// is added under that name too, for the FuncButton to find its methods.
	typ := *types.TypeByName("main.Simmer")
	typ.Name = "cogentcore.org/lab/examples/simmer.Simmer"
	types.AddType(&typ)
	t.Setenv("HOME", t.TempDir())
	homedir.Reset()
	t.Chdir(t.TempDir())
	b, sr := NewSimmerWindow(interpreter.NewInterpreter(interpreter.Options{}))
	sim := filepath.Join(t.TempDir(), "mysim")
	require.NoError(t, os.Mkdir(sim, 0755))
	require.NoError(t, sr.newProject(sim))
	assert.Equal(t, []string{"Epoch"}, sr.Project.GroupColumns, "the default")

	dt := table.New()
	run := dt.AddIntColumn("Run")
	epc := dt.AddIntColumn("TrainEpoch")
	er := dt.AddFloat64Column("Err")
	dt.SetNumRows(6)
	for i := range 6 {
		run.SetInt1D(i/3, i)
		epc.SetInt1D(i%3, i)
		er.SetFloat1D(float64(i), i)
	}
	sr.ResultsList = []*Result{{JobID: "j1", Label: "a", Path: "a_epoch.tsv", Table: dt}}
	core.RenderImage(b, core.Locked(b, func() {
		sr.ResultsTableView = sr.Tabs.AsLab().SliceTable("Results", &sr.ResultsList)
		sr.ResultsTableView.SelectIndex(0)

		var fb *core.FuncButton
		for _, c := range sr.Toolbar.Children {
			if w, ok := c.(*core.FuncButton); ok && w.Text == "Plot mean" {
				fb = w
			}
		}
		require.NotNil(t, fb)
		sr.Project.GroupColumns = []string{"Run", "TrainEpoch"}
		fb.Send(events.Click)
		assert.Equal(t, []string{"Run", "TrainEpoch"}, fb.Args[0].Value, "prompted with the saved columns")

		sr.PlotMean([]string{"Epoch"}) // not in the results
		assert.Equal(t, []string{"Run", "TrainEpoch"}, sr.Project.GroupColumns, "not saved")

		sr.PlotMean([]string{"TrainEpoch"})
		assert.Equal(t, []string{"TrainEpoch"}, sr.Project.GroupColumns)
		var pr Project
		require.NoError(t, jsonx.Open(&pr, filepath.Join(sim, ProjectFilename)))
		assert.Equal(t, []string{"TrainEpoch"}, pr.GroupColumns, "saved in the project")
	}))
}
