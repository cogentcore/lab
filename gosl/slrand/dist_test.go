// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slrand

import (
	"testing"

	"cogentcore.org/core/math32"
	"cogentcore.org/lab/gosl/sltype"
	"github.com/stretchr/testify/assert"
)

// TestVec2Consistency tests that the Vec2 functions return the values of
// the corresponding scalar conversions of the same two random uint32s.
func TestVec2Consistency(t *testing.T) {
	var counter uint64 = 12345
	for i := range uint32(20) {
		u := Uint32Vec2(counter, i, 7)
		assert.Equal(t, u.X, Uint32(counter, i, 7))
		assert.Equal(t, sltype.Uint32ToFloat32Vec2(u), Float32Vec2(counter, i, 7))
		assert.Equal(t, Float32(counter, i, 7), Float32Vec2(counter, i, 7).X)

		r11 := Float32Range11Vec2(counter, i, 7)
		assert.Equal(t, sltype.Uint32ToFloat32Range11(u.X), r11.X)
		assert.Equal(t, sltype.Uint32ToFloat32Range11(u.Y), r11.Y)
		assert.Equal(t, Float32Range11(counter, i, 7), r11.X)

		assert.Equal(t, Float32NormVec2(counter, i, 7).X, Float32Norm(counter, i, 7))
	}
}

// TestDistributions tests the ranges, means and variances of the uniform
// and normal distributions over many samples.
func TestDistributions(t *testing.T) {
	const n = 20000
	var counter uint64
	var sum01, sum11, sumN, sumN2, neg11 float64
	for i := range uint32(n) {
		f := Float32Vec2(counter, i, 3)
		r := Float32Range11Vec2(counter, i, 3)
		g := Float32NormVec2(counter, i, 3)
		for _, v := range []float32{f.X, f.Y} {
			assert.True(t, v > 0 && v < 1, "uniform 01 out of range: %g", v)
			sum01 += float64(v)
		}
		for _, v := range []float32{r.X, r.Y} {
			assert.True(t, v >= -1 && v <= 1, "uniform 11 out of range: %g", v)
			sum11 += float64(v)
			if v < 0 {
				neg11++
			}
		}
		for _, v := range []float32{g.X, g.Y} {
			sumN += float64(v)
			sumN2 += float64(v) * float64(v)
		}
	}
	m := float64(2 * n)
	assert.InDelta(t, 0.5, sum01/m, 0.01)
	assert.InDelta(t, 0, sum11/m, 0.02)
	assert.InDelta(t, 0.5, neg11/m, 0.02)
	mean := sumN / m
	assert.InDelta(t, 0, mean, 0.02)
	assert.InDelta(t, 1, sumN2/m-mean*mean, 0.03)
}

// TestBoolP tests the frequency of true values of BoolP, and the extremes.
func TestBoolP(t *testing.T) {
	var counter uint64
	ntrue := 0
	for i := range uint32(10000) {
		if BoolP(counter, i, 1, 0.3) {
			ntrue++
		}
		assert.False(t, BoolP(counter, i, 1, 0))
		assert.True(t, BoolP(counter, i, 1, 1))
	}
	assert.InDelta(t, 3000, ntrue, 150)
}

// TestSincosPi tests SincosPi at known points: X is cos and Y is sin of pi x.
func TestSincosPi(t *testing.T) {
	for _, tc := range []struct{ x, cos, sin float32 }{
		{0, 1, 0}, {0.5, 0, 1}, {1, -1, 0}, {-0.5, 0, -1}, {0.25, math32.Sqrt2 / 2, math32.Sqrt2 / 2},
	} {
		r := SincosPi(tc.x)
		assert.InDelta(t, tc.cos, r.X, 1e-6, "cos %g", tc.x)
		assert.InDelta(t, tc.sin, r.Y, 1e-6, "sin %g", tc.x)
	}
}

// TestCounter tests seeding, adding to and resetting a Counter.
func TestCounter(t *testing.T) {
	var ct Counter
	ct.Seed(5)
	assert.Equal(t, uint32(5), ct.HiSeed)
	assert.Equal(t, uint64(5)<<32, ct.Counter)
	ct.Add(10)
	ct.Add(0xffffffff)
	assert.Equal(t, uint64(5)<<32+10+0xffffffff, ct.Counter)
	ct.Reset()
	assert.Equal(t, uint64(5)<<32, ct.Counter)
}
