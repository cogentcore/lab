// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBroadcaster tests that the broadcaster gives the same indexes
// as Shape.IndexFrom1D and WrapIndex1D, for all of the index cases.
func TestBroadcaster(t *testing.T) {
	shapes := [][2][]int{
		{{3, 4}, {3, 4}},       // same
		{{3, 4}, {1}},          // scalar b
		{{1}, {3, 4}},          // scalar a
		{{3, 4}, {4}},          // row vector
		{{3, 1}, {1, 4}},       // outer product
		{{2, 3, 4}, {3, 1}},    // inner broadcast
		{{2, 1, 4}, {2, 3, 4}}, // a broadcast in the middle
		{{5}, {5}},
	}
	for _, sh := range shapes {
		a, b := NewFloat64(sh[0]...), NewFloat64(sh[1]...)
		as, bs, os, err := AlignShapes(a, b)
		assert.NoError(t, err)
		br := newBroadcaster(as, bs, os)
		for idx := range os.Len() {
			oi := os.IndexFrom1D(idx)
			ai, bi := br.index(idx)
			assert.Equal(t, WrapIndex1D(as, oi...), ai, fmt.Sprint(sh, idx))
			assert.Equal(t, WrapIndex1D(bs, oi...), bi, fmt.Sprint(sh, idx))
		}
	}
	// assignment
	a, b := NewFloat64(2, 3, 4), NewFloat64(3, 1)
	as, bs, err := AlignForAssign(a, b)
	assert.NoError(t, err)
	br := newBroadcaster(as, bs, as)
	for idx := range as.Len() {
		_, bi := br.index(idx)
		assert.Equal(t, WrapIndex1D(bs, as.IndexFrom1D(idx)...), bi)
	}
}

func BenchmarkFloatBinary(b *testing.B) {
	n := 10000
	x, y, out := NewFloat64(n), NewFloat64(n), NewFloat64(n)
	sc := NewFloat64Scalar(2)
	row := NewFloat64(100)
	x2 := NewFloat64(100, 100)
	sub := func(a, b float64) float64 { return a - b }
	b.Run("same", func(b *testing.B) {
		for b.Loop() {
			FloatBinaryFuncOut(1, sub, x, y, out)
		}
	})
	b.Run("scalar", func(b *testing.B) {
		for b.Loop() {
			FloatBinaryFuncOut(1, sub, x, sc, out)
		}
	})
	b.Run("broadcast", func(b *testing.B) {
		for b.Loop() {
			FloatBinaryFuncOut(1, sub, x2, row, out)
		}
	})
}
