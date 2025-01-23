// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

//go:generate core generate

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/iox/jsonx"
	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/base/keylist"
	"cogentcore.org/core/system"
	"cogentcore.org/lab/goal"
	"cogentcore.org/lab/goal/goalib"
)

// goalrun is needed for running goal commands.
var goalrun *goal.Goal

// BareMetal is the overall bare metal job manager.
type BareMetal struct {

	//	Servers is the ordered list of server machines.
	Servers Servers `json:"-"`

	// NextID is the next job ID to assign.
	NextID int

	// Active has all the active (pending, running) jobs being managed,
	// in the order submitted.
	// The unique key is the bare metal job ID (int).
	Active Jobs

	// Done has all the completed jobs that have been run.
	// This list can be purged by time as needed.
	// The unique key is the bare metal job ID (int).
	Done Jobs

	// ticker is the [time.Ticker] used to control the background update loop.
	ticker *time.Ticker `json:"-" toml:"-"`

	// Lock for responding to inputs.
	// everything below top-level input processing is assumed to be locked.
	sync.Mutex `json:"-" toml:"-"`
}

// Jobs is the ordered list of jobs, in order submitted.
type Jobs = keylist.List[int, *Job]

// Servers is the ordered list of servers, in order of use preference.
type Servers = keylist.List[string, *Server]

func NewBareMetal() *BareMetal {
	bm := &BareMetal{}
	return bm
}

// Init does the full initialization of the server: Config, OpenState,
// InitServers, using given goal.Goal instance.
func (bm *BareMetal) Init(gl *goal.Goal) {
	goalrun = gl
	bm.config()
	bm.openState()
	bm.initServers()
}

// config loads a toml format config file from
// TheApp.DataDir()/BareMetal/config.toml to load the servers.
// Use [[Servers.Values]] header for each server.
func (bm *BareMetal) config() {
	dir := filepath.Join(system.TheApp.DataDir(), "BareMetal")
	os.MkdirAll(dir, 0777)
	file := filepath.Join(dir, "config.toml")
	if !goalib.FileExists(file) {
		slog.Error("BareMetal config file not found: no servers will be configured", "File location:", file)
		return
	}
	errors.Log(tomlx.Open(bm, file))
	bm.updateServerIndexes()
}

// saveState saves the current active state to a JSON file:
// TheApp.DataDir()/BareMetal/state.json  A backup ~ file is
// made of any existing prior to saving.
func (bm *BareMetal) saveState() {
	dir := filepath.Join(system.TheApp.DataDir(), "BareMetal")
	os.MkdirAll(dir, 0777)
	file := filepath.Join(dir, "state.json")
	bkup := filepath.Join(dir, "state.json~")
	if goalib.FileExists(file) {
		if goalib.FileExists(bkup) {
			os.Remove(bkup)
		}
		os.Rename(file, bkup)
	}
	errors.Log(jsonx.Save(bm, file))
}

// openState opens the current active state from the file made by SaveState,
// to restore to prior running state.
func (bm *BareMetal) openState() {
	dir := filepath.Join(system.TheApp.DataDir(), "BareMetal")
	file := filepath.Join(dir, "state.json")
	if !goalib.FileExists(file) {
		return
	}
	errors.Log(jsonx.Open(bm, file))
	bm.updateServerIndexes()
	bm.Active.UpdateIndexes()
	bm.Done.UpdateIndexes()
	bm.setServerUsedFromJobs()
}

// initServers initializes the server state, including opening SSH connections.
func (bm *BareMetal) initServers() {
	for _, sv := range bm.Servers.Values {
		sv.OpenSSH()
	}
	goalrun.Run("@0")
}

// updateServerIndexes updates the indexes in the Servers ordered map,
// which is needed after loading new Server config.
func (bm *BareMetal) updateServerIndexes() {
	svs := &bm.Servers
	svs.Keys = make([]string, len(svs.Values))
	for i, v := range svs.Values {
		svs.Keys[i] = v.Name
	}
	svs.UpdateIndexes()
}
