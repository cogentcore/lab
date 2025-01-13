// Code generated by "goal build"; DO NOT EDIT.
//line jobs.goal:1
// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"bytes"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"cogentcore.org/core/base/errors"
)

// Status are the job status values.
type Status int32 //enums:enum

const (
	// NoStatus is the unknown status state.
	NoStatus Status = iota

	// Pending means the job has been submitted, but not yet run.
	Pending

	// Running means the job is running.
	Running

	// Completed means the job finished on its own, with no error status.
	Completed

	// Canceled means the job was canceled by the user.
	Canceled

	// Errored means the job quit with an error
	Errored
)

// Job is one bare metal job.
type Job struct {
	// ID is the overall baremetal unique ID number.
	ID int

	// Status is the current status of the job.
	Status Status

	// Source is info about the source of the job, e.g., simrun sim project.
	Source string

	// Path is the path from the SSH home directory to launch the job in.
	// This path will be
	Path string

	// Script is name of the job script to run, which must be at the top level
	// within the given tar file.
	Script string

	// Files is the gzipped tar file of the job files set at submission.
	Files []byte `display:"-"`

	// ResultsGlob is a glob expression for the result files to get back
	// from the server (e.g., *.tsv). job.out is automatically included as well,
	// which has the job stdout, stederr output.
	ResultsGlob string `display:"-"`

	// Results is the gzipped tar file of the job result files, gathered
	// at completion or when queried for results.
	Results []byte `display:"-"`

	// Submit is the time submitted.
	Submit time.Time

	// Start is the time actually started.
	Start time.Time

	// End is the time stopped running.
	End time.Time

	// ServerName is the name of the server it is running / ran on. Empty for pending.
	ServerName string

	// ServerGPU is the logical index of the GPU assigned to this job (0..N-1).
	ServerGPU int

	// pid is the process id of the job script.
	PID int
}

// Job returns the Job record for given job number; nil if not found
// in Active or Done;
func (bm *BareMetal) Job(jobno int) *Job {
	job, ok := bm.Active.AtTry(jobno)
	if ok {
		return job
	}
	job, ok = bm.Done.AtTry(jobno)
	if ok {
		return job
	}
	return nil
}

// Submit adds a new Active job with given parameters.
func (bm *BareMetal) Submit(src, path, script, results string, files []byte) *Job {
	job := &Job{ID: bm.NextID, Status: Pending, Source: src, Path: path, Script: script, Files: files, ResultsGlob: results, Submit: time.Now(), ServerGPU: -1}
	bm.NextID++
	bm.Active.Add(job.ID, job)
	bm.SaveState()
	return job
}

// RunJob runs the given job on the given server on given gpu number.
func (bm *BareMetal) RunJob(job *Job, sv *Server, gpu int) error {
	defer func() {
		goalrun.Run("cd")
		goalrun.Run("@0")
	}()
	sv.Use()
	goalrun.Run("cd")
	goalrun.Run("mkdir", "-p", job.Path)
	goalrun.Run("cd", job.Path)
	sshcl, err := goalrun.SSHByHost(sv.Name)
	if errors.Log(err) != nil {
		return err
	}
	b := bytes.NewReader(job.Files)
	sz := int64(len(job.Files))
	ctx := goalrun.StartContext()
	err = sshcl.CopyLocalToHost(ctx, b, sz, "job.files.tar.gz")
	goalrun.EndContext()
	if errors.Log(err) != nil {
		return err
	}
	goalrun.Run("tar", "-xzf", "job.files.tar.gz")
	// set BARE_GPU {gpu}
	gpus := strconv.Itoa(gpu)
	// $nohup {"./"+job.Script} > job.out 2>&1 & echo "$!" > job.pid $
	// note: anything with an & in it just doesn't work on our ssh client, for unknown reasons.
	// goalrun.Run("nohup", "./"+job.Script, ">&", "job.out", "&", "echo", "$!", ">", "job.pid")
	goalrun.Run("BARE_GPU="+gpus, "nohup", "./"+job.Script)
	for range 10 {
		if bm.GetJobPID(job) {
			break
		}
		time.Sleep(time.Second)
	}
	job.ServerName = sv.Name
	job.ServerGPU = gpu
	slog.Info("Job running on server", "Job:", job.ID, "Server:", sv.Name)
	return nil
}

// GetJobPID tries to get the job PID, returning true if obtained.
// Must already be in the ssh and directory for correct server.
func (bm *BareMetal) GetJobPID(job *Job) bool {
	pids := strings.TrimSpace(goalrun.Output("cat", "job.pid"))
	if pids != "" {
		pidn, err := strconv.Atoi(pids)
		if err == nil {
			job.PID = pidn
			return true
		}
	}
	return false
}

