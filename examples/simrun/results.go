// Code generated by "goal build"; DO NOT EDIT.
//line results.goal:1
// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/core"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
)

// OpenResultFiles opens the given result files.
func (sr *SimRun) OpenResultFiles(jobs []string, filter FilterResults) {
	ts := sr.Tabs.AsLab()
	for _, jid := range jobs {
		jpath := sr.JobPath(jid)
		message := sr.ValueForJob(jid, "Message")
		label := sr.ValueForJob(jid, "Label")
		args := sr.ValueForJob(jid, "Args")
		fls := fsx.Filenames(jpath, filter.Ext)
		for _, fn := range fls {
			if filter.FileContains != "" && !strings.Contains(fn, filter.FileContains) {
				continue
			}
			dt := table.New()
			fpath := filepath.Join(jpath, fn)
			err := dt.OpenCSV(core.Filename(fpath), tensor.Tab)
			if err != nil {
				fmt.Println(err.Error())
			}
			rpath := strings.TrimPrefix(fpath, sr.DataRoot)
			sr.ResultsList = append(sr.ResultsList, &Result{JobID: jid, Label: label, Message: message, Args: args, Path: rpath, Table: dt})
		}
	}
	if len(sr.ResultsList) == 0 {
		core.MessageSnackbar(sr, "No files containing: "+filter.FileContains+" with extension: "+filter.Ext)
		return
	}
	sr.ResultsTableView = ts.SliceTable("Results", &sr.ResultsList)
	sr.ResultsTableView.Update()
	sr.UpdateSims()
}

// Results loads specific .tsv data files from the jobs selected
// in the Jobs table, into the Results table.  There are often
// multiple result files per job, so this step is necessary to
// choose which such files to select for plotting.
func (sr *SimRun) Results() { //types:add
	tv := sr.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		fmt.Println("No Jobs rows selected")
		return
	}
	// fmt.Println(jobs)
	if sr.Config.Filter.Ext == "" {
		sr.Config.Filter.Ext = ".tsv"
	}
	lab.PromptStruct(sr, &sr.Config.Filter, "Open results data for files", func() {
		sr.OpenResultFiles(jobs, sr.Config.Filter)
	})
}

// Diff shows the differences between two selected jobs, or if only
// one job is selected, between that job and the current source directory.
func (sr *SimRun) Diff() { //types:add
	goalrun.Run("@0")
	tv := sr.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	nj := len(jobs)
	if nj == 0 || nj > 2 {
		core.MessageSnackbar(sr, "Diff requires two Job rows to be selected")
		return
	}
	if nj == 1 {
		ja := sr.JobPath(jobs[0])
		lab.NewDiffBrowserDirs(ja, sr.StartDir)
		return
	}
	ja := sr.JobPath(jobs[0])
	jb := sr.JobPath(jobs[1])
	lab.NewDiffBrowserDirs(ja, jb)
}

// Plot concatenates selected Results data files and generates a plot
// of the resulting data.
func (sr *SimRun) Plot() { //types:add
	ts := sr.Tabs.AsLab()
	tv := sr.ResultsTableView
	jis := tv.SelectedIndexesList(false)
	if len(jis) == 0 {
		fmt.Println("No Results rows selected")
		return
	}
	var AggTable *table.Table
	for _, i := range jis {
		res := sr.ResultsList[i]
		jid := res.JobID
		label := res.Label
		dt := res.Table.InsertKeyColumns("JobID", jid, "JobLabel", label)
		if AggTable == nil {
			AggTable = dt
		} else {
			AggTable.AppendRows(dt)
		}
	}
	ts.PlotTable("Plot", AggTable)
	sr.UpdateSims()
}

// Reset resets the Results table
func (sr *SimRun) Reset() { //types:add
	ts := sr.Tabs.AsLab()
	sr.ResultsList = []*Result{}
	sr.ResultsTableView = ts.SliceTable("Results", &sr.ResultsList)
	sr.UpdateSims()
}
