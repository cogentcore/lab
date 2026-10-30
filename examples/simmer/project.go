// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/iox/jsonx"
	"cogentcore.org/core/base/sshclient"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/table"
)

// ProjectFilename is the name of the file with the settings of a simulation
// project (see [Project]), in the directory of the simulation.
const ProjectFilename = "simmer.json"

var (
	// RecentProjects are the simmer.json files of the recently opened projects.
	RecentProjects core.FilePaths

	// recentProjectsFilename is the name of the file of the RecentProjects,
	// in the app data directory.
	recentProjectsFilename = "recent-projects.json"
)

// saveRecentProjects saves the [RecentProjects] in the app data directory.
func saveRecentProjects() {
	RecentProjects.Save(filepath.Join(core.TheApp.AppDataDir(), recentProjectsFilename))
}

// openRecentProjects opens the [RecentProjects] from the app data directory.
func openRecentProjects() {
	RecentProjects.Open(filepath.Join(core.TheApp.AppDataDir(), recentProjectsFilename))
}

// projectFile returns the project file for the given path, which is the
// path itself if it is a simmer.json file, and otherwise the simmer.json
// file in the directory of the project of the path: the path itself if it
// is a directory, or the directory of any other file, or the nearest
// directory above it with a simmer.json file or simdata directory, if any.
// A project that does not have a simmer.json file yet is opened in this
// way, by any file in it.
func projectFile(path string) string {
	path = errors.Log1(filepath.Abs(path))
	if filepath.Base(path) == ProjectFilename {
		return path
	}
	if !isDir(path) {
		path = filepath.Dir(path)
	}
	for dir := path; ; {
		if isFile(filepath.Join(dir, ProjectFilename)) || isDir(filepath.Join(dir, "simdata")) {
			return filepath.Join(dir, ProjectFilename)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Join(path, ProjectFilename)
}

// isDir returns whether the given path is a directory, without logging
// an error if it does not exist.
func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// isFile returns whether the given path is a file, without logging
// an error if it does not exist.
func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// projectArg removes and returns the project of the command line, if the
// first argument is a simmer.json file or a directory, so that the other
// arguments are those of the interpreter. A new app is opened with the
// project in this way (see [Simmer.OpenProject]).
func projectArg() string {
	if len(os.Args) < 2 {
		return ""
	}
	arg := os.Args[1]
	if strings.HasPrefix(arg, "-") || arg == "run" || arg == "build" || arg == "help" {
		return ""
	}
	if filepath.Base(arg) != ProjectFilename && !isDir(arg) {
		return ""
	}
	os.Args = append(os.Args[:1], os.Args[2:]...)
	return arg
}

// isOpen returns whether a project is open.
func (sr *Simmer) isOpen() bool {
	return sr.Project.StartDir != ""
}

// OpenProject opens the simulation project of the given simmer.json file,
// or of the directory with it, which can be given by any file in it, as
// for a project that does not have that file yet (see [Simmer.openProject]).
// If a project is already open, the project is opened in a new simmer app,
// as each app has one project.
func (sr *Simmer) OpenProject(path core.Filename) { //types:add
	fn := projectFile(string(path))
	if sr.isOpen() {
		if err := openInNewApp(fn); err != nil {
			core.ErrorSnackbar(sr, err)
		}
		return
	}
	if err := sr.openProject(fn); err != nil {
		errors.Log(err)
		core.ErrorSnackbar(sr, err)
	}
}

// openInNewApp opens the given project file in a new simmer app.
func openInNewApp(fn string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, fn)
	cmd.Dir = filepath.Dir(fn)
	return cmd.Start()
}

// openProject opens the project of the given simmer.json file. A project
// without that file, made before it was used, is opened if it has a
// simdata directory, and its settings are moved to the file from its
// defaults.goal script (see [Simmer.migrateProject]).
func (sr *Simmer) openProject(fn string) error {
	dir := filepath.Dir(fn)
	hasFile := isFile(fn)
	if !hasFile && !isDir(filepath.Join(dir, "simdata")) {
		return fmt.Errorf("%s is not a simmer project: it has no %s file, or simdata directory (see New project)", dir, ProjectFilename)
	}
	sr.StartDir = dir
	errors.Log(os.Chdir(dir))
	goalrun.Run("@0")
	goalrun.Run("cd", dir)
	sr.SetDataRoot(filepath.Join(dir, "simdata"))
	sr.SetScriptsDir(filepath.Join(sr.DataRoot, "labscripts"))

	pr := &sr.Project
	*pr = Project{StartDir: dir, DataRoot: sr.DataRoot}
	pr.Defaults()
	if hasFile {
		if err := jsonx.Open(pr, fn); err != nil {
			return err
		}
	} else {
		sr.migrateProject()
	}
	pr.StartDir, pr.DataRoot = dir, sr.DataRoot // where it is now
	pr.Update()
	if !hasFile {
		sr.SaveProject()
	}

	sr.JobsTable = table.New()
	sr.UpdateScripts() // automatically runs lowercase init scripts
	if err := sr.connectServer(); err != nil {
		core.ErrorSnackbar(sr, err)
	}
	RecentProjects.AddPath(fn, core.SystemSettings.SavedPathsMax)
	saveRecentProjects()
	sr.Scene.Body.SetTitle("Simmer • " + pr.Name)
	if tab := sr.Tabs.TabByName("Projects"); tab != nil {
		sr.Tabs.DeleteTabIndex(tab.IndexInParent())
	}
	sr.UpdateFiles()
	if sr.Toolbar != nil { // made when the window opens
		sr.Toolbar.Update()
	}
	return nil
}

// SaveProject saves the project in its simmer.json file.
func (sr *Simmer) SaveProject() {
	if !sr.isOpen() {
		return
	}
	errors.Log(jsonx.Save(&sr.Project, filepath.Join(sr.StartDir, ProjectFilename)))
}

// EditProject edits the settings of the project, saving them.
func (sr *Simmer) EditProject() { //types:add
	lab.PromptStruct(sr, &sr.Project, "Project settings", sr.SaveProject)
}

// NewProject makes a new simulation project for the simulation in the
// given directory, named as the directory, and opens it: it makes the
// ~/simdata/<name>/<user> directory for its data, with a labscripts
// directory for its scripts, and a simdata symbolic link to it in the
// simulation directory, with a simmer.json file for its settings.
// The parts that exist already are kept.
func (sr *Simmer) NewProject(dir core.Filename) { //types:add
	if err := sr.newProject(string(dir)); err != nil {
		core.ErrorSnackbar(sr, err)
	}
}

func (sr *Simmer) newProject(dir string) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if !isDir(dir) {
		return fmt.Errorf("NewProject: the simulation directory %s does not exist", dir)
	}
	usr, err := user.Current()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	name := filepath.Base(dir)
	data := filepath.Join(home, "simdata", name, usr.Username)
	if err := os.MkdirAll(filepath.Join(data, "labscripts"), 0755); err != nil {
		return err
	}
	link := filepath.Join(dir, "simdata")
	if _, err := os.Lstat(link); err != nil {
		if err := os.Symlink(data, link); err != nil {
			return err
		}
	}
	fn := filepath.Join(dir, ProjectFilename)
	if !isFile(fn) {
		if sr.isOpen() { // Defaults changes the directory
			defer goalrun.Run("cd", sr.StartDir)
		}
		pr := &Project{StartDir: dir, DataRoot: link}
		pr.Defaults()
		pr.Update()
		if err := jsonx.Save(pr, fn); err != nil {
			return err
		}
	}
	sr.OpenProject(core.Filename(fn))
	return nil
}

