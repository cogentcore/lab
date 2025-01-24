// Code generated by "goal build"; DO NOT EDIT.
//line bare.goal:1
// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/core"
	"cogentcore.org/lab/examples/baremetal"
	"cogentcore.org/lab/goal/goalib"
)

// bare supports the baremetal platform without slurm or other job management infra.

// SubmitBare submits a bare metal run job, returning the pid of the resulting process.
func (sr *SimRun) SubmitBare(jid, args string) string {
	goalrun.Run("@0")
	script := "job.sbatch"
	f, _ := os.Create(script)
	sr.WriteBare(f, jid, args)
	f.Close()
	goalrun.Run("chmod", "+x", script)

	files, err := baremetal.AllFiles("./", ".*")
	if errors.Log(err) != nil {
		return ""
	}
	// fmt.Println("files:", files)
	// todo: this is triggering a type defn mode:
	var b bytes.Buffer
	err = baremetal.TarFiles(&b, "./", true, files...)
	if errors.Log(err) != nil {
		return ""
	}
	bm := sr.BareMetal

	spath := sr.ServerJobPath(jid)
	job, err := bm.Submit(sr.Config.Project, spath, script, sr.Config.FetchFiles, b.Bytes())
	goalrun.Run("@0")
	if err != nil {
		core.ErrorSnackbar(sr, err)
		return "-1"
	}
	bid := strconv.Itoa(job.ID)
	goalib.WriteFile("job.job", bid)
	bm.UpdateJobs()
	sr.GetMeta(jid)
	return bid
}

// WriteBare writes the bash script to run a "bare metal" run.
func (sr *SimRun) WriteBare(w io.Writer, jid, args string) {
	if sr.Config.JobScript != "" {
		js := sr.Config.JobScript
		strings.ReplaceAll(js, "$JOB_ARGS", args)
		fmt.Fprintln(w, js)
		return
	}
	fmt.Fprintf(w, "#!/bin/bash -l\n") //  -l = login session, sources your .bash_profile

	fmt.Fprintf(w, "\n\n")
	if sr.Config.SetupScript != "" {
		fmt.Fprintln(w, sr.Config.SetupScript)
	}

	// fmt.Fprintf(w, "go build -mod=mod -tags mpi\n")
	fmt.Fprintf(w, "go build -mod=mod\n")
	cmd := `date '+%Y-%m-%d %T %Z' > job.start`
	fmt.Fprintln(w, cmd)

	fmt.Fprintf(w, "./%s -nogui -cfg config_job.toml -gpu-device $BARE_GPU %s >& job.out & echo $! > job.pid", sr.Config.Project, args)
}

func (sr *SimRun) QueueBare() {
	sr.UpdateBare()
	ts := sr.Tabs.AsLab()
	goalrun.Run("@1")
	goalrun.Run("cd")
	smi := goalrun.Output("nvidia-smi")
	goalrun.Run("@0")
	ts.EditorString("Queue", smi)
}

// UpdateBare updates the BareMetal jobs
func (sr *SimRun) UpdateBare() { //types:add
	// nrun, nfin := errors.Log2(sr.BareMetal.UpdateJobs())
	// core.MessageSnackbar(sr, fmt.Sprintf("BareMetal jobs run: %d finished: %d", nrun, nfin))
}

// FetchJobBare downloads results files from bare metal server.
func (sr *SimRun) FetchJobBare(jid string, force bool) {
	jpath := sr.JobPath(jid)
	goalrun.Run("@0")
	goalrun.Run("cd", jpath)
	sstat := goalib.ReadFile("job.status")
	if !force && sstat == "Fetched" {
		return
	}
	sjob := sr.ValueForJob(jid, "ServerJob")
	sj := errors.Log1(strconv.Atoi(sjob))
	jobs, err := sr.BareMetal.FetchResults(sj)
	if err != nil {
		core.ErrorSnackbar(sr, err)
		return
	}
	job := jobs[0]
	baremetal.Untar(bytes.NewReader(job.Results), jpath, true) // gzip
	// note: we don't do any post-processing here -- see slurm version for combining separate runs
	if sstat == "Finalized" {
		// fmt.Println("status finalized")
		goalib.WriteFile("job.status", "Fetched")
		goalib.ReplaceInFile("dbmeta.toml", "\"Finalized\"", "\"Fetched\"")
	} else {
		fmt.Println("status:", sstat)
	}
}

func (sr *SimRun) CancelJobsBare(jobs []string) {
	jnos := make([]int, 0, len(jobs))
	for _, jid := range jobs {
		sjob := sr.ValueForJob(jid, "ServerJob")
		if sjob != "" {
			jno := errors.Log1(strconv.Atoi(sjob))
			jnos = append(jnos, jno)
		}
	}
	sr.BareMetal.CancelJobs(jnos...)
}
