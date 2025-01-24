// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"fmt"
	"time"

	"cogentcore.org/core/base/errors"
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

// Job returns the Job record for given job number; nil if not found
// in Active or Done;
func (bm *BareMetal) Job(jobno int) *Job {
	bm.Lock()
	defer bm.Unlock()

	return bm.job(jobno)
}

// Submit adds a new Active job with given parameters.
func (bm *BareMetal) Submit(src, path, script, results string, files []byte) *Job {
	bm.Lock()
	defer bm.Unlock()

	return bm.submit(src, path, script, results, files)
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

// CancelJobs cancels list of job IDs. Returns error for jobs not found.
func (bm *BareMetal) CancelJobs(jobs ...int) error {
	bm.Lock()
	defer bm.Unlock()

	return bm.cancelJobs(jobs...)
}

// FetchResults gets job results back from server for given job id(s).
// Results are available as job.Results as a compressed tar file.
func (bm *BareMetal) FetchResults(ids ...int) ([]*Job, error) {
	bm.Lock()
	defer bm.Unlock()

	return bm.fetchResults(ids...)
}

// bgLoop is the background update loop
func (bm *BareMetal) bgLoop() {
	for {
		bm.Lock()
		if bm.ticker == nil {
			bm.Unlock()
			return
		}
		bm.Unlock()
		<-bm.ticker.C
		nrun, nfin, err := bm.UpdateJobs()
		if err != nil {
			errors.Log(err)
		} else {
			fmt.Printf("BareMetal: N Run: %d  N Finished: %d\n", nrun, nfin)
		}
	}
}
