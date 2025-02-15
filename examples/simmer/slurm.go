// Code generated by "goal build"; DO NOT EDIT.
//line slurm.goal:1
// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cogentcore.org/core/core"
	"cogentcore.org/lab/goal/goalib"
	"cogentcore.org/lab/table"
)

// WriteSBatchHeader writes the header of a SLURM SBatch script
// that is common across all three scripts.
// IMPORTANT: set the job parameters here!
func (sr *Simmer) WriteSBatchHeader(w io.Writer, jid string) {
	fmt.Fprintf(w, "#SBATCH --job-name=%s_%s\n", sr.Config.Project, jid)
	fmt.Fprintf(w, "#SBATCH --mem-per-cpu=%dG\n", sr.Config.Job.Memory)
	fmt.Fprintf(w, "#SBATCH --time=%d:00:00\n", sr.Config.Job.Hours)
	fmt.Fprintf(w, "#SBATCH --ntasks=%d\n", sr.Config.Job.Tasks)
	fmt.Fprintf(w, "#SBATCH --cpus-per-task=%d\n", sr.Config.Job.CPUsPerTask)
	fmt.Fprintf(w, "#SBATCH --ntasks-per-node=%d\n", sr.Config.Job.TasksPerNode)
	if sr.Config.ExcludeNodes != "" {
		fmt.Fprintf(w, "#SBATCH --exclude=%s\n", sr.Config.ExcludeNodes)
	}
	// fmt.Fprint(w, "#SBATCH --nodelist=agate-[2,19]\n")
	// fmt.Fprintf(w, "#SBATCH --qos=%s\n", qos)
	// fmt.Fprintf(w, "#SBATCH --partition=%s\n", qosShort)
	fmt.Fprintf(w, "#SBATCH --mail-type=FAIL\n")
	fmt.Fprintf(w, "#SBATCH --mail-user=%s\n", sr.Config.User)
	// these might be needed depending on environment in head node vs. compute nodes
	// fmt.Fprintf(w, "#SBATCH --export=NONE\n")
	// fmt.Fprintf(w, "unset SLURM_EXPORT_ENV\n")
}

func (sr *Simmer) WriteSBatchSetup(w io.Writer, jid string) {
	fmt.Fprintf(w, "#!/bin/bash -l\n") //  -l = login session, sources your .bash_profile
	fmt.Fprint(w, "#SBATCH --output=job.setup.out\n")
	fmt.Fprint(w, "#SBATCH --error=job.setup.err\n")
	sr.WriteSBatchHeader(w, jid)

	//////////////////////////////////////////////////////////
	// now we do all the setup, like building the executable

	fmt.Fprintf(w, "\n\n")
	// fmt.Fprintf(w, "go build -mod=mod -tags mpi\n")
	fmt.Fprintf(w, "go build -mod=mod\n")
	// fmt.Fprintf(w, "/bin/rm images\n")
	// fmt.Fprintf(w, "ln -s $HOME/ccn_images/CU3D100_20obj8inst_8tick4sac images\n")
	cmd := "date '+%Y-%m-%d %T %Z' > job.start"
	fmt.Fprintln(w, cmd)
}

func (sr *Simmer) WriteSBatchArray(w io.Writer, jid, setup_id, args string) {
	fmt.Fprintf(w, "#!/bin/bash -l\n") //  -l = login session, sources your .bash_profile
	fmt.Fprintf(w, "#SBATCH --array=0-%d\n", sr.Config.Job.NRuns-1)
	fmt.Fprint(w, "#SBATCH --output=job.%A_%a.out\n")
	// fmt.Fprint(w, "#SBATCH --error=job.%A_%a.err\n")
	fmt.Fprintf(w, "#SBATCH --dependency=afterany:%s\n", setup_id)
	sr.WriteSBatchHeader(w, jid)

	//////////////////////////////////////////////////////////
	// now we run the job

	fmt.Fprintf(w, "echo $SLURM_ARRAY_JOB_ID\n")
	fmt.Fprintf(w, "\n\n")
	// note: could use srun to run job; -runs = 1 is number to run from run start
	fmt.Fprintf(w, "./%s -nogui -cfg config_job.toml -run $SLURM_ARRAY_TASK_ID -runs 1 %s\n", sr.Config.Project, args)
}

