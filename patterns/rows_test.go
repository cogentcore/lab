// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patterns

import (
	"testing"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedRand gives the test its own seeded random source and restores
// the package default afterwards, so test order does not matter.
func seedRand(t *testing.T, seed int64) {
	NewRand(seed)
	t.Cleanup(func() { NewRand(5711) })
}

// TestSetRandSeed tests that SetRandSeed and RestoreSeed repeat the
// random sequence.
func TestSetRandSeed(t *testing.T) {
	seedRand(t, 1)
	SetRandSeed(33)
	assert.Equal(t, int64(33), RandSeed)
	a := RandSource.Perm(10)
	RestoreSeed()
	assert.Equal(t, a, RandSource.Perm(10))
}

// TestPermutedBinary tests that PermutedBinary sets exactly nOn on
// values, and that empty tensors are left alone.
func TestPermutedBinary(t *testing.T) {
	seedRand(t, 2)
	tsr := tensor.NewFloat32(4, 5)
	PermutedBinary(tsr, 6, 1, -1)
	assert.Equal(t, 6, countVal(tsr, 1))
	assert.Equal(t, 14, countVal(tsr, -1))

	PermutedBinary(tensor.NewFloat32(0), 3, 1, 0) // no panic

	assert.Equal(t, 3, NFromPct(0.25, 12))
}

// TestPermutedBinaryRows tests that each row gets nOn values and that
// the rows differ (permutation is re-drawn per row).
func TestPermutedBinaryRows(t *testing.T) {
	seedRand(t, 3)
	tsr := tensor.NewFloat64(5, 4, 4)
	PermutedBinaryRows(tsr, 4, 1, 0)
	distinct := map[string]bool{}
	for r := range 5 {
		row := tsr.SubSpace(r)
		assert.Equal(t, 4, NOnInTensor(row))
		assert.Equal(t, float32(0.25), PctActInTensor(row))
		distinct[row.String()] = true
	}
	assert.Greater(t, len(distinct), 1)
	PermutedBinaryRows(tensor.NewFloat64(0, 4), 1, 1, 0) // no panic
}

// TestPermutedBinaryMinDiffErrors tests the error results of
// PermutedBinaryMinDiff for an empty tensor and an unsatisfiable
// minimum difference.
func TestPermutedBinaryMinDiffErrors(t *testing.T) {
	seedRand(t, 4)
	assert.Error(t, PermutedBinaryMinDiff(tensor.NewFloat32(0, 3), 1, 1, 0, 1))

	MinDiffPrintIterations = true
	defer func() { MinDiffPrintIterations = false }()
	// 4 rows of 3 cells with 1 on: two rows must share their bit,
	// so a min diff of 1 can never be met.
	tsr := tensor.NewFloat32(4, 3)
	assert.Error(t, PermutedBinaryMinDiff(tsr, 1, 1, 0, 1))
	for r := range 4 {
		assert.Equal(t, 1, NOnInTensor(tsr.SubSpace(r)))
	}
}

// TestFlipBits tests that FlipBits turns the requested number of bits
// off and on, clipping to what is available.
func TestFlipBits(t *testing.T) {
	seedRand(t, 5)
	tsr := tensor.NewFloat32(10)
	for i := range 4 {
		tsr.SetFloat1D(1, i)
	}
	FlipBits(tsr, 1, 3, 1, 0)
	assert.Equal(t, 6, countVal(tsr, 1))

	// flip more than available: all on go off, all off go on.
	FlipBits(tsr, 100, 100, 1, 0)
	assert.Equal(t, 4, countVal(tsr, 1))

	FlipBits(tensor.NewFloat32(0), 1, 1, 1, 0) // no panic

	rows := tensor.NewFloat32(3, 6)
	FlipBitsRows(rows, 0, 2, 1, 0)
	for r := range 3 {
		assert.Equal(t, 2, NOnInTensor(rows.SubSpace(r)))
	}
}

// TestSplitRowsNames tests the default naming and the error paths of
// SplitRows.
func TestSplitRowsNames(t *testing.T) {
	dir, err := tensorfs.NewDir("split")
	require.NoError(t, err)
	src := tensor.NewFloat32(5, 2)
	for i := range 10 {
		src.SetFloat1D(float64(i), i)
	}
	// unnamed source: names are numbers.
	require.NoError(t, SplitRows(dir, src, nil, 2))
	s0 := dir.Float32("0")
	s1 := dir.Float32("1")
	assert.Equal(t, []int{2, 2}, s0.ShapeSizes())
	assert.Equal(t, []int{3, 2}, s1.ShapeSizes())
	assert.Equal(t, 4.0, s1.Float1D(0))

	// named source: names are name_N.
	metadata.SetName(src, "src")
	require.NoError(t, SplitRows(dir, src, nil, 1, 4))
	assert.Equal(t, []int{3, 2}, dir.Float32("src_1").ShapeSizes())
	assert.Equal(t, 9.0, dir.Float32("src_2").Float1D(1))

	assert.Error(t, SplitRows(dir, src, []string{"a"}, 2))
	assert.Error(t, SplitRows(dir, src, nil, 3, 2))
}

// TestSplitRowsAgain tests that splitting again into the same
// directory reshapes the existing values, and that the rows argument
// is not written into.
func TestSplitRowsAgain(t *testing.T) {
	dir, err := tensorfs.NewDir("split")
	require.NoError(t, err)
	src := tensor.NewFloat32(5, 2)
	for i := range 10 {
		src.SetFloat1D(float64(i), i)
	}
	names := []string{"a", "b"}
	require.NoError(t, SplitRows(dir, src, names, 1))
	assert.Equal(t, []int{4, 2}, dir.Float32("b").ShapeSizes())

	rows := make([]int, 1, 4)
	rows[0] = 4
	require.NoError(t, SplitRows(dir, src, names, rows...))
	assert.Equal(t, []int{4, 2}, dir.Float32("a").ShapeSizes())
	assert.Equal(t, []int{1, 2}, dir.Float32("b").ShapeSizes())
	assert.Equal(t, 8.0, dir.Float32("b").Float1D(0))
	assert.Equal(t, 0, rows[:2][1])
}

// countVal counts the elements of tsr equal to v.
func countVal(tsr tensor.Values, v float64) int {
	n := 0
	for i := range tsr.Len() {
		if tsr.Float1D(i) == v {
			n++
		}
	}
	return n
}
