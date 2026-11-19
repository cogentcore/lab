// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package histogram

import (
	"math"
	"slices"
	"sort"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/lab/tensor"
)

// kdeCutoff is the number of bandwidths beyond which the gaussian kernel
// is not computed: exp(-0.5 * 6^2) = 1.5e-8 of the peak.
const kdeCutoff = 6

// sortedValues returns the sorted non-NaN values of the given tensor.
func sortedValues(vals tensor.Tensor) []float64 {
	n := vals.Len()
	sv := make([]float64, 0, n)
	for i := range n {
		v := vals.Float1D(i)
		if !math.IsNaN(v) {
			sv = append(sv, v)
		}
	}
	slices.Sort(sv)
	return sv
}

// bandwidth returns Silverman's rule of thumb bandwidth for the
// given sorted values.
func bandwidth(sv []float64) float64 {
	n := len(sv)
	if n < 2 {
		return 1
	}
	mean, ss := 0.0, 0.0
	for _, v := range sv {
		mean += v
	}
	mean /= float64(n)
	for _, v := range sv {
		ss += (v - mean) * (v - mean)
	}
	sd := math.Sqrt(ss / float64(n-1))
	q := func(p float64) float64 { // linear interpolation
		qi := p * float64(n-1)
		lw := math.Floor(qi)
		li := int(lw)
		if li >= n-1 {
			return sv[n-1]
		}
		return sv[li] + (qi-lw)*(sv[li+1]-sv[li])
	}
	iqr := (q(.75) - q(.25)) / 1.34
	a := min(sd, iqr)
	if a == 0 {
		a = max(sd, iqr)
	}
	if a == 0 { // all the same value
		return 1
	}
	return 0.9 * a * math.Pow(float64(n), -0.2)
}

// KDEBandwidth returns the bandwidth (the standard deviation of the
// gaussian kernel) for a kernel density estimate of the non-NaN values
// of the given tensor, using Silverman's rule of thumb:
// 0.9 * min(sd, IQR / 1.34) * n^(-1/5), which is robust to outliers and
// skewed distributions, and works well for unimodal distributions
// (multimodal ones are better shown with a smaller bandwidth).
// It is 1 if there are fewer than 2 values, or they are all the same.
func KDEBandwidth(vals tensor.Tensor) float64 {
	return bandwidth(sortedValues(vals))
}

// KDE returns the gaussian kernel density estimate of the distribution
// of the non-NaN values of the given tensor, at each of the given grid
// values (e.g., from [KDEGrid]), with the given bandwidth (the standard
// deviation of the gaussian kernel; [KDEBandwidth] if <= 0).
// The density integrates to 1 over the values. It is the smooth
// version of a histogram, e.g., for the body of a violin plot.
// The output has the shape of the grid.
func KDE(vals, grid tensor.Tensor, bandwidth float64) tensor.Values {
	out := tensor.NewFloat64()
	errors.Log(KDEOut(vals, grid, bandwidth, out))
	return out
}

// KDEOut computes the gaussian kernel density estimate into the given
// output tensor. See [KDE] for details.
func KDEOut(vals, grid tensor.Tensor, bw float64, out tensor.Values) error {
	sv := sortedValues(vals)
	if bw <= 0 {
		bw = bandwidth(sv)
	}
	tensor.SetShapeFrom(out, grid)
	n := len(sv)
	ng := grid.Len()
	if n == 0 {
		out.SetZeros()
		return nil
	}
	norm := 1 / (float64(n) * bw * math.Sqrt(2*math.Pi))
	cut := kdeCutoff * bw
	tensor.VectorizeThreaded(min(n, 100)*10, func(tsr ...tensor.Tensor) int { return ng },
		func(idx int, tsr ...tensor.Tensor) {
			g := tsr[0].Float1D(idx)
			// only the values within the cutoff of g, in sorted order
			st := sort.SearchFloat64s(sv, g-cut)
			d := 0.0
			for _, v := range sv[st:] {
				if v > g+cut {
					break
				}
				z := (g - v) / bw
				d += math.Exp(-0.5 * z * z)
			}
			tsr[1].SetFloat1D(d*norm, idx)
		}, grid, out)
	return nil
}

// KDEGrid returns n evenly-spaced grid values for a kernel density
// estimate of the non-NaN values of the given tensor, from the min to
// the max value, extended by cut times the bandwidth on each side
// (seaborn uses a cut of 2 for violin plots, and 0 is the range of the
// data, as in matplotlib violin plots). The bandwidth is [KDEBandwidth]
// if <= 0. The grid is empty if there are no values.
func KDEGrid(vals tensor.Tensor, n int, bw, cut float64) *tensor.Float64 {
	sv := sortedValues(vals)
	if len(sv) == 0 || n <= 0 {
		return tensor.NewFloat64(0)
	}
	if bw <= 0 {
		bw = bandwidth(sv)
	}
	lo, hi := sv[0]-cut*bw, sv[len(sv)-1]+cut*bw
	grid := tensor.NewFloat64(n)
	if n == 1 {
		grid.SetFloat1D(0.5*(lo+hi), 0)
		return grid
	}
	for i := range n {
		grid.SetFloat1D(lo+(hi-lo)*float64(i)/float64(n-1), i)
	}
	return grid
}
