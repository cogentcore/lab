// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"cmp"
	"math"
	"slices"
)

// Sorted returns a sorted copy of the given tensor (as sort does in NumPy),
// with the values along the given axis sorted in ascending order, for each
// index of the other dimensions: by default the last axis, e.g., the values
// of each row of a 2D tensor, and axis 0 sorts each column. NaN values are
// sorted last, and strings are sorted as strings.
func Sorted(tsr Tensor, axis ...int) Values {
	out := Clone(tsr)
	idx := ArgSort(tsr, axis...)
	AxisLanes(tsr.Shape(), axisArg(tsr, axis), func(lane []int) {
		for _, i := range lane {
			setFrom1D(out, i, tsr, lane[idx.Values[i]])
		}
	})
	return out
}

// ArgSort returns the indexes that sort the values of the given tensor
// along the given axis (as argsort does in NumPy), with the shape of the
// tensor: by default the last axis, e.g., the indexes of the columns of the
// sorted values of each row of a 2D tensor. The sort is stable, and NaN
// values are sorted last. For example, the indexes of a 1D tensor of the
// values of a column select the rows of a matrix in the order of the column,
// as in a[argsort(a[:, 0]), :] in Goal math mode.
func ArgSort(tsr Tensor, axis ...int) *Int {
	out := NewInt(tsr.ShapeSizes()...)
	str := tsr.IsString()
	AxisLanes(tsr.Shape(), axisArg(tsr, axis), func(lane []int) {
		order := make([]int, len(lane))
		for k := range order {
			order[k] = k
		}
		slices.SortStableFunc(order, func(a, b int) int {
			if str {
				return cmp.Compare(tsr.String1D(lane[a]), tsr.String1D(lane[b]))
			}
			return cmp.Compare(tsr.Float1D(lane[a]), tsr.Float1D(lane[b])) // NaN first
		})
		if !str { // NaN last
			nan := 0
			for nan < len(order) && math.IsNaN(tsr.Float1D(lane[order[nan]])) {
				nan++
			}
			order = append(order[nan:], order[:nan]...)
		}
		for k, i := range lane {
			out.Values[i] = order[k]
		}
	})
	return out
}

// ArgWhere returns the indexes of the non-zero (true) values of the given
// tensor (as argwhere does in NumPy), as a 2D tensor with a row for each
// value, of its index in each dimension of the tensor. These are indexes
// of the arbitrary indexing of Goal math mode, so that a[argwhere(a > 5)]
// is the values of a that are greater than 5. For a 1D tensor, the indexes
// are the column a[argwhere(a > 5)][:, 0], as nonzero is in NumPy.
func ArgWhere(tsr Tensor) *Int {
	nd := tsr.NumDims()
	sh := tsr.Shape()
	var idxs []int
	for i := range tsr.Len() {
		if tsr.Float1D(i) != 0 {
			idxs = append(idxs, sh.IndexFrom1D(i)...)
		}
	}
	out := NewIntFromValues(idxs...)
	out.SetShapeSizes(len(idxs)/max(nd, 1), nd)
	return out
}

// axisArg returns the axis given as an optional argument for a tensor:
// its last axis by default, and counting from the last if negative.
func axisArg(tsr Tensor, axis []int) int {
	if len(axis) == 0 {
		return max(tsr.NumDims()-1, 0)
	}
	return axisIndex(axis[0], tsr.NumDims())
}

// AxisLanes calls the given function with the 1D indexes of the values of
// each lane of a tensor of the given shape along the given axis: the values
// along the axis, for each index of the other dimensions, e.g., each row of
// a 2D tensor for axis 1, and each column for axis 0. The lane slice is
// reused for each call.
func AxisLanes(sh *Shape, axis int, fun func(lane []int)) {
	if sh.Len() == 0 {
		return
	}
	n := sh.DimSize(axis)
	inner := 1 // the stride of the axis
	for d := axis + 1; d < sh.NumDims(); d++ {
		inner *= sh.DimSize(d)
	}
	outer := sh.Len() / (n * inner)
	lane := make([]int, n)
	for o := range outer {
		for in := range inner {
			for k := range n {
				lane[k] = (o*n+k)*inner + in
			}
			fun(lane)
		}
	}
}
