// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package threading

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// runRanges runs ParallelRun and returns the sorted [start, end) ranges
// passed to the function, and the per-item visit counts.
func runRanges(total, nThreads int) ([][2]int, []int) {
	var mu sync.Mutex
	var ranges [][2]int
	counts := make([]int, total)
	ParallelRun(func(st, ed int) {
		mu.Lock()
		defer mu.Unlock()
		ranges = append(ranges, [2]int{st, ed})
		for i := st; i < ed; i++ {
			counts[i]++
		}
	}, total, nThreads)
	sort.Slice(ranges, func(i, j int) bool { return ranges[i][0] < ranges[j][0] })
	return ranges, counts
}

// TestParallelRun tests that the range is split into the expected chunks,
// with every item visited exactly once.
func TestParallelRun(t *testing.T) {
	ranges, counts := runRanges(10, 3)
	assert.Equal(t, [][2]int{{0, 4}, {4, 8}, {8, 10}}, ranges)
	for i, c := range counts {
		assert.Equal(t, 1, c, "item %d", i)
	}

	ranges, _ = runRanges(9, 3)
	assert.Equal(t, [][2]int{{0, 3}, {3, 6}, {6, 9}}, ranges)

	// more threads than items: one item each
	ranges, _ = runRanges(2, 8)
	assert.Equal(t, [][2]int{{0, 1}, {1, 2}}, ranges)

	ranges, _ = runRanges(5, 1)
	assert.Equal(t, [][2]int{{0, 5}}, ranges)
}

// TestParallelRunEmpty tests that nothing runs for a total of 0.
func TestParallelRunEmpty(t *testing.T) {
	ranges, _ := runRanges(0, 4)
	assert.Empty(t, ranges)
}
