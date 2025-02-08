// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"time"
)

// this file has the exported API for direct usage,
// which wraps calls in locks.

// OpenLog opens a log file for recording actions.
func (bm *BareMetal) OpenLog(filename string) error {
	// todo: openlog file on slog
	return nil
}

// StartBGUpdates starts a ticker to update job status periodically.
func (bm *BareMetal) StartBGUpdates() {
	bm.ticker = time.NewTicker(10 * time.Second)
	go bm.bgLoop()
}

// Submit adds a new Active job with given parameters.
func (bm *BareMetal) Submit(src, path, script, results string, files []byte) *Job {
	bm.Lock()
	defer bm.Unlock()

	return bm.submit(src, path, script, results, files)
}

// JobStatus gets current job data for given job id(s).
// An empty list returns all of the currently Active jobs.
func (bm *BareMetal) JobStatus(ids ...int) []*Job {
	bm.Lock()
	defer bm.Unlock()

	if len(ids) == 0 {
		return bm.Active.Values
	}
	jobs := make([]*Job, 0, len(ids))
	for _, id := range ids {
		job := bm.job(id)
		if job == nil {
			continue
		}
		jobs = append(jobs, job)
	}
	return jobs
}

// CancelJobs cancels list of job IDs. Returns error for jobs not found.
func (bm *BareMetal) CancelJobs(ids ...int) error {
	bm.Lock()
	defer bm.Unlock()

	return bm.cancelJobs(ids...)
}

// FetchResults gets job results back from server for given job id(s).
// Results are available as job.Results as a compressed tar file.
func (bm *BareMetal) FetchResults(ids ...int) ([]*Job, error) {
	bm.Lock()
	defer bm.Unlock()

	return bm.fetchResults(ids...)
}

// UpdateJobs runs any pending jobs if there are available GPUs to run on.
// returns number of jobs started, and any errors incurred in starting jobs.
func (bm *BareMetal) UpdateJobs() (nrun, nfinished int, err error) {
	bm.Lock()
	defer bm.Unlock()

	nfinished, err = bm.pollJobs()
	nrun, err = bm.runPendingJobs()
	bm.saveState()
	return
}
