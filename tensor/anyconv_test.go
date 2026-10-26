// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToTensor(t *testing.T) {
	f := NewFloat64(2)
	assert.Same(t, f, ToTensor(f))
	assert.Equal(t, 1.5, ToTensor(1.5).Float1D(0))
	assert.Equal(t, 3, ToTensor(3).Int1D(0))
	assert.Equal(t, "a", ToTensor("a").String1D(0))
	assert.Equal(t, 1.0, ToTensor(true).Float1D(0))
	assert.Equal(t, 2.0, ToTensor(float32(2)).Float1D(0))
	assert.Equal(t, 7.0, ToTensor(int64(7)).Float1D(0))

	// a slice is wrapped, not copied
	vals := []float64{1, 2, 3}
	ts := ToTensor(vals)
	assert.Equal(t, 3, ts.Len())
	ts.SetFloat1D(5, 1)
	assert.Equal(t, 5.0, vals[1])
	assert.Panics(t, func() { ToTensor(struct{}{}) })
}

func TestToValues(t *testing.T) {
	ts := NewFloat64FromValues(2.5, 3)
	assert.Equal(t, 2.5, ToFloat64(ts))
	assert.Equal(t, 2.5, ToFloat64(2.5))
	assert.Equal(t, 3.0, ToFloat64(3))
	assert.Equal(t, 2, ToInt(ts))
	assert.Equal(t, 4, ToInt(4.7))
	assert.Equal(t, 4, ToInt(int32(4)))
	assert.Equal(t, "2.5", ToString(ts))
	assert.Equal(t, "x", ToString("x"))

	assert.True(t, ToBool(true))
	assert.True(t, ToBool(NewBoolFromValues(true, false)))
	assert.False(t, ToBool(NewBoolFromValues(false)))
	assert.True(t, ToBool(ts))
	assert.False(t, ToBool(NewFloat64Scalar(0)))
	assert.False(t, ToBool(NewFloat64(0)))
	assert.True(t, ToBool(1))

	assert.Equal(t, []int{3}, ToInts(3))
	assert.Equal(t, []int{3, 4}, ToInts([]int{3, 4}))
	assert.Equal(t, []int{2, 3}, ToInts(NewIntFromValues(2, 3)))
	assert.Equal(t, []int{2}, ToInts(2.0))
}

func TestIter(t *testing.T) {
	collect := func(v any) (idxs []int, vals []float64) {
		for i, v := range Iter(v) {
			idxs = append(idxs, i)
			vals = append(vals, v)
		}
		return
	}
	idxs, vals := collect(3)
	assert.Equal(t, []int{0, 1, 2}, idxs)
	assert.Equal(t, []float64{0, 1, 2}, vals)
	idxs, vals = collect(NewFloat64FromValues(5, 6))
	assert.Equal(t, []int{0, 1}, idxs)
	assert.Equal(t, []float64{5, 6}, vals)
	_, vals = collect([]float64{7, 8})
	assert.Equal(t, []float64{7, 8}, vals)
	_, vals = collect([]int{1, 2})
	assert.Equal(t, []float64{1, 2}, vals)

	// stopping early
	n := 0
	for range Iter(10) {
		n++
		if n == 3 {
			break
		}
	}
	assert.Equal(t, 3, n)
}

func TestSetVar(t *testing.T) {
	var tsr Tensor
	SetVar(&tsr, 2.5)
	assert.Equal(t, 2.5, tsr.Float1D(0))
	f := NewFloat64(2)
	SetVar(&tsr, f)
	assert.Same(t, f, tsr)

	var x float64
	SetVar(&x, NewFloat64Scalar(1.5))
	assert.Equal(t, 1.5, x)
	var i int
	SetVar(&i, 3.9)
	assert.Equal(t, 3, i)
	var b bool
	SetVar(&b, NewBoolFromValues(true))
	assert.True(t, b)
	var s string
	SetVar(&s, NewStringScalar("a"))
	assert.Equal(t, "a", s)

	// other types, by reflection
	var f32 float32
	SetVar(&f32, NewFloat64Scalar(2))
	assert.Equal(t, float32(2), f32)
	var ft *Float64
	SetVar(&ft, Tensor(f))
	assert.Same(t, f, ft)
	assert.Panics(t, func() { SetVar(x, 1) })
}
