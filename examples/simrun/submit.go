// Code generated by "goal build"; DO NOT EDIT.
//line submit.goal:1
// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/core"
	"cogentcore.org/lab/goal/goalib"
	"cogentcore.org/lab/lab"
)

// NextJobNumber returns the next sequential job number to use,
// incrementing value saved in last_job.number file
func (br *SimRun) NextJobNumber() int {
	jf := "last_job.number"
	jnf := goalib.ReadFile(jf)
	jn := 0
	if jnf != "" {
		jn, _ = strconv.Atoi(strings.TrimSpace(jnf))
	}
	jn++
	goalib.WriteFile(jf, strconv.Itoa(jn))
	return jn
}

func (br *SimRun) NextJobID() string {
	jn := br.NextJobNumber()
	jstr := fmt.Sprintf("%s%05d", br.Config.UserShort, jn)
	return jstr
}

// FindGoMod finds the go.mod file starting from the given directory
func (br *SimRun) FindGoMod(dir string) string {
	for {
		if goalib.FileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		dir = filepath.Dir(dir)
		if dir == "" {
			return ""
		}
	}
	return ""
}

// GoModulePath returns the overall module path for project
// in given directory, and the full module path to the current
// project, which is a subdirectory within the module.
func (br *SimRun) GoModulePath(dir string) (modpath, fullpath string) {
	goalrun.Run("@0")
	os.Chdir(dir)
	goalrun.Run("cd", dir)
	gg := goalib.SplitLines(goalrun.Output("go", "mod", "graph"))
	gg = strings.Fields(gg[0])
	modpath = gg[0]

	// strategy: go up the dir until the dir name matches the last element of modpath
	dirsp := strings.Split(dir, "/")
	n := len(dirsp)
	for i := n - 1; i >= 0; i-- {
		d := dirsp[i]
		if strings.HasSuffix(modpath, d) {
			fullpath = filepath.Join(modpath, strings.Join(dirsp[i+1:], "/"))
			break
		}
	}
	return
}

// CopyFilesToJob copies files with given extensions (none for all),
// from localSrc to localJob and remote hostJob (@1).
// Ensures directories are made in the job locations
func (br *SimRun) CopyFilesToJob(localSrc, localJob, hostJob string, exts ...string) {
	goalrun.Run("@0")
	goalrun.Run("mkdir", "-p", localJob)
	goalrun.Run("cd", localJob)
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("mkdir", "-p", hostJob)
	goalrun.Run("cd", hostJob)
	goalrun.Run("@0")
	efls := fsx.Filenames(localSrc, exts...)
	for _, f := range efls {
		sfn := filepath.Join(localSrc, f)
		goalrun.Run("/bin/cp", sfn, f)
		goalrun.Run("scp", sfn, "@1:"+f)
	}
}

