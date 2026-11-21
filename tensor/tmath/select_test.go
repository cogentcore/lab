// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"math"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestIsIn(t *testing.T) {
	s := tensor.NewStringFromValues("a", "b", "c", "a")
	assert.Equal(t, []bool{true, false, true, true}, IsIn(s, tensor.NewStringFromValues("a", "c")).Values.ToBools())
	f := tensor.NewFloat64FromValues(1, 2, math.NaN(), 3)
	m := IsIn(f, tensor.NewFloat64FromValues(2, 3, math.NaN()))
	assert.Equal(t, []bool{false, true, false, true}, m.Values.ToBools(), "NaN is never in the values")
	m2 := IsIn(tensor.Reshape(tensor.NewIntFromValues(1, 2, 3, 4), 2, 2), tensor.NewIntFromValues(4))
	assert.Equal(t, []int{2, 2}, m2.ShapeSizes())
}

func TestWhere(t *testing.T) {
	x := tensor.NewFloat64FromValues(-1, 2, -3, 4)
	w := Where(Greater(x, tensor.NewFloat64Scalar(0)), x, tensor.NewFloat64Scalar(0))
	assert.Equal(t, []float64{0, 2, 0, 4}, w.(*tensor.Float64).Values)

	// strings, with a scalar for the false values
	s := tensor.NewStringFromValues("a", "b", "c")
	ws := Where(IsIn(s, tensor.NewStringFromValues("a")), s, tensor.NewStringScalar("other"))
	assert.Equal(t, []string{"a", "other", "other"}, ws.(*tensor.String).Values)

	// a row condition broadcast over the rows of a matrix
	mat := tensor.Reshape(tensor.NewFloat64FromValues(1, 2, 3, 4, 5, 6), 2, 3)
	cond := tensor.NewBoolShape(tensor.NewShape(3))
	cond.SetBool1D(true, 1)
	wm := Where(cond, mat, tensor.NewFloat64Scalar(-1))
	assert.Equal(t, []int{2, 3}, wm.ShapeSizes())
	assert.Equal(t, []float64{-1, 2, -1, -1, 5, -1}, wm.(*tensor.Float64).Values)
}
