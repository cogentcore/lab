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
func (br *SimRun) OpenResultFiles(jobs []string, filter FilterResults) {
	ts := br.Tabs.AsLab()
	for _, jid := range jobs {
		jpath := br.JobPath(jid)
		message := br.ValueForJob(jid, "Message")
		label := br.ValueForJob(jid, "Label")
		args := br.ValueForJob(jid, "Args")
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
			rpath := strings.TrimPrefix(fpath, br.DataRoot)
			br.ResultsList = append(br.ResultsList, &Result{JobID: jid, Label: label, Message: message, Args: args, Path: rpath, Table: dt})
		}
	}
	if len(br.ResultsList) == 0 {
		core.MessageSnackbar(br, "No files containing: "+filter.FileContains+" with extension: "+filter.Ext)
		return
	}
	br.ResultsTableView = ts.SliceTable("Results", &br.ResultsList)
	br.ResultsTableView.Update()
	br.UpdateSims()
}

// Results loads specific .tsv data files from the jobs selected
// in the Jobs table, into the Results table.  There are often
// multiple result files per job, so this step is necessary to
// choose which such files to select for plotting.
func (br *SimRun) Results() { //types:add
	tv := br.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		fmt.Println("No Jobs rows selected")
		return
	}
	// fmt.Println(jobs)
	if br.Config.Filter.Ext == "" {
		br.Config.Filter.Ext = ".tsv"
	}
	lab.PromptStruct(br, &br.Config.Filter, "Open results data for files", func() {
		br.OpenResultFiles(jobs, br.Config.Filter)
	})
}

// Diff shows the differences between two selected jobs, or if only
// one job is selected, between that job and the current source directory.
func (br *SimRun) Diff() { //types:add
	goalrun.Run("@0")
	tv := br.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	nj := len(jobs)
	if nj == 0 || nj > 2 {
		core.MessageSnackbar(br, "Diff requires two Job rows to be selected")
		return
	}
	if nj == 1 {
		ja := br.JobPath(jobs[0])
		lab.NewDiffBrowserDirs(ja, br.StartDir)
		return
	}
	ja := br.JobPath(jobs[0])
	jb := br.JobPath(jobs[1])
	lab.NewDiffBrowserDirs(ja, jb)
}

// Plot concatenates selected Results data files and generates a plot
// of the resulting data.
func (br *SimRun) Plot() { //types:add
	ts := br.Tabs.AsLab()
	tv := br.ResultsTableView
	jis := tv.SelectedIndexesList(false)
	if len(jis) == 0 {
		fmt.Println("No Results rows selected")
		return
	}
	var AggTable *table.Table
	for _, i := range jis {
		res := br.ResultsList[i]
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
	br.UpdateSims()
}

// Reset resets the Results table
func (br *SimRun) Reset() { //types:add
	ts := br.Tabs.AsLab()
	br.ResultsList = []*Result{}
	br.ResultsTableView = ts.SliceTable("Results", &br.ResultsList)
	br.UpdateSims()
}
