// Code generated by "goal build"; DO NOT EDIT.
//line jobs.goal:1
// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"cogentcore.org/core/base/elide"
	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/base/strcase"
	"cogentcore.org/core/core"
	"cogentcore.org/lab/goal/goalib"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
)

// Jobs updates the Jobs tab with a Table showing all the Jobs
// with their meta data.  Uses the dbmeta.toml data compiled from
// the Status function.
func (br *SimRun) Jobs() { //types:add
	ts := br.Tabs.AsLab()
	tv := ts.TensorTable("Jobs", br.JobsTable)
	dt := br.JobsTable
	br.JobsTableView = tv
	dpath := filepath.Join(br.DataRoot, "jobs")
	// fmt.Println("opening data at:", dpath)

	if dt.NumColumns() == 0 {
		dbfmt := filepath.Join(br.DataRoot, "dbformat.csv")
		fdt := table.New()
		if errors.Log1(fsx.FileExists(dbfmt)) {
			fdt.OpenCSV(fsx.Filename(dbfmt), tensor.Comma)
		} else {
			fdt.ReadCSV(bytes.NewBuffer([]byte(defaultJobFormat)), tensor.Comma)
		}
		dt.ConfigFromTable(fdt)
	}

	ds := fsx.Dirs(dpath)
	dt.SetNumRows(len(ds))
	for i, d := range ds {
		dt.Column("JobID").SetString(d, i)
		dp := filepath.Join(dpath, d)
		meta := filepath.Join(dp, "dbmeta.toml")
		if goalib.FileExists(meta) {
			md := make(map[string]string)
			tomlx.Open(&md, meta)
			for k, v := range md {
				dc := dt.Column(k)
				if dc != nil {
					dc.SetString(v, i)
					//	} else {
					//		fmt.Println("warning: job column named:", k, "not found")
				}
			}
		}
	}
	tv.Table.Sequential()
	br.Update()
	nrows := dt.NumRows()
	if nrows > 0 && br.Config.Submit.Message == "" {
		br.Config.Submit.Message = dt.Column("Message").String1D(nrows - 1)
		br.Config.Submit.Args = dt.Column("Args").String1D(nrows - 1)
		br.Config.Submit.Label = dt.Column("Label").String1D(nrows - 1)
	}
}

// Default update function
func (br *SimRun) UpdateSims() {
	br.Jobs()
}

func (br *SimRun) JobPath(jid string) string {
	return filepath.Join(br.DataRoot, "jobs", jid)
}

func (br *SimRun) ServerJobPath(jid string) string {
	return filepath.Join(br.Config.ServerRoot, "jobs", jid)
}

func (br *SimRun) JobRow(jid string) int {
	jt := br.JobsTable.Column("JobID")
	nr := jt.DimSize(0)
	for i := range nr {
		if jt.String1D(i) == jid {
			return i
		}
	}
	fmt.Println("JobRow ERROR: job id:", jid, "not found")
	return -1
}

// ValueForJob returns value in given column for given job id
func (br *SimRun) ValueForJob(jid, column string) string {
	if jrow := br.JobRow(jid); jrow >= 0 {
		return br.JobsTable.Column(column).String1D(jrow)
	}
	return ""
}

// Queue runs a queue query command on the server and shows the results.
func (br *SimRun) Queue() { //types:add
	ts := br.Tabs.AsLab()
	goalrun.Run("@1")
	goalrun.Run("cd")
	myq := goalrun.Output("squeue", "-l", "-u", "$USER")
	sinfoall := goalrun.Output("sinfo")
	goalrun.Run("@0")
	sis := []string{}
	for _, l := range goalib.SplitLines(sinfoall) {
		if strings.HasPrefix(l, "low") || strings.HasPrefix(l, "med") {
			continue
		}
		sis = append(sis, l)
	}
	sinfo := strings.Repeat("#", 60) + "\n" + strings.Join(sis, "\n")
	qstr := myq + "\n" + sinfo
	ts.EditorString("Queue", qstr)
}

