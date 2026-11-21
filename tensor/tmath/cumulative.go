// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"slices"

	"cogentcore.org/lab/tensor"
)

// Clip returns the values of a limited to the range from lo to hi (as clip
// does in NumPy): lo where a is less than lo, and hi where it is greater than
// hi, with broadcasting (e.g., scalar limits).
func Clip(a, lo, hi tensor.Tensor) tensor.Values {
	return Min(Max(a, lo), hi)
}

// CumSum returns the cumulative sums of the values of the given tensor (as
// cumsum does in NumPy): of all of its values in order, as a 1D tensor, or,
// with an axis, of the values along that axis, with the shape of the tensor.
func CumSum(a tensor.Tensor, axis ...int) *tensor.Float64 {
	return cumulative(a, 0, func(agg, v float64) float64 { return agg + v }, axis...)
}

// CumProd returns the cumulative products of the values of the given tensor
// (as cumprod does in NumPy), as [CumSum] does sums.
func CumProd(a tensor.Tensor, axis ...int) *tensor.Float64 {
	return cumulative(a, 1, func(agg, v float64) float64 { return agg * v }, axis...)
}

// cumulative returns the cumulative values of the given function of the
// given tensor, starting from the given initial value, as [CumSum] does.
func cumulative(a tensor.Tensor, ini float64, fun func(agg, v float64) float64, axis ...int) *tensor.Float64 {
	if len(axis) == 0 {
		a = tensor.As1D(a)
		axis = []int{0}
	}
	out := tensor.NewFloat64(a.ShapeSizes()...)
	if a.Len() == 0 {
		return out
	}
	tensor.AxisLanes(a.Shape(), axis[0], func(lane []int) {
		agg := ini
		for _, i := range lane {
			agg = fun(agg, a.Float1D(i))
			out.Values[i] = agg
		}
	})
	return out
}

// Diff returns the differences between the successive values of the given
// tensor along the given axis, by default its last (as diff does in NumPy),
// with one value less along that axis.
func Diff(a tensor.Tensor, axis ...int) *tensor.Float64 {
	sizes := slices.Clone(a.ShapeSizes())
	if len(sizes) == 0 || a.Len() == 0 {
		return tensor.NewFloat64(0)
	}
	ax := len(sizes) - 1
	if len(axis) > 0 {
		ax = (axis[0] + len(sizes)) % len(sizes)
	}
	sizes[ax] = max(sizes[ax]-1, 0)
	out := tensor.NewFloat64(sizes...)
	osh := out.Shape()
	ash := a.Shape()
	for i := range out.Len() {
		idx := osh.IndexFrom1D(i)
		v0 := a.Float1D(ash.IndexTo1D(idx...))
		idx[ax]++
		out.Values[i] = a.Float1D(ash.IndexTo1D(idx...)) - v0
	}
	return out
}