func (sr *Simmer) WriteSBatchCleanup(w io.Writer, jid, array_id string) {
	fmt.Fprintf(w, "#!/bin/bash -l\n") //  -l = login session, sources your .bash_profile
	fmt.Fprint(w, "#SBATCH --output=job.cleanup.out\n")
	// fmt.Fprint(w, "#SBATCH --error=job.cleanup.err")
	fmt.Fprintf(w, "#SBATCH --dependency=afterany:%s\n", array_id)
	sr.WriteSBatchHeader(w, jid)
	fmt.Fprintf(w, "\n\n")

	//////////////////////////////////////////////////////////
	// now we cleanup after all the jobs have run
	//	can cat results files etc.

	fmt.Fprintf(w, "cat job.*.out > job.out\n")
	fmt.Fprintf(w, "/bin/rm job.*.out\n")

	fmt.Fprintf(w, "cat *_train_run.tsv > all_run.tsv\n")
	fmt.Fprintf(w, "/bin/rm *_train_run.tsv\n")

	fmt.Fprintf(w, "cat *_train_epoch.tsv > all_epc.tsv\n")
	fmt.Fprintf(w, "/bin/rm *_train_epoch.tsv\n")

	cmd := "date '+%Y-%m-%d %T %Z' > job.end"
	fmt.Fprintln(w, cmd)
}

func (sr *Simmer) SubmitSBatch(jid, args string) string {
	goalrun.Run("@0")
	f, _ := os.Create("job.setup.sbatch")
	sr.WriteSBatchSetup(f, jid)
	f.Close()
	goalrun.Run("scp", "job.setup.sbatch", "@1:job.setup.sbatch")
	sid := sr.RunSBatch("job.setup.sbatch")

	f, _ = os.Create("job.sbatch")
	sr.WriteSBatchArray(f, jid, sid, args)
	f.Close()
	goalrun.Run("scp", "job.sbatch", "@1:job.sbatch")
	aid := sr.RunSBatch("job.sbatch")

	f, _ = os.Create("job.cleanup.sbatch")
	sr.WriteSBatchCleanup(f, jid, aid)
	f.Close()
	goalrun.Run("scp", "job.cleanup.sbatch", "@1:job.cleanup.sbatch")
	sr.RunSBatch("job.cleanup.sbatch")
	sr.GetMeta(jid)
	return aid
}

// RunSBatch runs sbatch on the given sbatch file,
// returning the resulting job id.
func (sr *Simmer) RunSBatch(sbatch string) string {
	goalrun.Run("@1")
	goalrun.Run("sbatch", sbatch, ">", "job.slurm")
	goalrun.Run("@0")
	ss := goalrun.Output("@1", "cat", "job.slurm")
	if ss == "" {
		fmt.Println("JobStatus ERROR: no server job.slurm file to get server job id from")
		goalrun.Run("@1", "cd")
		goalrun.Run("@0")
		return ""
	}
	ssf := strings.Fields(ss)
	sj := ssf[len(ssf)-1]
	return sj
}

// QueueSlurm runs a queue query command on the server and shows the results.
func (sr *Simmer) QueueSlurm() {
	ts := sr.Tabs.AsLab()
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

func (sr *Simmer) FetchJobSlurm(jid string, force bool) {
	spath := sr.ServerJobPath(jid)
	jpath := sr.JobPath(jid)
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
	ffiles := goalrun.Output("@1", "/bin/ls", "-1", sr.Config.FetchFiles)
	if len(ffiles) > 0 {
		core.MessageSnackbar(sr, fmt.Sprintf("Fetching %d data files for job: %s", len(ffiles), jid))
	}
	for _, ff := range goalib.SplitLines(ffiles) {
		// fmt.Println(ff)
		rfn := "@1:" + ff
		goalrun.Run("scp", rfn, ff)
		if (sstat == "Finalized" || sstat == "Fetched") && strings.HasSuffix(ff, ".tsv") {
			if ff == "all_epc.tsv" {
				table.CleanCatTSV(ff, "Run", "Epoch")
				idx := strings.Index(ff, "_epc.tsv")
				goalrun.Run("tablecat", "-colavg", "-col", "Epoch", "-o", ff[:idx+1]+"avg"+ff[idx+1:], ff)
			} else if ff == "all_run.tsv" {
				table.CleanCatTSV(ff, "Run")
				idx := strings.Index(ff, "_run.tsv")
				goalrun.Run("tablecat", "-colavg", "-o", ff[:idx+1]+"avg"+ff[idx+1:], ff)
				//	} else {
				//		table.CleanCatTSV(ff, "Run")
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

// CancelJobsSlurm cancels the given jobs, for slurm
func (sr *Simmer) CancelJobsSlurm(jobs []string) {
	goalrun.Run("@0")
	filepath.Join(sr.DataRoot, "jobs")
	filepath.Join(sr.Config.Server.Root, "jobs")
	goalrun.Run("@1")
	for _, jid := range jobs {
		sjob := sr.ValueForJob(jid, "ServerJob")
		if sjob != "" {
			goalrun.Run("scancel", sjob)
		}
	}
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("@0")
	core.MessageSnackbar(sr, "Done canceling jobs")
}
