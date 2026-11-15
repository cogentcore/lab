// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package convolve

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSlice32Errors tests the errors for empty, even-sized and too-large kernels.
func TestSlice32Errors(t *testing.T) {
	var dest []float32
	assert.Error(t, Slice32(&dest, nil, []float32{1}))
	assert.Error(t, Slice32(&dest, []float32{1, 2, 3}, nil))
	assert.Error(t, Slice32(&dest, []float32{1, 2, 3}, []float32{.5, .5}))
	assert.Error(t, Slice32(&dest, []float32{1, 2}, []float32{.25, .5, .25}))
	assert.Nil(t, dest)
}

// TestSlice64Errors tests the errors for empty, even-sized and too-large kernels.
func TestSlice64Errors(t *testing.T) {
	var dest []float64
	assert.Error(t, Slice64(&dest, nil, []float64{1}))
	assert.Error(t, Slice64(&dest, []float64{1, 2, 3}, nil))
	assert.Error(t, Slice64(&dest, []float64{1, 2, 3}, []float64{.5, .5}))
	assert.Error(t, Slice64(&dest, []float64{1, 2}, []float64{.25, .5, .25}))
	assert.Nil(t, dest)
}

// TestSlice32Known tests the interior and renormalized edge values
// of a small convolution computed by hand.
func TestSlice32Known(t *testing.T) {
	kern := []float32{.25, .5, .25}
	src := []float32{1, 2, 3, 4, 5}
	dest := make([]float32, 0, 10) // capacity is reused
	require.NoError(t, Slice32(&dest, src, kern))
	ex := []float32{4.0 / 3, 2, 3, 4, 14.0 / 3}
	assert.Len(t, dest, len(src))
	assert.Equal(t, 10, cap(dest))
	assert.InDeltaSlice(t, ex, dest, 1e-6)
}

// TestSlice64Known tests the interior and renormalized edge values
// of a small convolution computed by hand.
func TestSlice64Known(t *testing.T) {
	kern := []float64{.25, .5, .25}
	src := []float64{1, 2, 3, 4, 5}
	var dest []float64
	require.NoError(t, Slice64(&dest, src, kern))
	ex := []float64{4.0 / 3, 2, 3, 4, 14.0 / 3}
	assert.InDeltaSlice(t, ex, dest, 1e-12)
}

// TestSliceConstant tests that a constant source stays constant everywhere,
// including the renormalized edges, with a wider kernel.
func TestSliceConstant(t *testing.T) {
	k32 := GaussianKernel32(3, .5)
	k64 := GaussianKernel64(3, .5)
	src32 := make([]float32, 12)
	src64 := make([]float64, 12)
	for i := range src32 {
		src32[i] = 7
		src64[i] = 7
	}
	var d32 []float32
	var d64 []float64
	require.NoError(t, Slice32(&d32, src32, k32))
	require.NoError(t, Slice64(&d64, src64, k64))
	for i := range src32 {
		assert.InDelta(t, 7, d32[i], 1e-5, "index %d", i)
		assert.InDelta(t, 7, d64[i], 1e-12, "index %d", i)
	}
}

// TestSlice64Ramp tests that the interior of a linear ramp is unchanged
// by a symmetric kernel, and that the edges are pulled inward.
func TestSlice64Ramp(t *testing.T) {
	kern := GaussianKernel64(2, .5)
	src := make([]float64, 10)
	for i := range src {
		src[i] = float64(i)
	}
	var dest []float64
	require.NoError(t, Slice64(&dest, src, kern))
	for i := 2; i < 8; i++ {
		assert.InDelta(t, float64(i), dest[i], 1e-12)
	}
	assert.Greater(t, dest[0], 0.0)
	assert.Less(t, dest[9], 9.0)
}

// TestGaussianKernel tests that the kernels are normalized, symmetric and
// have the known values for a half-width of 1 and sigma of 1.
func TestGaussianKernel(t *testing.T) {
	k64 := GaussianKernel64(1, 1)
	e := math.Exp(-0.5)
	sum := 1 + 2*e
	assert.InDeltaSlice(t, []float64{e / sum, 1 / sum, e / sum}, k64, 1e-12)

	k32 := GaussianKernel32(1, 1)
	assert.InDeltaSlice(t, []float32{float32(e / sum), float32(1 / sum), float32(e / sum)}, k32, 1e-6)

	for _, kh := range []int{1, 2, 5} {
		k := GaussianKernel64(kh, .5)
		assert.Len(t, k, 2*kh+1)
		s := 0.0
		for i, v := range k {
			s += v
			assert.InDelta(t, k[len(k)-1-i], v, 1e-15)
			if i < kh {
				assert.Less(t, v, k[i+1])
			}
		}
		assert.InDelta(t, 1, s, 1e-12)

		k32 := GaussianKernel32(kh, .5)
		var s32 float32
		for _, v := range k32 {
			s32 += v
		}
		assert.InDelta(t, 1, s32, 1e-6)
	}
}
