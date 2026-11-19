// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"math"
	"slices"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/lab/tensor"
)

// Quantiles returns the given quantile(s) of the non-NaN values of the
// given tensor. qs are 0-1 values, 0 = min, 1 = max, .5 = median, etc.,
// using linear interpolation between the sorted values. For a 1D tensor,
// the output has the shape of qs. For a higher-dimensional tensor, the
// quantiles are computed for each cell across the rows (the outermost
// dimension), as for the other stats, with an output of the cells shape
// for one quantile, and nq x the cells shape for nq quantiles.
// The quantiles are 0 if there are no values.
// Because this requires a sort, it is more efficient to get as many
// quantiles as needed in one pass.
func Quantiles(in, qs tensor.Tensor) tensor.Values {
	return tensor.CallOut2(QuantilesOut, in, qs)
}

// QuantilesOut returns the given quantile(s) of the non-NaN values of the
// given tensor into the given output. See [Quantiles] for details.
func QuantilesOut(in, qs tensor.Tensor, out tensor.Values) error {
	if qs.NumDims() != 1 {
		return errors.Log(errors.New("stats.QuantilesFunc: only 1D quantile tensors allowed"))
	}
	nq := qs.Len()
	if in.NumDims() <= 1 {
		tensor.SetShapeFrom(out, qs)
		vals := sortedValues(nil, in, 0, 1, in.Len())
		for i := range nq {
			out.SetFloat1D(quantile(vals, qs.Float1D(i)), i)
		}
		return nil
	}
	rows, cells := in.Shape().RowCellSize()
	csz := tensor.CellsSize(in.ShapeSizes())
	if nq == 1 {
		out.SetShapeSizes(csz...)
	} else {
		out.SetShapeSizes(append([]int{nq}, csz...)...)
	}
	var vals []float64
	for c := range cells {
		vals = sortedValues(vals, in, c, cells, rows)
		for i := range nq {
			out.SetFloat1D(quantile(vals, qs.Float1D(i)), i*cells+c)
		}
	}
	return nil
}

// sortedValues returns the sorted non-NaN values of the given tensor,
// at n 1D indexes from start, every stride, reusing the given slice.
func sortedValues(vals []float64, in tensor.Tensor, start, stride, n int) []float64 {
	vals = vals[:0]
	for r := range n {
		v := in.Float1D(start + r*stride)
		if !math.IsNaN(v) {
			vals = append(vals, v)
		}
	}
	slices.Sort(vals)
	return vals
}

// quantile returns the q quantile of the given sorted values,
// using linear interpolation, and 0 if there are none.
func quantile(vals []float64, q float64) float64 {
	sz := len(vals) - 1
	if sz < 0 {
		return 0
	}
	qi := q * float64(sz)
	lwi := math.Floor(qi)
	lwii := int(lwi)
	switch {
	case lwii >= sz:
		return vals[sz]
	case lwii < 0:
		return vals[0]
	}
	phi := qi - lwi
	return (1-phi)*vals[lwii] + phi*vals[lwii+1]
}

// BoxStats returns the five values of a box plot of the non-NaN values
// of the given tensor, in this order: the low whisker, the first quartile
// (Q1), the median, the third quartile (Q3), and the high whisker.
// The whiskers are the most extreme values within whisker times the
// interquartile range (Q3 - Q1) below Q1 and above Q3 (Tukey's box plot,
// typically with whisker = 1.5), or the min and max values if whisker is 0.
// Values beyond the whiskers are outliers. For a 1D tensor, the output
// has 5 values; for a higher-dimensional tensor, the box stats are
// computed for each cell across the rows, with an output of 5 x the
// cells shape. The values are all 0 if there are no values.
func BoxStats(in tensor.Tensor, whisker float64) tensor.Values {
	out := tensor.NewFloat64()
	errors.Log(BoxStatsOut(in, whisker, out))
	return out
}

// BoxStatsOut returns the five values of a box plot of the non-NaN values
// of the given tensor into the given output. See [BoxStats] for details.
func BoxStatsOut(in tensor.Tensor, whisker float64, out tensor.Values) error {
	rows, cells := 0, 1
	if in.NumDims() <= 1 {
		rows = in.Len()
		out.SetShapeSizes(5)
	} else {
		rows, cells = in.Shape().RowCellSize()
		out.SetShapeSizes(append([]int{5}, tensor.CellsSize(in.ShapeSizes())...)...)
	}
	var vals []float64
	for c := range cells {
		vals = sortedValues(vals, in, c, cells, rows)
		q1, med, q3 := quantile(vals, .25), quantile(vals, .5), quantile(vals, .75)
		lo, hi := q1, q3
		if len(vals) > 0 {
			lo, hi = vals[0], vals[len(vals)-1]
		}
		if whisker > 0 {
			iqr := q3 - q1
			lmin, hmax := q1-whisker*iqr, q3+whisker*iqr
			lo, hi = q1, q3
			for _, v := range vals {
				if v >= lmin {
					lo = min(lo, v)
					break
				}
			}
			for k := len(vals) - 1; k >= 0; k-- {
				if vals[k] <= hmax {
					hi = max(hi, vals[k])
					break
				}
			}
		}
		for i, v := range []float64{lo, q1, med, q3, hi} {
			out.SetFloat1D(v, i*cells+c)
		}
	}
	return nil
}

// Median computes the median (50% quantile) of tensor values.
// See [StatsFunc] for general information.
func Median(in tensor.Tensor) tensor.Values {
	return Quantiles(in, tensor.NewFloat64Scalar(.5))
}

// Q1 computes the first quantile (25%) of tensor values.
// See [StatsFunc] for general information.
func Q1(in tensor.Tensor) tensor.Values {
	return Quantiles(in, tensor.NewFloat64Scalar(.25))
}

// Q3 computes the third quantile (75%) of tensor values.
// See [StatsFunc] for general information.
func Q3(in tensor.Tensor) tensor.Values {
	return Quantiles(in, tensor.NewFloat64Scalar(.75))
}

// MedianOut computes the median (50% quantile) of tensor values.
// See [StatsFunc] for general information.
func MedianOut(in tensor.Tensor, out tensor.Values) error {
	return QuantilesOut(in, tensor.NewFloat64Scalar(.5), out)
}

// Q1Out computes the first quantile (25%) of tensor values.
// See [StatsFunc] for general information.
func Q1Out(in tensor.Tensor, out tensor.Values) error {
	return QuantilesOut(in, tensor.NewFloat64Scalar(.25), out)
}

// Q3Out computes the third quantile (75%) of tensor values.
// See [StatsFunc] for general information.
func Q3Out(in tensor.Tensor, out tensor.Values) error {
	return QuantilesOut(in, tensor.NewFloat64Scalar(.75), out)
}