// migrateProject moves the settings of a project made before simmer.json,
// from its labscripts/defaults.goal script, which set them and connected
// to the server with gossh, to the project, with the server as
// [ServerParams.SSH]. The script is renamed to defaults.goal.old, so that
// it is not run again.
func (sr *Simmer) migrateProject() {
	fn := filepath.Join(sr.ScriptsDir, "defaults.goal")
	b, err := os.ReadFile(fn)
	if err != nil {
		return
	}
	var code []string
	for _, ln := range strings.Split(string(b), "\n") {
		f := strings.Fields(ln)
		switch {
		case len(f) >= 2 && f[0] == "gossh":
			if sr.Project.Server.SSH == "" {
				sr.Project.Server.SSH = f[1]
			}
			continue
		case len(f) == 1 && f[0] == "@0":
			continue
		}
		code = append(code, ln)
	}
	sc := strings.Join(code, "\n")
	sc = strings.ReplaceAll(sc, "lab.LabBrowser.Config", "lab.LabBrowser.Project")
	sc = regexp.MustCompile(`\bcf\.Project\b`).ReplaceAllString(sc, "cf.Name")
	in := sr.Interpreter.(*interpreter.Interpreter)
	clearErrors() // of the commands before, which would stop it
	if _, _, err := in.Eval(sc); err != nil {
		slog.Error("Simmer: could not move the settings of defaults.goal to "+ProjectFilename, "Error", err)
		return
	}
	errors.Log(os.Rename(fn, fn+".old"))
	slog.Info("Simmer: moved the settings of defaults.goal to "+ProjectFilename, "Project", sr.Project.Name)
}

// connectServer connects to the server of the project, [ServerParams.SSH],
// as @1, as gossh does, without making it the active host. If it could not
// be connected, as when the network is down, it tries again every
// [sshclient.Config.Timeout], and if the connection was lost, it connects
// again (see [sshclient.Client.CheckConnection]). It is called before the
// commands on the server.
func (sr *Simmer) connectServer() error {
	ssh := sr.Project.Server.SSH
	if ssh == "" {
		return nil
	}
	if cl, ok := goalrun.SSHClients["1"]; ok {
		return cl.CheckConnection()
	}
	if time.Since(sr.lastConnect) < goalrun.SSH.Timeout {
		return fmt.Errorf("Simmer: not connected to the server %s", ssh)
	}
	sr.lastConnect = time.Now()
	cl := sshclient.NewClient(goalrun.SSH)
	if err := cl.Connect(ssh); err != nil {
		return err
	}
	goalrun.SSHClients["1"] = cl
	return nil
}

// projectsTab shows the recent projects in a Projects tab, to open one,
// when no project is open.
func (sr *Simmer) projectsTab() {
	lab.NewTab(sr.Tabs, "Projects", func(tab *core.Frame) *core.Frame {
		fr := core.NewFrame(tab)
		fr.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Grow.Set(1, 1)
		})
		core.NewText(fr).SetType(core.TextHeadlineSmall).SetText("Open a recent project")
		for _, rp := range RecentProjects {
			core.NewButton(fr).SetText(filepath.Dir(rp)).SetIcon(icons.Folder).SetType(core.ButtonText).
				OnClick(func(e events.Event) {
					sr.OpenProject(core.Filename(rp))
				})
		}
		core.NewText(fr).SetType(core.TextSupporting).SetText("or use Open or New project in the toolbar")
		return fr
	})
}
