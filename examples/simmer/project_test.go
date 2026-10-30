// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"cogentcore.org/core/base/iox/jsonx"
	"cogentcore.org/lab/goal/interpreter"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestSimmer returns a new Simmer, with a temporary home directory
// and working directory.
func newTestSimmer(t *testing.T) *Simmer {
	t.Setenv("HOME", t.TempDir())
	homedir.Reset() // which goal uses, caching the home directory
	t.Chdir(t.TempDir())
	in := interpreter.NewInterpreter(interpreter.Options{})
	_, sr := NewSimmerWindow(in)
	return sr
}

// TestNewProject tests making a new project, which is opened, and saving
// its settings.
func TestNewProject(t *testing.T) {
	sr := newTestSimmer(t)
	sim := filepath.Join(t.TempDir(), "mysim")
	require.NoError(t, os.Mkdir(sim, 0755))
	require.NoError(t, sr.newProject(sim))

	usr, err := user.Current()
	require.NoError(t, err)
	data := filepath.Join(os.Getenv("HOME"), "simdata", "mysim", usr.Username)
	assert.DirExists(t, filepath.Join(data, "labscripts"))
	link, err := os.Readlink(filepath.Join(sim, "simdata"))
	require.NoError(t, err)
	assert.Equal(t, data, link)
	fn := filepath.Join(sim, ProjectFilename)
	assert.FileExists(t, fn)

	assert.True(t, sr.isOpen())
	assert.Equal(t, "mysim", sr.Project.Name)
	assert.Equal(t, sim, sr.Project.StartDir)
	assert.Equal(t, filepath.Join(sim, "simdata"), sr.DataRoot)
	assert.Equal(t, filepath.Join("simserver", "mysim", sr.Project.User), sr.Project.Server.Root)
	assert.Contains(t, []string(RecentProjects), fn)

	sr.Project.Job.NRuns = 7
	sr.SaveProject()
	var pr Project
	require.NoError(t, jsonx.Open(&pr, fn))
	assert.Equal(t, 7, pr.Job.NRuns)
}

// TestMigrateProject tests opening a project made before simmer.json,
// whose settings are moved from its defaults.goal script.
func TestMigrateProject(t *testing.T) {
	sr := newTestSimmer(t)
	sim := filepath.Join(t.TempDir(), "oldsim")
	scripts := filepath.Join(sim, "simdata", "labscripts")
	require.NoError(t, os.MkdirAll(scripts, 0755))
	defaults := `// primary remote server: avail as @1
gossh ubuntu@127.0.0.1:1
@0

func defaults() {
	cf := &lab.LabBrowser.Config
	cf.Package = "github.com/emer/axon/v2"
	cf.Server.Name = "h100"
	cf.Job.NRuns = 25
	cf.ExtraFiles = []string{"config_job.toml", "random_5x5_25.tsv"}
	cf.Update()
}

defaults()
`
	dfn := filepath.Join(scripts, "defaults.goal")
	require.NoError(t, os.WriteFile(dfn, []byte(defaults), 0644))

	// opened by any file in it
	require.NoError(t, os.WriteFile(filepath.Join(sim, "main.go"), nil, 0644))
	fn := projectFile(filepath.Join(sim, "main.go"))
	assert.Equal(t, filepath.Join(sim, ProjectFilename), fn)
	assert.Equal(t, fn, projectFile(sim))
	assert.Equal(t, fn, projectFile(fn))
	assert.Equal(t, fn, projectFile(scripts), "in a directory of the project")
	require.NoError(t, sr.openProject(fn))
	pr := &sr.Project
	assert.Equal(t, "oldsim", pr.Name)
	assert.Equal(t, "ubuntu@127.0.0.1:1", pr.Server.SSH)
	assert.Equal(t, "github.com/emer/axon/v2", pr.Package)
	assert.Equal(t, "h100", pr.Server.Name)
	assert.Equal(t, 25, pr.Job.NRuns)
	assert.Equal(t, []string{"config_job.toml", "random_5x5_25.tsv"}, pr.ExtraFiles)
	assert.NoFileExists(t, dfn, "not run again")
	assert.FileExists(t, dfn+".old")

	var saved Project
	require.NoError(t, jsonx.Open(&saved, filepath.Join(sim, ProjectFilename)))
	assert.Equal(t, 25, saved.Job.NRuns)
	assert.Equal(t, "ubuntu@127.0.0.1:1", saved.Server.SSH)
}

// TestProjectArg tests taking the project from the command line.
func TestProjectArg(t *testing.T) {
	dir := t.TempDir()
	defer func(args []string) { os.Args = args }(os.Args)
	os.Args = []string{"simmer", dir, "-e", "x"}
	assert.Equal(t, dir, projectArg())
	assert.Equal(t, []string{"simmer", "-e", "x"}, os.Args)
	os.Args = []string{"simmer", "-e", "x"}
	assert.Equal(t, "", projectArg())
	os.Args = []string{"simmer", "script.goal"}
	assert.Equal(t, "", projectArg())
}