// NewJob runs a new job with given parameters.
// This is run as a separate goroutine!
func (br *SimRun) NewJob(jp SubmitParams) {
	message := jp.Message
	args := jp.Args
	label := jp.Label

	os.Chdir(br.DataRoot)
	jid := br.NextJobID()
	spath := br.ServerJobPath(jid)
	jpath := br.JobPath(jid)
	// this might cause crashing:
	br.AsyncLock()
	core.MessageSnackbar(br, "Submitting Job: "+jid)
	br.AsyncUnlock()

	gomodDir := br.FindGoMod(br.StartDir)
	_, fullPath := br.GoModulePath(br.StartDir)
	projPath := filepath.Join("emer", br.Config.Project)

	// fmt.Println("go.mod:", gomodDir, "\nmodule:", modulePath, "\nfull path:", fullPath, "\njob proj:", projPath)

	goalrun.Run("@0")
	// fmt.Println(jpath)
	os.MkdirAll(jpath, 0750)
	os.Chdir(jpath)
	goalib.WriteFile("job.message", message)
	goalib.WriteFile("job.args", args)
	goalib.WriteFile("job.label", label)
	goalib.WriteFile("job.submit", time.Now().Format(br.Config.TimeFormat))
	goalib.WriteFile("job.status", "Submitted")

	// need to do sub-code first and update paths in copied files
	codepaths := make([]string, len(br.Config.CodeDirs))
	for i, ed := range br.Config.CodeDirs {
		goalrun.Run("@0")
		loce := filepath.Join(br.StartDir, ed)
		codepaths[i] = filepath.Join(fullPath, ed)
		jpathe := filepath.Join(jpath, ed)
		spathe := filepath.Join(spath, ed)
		br.CopyFilesToJob(loce, jpathe, spathe, ".go")
	}
	// copy local files:
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("mkdir", "-p", spath)
	goalrun.Run("cd", spath)
	goalrun.Run("@0")
	goalrun.Run("cd", jpath)
	fls := fsx.Filenames(br.StartDir, ".go")
	for _, f := range fls {
		sfn := filepath.Join(br.StartDir, f)
		goalrun.Run("/bin/cp", sfn, f)
		for i, ed := range br.Config.CodeDirs {
			subpath := filepath.Join(projPath, ed)
			// fmt.Println("replace in:", f, codepaths[i], "->", subpath)
			goalib.ReplaceInFile(f, codepaths[i], subpath)
		}
		goalrun.Run("scp", f, "@1:"+f)
	}
	for _, f := range br.Config.ExtraFiles {
		sfn := filepath.Join(br.StartDir, f)
		goalrun.Run("/bin/cp", sfn, f)
		goalrun.Run("scp", sfn, "@1:"+f)
	}
	for _, ed := range br.Config.ExtraDirs {
		jpathe := filepath.Join(jpath, ed)
		spathe := filepath.Join(spath, ed)
		loce := filepath.Join(br.StartDir, ed)
		br.CopyFilesToJob(loce, jpathe, spathe)
	}
	goalrun.Run("@1")
	goalrun.Run("cd")
	goalrun.Run("cd", spath)
	goalrun.Run("@0")
	goalrun.Run("cd", jpath)

	br.AsyncLock()
	core.MessageSnackbar(br, "Job: "+jid+" files copied")
	br.AsyncUnlock()

	if gomodDir != "" {
		sfn := filepath.Join(gomodDir, "go.mod")
		// fmt.Println("go.mod dir:", gomodDir, sfn)
		goalrun.Run("scp", sfn, "@1:go.mod")
		sfn = filepath.Join(gomodDir, "go.sum")
		goalrun.Run("scp", sfn, "@1:go.sum")
		goalrun.Run("@1")
		goalrun.Run("go", "mod", "edit", "-module", projPath)
		if br.Config.Package != "" {
			goalrun.Run("go", "get", br.Config.Package+"@"+br.Config.Version)
		}
		if br.Config.ExtraGoGet != "" {
			goalrun.Run("go", "get", br.Config.ExtraGoGet)
		}
		goalrun.Run("go", "mod", "tidy")
		goalrun.Run("@0")
		goalrun.Run("scp", "@1:go.mod", "go.mod")
		goalrun.Run("scp", "@1:go.sum", "go.sum")
	} else {
		fmt.Println("go.mod file not found!")
	}

	if br.Config.Server.Slurm {
		sid := br.SubmitSBatch(jid, args)
		goalib.WriteFile("job.job", sid)
		fmt.Println("server job id:", sid)
		goalrun.Run("scp", "job.job", "@1:job.job")
		core.MessageSnackbar(br, "Job: "+jid+" server job: "+sid+" successfully submitted")
	} else {
		sid := br.SubmitRun(jid, args, jp.GPU)
		goalib.WriteFile("job.job", sid)
		fmt.Println("server job id:", sid)
		goalrun.Run("scp", "job.job", "@1:job.job")
		br.AsyncLock()
		core.MessageSnackbar(br, "Job: "+jid+" server job: "+sid+" successfully submitted")
		br.AsyncUnlock()
	}
	goalrun.Run("@1", "cd")
	goalrun.Run("@0")
	br.UpdateSimsAsync()
}

// Submit submits a job to SLURM on the server, using an array
// structure, with an outer startup job that calls the main array
// jobs and a final cleanup job.  Creates a new job dir based on
// incrementing counter, synchronizing the job files.
func (br *SimRun) Submit() { //types:add
	lab.PromptStruct(br, &br.Config.Submit, "Submit a new job", func() {
		go br.NewJob(br.Config.Submit)
	})
}
