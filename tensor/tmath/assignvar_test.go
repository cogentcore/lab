// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestAssignVar(t *testing.T) {
	// a tensor variable is updated in place
	a := tensor.NewFloat64FromValues(1, 2)
	ta := tensor.Tensor(a)
	assert.NoError(t, AddAssignVar(&ta, 1))
	assert.Equal(t, []float64{2, 3}, a.Values)
	assert.NoError(t, MulAssignVar(&ta, tensor.NewFloat64Scalar(2)))
	assert.Equal(t, []float64{4, 6}, a.Values)
	assert.NoError(t, SubAssignVar(&a, 1)) // a *tensor.Float64, by reflection
	assert.Equal(t, []float64{3, 5}, a.Values)

	// a Go variable is set to the result
	x := 1.5
	assert.NoError(t, AddAssignVar(&x, tensor.NewFloat64Scalar(2)))
	assert.Equal(t, 3.5, x)
	assert.NoError(t, DivAssignVar(&x, 2))
	assert.Equal(t, 1.75, x)
	i := 7
	assert.NoError(t, DivAssignVar(&i, 2))
	assert.Equal(t, 3, i)
	assert.NoError(t, ModAssignVar(&i, 2))
	assert.Equal(t, 1, i)
	assert.NoError(t, SubAssignVar(&i, 1))
	assert.Equal(t, 0, i)
	f32 := float32(1)
	assert.NoError(t, MulAssignVar(&f32, 3))
	assert.Equal(t, float32(3), f32)
	s := "a"
	assert.NoError(t, AddAssignVar(&s, "b"))
	assert.Equal(t, "ab", s)
	assert.Error(t, AddAssignVar(x, 1))
	b := true
	assert.Error(t, AddAssignVar(&b, 1))
}
