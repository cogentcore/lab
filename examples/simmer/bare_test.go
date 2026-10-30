// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	"cogentcore.org/lab/table"
	"github.com/stretchr/testify/assert"
)

// TestActiveJobsBare tests that the jobs that are still pending or running
// are found from the jobs table when bare metal cannot be reached, so
// that they are canceled instead of deleted.
func TestActiveJobsBare(t *testing.T) {
	sr := &Simmer{}
	sr.JobsTable = table.New()
	jobs := map[string]string{"j1": "Running", "j2": "Fetched", "j3": "Pending", "j4": "Canceled"}
	ids := []string{"j1", "j2", "j3", "j4"}
	sr.JobsTable.SetNumRows(len(ids))
	jc := sr.JobsTable.AddStringColumn("JobID")
	sc := sr.JobsTable.AddStringColumn("Status")
	sj := sr.JobsTable.AddStringColumn("ServerJob")
	for i, jid := range ids {
		jc.SetString1D(jid, i)
		sc.SetString1D(jobs[jid], i)
		sj.SetString1D("1", i)
	}
	active, others := sr.activeJobsBare(ids)
	assert.Equal(t, []string{"j1", "j3"}, active)
	assert.Equal(t, []string{"j2", "j4"}, others)
}

// TestCancelActiveJobs tests that the jobs that are still pending or
// running are not deleted or archived, and that the message says that
// they could not be canceled when bare metal is not connected.
func TestCancelActiveJobs(t *testing.T) {
	sr := &Simmer{}
	sr.JobsTable = table.New()
	sr.JobsTable.SetNumRows(2)
	jc := sr.JobsTable.AddStringColumn("JobID")
	sc := sr.JobsTable.AddStringColumn("Status")
	sj := sr.JobsTable.AddStringColumn("ServerJob")
	for i, st := range []string{"Running", "Fetched"} {
		jc.SetString1D([]string{"j1", "j2"}[i], i)
		sc.SetString1D(st, i)
		sj.SetString1D("1", i)
	}
	others, msg := sr.cancelActiveJobs([]string{"j1", "j2"}, "deleted")
	assert.Equal(t, []string{"j2"}, others)
	assert.Contains(t, msg, "not deleted, and could not be canceled: j1")
}
