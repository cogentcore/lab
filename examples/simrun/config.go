// Code generated by "goal build"; DO NOT EDIT.
//line config.goal:1
// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"strings"

	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/table"
)

// FilterResults specifies which results files to open.
type FilterResults struct {
	// File name contains this string, e.g., "_epc" or "_run"
	FileContains string `width:"60"`

	// extension of files, e.g., .tsv
	Ext string
}

func (fp *FilterResults) Defaults() {
	fp.FileContains = "_avgepc"
	fp.Ext = ".tsv"
}

// SubmitParams specifies the parameters for submitting a job.
type SubmitParams struct {

	// Message describing the simulation:
	// this is key info for what is special about this job, like a github commit message
	Message string `width:"80"`

	// Label is brief, unique label used for plots to label this job
	Label string `width:"80"`

	//	arguments to pass on the command line.
	//
	// -nogui is already passed by default
	Args string `width:"80"`
}

// JobParams are parameters for running the job
type JobParams struct {

	// number of parallel runs; can also set to 1 and run multiple runs per job using args
	NRuns int

	// max number of hours: slurm will terminate if longer, so be generous
	// 2d = 48, 3d = 72, 4d = 96, 5d = 120, 6d = 144, 7d = 168
	Hours int

	// memory per CPU in gigabytes
	Memory int

	// number of mpi "tasks" (procs in MPI terminology)
	Tasks int

	// number of cpu cores (threads) per task
	CPUsPerTask int

	// how to allocate tasks within compute nodes
	// cpus_per_task * tasks_per_node <= total cores per node
	TasksPerNode int

	// qos is the queue "quality of service" name
	Qos string
}

func (jp *JobParams) Defaults() {
	jp.NRuns = 10
	jp.Hours = 1
	jp.Memory = 1
	jp.Tasks = 1
	jp.CPUsPerTask = 8
	jp.TasksPerNode = 1
}

// Configuration holds all of the user-settable parameters
type Configuration struct {

	// DataRoot is the path to the root of the data to browse.
	DataRoot string

	// StartDir is the starting directory, where the app was originally started.
	StartDir string

	// User id as in system login name (i.e., user@system)
	User string

	// first 3 letters of User, for naming jobs (auto-set from User)
	UserShort string

	// name of simulation project, lowercase (should be name of source dir)
	Project string

	// current git version string, from git describe --tags
	Version string

	// parameters for job resources etc
	Job JobParams `display:"inline"`

	// glob expression for files to fetch from server, for Fetch command,
	// is *.tsv by default
	FetchFiles string

	// nodes to exclude from job, based on what is slow
	ExcludeNodes string

	// extra files to upload with job submit, from same dir
	ExtraFiles []string

	// subdirs with other files to upload with job submit (non-code -- see CodeDirs)
	ExtraDirs []string

	// subdirs with code to upload with job submit; go.mod auto-updated to use
	CodeDirs []string

	// name of current server using to run jobs; gets recorded with each job
	ServerName string

	// ExtraGoGet is an extra package to do "go get" with, for launching the job.
	// Typically set this to the parent packge if running within a larger package
	// upon which this simulation depends, e.g., "github.com/emer/axon/v2@main"
	ExtraGoGet string

	// root path from user home dir on server.
	// is auto-set to: filepath.Join("simdata", Project, User)
	ServerRoot string

	// format for timestamps, defaults to "2006-01-02 15:04:05 MST"
	TimeFormat string

	// parameters for filtering results
	Filter FilterResults

	// parameters for submitting jobs; set from last job run
	Submit SubmitParams
}

func (cf *Configuration) Defaults() {
	goalrun.Run("@0")
	goalrun.Run("cd")
	goalrun.Run("cd", cf.StartDir)
	cf.Version = strings.TrimSpace(goalrun.Output("git", "describe", "--tags"))
	goalrun.Run("cd", cf.DataRoot)
	cf.User = strings.TrimSpace(goalrun.Output("echo", "$USER"))
	_, pj := filepath.Split(cf.StartDir)
	cf.Project = pj
	cf.Job.Defaults()
	cf.FetchFiles = "*.tsv"
	cf.Filter.Defaults()
	cf.TimeFormat = "2006-01-02 15:04:05 MST"
}

func (cf *Configuration) Update() {
	cf.UserShort = cf.User[:3]
	cf.ServerRoot = filepath.Join("simdata", cf.Project, cf.User)
}

// Result has info for one loaded result, as a table.Table
type Result struct {

	// job id for results
	JobID string

	// short label used as a legend in the plot
	Label string

	// description of job
	Message string

	// args used in running job
	Args string

	// path to data
	Path string

	// result data
	Table *table.Table
}

// EditConfig edits the configuration
func (br *SimRun) EditConfig() { //types:add
	lab.PromptStruct(br, &br.Config, "Configuration parameters", nil)
}
