// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ticks

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForRangeKnown tests ForRange on ranges with known nice tick values.
func TestForRangeKnown(t *testing.T) {
	tests := []struct {
		mn, mx  float64
		contain Containments
		want    []float64
		step, q float64
		mag     int
	}{
		{0, 10, Free, []float64{0, 2.5, 5, 7.5, 10}, 2.5, 2.5, 0},
		{0, 1, ContainData, []float64{0, 0.25, 0.5, 0.75, 1}, 2.5, 2.5, -1},
		{-3.7, 12.2, Free, []float64{-4, 0, 4, 8, 12}, 4, 4, 0},
		{-3.7, 12.2, ContainData, []float64{-5, 0, 5, 10, 15}, 5, 5, 0},
		{-3.7, 12.2, WithinData, []float64{-3, 0, 3, 6, 9, 12}, 3, 3, 0},
		{1e6, 5e7, WithinData, []float64{1e7, 2e7, 3e7, 4e7, 5e7}, 1, 1, 7},
		{1e6, 5e7, ContainData, []float64{0, 1e7, 2e7, 3e7, 4e7, 5e7}, 1, 1, 7},
	}
	for _, tt := range tests {
		v, step, q, mag := ForRange(tt.mn, tt.mx, 5, tt.contain)
		require.Len(t, v, len(tt.want), "%v %v", tt.mn, tt.mx)
		for i := range v {
			assert.InDelta(t, tt.want[i], v[i], 1.0e-9*math.Abs(tt.mx))
		}
		assert.Equal(t, tt.step, step)
		assert.Equal(t, tt.q, q)
		assert.Equal(t, tt.mag, mag)
	}
}

// TestForRangeContainment tests the containment guarantees and even
// spacing of tick values over a variety of ranges.
func TestForRangeContainment(t *testing.T) {
	ranges := [][2]float64{{0, 10}, {-3.7, 12.2}, {0.0012, 0.0047}, {100, 100.5}, {-1e5, -2}, {1, 1e9}, {-0.3, 0.2}}
	for _, r := range ranges {
		for _, n := range []int{3, 5, 10} {
			for _, c := range ContainmentsValues() {
				v, step, q, mag := ForRange(r[0], r[1], n, c)
				require.GreaterOrEqual(t, len(v), 2, "%v %d %s", r, n, c)
				d := step * math.Pow10(mag)
				tol := 1.0e-9 * (r[1] - r[0])
				for i := 1; i < len(v); i++ {
					assert.InDelta(t, d, v[i]-v[i-1], tol, "%v %d %s", r, n, c)
				}
				switch c {
				case ContainData:
					assert.LessOrEqual(t, v[0], r[0]+tol)
					assert.GreaterOrEqual(t, v[len(v)-1], r[1]-tol)
				case WithinData:
					assert.GreaterOrEqual(t, v[0], r[0]-tol)
					assert.LessOrEqual(t, v[len(v)-1], r[1]+tol)
				}
				assert.Contains(t, []float64{1, 5, 2, 2.5, 4, 3, 10, 20, 25, 30, 40, 50, 15, 6, 8, 12, 7.5, 9, 4.5}, step, "%v %d %s", r, n, c)
				assert.NotZero(t, q)
			}
		}
	}
}

// TestForRangeDegenerate tests a zero-width range, which returns
// the requested number of identical values, and an invalid range.
func TestForRangeDegenerate(t *testing.T) {
	v, step, q, mag := ForRange(5, 5, 4, Free)
	assert.Equal(t, []float64{5, 5, 5, 5}, v)
	assert.Equal(t, 0.0, step)
	assert.Equal(t, 0.0, q)
	assert.Equal(t, 0, mag)

	assert.Panics(t, func() { ForRange(2, 1, 5, Free) })
}