// RunPendingJobs runs any pending jobs if there are available GPUs to run on.
// returns number of jobs started, and any errors incurred in starting jobs.
func (bm *BareMetal) RunPendingJobs() (int, error) {
	avail := bm.AvailableGPUs()
	if len(avail) == 0 {
		return 0, nil
	}
	nRun := 0
	var errs []error
	for _, job := range bm.Active.Values {
		if job.Status != Pending {
			continue
		}
		fmt.Println("job status:", job.Status, "jobno:", job.ID)
		av := avail[0]
		sv := bm.Servers.At(av.Name)
		next := sv.NextGPU()
		for next < 0 {
			if len(avail) == 1 {
				return nRun, errors.Join(errs...)
			}
			avail = avail[1:]
			av = avail[0]
			sv = bm.Servers.At(av.Name)
			next = sv.NextGPU()
		}
		err := bm.RunJob(job, sv, next)
		if err != nil { // note: errors are server errors, not job errors, so don't affect job status
			sv.FreeGPU(next)
			errs = append(errs, err)
		} else {
			job.Status = Running
			nRun++
		}
	}
	if nRun > 0 {
		bm.SaveState()
	}
	return nRun, errors.Join(errs...)
}

// CancelJobs cancels list of job IDs. Returns error for jobs not found.
func (bm *BareMetal) CancelJobs(jobs ...int) error {
	var errs []error
	for _, jid := range jobs {
		job, ok := bm.Active.AtTry(jid)
		if !ok {
			err := errors.Log(fmt.Errorf("CancelJobs: job id not found in Active job list: %d", jid))
			errs = append(errs, err)
		} else {
			err := bm.CancelJob(job)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	bm.SaveState()
	return errors.Join(errs...)
}

// CancelJob cancels the running of the given job (killing process if Running).
func (bm *BareMetal) CancelJob(job *Job) error {
	if job.Status == Pending {
		job.Status = Canceled
		job.End = time.Now()
		bm.Done.Add(job.ID, job)
		bm.Active.DeleteByKey(job.ID)
		return nil
	}
	sv, err := bm.Server(job.ServerName)
	if errors.Log(err) != nil {
		return err
	}
	sv.Use()
	if job.PID == 0 {
		goalrun.Run("cd", job.Path)
		if !bm.GetJobPID(job) {
			return fmt.Errorf("CancelJob: Job %d PID is 0 and could not get it from job.pid file: must cancel manually", job.ID)
		}
	}
	goalrun.RunErrOK("kill", "-9", job.PID)
	job.Status = Canceled
	goalrun.Run("@0")
	bm.JobDone(job, sv)
	return nil
}

// PollJobs checks to see if any running jobs have finished.
// Returns number of jobs that finished.
func (bm *BareMetal) PollJobs() (int, error) {
	var errs []error
	nDone := 0
	for _, job := range bm.Active.Values {
		fmt.Println("job status:", job.Status, "jobno:", job.ID)
		if job.Status != Running {
			continue
		}
		sv, err := bm.Server(job.ServerName)
		if errors.Log(err) != nil {
			errs = append(errs, err)
			continue
		}
		sv.Use()
		goalrun.Output("cd", job.Path)
		if job.PID == 0 {
			if !bm.GetJobPID(job) {
				err := fmt.Errorf("PollJobs: Job %d PID is 0 and could not get it from job.pid file: must cancel manually", job.ID)
				errs = append(errs, errors.Log(err))
				continue
			}
		}
		// psout := $ps -p {job.PID} >/dev/null; echo "$?"$ // todo: don't parse ; ourselves!
		psout := strings.TrimSpace(goalrun.Output("ps", "-p", job.PID, ">", "/dev/null", ";", "echo", "$?"))
		// fmt.Println("status:", psout)
		if psout == "1" {
			job.Status = Completed
			bm.GetResults(job, sv)
			bm.JobDone(job, sv)
			nDone++
		}
	}
	goalrun.Output("@0")
	if nDone > 0 {
		bm.SaveState()
	}
	return nDone, errors.Join(errs...)
}

// JobDone sets job to be completed and moves to Done category.
func (bm *BareMetal) JobDone(job *Job, sv *Server) {
	job.End = time.Now()
	if job.ServerGPU >= 0 {
		sv.FreeGPU(job.ServerGPU)
	}
	bm.Done.Add(job.ID, job)
	bm.Active.DeleteByKey(job.ID)
}

// GetResults gets job results back from server.
func (bm *BareMetal) GetResults(job *Job, sv *Server) error {
	defer func() {
		goalrun.Run("cd")
		goalrun.Run("@0")
	}()
	sv.Use()
	goalrun.Run("cd")
	goalrun.Run("cd", job.Path)
	goalrun.Run("tar", "-czf", "job.results.tar.gz", job.ResultsGlob)
	var b bytes.Buffer
	sshcl, err := goalrun.SSHByHost(sv.Name)
	if errors.Log(err) != nil {
		return err
	}
	ctx := goalrun.StartContext()
	err = sshcl.CopyHostToLocal(ctx, "job.results.tar.gz", &b)
	goalrun.EndContext()
	if errors.Log(err) != nil {
		return err
	}
	job.Results = b.Bytes()
	return nil
}

// SetServerUsedFromJobs is called at startup to set the server Used status
// based on the current Active jobs, loaded from State.
func (bm *BareMetal) SetServerUsedFromJobs() error {
	for _, sv := range bm.Servers.Values {
		sv.Used = make(map[int]bool)
	}
	var errs []error
	for _, job := range bm.Active.Values {
		if job.Status != Running {
			continue
		}
		sv, err := bm.Server(job.ServerName)
		if errors.Log(err) != nil {
			errs = append(errs, err)
			continue
		}
		sv.Used[job.ServerGPU] = true
	}
	return errors.Join(errs...)
}
