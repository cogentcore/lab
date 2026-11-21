// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"fmt"
	"reflect"
	"slices"
)

// Concatenate returns a new tensor with the values of the given tensors
// joined along the given axis (as concatenate does in NumPy), which can be
// negative to count from the last. The tensors must have the same number of
// dimensions, and the same sizes in all of the other dimensions. The result
// has the type of the tensors if they all have the same type, strings if
// any of them has strings, and float64 values otherwise. It panics if the
// shapes do not match.
func Concatenate(axis int, tsrs ...Tensor) Values {
	if len(tsrs) == 0 {
		return NewFloat64(0)
	}
	nd := tsrs[0].NumDims()
	ax := axisIndex(axis, nd)
	sizes := slices.Clone(tsrs[0].ShapeSizes())
	starts := make([]int, len(tsrs)) // the start of each tensor along the axis
	sizes[ax] = 0
	for i, t := range tsrs {
		ts := t.ShapeSizes()
		if len(ts) != nd {
			panic(fmt.Sprintf("tensor.Concatenate: tensor %d has %d dimensions, not %d", i, len(ts), nd))
		}
		for d := range nd {
			if d != ax && ts[d] != sizes[d] {
				panic(fmt.Sprintf("tensor.Concatenate: tensor %d has size %d in dimension %d, not %d", i, ts[d], d, sizes[d]))
			}
		}
		starts[i] = sizes[ax]
		sizes[ax] += ts[ax]
	}
	out := NewOfType(commonType(tsrs...), sizes...)
	osh := out.Shape()
	for i, t := range tsrs {
		tsh := t.Shape()
		for j := range t.Len() {
			idx := tsh.IndexFrom1D(j)
			idx[ax] += starts[i]
			setFrom1D(out, osh.IndexTo1D(idx...), t, j)
		}
	}
	return out
}

// Stack returns a new tensor with the values of the given tensors, which
// must have the same shape, stacked along a new axis at the given position
// (as stack does in NumPy), e.g., axis 0 makes a tensor of each of them.
func Stack(axis int, tsrs ...Tensor) Values {
	if len(tsrs) == 0 {
		return NewFloat64(0)
	}
	ax := axisIndex(axis, tsrs[0].NumDims()+1)
	exp := make([]Tensor, len(tsrs))
	for i, t := range tsrs {
		exp[i] = Reshape(t, slices.Insert(slices.Clone(t.ShapeSizes()), ax, 1)...)
	}
	return Concatenate(ax, exp...)
}

// VStack returns a new tensor with the given tensors stacked vertically,
// as rows (as vstack does in NumPy): 1D tensors are each one row, and
// other tensors are concatenated along their first axis.
func VStack(tsrs ...Tensor) Values {
	exp := make([]Tensor, len(tsrs))
	for i, t := range tsrs {
		exp[i] = t
		if t.NumDims() < 2 {
			exp[i] = Reshape(t, 1, t.Len())
		}
	}
	return Concatenate(0, exp...)
}

// HStack returns a new tensor with the given tensors stacked horizontally,
// as columns (as hstack does in NumPy): 1D tensors are concatenated, and
// other tensors are concatenated along their second axis.
func HStack(tsrs ...Tensor) Values {
	if len(tsrs) > 0 && tsrs[0].NumDims() < 2 {
		return Concatenate(0, tsrs...)
	}
	return Concatenate(1, tsrs...)
}

// ColumnStack returns a new 2D tensor with the given 1D tensors as its
// columns (as column_stack does in NumPy); 2D tensors are concatenated as
// columns, as in [HStack].
func ColumnStack(tsrs ...Tensor) Values {
	exp := make([]Tensor, len(tsrs))
	for i, t := range tsrs {
		exp[i] = t
		if t.NumDims() < 2 {
			exp[i] = Reshape(t, t.Len(), 1)
		}
	}
	return Concatenate(1, exp...)
}

// Tile returns a new tensor with the given tensor repeated the given number
// of times along each dimension (as tile does in NumPy), e.g., Tile(a, 2, 3)
// of a 2D tensor a is 2 copies of a down and 3 across. With fewer repeats
// than dimensions, the repeats are of the last dimensions, and with more,
// the tensor has new dimensions of size 1 before its first one.
func Tile(tsr Tensor, reps ...int) Values {
	sizes := slices.Clone(tsr.ShapeSizes())
	for len(sizes) < len(reps) {
		sizes = slices.Insert(sizes, 0, 1)
	}
	for len(reps) < len(sizes) {
		reps = slices.Insert(slices.Clone(reps), 0, 1)
	}
	src := Reshape(tsr, sizes...)
	osizes := make([]int, len(sizes))
	for d := range sizes {
		osizes[d] = sizes[d] * reps[d]
	}
	out := NewOfType(tsr.DataType(), osizes...)
	osh, ssh := out.Shape(), src.Shape()
	for i := range out.Len() {
		idx := osh.IndexFrom1D(i)
		for d := range idx {
			idx[d] %= sizes[d]
		}
		setFrom1D(out, i, src, ssh.IndexTo1D(idx...))
	}
	return out
}

// Meshgrid returns two 2D tensors with the coordinates of a grid of points
// with the x values of the given 1D x tensor and the y values of the y
// tensor (as meshgrid does in NumPy): x values vary along the columns of
// the first, and y values along the rows of the second, each with a row for
// each y value and a column for each x value, for evaluating a function of
// x and y at each point of the grid with element-wise operations.
func Meshgrid(x, y Tensor) (xs, ys Values) {
	nx, ny := x.Len(), y.Len()
	xs = NewOfType(x.DataType(), ny, nx)
	ys = NewOfType(y.DataType(), ny, nx)
	for i := range ny {
		for j := range nx {
			setFrom1D(xs, i*nx+j, x, j)
			setFrom1D(ys, i*nx+j, y, i)
		}
	}
	return
}

// axisIndex returns the index of the given axis of a tensor with the given
// number of dimensions, counting from the last if negative, and panics if
// it is out of range.
func axisIndex(axis, ndims int) int {
	ax := axis
	if ax < 0 {
		ax += ndims
	}
	if ax < 0 || ax >= ndims {
		panic(fmt.Sprintf("tensor: axis %d is out of range for %d dimensions", axis, ndims))
	}
	return ax
}

// commonType returns the type of values of the given tensors: their type if
// they all have the same type, strings if any has strings, and float64
// otherwise.
func commonType(tsrs ...Tensor) reflect.Kind {
	typ := tsrs[0].DataType()
	for _, t := range tsrs {
		if t.IsString() {
			return reflect.String
		}
		if t.DataType() != typ {
			typ = reflect.Float64
		}
	}
	return typ
}

// setFrom1D sets the given 1D value of out from the given 1D value of src,
// as a string if out has strings, and as a number otherwise.
func setFrom1D(out Values, oi int, src Tensor, si int) {
	if out.IsString() {
		out.SetString1D(src.String1D(si), oi)
	} else {
		out.SetFloat1D(src.Float1D(si), oi)
	}
}
