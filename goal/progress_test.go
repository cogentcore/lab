// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"bytes"
	"testing"
	"time"

	"cogentcore.org/core/base/sshclient"
	"github.com/stretchr/testify/assert"
)

func TestProgressLine(t *testing.T) {
	p := sshclient.Progress{File: "/data/run/all_epc.tsv", TotalDone: 9 << 20, Total: 20 << 20, FileNum: 2, NFiles: 5}
	assert.Equal(t, "[2/5] all_epc.tsv [=========>          ]  45%  9.0 MB/20.0 MB  4.5 MB/s  0:02",
		progressLine(p, 2*time.Second, 80))
	// a long name is shortened to fit
	p.File = "/data/a_very_long_file_name_of_simulation_results_and_more.tsv"
	assert.Equal(t, "[2/5] a_very_lon... [=========>          ]  45%  9.0 MB/20.0 MB  4.5 MB/s  0:02",
		progressLine(p, 2*time.Second, 80))
	// a narrow terminal drops the bytes, the rate, and then the bar
	assert.Equal(t, "[2/5] a_very_long_file... [=========>          ]  45%  4.5 MB/s  0:02",
		progressLine(p, 2*time.Second, 70))
	assert.Equal(t, "[2/5] a_very_long_file_name_of_s...  45%  0:02", progressLine(p, 2*time.Second, 47))
	for _, w := range []int{80, 70, 47} {
		assert.LessOrEqual(t, len(progressLine(p, 2*time.Second, w)), w-1)
	}
	// one file, all done
	p = sshclient.Progress{File: "x.tsv", TotalDone: 100, Total: 100, FileNum: 1, NFiles: 1}
	assert.Equal(t, "x.tsv [====================] 100%  100 B/100 B  100 B/s  0:00", progressLine(p, time.Second, 80))

	assert.Equal(t, "1023 B", formatBytes(1023))
	assert.Equal(t, "1.5 KB", formatBytes(1536))
	assert.Equal(t, "2.0 GB", formatBytes(2<<30))

	// no progress if not a terminal
	assert.Nil(t, scpProgress(&bytes.Buffer{}))
}
