// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cogentcore.org/core/base/sshclient/sshtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jobScript is a job script as simmer makes: it runs the job in the
// background, writing its pid to job.pid. The job runs until a finish
// file is made.
const jobScript = `#!/bin/sh
sh -c 'while [ ! -f finish ]; do sleep 0.05; done; echo results > data.tsv' > job.out 2>&1 &
echo $! > job.pid
`

// TestDisconnect tests that baremetal does not hang when the network to
// a server is lost, does not change the state of the jobs on the server
// while it is lost, and goes on with them when the network comes back.
func TestDisconnect(t *testing.T) {
	s := sshtest.NewServer(t)
	bm := NewBareMetal()
	bm.DataDir = t.TempDir()
	bm.newInterpreter()
	goalrun.SSH.User.KeyPath = s.KeyPath
	goalrun.SSH.KeepAlive = 100 * time.Millisecond
	goalrun.SSH.Timeout = 300 * time.Millisecond
	bm.Servers.Add("test", &Server{Name: "test", SSH: s.Addr, NGPUs: 1})
	bm.initServers()
	t.Cleanup(goalrun.CloseSSH)

	jobs := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(jobs, "job.sh"), []byte(jobScript), 0755))
	var b bytes.Buffer
	require.NoError(t, TarFiles(&b, jobs, true, "job.sh"))
	job := bm.Submit("test", "jobs/1", "job.sh", "*.tsv", b.Bytes())

	update := func(what string) {
		t.Helper()
		done := make(chan struct{})
		go func() {
			bm.UpdateJobs()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatalf("%s: UpdateJobs hangs", what)
		}
	}
	update("start")
	require.Equal(t, Running, job.Status)
	assert.Greater(t, job.PID, 0)

	// the job finishes while the network is lost, which is not seen
	s.NetworkDown()
	require.NoError(t, os.WriteFile(filepath.Join(s.Root, "jobs/1/finish"), nil, 0644))
	for range 3 {
		update("network lost")
		assert.Equal(t, Running, job.Status, "running while the network is lost")
		assert.Equal(t, 1, bm.Active.Len())
	}
	assert.Error(t, bm.CancelJobs(job.ID), "not canceled while the network is lost")
	assert.Equal(t, Running, job.Status)

	// and is seen as completed when it comes back, with its results
	s.NetworkUp()
	deadline := time.Now().Add(10 * time.Second)
	for job.Status == Running {
		if time.Now().After(deadline) {
			t.Fatal("the job is not completed after the network comes back")
		}
		update("network back")
		time.Sleep(100 * time.Millisecond)
	}
	assert.Equal(t, Completed, job.Status)
	assert.Equal(t, 0, bm.Active.Len())
	assert.NotEmpty(t, job.Results)
}