// JobStatus gets job status from server for given job id.
// jobs that are already Finalized are skipped, unless force is true.
func (br *SimRun) JobStatus(jid string, force bool) {
	// fmt.Println("############\nStatus of Job:", jid)
	spath := br.ServerJobPath(jid)
	jpath := br.JobPath(jid)
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("@0")
	goalrun.Run("cd", jpath)
	sstat := goalib.ReadFile("job.status")
	if !force && (sstat == "Finalized" || sstat == "Fetched") {
		return
	}
	goalrun.Run("@1", "cd", spath)
	goalrun.Run("@0")
	sj := goalrun.Output("@1", "cat", "job.job")
	// fmt.Println("server job:", sj)
	if sstat != "Done" && !force {
		goalrun.RunErrOK("@1", "squeue", "-j", sj, "-o", "%T", ">&", "job.squeue")
		stat := goalrun.Output("@1", "cat", "job.squeue")
		// fmt.Println("server status:", stat)
		switch {
		case strings.Contains(stat, "Invalid job id"):
			goalrun.Run("@1", "echo", "Invalid job id", ">", "job.squeue")
			sstat = "Done"
		case strings.Contains(stat, "RUNNING"):
			nrep := strings.Count(stat, "RUNNING")
			sstat = fmt.Sprintf("Running:%d", nrep)
		case strings.Contains(stat, "PENDING"):
			nrep := strings.Count(stat, "PENDING")
			sstat = fmt.Sprintf("Pending:%d", nrep)
		case strings.Contains(stat, "STATE"): // still visible in queue but done
			sstat = "Done"
		}
		goalib.WriteFile("job.status", sstat)
	}
	goalrun.Run("@1", "/bin/ls", "-1", ">", "job.files")
	goalrun.Run("@0")
	core.MessageSnackbar(br, "Retrieving job files for: "+jid)
	jfiles := goalrun.Output("@1", "/bin/ls", "-1", "job.*")
	for _, jf := range goalib.SplitLines(jfiles) {
		// fmt.Println(jf)
		rfn := "@1:" + jf
		if !force {
			goalrun.Run("scp", rfn, jf)
		}
	}
	goalrun.Run("@0")
	if sstat == "Done" {
		sstat = "Finalized"
		goalib.WriteFile("job.status", sstat)
		goalrun.RunErrOK("/bin/rm", "job.*.out")
	}
	jfiles = goalrun.Output("/bin/ls", "-1", "job.*") // local
	meta := fmt.Sprintf("%s = %q\n", "Version", br.Config.Version) + fmt.Sprintf("%s = %q\n", "Server", br.Config.ServerName)
	for _, jf := range goalib.SplitLines(jfiles) {
		if strings.Contains(jf, "sbatch") || strings.HasSuffix(jf, ".out") {
			continue
		}
		key := strcase.ToCamel(strings.TrimPrefix(jf, "job."))
		switch key {
		case "Job":
			key = "ServerJob"
		case "Squeue":
			key = "ServerStatus"
		}
		val := strings.TrimSpace(goalib.ReadFile(jf))
		if key == "ServerStatus" {
			val = strings.ReplaceAll(elide.Middle(val, 50), "…", "...")
		}
		ln := fmt.Sprintf("%s = %q\n", key, val)
		// fmt.Println(ln)
		meta += ln
	}
	goalib.WriteFile("dbmeta.toml", meta)
	core.MessageSnackbar(br, "Job: "+jid+" updated with status: "+sstat)
}

// Status gets updated job.* files from the server for any job that
// doesn't have a Finalized or Fetched status.  It updates the
// status based on the server job status query, assigning a
// status of Finalized if job is done.  Updates the dbmeta.toml
// data based on current job data.
func (br *SimRun) Status() { //types:add
	goalrun.Run("@0")
	br.UpdateFiles()
	dpath := filepath.Join(br.DataRoot, "jobs")
	ds := fsx.Dirs(dpath)
	for _, jid := range ds {
		br.JobStatus(jid, false) // true = update all -- for format and status edits
	}
	core.MessageSnackbar(br, "Jobs Status completed")
	br.UpdateSims()
}

// FetchJob downloads results files from server.
// if force == true then will re-get already-Fetched jobs,
// otherwise these are skipped.
func (br *SimRun) FetchJob(jid string, force bool) {
	spath := br.ServerJobPath(jid)
	jpath := br.JobPath(jid)
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("@0")
	goalrun.Run("cd", jpath)
	sstat := goalib.ReadFile("job.status")
	if !force && sstat == "Fetched" {
		return
	}
	goalrun.Run("@1", "cd", spath)
	goalrun.Run("@0")
	ffiles := goalrun.Output("@1", "/bin/ls", "-1", br.Config.FetchFiles)
	if len(ffiles) > 0 {
		core.MessageSnackbar(br, fmt.Sprintf("Fetching %d data files for job: %s", len(ffiles), jid))
	}
	for _, ff := range goalib.SplitLines(ffiles) {
		// fmt.Println(ff)
		rfn := "@1:" + ff
		goalrun.Run("scp", rfn, ff)
		if (sstat == "Finalized" || sstat == "Fetched") && strings.HasSuffix(ff, ".tsv") {
			if strings.Contains(ff, "_epc.tsv") {
				table.CleanCatTSV(ff, "Run", "Epoch")
				idx := strings.Index(ff, "_epc.tsv")
				goalrun.Run("tablecat", "-colavg", "-col", "Epoch", "-o", ff[:idx+1]+"avg"+ff[idx+1:], ff)
			} else if strings.Contains(ff, "_run.tsv") {
				table.CleanCatTSV(ff, "Run")
				idx := strings.Index(ff, "_run.tsv")
				goalrun.Run("tablecat", "-colavg", "-o", ff[:idx+1]+"avg"+ff[idx+1:], ff)
			} else {
				table.CleanCatTSV(ff, "Run")
			}
		}
	}
	goalrun.Run("@0")
	if sstat == "Finalized" {
		// fmt.Println("status finalized")
		goalib.WriteFile("job.status", "Fetched")
		goalib.ReplaceInFile("dbmeta.toml", "\"Finalized\"", "\"Fetched\"")
	} else {
		fmt.Println("status:", sstat)
	}
}

