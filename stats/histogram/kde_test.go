// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package histogram

import (
	"math"
	"math/rand"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

// normalSamples returns n gaussian samples with the given mean and sd.
func normalSamples(n int, mean, sd float64) *tensor.Float64 {
	rng := rand.New(rand.NewSource(1))
	t := tensor.NewFloat64(n)
	for i := range n {
		t.SetFloat1D(mean+sd*rng.NormFloat64(), i)
	}
	return t
}

func TestKDE(t *testing.T) {
	vals := normalSamples(2000, 3, 2)
	vals.SetFloat1D(math.NaN(), 0)
	bw := KDEBandwidth(vals)
	// Silverman: 0.9 * ~2 * 1999^-0.2 = ~0.39
	assert.InDelta(t, 0.39, bw, 0.03)

	grid := KDEGrid(vals, 200, 0, 3)
	assert.Equal(t, 200, grid.Len())
	mn, mx := math.Inf(1), math.Inf(-1)
	for i := range vals.Len() {
		if v := vals.Float1D(i); !math.IsNaN(v) {
			mn, mx = min(mn, v), max(mx, v)
		}
	}
	assert.InDelta(t, mn-3*bw, grid.Float1D(0), 1e-9)
	assert.InDelta(t, mx+3*bw, grid.Float1D(199), 1e-9)

	dens := KDE(vals, grid, 0)
	assert.Equal(t, []int{200}, dens.ShapeSizes())
	// it integrates to 1 (trapezoid rule)
	dx := grid.Float1D(1) - grid.Float1D(0)
	sum, peak, peakX := 0.0, 0.0, 0.0
	for i := range dens.Len() {
		d := dens.Float1D(i)
		sum += d * dx
		if d > peak {
			peak, peakX = d, grid.Float1D(i)
		}
	}
	assert.InDelta(t, 1, sum, 0.01)
	assert.InDelta(t, 3, peakX, 0.4)
	// the peak of a normal with sd 2 is 0.2, a bit lower with smoothing
	assert.InDelta(t, 0.19, peak, 0.02)

	// a single value with a given bandwidth is that gaussian
	one := KDE(tensor.NewFloat64FromValues(1), tensor.NewFloat64FromValues(1, 2), 1)
	assert.InDelta(t, 1/math.Sqrt(2*math.Pi), one.Float1D(0), 1e-12)
	assert.InDelta(t, math.Exp(-0.5)/math.Sqrt(2*math.Pi), one.Float1D(1), 1e-12)

	// degenerate cases
	assert.Equal(t, 1.0, KDEBandwidth(tensor.NewFloat64FromValues(2, 2, 2)))
	assert.Equal(t, 1.0, KDEBandwidth(tensor.NewFloat64FromValues(2)))
	assert.Equal(t, 0, KDEGrid(tensor.NewFloat64FromValues(math.NaN()), 10, 0, 2).Len())
	assert.Equal(t, 0.0, KDE(tensor.NewFloat64FromValues(math.NaN()), grid, 0).Float1D(5))
	assert.Equal(t, 1, KDEGrid(vals, 1, 1, 0).Len())
}

func BenchmarkKDE(b *testing.B) {
	vals := normalSamples(10000, 0, 1)
	grid := KDEGrid(vals, 100, 0, 2)
	for b.Loop() {
		KDE(vals, grid, 0)
	}
}
