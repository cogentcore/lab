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
	"strconv"
	"strings"

	"cogentcore.org/core/base/elide"
	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/base/strcase"
	"cogentcore.org/core/core"
	"cogentcore.org/core/styles"
	"cogentcore.org/lab/goal/goalib"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
)

// Jobs updates the Jobs tab with a Table showing all the Jobs
// with their meta data. Uses the dbmeta.toml data compiled from
// the Status function.
func (sr *SimRun) Jobs() { //types:add
	ts := sr.Tabs.AsLab()
	if !sr.IsSlurm() {
		at := ts.SliceTable("Bare Active", &sr.BareMetal.Active.Values)
		if sr.BareMetalActiveTable != at {
			sr.BareMetalActiveTable = at
			at.Styler(func(s *styles.Style) {
				s.SetReadOnly(true)
			})
		}
		at.Update()
		dt := ts.SliceTable("Bare Done", &sr.BareMetal.Done.Values)
		if sr.BareMetalDoneTable != dt {
			sr.BareMetalDoneTable = dt
			dt.Styler(func(s *styles.Style) {
				s.SetReadOnly(true)
			})
		}
		dt.Update()
	}

	tv := ts.TensorTable("Jobs", sr.JobsTable)
	dt := sr.JobsTable
	if sr.JobsTableView != tv {
		sr.JobsTableView = tv
		tv.ShowIndexes = true
		tv.ReadOnlyMultiSelect = true
		tv.Styler(func(s *styles.Style) {
			s.SetReadOnly(true)
		})
	}
	dpath := filepath.Join(sr.DataRoot, "jobs")
	// fmt.Println("opening data at:", dpath)

	if dt.NumColumns() == 0 {
		dbfmt := filepath.Join(sr.DataRoot, "dbformat.csv")
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
				}
			}
		}
	}
	tv.Table.Sequential()
	nrows := dt.NumRows()
	if nrows > 0 && sr.Config.Submit.Message == "" {
		sr.Config.Submit.Message = dt.Column("Message").String1D(nrows - 1)
		sr.Config.Submit.Args = dt.Column("Args").String1D(nrows - 1)
		sr.Config.Submit.Label = dt.Column("Label").String1D(nrows - 1)
	}
}

// Jobs updates the Jobs tab with a Table showing all the Jobs
// with their meta data. Uses the dbmeta.toml data compiled from
// the Status function.
func (sr *SimRun) UpdateSims() { //types:add
	sr.Jobs()
	sr.Update()
}

// UpdateSims updates the sim status info, for async case.
func (sr *SimRun) UpdateSimsAsync() {
	sr.AsyncLock()
	sr.Jobs()
	sr.Update()
	sr.AsyncUnlock()
}

func (sr *SimRun) JobPath(jid string) string {
	return filepath.Join(sr.DataRoot, "jobs", jid)
}

func (sr *SimRun) ServerJobPath(jid string) string {
	return filepath.Join(sr.Config.Server.Root, "jobs", jid)
}