// Fetch retrieves all the .tsv data files from the server
// for any jobs not already marked as Fetched.
// Operates on the jobs selected in the Jobs table,
// or on all jobs if none selected.
func (br *SimRun) Fetch() { //types:add
	goalrun.Run("@0")
	tv := br.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		dpath := filepath.Join(br.DataRoot, "jobs")
		jobs = fsx.Dirs(dpath)
	}
	for _, jid := range jobs {
		br.FetchJob(jid, false)
	}
	core.MessageSnackbar(br, "Fetch Jobs completed")
	br.UpdateSims()
}

// CancelJobs cancels the given jobs
func (br *SimRun) CancelJobs(jobs []string) {
	goalrun.Run("@0")
	filepath.Join(br.DataRoot, "jobs")
	filepath.Join(br.Config.ServerRoot, "jobs")
	goalrun.Run("@1")
	for _, jid := range jobs {
		sjob := br.ValueForJob(jid, "ServerJob")
		if sjob != "" {
			goalrun.Run("scancel", sjob)
		}
	}
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("@0")
	core.MessageSnackbar(br, "Done canceling jobs")
}

// Cancel cancels the jobs selected in the Jobs table,
// with a confirmation prompt.
func (br *SimRun) Cancel() { //types:add
	tv := br.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		core.MessageSnackbar(br, "No jobs selected for cancel")
		return
	}
	lab.PromptOKCancel(br, "Ok to cancel these jobs: "+strings.Join(jobs, " "), func() {
		br.CancelJobs(jobs)
		br.UpdateSims()
	})
}

// DeleteJobs deletes the given jobs
func (br *SimRun) DeleteJobs(jobs []string) {
	goalrun.Run("@0")
	dpath := filepath.Join(br.DataRoot, "jobs")
	spath := filepath.Join(br.Config.ServerRoot, "jobs")
	for _, jid := range jobs {
		goalrun.Run("@1")
		goalrun.Run("cd")
		goalrun.Run("cd", spath)
		goalrun.RunErrOK("/bin/rm", "-rf", jid)
		goalrun.Run("@0")
		goalrun.Run("cd", dpath)
		goalrun.RunErrOK("/bin/rm", "-rf", jid)
	}
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("@0")
	core.MessageSnackbar(br, "Done deleting jobs")
}

// Delete deletes the selected Jobs, with a confirmation prompt.
func (br *SimRun) Delete() { //types:add
	tv := br.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		core.MessageSnackbar(br, "No jobs selected for deletion")
		return
	}
	lab.PromptOKCancel(br, "Ok to delete these jobs: "+strings.Join(jobs, " "), func() {
		br.DeleteJobs(jobs)
		br.UpdateSims()
	})
}

// ArchiveJobs archives the given jobs
func (br *SimRun) ArchiveJobs(jobs []string) {
	goalrun.Run("@0")
	dpath := filepath.Join(br.DataRoot, "jobs")
	apath := filepath.Join(br.DataRoot, "archive", "jobs")
	goalrun.Run("mkdir", "-p", apath)
	spath := filepath.Join(br.Config.ServerRoot, "jobs")
	for _, jid := range jobs {
		goalrun.Run("@1")
		goalrun.Run("cd")
		goalrun.Run("cd", spath)
		goalrun.RunErrOK("/bin/rm", "-rf", jid)
		goalrun.Run("@0")
		dj := filepath.Join(dpath, jid)
		aj := filepath.Join(apath, jid)
		goalrun.Run("/bin/mv", dj, aj)
	}
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("@0")
	core.MessageSnackbar(br, "Done archiving jobs")
}

// Archive moves the selected Jobs to the Archive directory,
// locally, and deletes them from the server,
// for results that are useful but not immediately relevant.
func (br *SimRun) Archive() { //types:add
	tv := br.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		core.MessageSnackbar(br, "No jobs selected for archiving")
		return
	}
	lab.PromptOKCancel(br, "Ok to archive these jobs: "+strings.Join(jobs, " "), func() {
		br.ArchiveJobs(jobs)
		br.UpdateSims()
	})
}
