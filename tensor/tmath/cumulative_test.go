// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestClip(t *testing.T) {
	a := tensor.NewFloat64FromValues(-1, 2, 5, 9)
	c := Clip(a, tensor.NewFloat64Scalar(0), tensor.NewFloat64Scalar(5))
	assert.Equal(t, []float64{0, 2, 5, 5}, tensor.AsFloat64Slice(c))
}

func TestCumulative(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, 2, 3, 4, 5, 6)
	a.SetShapeSizes(2, 3)
	cs := CumSum(a)
	assert.Equal(t, []int{6}, cs.ShapeSizes(), "all of the values")
	assert.Equal(t, []float64{1, 3, 6, 10, 15, 21}, cs.Values)
	cs = CumSum(a, 0)
	assert.Equal(t, []int{2, 3}, cs.ShapeSizes())
	assert.Equal(t, []float64{1, 2, 3, 5, 7, 9}, cs.Values)
	cs = CumSum(a, 1)
	assert.Equal(t, []float64{1, 3, 6, 4, 9, 15}, cs.Values)
	assert.Equal(t, []float64{1, 2, 6, 24, 120, 720}, CumProd(a).Values)
}

func TestDiff(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, 4, 9, 2, 4, 8)
	a.SetShapeSizes(2, 3)
	d := Diff(a)
	assert.Equal(t, []int{2, 2}, d.ShapeSizes())
	assert.Equal(t, []float64{3, 5, 2, 4}, d.Values)
	d = Diff(a, 0)
	assert.Equal(t, []int{1, 3}, d.ShapeSizes())
	assert.Equal(t, []float64{1, 0, -1}, d.Values)
}