// TestTalbotLinHanrahanFallback tests the fallback of evenly spaced values
// when no candidate scores above the minimum, using negative weights,
// along with custom nice numbers and legibility functions.
func TestTalbotLinHanrahanFallback(t *testing.T) {
	w := &weights{simplicity: -10, coverage: -10, density: -10, legibility: -10}
	v, step, q, mag := TalbotLinHanrahan(10, 40, 4, Free, nil, w, nil)
	assert.Equal(t, []float64{10, 20, 30, 40}, v)
	assert.Equal(t, 10.0, step)
	assert.Equal(t, 0.0, q)
	assert.Equal(t, 1, mag)

	d, fc, prec := Format(v, step, q, mag)
	assert.Equal(t, 10.0, d) // label distance for the fallback
	assert.Equal(t, byte('f'), fc)
	assert.Equal(t, 0, prec)

	legible := func(lMin, lMax, lStep float64) float64 {
		if lStep == 2 {
			return 1
		}
		return 0
	}
	w = &weights{simplicity: 0.25, coverage: 0.2, density: 0.5, legibility: 10}
	v, step, q, _ = TalbotLinHanrahan(0, 10, 5, ContainData, []float64{1, 2, 5}, w, legible)
	assert.Equal(t, []float64{0, 2, 4, 6, 8, 10}, v)
	assert.Equal(t, 2.0, step)
	assert.Equal(t, 2.0, q)
}

// TestFormat tests the format and precision chosen for tick labels.
func TestFormat(t *testing.T) {
	v, step, q, mag := ForRange(-3.7, 12.2, 5, Free)
	d, fc, prec := Format(v, step, q, mag)
	assert.Equal(t, 4.0, d)
	assert.Equal(t, byte('f'), fc)
	assert.Equal(t, 0, prec)

	v, step, q, mag = ForRange(1e6, 5e7, 5, WithinData)
	d, fc, prec = Format(v, step, q, mag)
	assert.Equal(t, 1e7, d)
	assert.Equal(t, byte('g'), fc)
	assert.Equal(t, 1, prec)

	v, step, q, mag = ForRange(0.0012, 0.0047, 5, Free)
	d, fc, prec = Format(v, step, q, mag)
	assert.InDelta(t, 0.001, d, 1.0e-15)
	assert.Equal(t, byte('g'), fc)
	assert.Equal(t, 3, prec)
}

// TestScoreHelpers tests the individual scoring helper functions.
func TestScoreHelpers(t *testing.T) {
	Q := []float64{1, 5, 2}
	assert.Equal(t, 2.0, maxSimplicity(1, Q, 0))
	assert.Equal(t, 2.0, simplicity(1, Q, 0, 0, 10, 1)) // includes zero
	assert.Equal(t, -0.5, simplicity(5, Q, 1, 1, 10, 5))
	assert.Panics(t, func() { simplicity(7, Q, 1, 0, 1, 1) })
	assert.Panics(t, func() { maxSimplicity(7, Q, 1) })

	assert.Equal(t, 1.0, coverage(0, 10, 0, 10))
	assert.Equal(t, 1.0, maxCoverage(0, 10, 5))
	assert.InDelta(t, 0.75, maxCoverage(0, 10, 11), 1.0e-12)

	assert.Equal(t, 1.0, density(5, 5, 0, 4, 0, 4))
	assert.Equal(t, 1.0, maxDensity(3, 5))
	assert.Equal(t, 0.0, maxDensity(9, 5))
	assert.Equal(t, 1.0, unitLegibility(0, 1, 1))
	assert.Equal(t, -1, minAbsMag(0.5, 100))
}

// TestContainmentsEnum tests the generated Containments enum methods.
func TestContainmentsEnum(t *testing.T) {
	var c Containments
	require.NoError(t, c.SetString("WithinData"))
	assert.Equal(t, WithinData, c)
	assert.Equal(t, "WithinData", c.String())
	assert.Error(t, c.SetString("Nope"))
	assert.Len(t, c.Values(), int(ContainmentsN))
	txt, err := ContainData.MarshalText()
	require.NoError(t, err)
	require.NoError(t, c.UnmarshalText(txt))
	assert.Equal(t, ContainData, c)
	assert.NotEmpty(t, ContainData.Desc())
	assert.Equal(t, int64(2), WithinData.Int64())
	c.SetInt64(0)
	assert.Equal(t, Free, c)
}
