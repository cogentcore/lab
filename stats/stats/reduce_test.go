// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestReduce(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, 2, 3, 4, 5, 6)
	a.SetShapeSizes(2, 3)
	assert.Equal(t, 6.0, Reduce(a, StatMax).Float1D(0), "all of the values")
	assert.Equal(t, 21.0, Reduce(a, StatSum).Float1D(0))
	r := Reduce(a, StatSum, 0)
	assert.Equal(t, []int{3}, r.ShapeSizes())
	assert.Equal(t, []float64{5, 7, 9}, tensor.AsFloat64Slice(r))
	r = Reduce(a, StatMax, -1)
	assert.Equal(t, []float64{3, 6}, tensor.AsFloat64Slice(r))

	b := tensor.NewFloat64(2, 3, 4)
	for i := range b.Len() {
		b.Values[i] = float64(i)
	}
	r = Reduce(b, StatMin, 1)
	assert.Equal(t, []int{2, 4}, r.ShapeSizes())
	assert.Equal(t, []float64{0, 1, 2, 3, 12, 13, 14, 15}, tensor.AsFloat64Slice(r))
	assert.Panics(t, func() { Reduce(a, StatSum, 2) })
}
