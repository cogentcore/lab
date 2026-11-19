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

// TestRowsIndex1D tests the 1D accessors of Rows with Indexes,
// against the n-dimensional ones.
func TestRowsIndex1D(t *testing.T) {
	ft := NewFloat64(4, 2, 3)
	for i := range ft.Len() {
		ft.SetFloat1D(float64(i), i)
	}
	st := NewString(4, 2)
	it := NewInt(5)
	for i := range st.Len() {
		st.SetString1D(fmt.Sprint(i), i)
	}
	for i := range it.Len() {
		it.SetInt1D(i*10, i)
	}
	idx := []int{3, 1, 2}
	fr := NewRows(ft, idx...)
	for i := range fr.Len() {
		r, c := i/6, i%6
		assert.Equal(t, ft.Float(idx[r], c/3, c%3), fr.Float1D(i))
	}
	fr.SetFloat1D(-1, 7) // row 1 = 1, cell 1
	assert.Equal(t, -1.0, ft.Float(1, 0, 1))
	sr := NewRows(st, idx...)
	assert.Equal(t, st.StringValue(2, 1), sr.String1D(5))
	sr.SetString1D("x", 0)
	assert.Equal(t, "x", st.StringValue(3, 0))
	ir := NewRows(it, 4, 0)
	assert.Equal(t, 40, ir.Int1D(0))
	ir.SetInt1D(7, 1)
	assert.Equal(t, 7, it.Int1D(0))
}
