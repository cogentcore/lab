// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"slices"

	"cogentcore.org/lab/tensor"
)

// Reduce returns the given statistic of the values of the given tensor, as
// the reductions of NumPy do (e.g., max(a) and a.max(axis)): with no axis,
// of all of its values, as a tensor of one value; with an axis (which can be
// negative to count from the last), of the values along that axis, for each
// index of the other dimensions, as a tensor of their shape, e.g., of each
// column of a 2D tensor for axis 0, and of each row for axis 1. In
// comparison, the stats functions such as [Max] compute the statistic over
// the rows (the first dimension) of a tensor for each of its cells, as for
// the columns of a table, which is the same as axis 0.
func Reduce(a tensor.Tensor, stat Stats, axis ...int) tensor.Values {
	if len(axis) == 0 {
		return stat.Call(tensor.As1D(a))
	}
	sizes := a.ShapeSizes()
	nd := len(sizes)
	ax := axis[0]
	if ax < 0 {
		ax += nd
	}
	if ax < 0 || ax >= nd {
		panic(fmt.Sprintf("stats.Reduce: axis %d is out of range for %d dimensions", axis[0], nd))
	}
	n := sizes[ax]
	osizes := slices.Delete(slices.Clone(sizes), ax, ax+1)
	cells := 1
	for _, s := range osizes {
		cells *= s
	}
	vals := tensor.NewFloat64(n, cells) // the values along the axis as the rows
	li := 0
	tensor.AxisLanes(a.Shape(), ax, func(lane []int) {
		for k, i := range lane {
			vals.Values[k*cells+li] = a.Float1D(i)
		}
		li++
	})
	out := stat.Call(vals)
	if len(osizes) > 0 {
		out.SetShapeSizes(osizes...)
	}
	return out
}
