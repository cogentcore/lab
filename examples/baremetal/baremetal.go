// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

//go:generate core generate

import (
	"sync"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/iox/jsonx"
	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/base/keylist"
	"cogentcore.org/lab/goal"
)

// goalrun is needed for running goal commands.
var goalrun *goal.Goal

// BareMetal is the overall bare metal job manager.
type BareMetal struct {

	//	Servers is the ordered list of server machines.
	Servers Servers

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

// Config loads a toml format config file.
// Use [[Servers.Values]] header for each server.
func (bm *BareMetal) Config(filename string) {
	errors.Log(tomlx.Open(bm, filename))
	bm.updateServerIndexes()
}

// SaveState saves the current active state to a JSON file.
func (bm *BareMetal) SaveState(filename string) {
	errors.Log(jsonx.Save(bm, filename))
}

// OpenState opens the current active state from a JSON file,
// to restore to prior running state.
func (bm *BareMetal) OpenState(filename string) {
	errors.Log(jsonx.Open(bm, filename))
}

// InitGoal makes a new Goal transpiler system, mainly for testing.
func (bm *BareMetal) InitGoal() {
	goalrun = goal.NewGoal()
}

// InitServers initializes the server state, including opening SSH connections.
func (bm *BareMetal) InitServers() {
	for _, sv := range bm.Servers.Values {
		sv.OpenSSH()
	}
}

// OpenLog opens a log file for recording actions.
func (bm *BareMetal) OpenLog(filename string) error {
	// todo: openlog file on slog
	return nil
}