func (sr *SimRun) JobRow(jid string) int {
	jt := sr.JobsTable.Column("JobID")
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
func (sr *SimRun) ValueForJob(jid, column string) string {
	if jrow := sr.JobRow(jid); jrow >= 0 {
		return sr.JobsTable.Column(column).String1D(jrow)
	}
	return ""
}

// Queue runs a queue query command on the server and shows the results.
func (sr *SimRun) Queue() { //types:add
	if sr.IsSlurm() {
		sr.QueueSlurm()
	} else {
		sr.QueueBare()
	}
}

// JobStatus gets job status from server for given job id.
// jobs that are already Finalized are skipped, unless force is true.
func (sr *SimRun) JobStatus(jid string, force bool) {
	// fmt.Println("############\nStatus of Job:", jid)
	spath := sr.ServerJobPath(jid)
	jpath := sr.JobPath(jid)
	goalrun.Run("@0")
	goalrun.Run("cd", jpath)
	if !goalib.FileExists("job.status") {
		goalib.WriteFile("job.status", "Unknown")
	}
	sstat := goalib.ReadFile("job.status")
	if !force && (sstat == "Finalized" || sstat == "Fetched") {
		return
	}
	if sr.IsSlurm() {
		goalrun.Run("@1")
		goalrun.Run("cd")
		goalrun.Run("cd", spath)
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
		core.MessageSnackbar(sr, "Retrieving job files for: "+jid)
		jfiles := goalrun.Output("@1", "/bin/ls", "-1", "job.*")
		for _, jf := range goalib.SplitLines(jfiles) {
			if !sr.IsSlurm() && jf == "job.status" {
				continue
			}
			// fmt.Println(jf)
			rfn := "@1:" + jf
			if !force {
				goalrun.Run("scp", rfn, jf)
			}
		}
	} else {
		jstr := strings.TrimSpace(goalrun.OutputErrOK("cat", "job.job"))
		if jstr == "" {
			msg := fmt.Sprintf("Status for Job: %s: job.job is empty, so can't proceed with BareMetal", jid)
			fmt.Println(msg)
			core.MessageSnackbar(sr, msg)
			return
		}
		sj := errors.Log1(strconv.Atoi(jstr))
		// fmt.Println(jid, "jobno:", sj)
		job := sr.BareMetal.Job(sj)
		if job == nil {
			core.MessageSnackbar(sr, fmt.Sprintf("Could not get BareMetal Job for: %s at job ID: %d", jid, sj))
		} else {
			sstat = job.Status.String()
			goalib.WriteFile("job.status", sstat)
			goalib.WriteFile("job.squeue", sstat)
			if !job.Start.IsZero() {
				goalib.WriteFile("job.start", job.Start.Format(sr.Config.TimeFormat))
			}
			if !job.End.IsZero() {
				goalib.WriteFile("job.end", job.End.Format(sr.Config.TimeFormat))
			}
			// fmt.Println(jid, sstat)
		}
		goalrun.Run("@0")
		if sstat == "Running" {
			// core.MessageSnackbar(sr, "Retrieving job files for: " + jid)
			// @1
			// cd
			// todo: need more robust ways of testing for files and error recovery on remote ssh con
			// cd {spath}
			// scp @1:job.out job.out
			// scp @1:nohup.out nohup.out
		}
	}
	goalrun.Run("@0")
	if sstat == "Done" || sstat == "Completed" {
		sstat = "Finalized"
		goalib.WriteFile("job.status", sstat)
		if sr.IsSlurm() {
			goalrun.RunErrOK("/bin/rm", "job.*.out")
		}
		sr.FetchJob(jid, false)
	}
	sr.GetMeta(jid)
	core.MessageSnackbar(sr, "Job: "+jid+" updated with status: "+sstat)
}

// GetMeta gets the dbmeta.toml file from all job.* files in job dir.
func (sr *SimRun) GetMeta(jid string) {
	goalrun.Run("@0")
	goalrun.Run("cd")
	jpath := sr.JobPath(jid)
	goalrun.Run("cd", jpath)
	// fmt.Println("getting meta for", jid)
	jfiles := goalrun.Output("/bin/ls", "-1", "job.*") // local
	meta := fmt.Sprintf("%s = %q\n", "Version", sr.Config.Version) + fmt.Sprintf("%s = %q\n", "Server", sr.Config.Server.Name)
	for _, jf := range goalib.SplitLines(jfiles) {
		if strings.Contains(jf, "sbatch") || strings.HasSuffix(jf, ".out") || strings.HasSuffix(jf, ".gz") {
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
}

// Status gets updated job.* files from the server for any job that
// doesn't have a Finalized or Fetched status.  It updates the
// status based on the server job status query, assigning a
// status of Finalized if job is done.  Updates the dbmeta.toml
// data based on current job data.
func (sr *SimRun) Status() { //types:add
	goalrun.Run("@0")
	sr.UpdateFiles()
	dpath := filepath.Join(sr.DataRoot, "jobs")
	ds := fsx.Dirs(dpath)
	for _, jid := range ds {
		sr.JobStatus(jid, false) // true = update all -- for format and status edits
	}
	core.MessageSnackbar(sr, "Jobs Status completed")
	sr.UpdateSims()
}

// FetchJob downloads results files from server.
// if force == true then will re-get already-Fetched jobs,
// otherwise these are skipped.
func (sr *SimRun) FetchJob(jid string, force bool) {
	if sr.IsSlurm() {
		sr.FetchJobSlurm(jid, force)
	} else {
		sr.FetchJobBare(jid, force)
	}
}

// Fetch retrieves all the .tsv data files from the server
// for any jobs not already marked as Fetched.
// Operates on the jobs selected in the Jobs table,
// or on all jobs if none selected.
func (sr *SimRun) Fetch() { //types:add
	goalrun.Run("@0")
	tv := sr.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		dpath := filepath.Join(sr.DataRoot, "jobs")
		jobs = fsx.Dirs(dpath)
	}
	for _, jid := range jobs {
		sr.FetchJob(jid, false)
	}
	core.MessageSnackbar(sr, "Fetch Jobs completed")
	sr.UpdateSims()
}

// Cancel cancels the jobs selected in the Jobs table,
// with a confirmation prompt.
func (sr *SimRun) Cancel() { //types:add
	tv := sr.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		core.MessageSnackbar(sr, "No jobs selected for cancel")
		return
	}
	lab.PromptOKCancel(sr, "Ok to cancel these jobs: "+strings.Join(jobs, " "), func() {
		if sr.IsSlurm() {
			sr.CancelJobsSlurm(jobs)
		} else {
			sr.CancelJobsBare(jobs)
		}
		sr.UpdateSims()
	})
}

// DeleteJobs deletes the given jobs
func (sr *SimRun) DeleteJobs(jobs []string) {
	goalrun.Run("@0")
	dpath := filepath.Join(sr.DataRoot, "jobs")
	spath := filepath.Join(sr.Config.Server.Root, "jobs")
	for _, jid := range jobs {
		goalrun.Run("@0")
		goalrun.Run("cd", dpath)
		goalrun.RunErrOK("/bin/rm", "-rf", jid)
		goalrun.Run("@1")
		goalrun.Run("cd")
		// todo: [cd {spath} && /bin/rm -rf {jid}]
		goalrun.Run("cd", spath, "&&", "/bin/rm", "-rf", jid)
		goalrun.Run("@0")
	}
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("@0")
	core.MessageSnackbar(sr, "Done deleting jobs")
}

// Delete deletes the selected Jobs, with a confirmation prompt.
func (sr *SimRun) Delete() { //types:add
	tv := sr.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		core.MessageSnackbar(sr, "No jobs selected for deletion")
		return
	}
	lab.PromptOKCancel(sr, "Ok to delete these jobs: "+strings.Join(jobs, " "), func() {
		sr.DeleteJobs(jobs)
		sr.UpdateSims()
	})
}

// ArchiveJobs archives the given jobs
func (sr *SimRun) ArchiveJobs(jobs []string) {
	goalrun.Run("@0")
	dpath := filepath.Join(sr.DataRoot, "jobs")
	apath := filepath.Join(sr.DataRoot, "archive", "jobs")
	goalrun.Run("mkdir", "-p", apath)
	spath := filepath.Join(sr.Config.Server.Root, "jobs")
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
	core.MessageSnackbar(sr, "Done archiving jobs")
}

// Archive moves the selected Jobs to the Archive directory,
// locally, and deletes them from the server,
// for results that are useful but not immediately relevant.
func (sr *SimRun) Archive() { //types:add
	tv := sr.JobsTableView
	jobs := tv.SelectedColumnStrings("JobID")
	if len(jobs) == 0 {
		core.MessageSnackbar(sr, "No jobs selected for archiving")
		return
	}
	lab.PromptOKCancel(sr, "Ok to archive these jobs: "+strings.Join(jobs, " "), func() {
		sr.ArchiveJobs(jobs)
		sr.UpdateSims()
	})
}
