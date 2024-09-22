// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import "cogentcore.org/core/base/errors"

// Clone returns a copy of the given tensor.
// If it is raw [Values] then a [Values.Clone] is returned.
// Otherwise if it is a view, then [Tensor.AsValues] is returned.
func Clone(tsr Tensor) Values {
	if vl, ok := tsr.(Values); ok {
		return vl.Clone()
	}
	return tsr.AsValues()
}

// MustBeValues returns the given tensor as a [Values] subtype, or nil and
// an error if it is not one. Typically outputs of compute operations must
// be values, and are reshaped to hold the results as needed.
func MustBeValues(tsr Tensor) (Values, error) {
	vl, ok := tsr.(Values)
	if !ok {
		return nil, errors.New("tensor.MustBeValues: tensor must be a Values type")
	}
	return vl, nil
}

// SetShape sets the dimension sizes from given Shape
func SetShape(vals Values, sh *Shape) {
	vals.SetShapeSizes(sh.Sizes...)
}

// SetShapeMustBeValues sets the dimension sizes from given Shape,
// calling [MustBeValues] on the destination tensor to ensure it is a [Values],
// type, returning an error if not. This is used extensively for output
// tensors in functions, and all such output tensors _must_ be Values tensors.
func SetShapeMustBeValues(tsr Tensor, sh *Shape) error {
	vals, err := MustBeValues(tsr)
	if err != nil {
		return err
	}
	vals.SetShapeSizes(sh.Sizes...)
	return nil
}

// SetShapeSizesFromTensor sets the dimension sizes as 1D int values from given tensor.
// The backing storage is resized appropriately, retaining all existing data that fits.
func SetShapeSizesFromTensor(vals Values, sizes Tensor) {
	vals.SetShapeSizes(AsIntSlice(sizes)...)
}

// SetShapeFrom sets shape of given tensor from a source tensor.
func SetShapeFrom(vals Values, from Tensor) {
	vals.SetShapeSizes(from.ShapeSizes()...)
}

// SetShapeFromMustBeValues sets shape of given tensor from a source tensor,
// calling [MustBeValues] on the destination tensor to ensure it is a [Values],
// type, returning an error if not. This is used extensively for output
// tensors in functions, and all such output tensors _must_ be Values tensors.
func SetShapeFromMustBeValues(tsr, from Tensor) error {
	vals, err := MustBeValues(tsr)
	if err != nil {
		return err
	}
	vals.SetShapeSizes(from.ShapeSizes()...)
	return nil
}

// New1DViewOf returns a 1D view into the given tensor, using the same
// underlying values, and just changing the shape to a 1D view.
// This can be useful e.g., for stats and metric functions that report
// on the 1D list of values.
func New1DViewOf(tsr Values) Values {
	vw := tsr.View()
	vw.SetShapeSizes(tsr.Len())
	return vw
}

// Cells1D returns a flat 1D [Values] view of the cells for given row index.
// This is useful for passing to other functions e.g.,
// in stats or metrics that process a 1D tensor.
func Cells1D(tsr RowMajor, row int) Values {
	return New1DViewOf(tsr.SubSpace(row))
}

// RowCellSplit splits the given tensor into a standard 2D row, cell
// shape at the given split dimension index.  All dimensions prior to
// split are collapsed into the row dimension, and from split onward
// form the cells dimension.  The resulting tensor is a re-shaped view
// of the original tensor, sharing the same underlying data.
func RowCellSplit(tsr Values, split int) Values {
	sizes := tsr.ShapeSizes()
	rows := sizes[:split]
	cells := sizes[split:]
	nr := 1
	for _, r := range rows {
		nr *= r
	}
	nc := 1
	for _, c := range cells {
		nc *= c
	}
	vw := tsr.View()
	vw.SetShapeSizes(nr, nc)
	return vw
}

// NewFloat64Scalar is a convenience method for a Tensor
// representation of a single float64 scalar value.
func NewFloat64Scalar(val float64) *Float64 {
	return NewNumberFromValues(val)
}

// NewIntScalar is a convenience method for a Tensor
// representation of a single int scalar value.
func NewIntScalar(val int) *Int {
	return NewNumberFromValues(val)
}

// NewStringScalar is a convenience method for a Tensor
// representation of a single string scalar value.
func NewStringScalar(val string) *String {
	return NewStringFromValues(val)
}

// NewFloat64FromValues returns a new 1-dimensional tensor of given value type
// initialized directly from the given slice values, which are not copied.
// The resulting Tensor thus "wraps" the given values.
func NewFloat64FromValues(vals ...float64) *Float64 {
	return NewNumberFromValues(vals...)
}

// NewIntFromValues returns a new 1-dimensional tensor of given value type
// initialized directly from the given slice values, which are not copied.
// The resulting Tensor thus "wraps" the given values.
func NewIntFromValues(vals ...int) *Int {
	return NewNumberFromValues(vals...)
}

// NewStringFromValues returns a new 1-dimensional tensor of given value type
// initialized directly from the given slice values, which are not copied.
// The resulting Tensor thus "wraps" the given values.
func NewStringFromValues(vals ...string) *String {
	n := len(vals)
	tsr := &String{}
	tsr.Values = vals
	tsr.SetShapeSizes(n)
	return tsr
}

// AsFloat64Scalar returns the first value of tensor as a float64 scalar.
// Returns 0 if no values.
func AsFloat64Scalar(tsr Tensor) float64 {
	if tsr.Len() == 0 {
		return 0
	}
	return tsr.Float1D(0)
}

// AsIntScalar returns the first value of tensor as an int scalar.
// Returns 0 if no values.
func AsIntScalar(tsr Tensor) int {
	if tsr.Len() == 0 {
		return 0
	}
	return tsr.Int1D(0)
}

// AsStringScalar returns the first value of tensor as a string scalar.
// Returns "" if no values.
func AsStringScalar(tsr Tensor) string {
	if tsr.Len() == 0 {
		return ""
	}
	return tsr.String1D(0)
}

// AsFloat64Slice returns all the tensor values as a slice of float64's.
// This allocates a new slice for the return values, and is not
// a good option for performance-critical code.
func AsFloat64Slice(tsr Tensor) []float64 {
	if tsr.Len() == 0 {
		return nil
	}
	sz := tsr.Len()
	slc := make([]float64, sz)
	for i := range sz {
		slc[i] = tsr.Float1D(i)
	}
	return slc
}

// AsIntSlice returns all the tensor values as a slice of ints.
// This allocates a new slice for the return values, and is not
// a good option for performance-critical code.
func AsIntSlice(tsr Tensor) []int {
	if tsr.Len() == 0 {
		return nil
	}
	sz := tsr.Len()
	slc := make([]int, sz)
	for i := range sz {
		slc[i] = tsr.Int1D(i)
	}
	return slc
}

// AsStringSlice returns all the tensor values as a slice of strings.
// This allocates a new slice for the return values, and is not
// a good option for performance-critical code.
func AsStringSlice(tsr Tensor) []string {
	if tsr.Len() == 0 {
		return nil
	}
	sz := tsr.Len()
	slc := make([]string, sz)
	for i := range sz {
		slc[i] = tsr.String1D(i)
	}
	return slc
}

// AsFloat64Tensor returns the tensor as a [Float64] tensor.
// If already is a Float64, it is returned as such.
// Otherwise, a new Float64 tensor is created and values are copied.
// Use this function for interfacing with gonum or other apis that
// only operate on float64 types.
func AsFloat64Tensor(tsr Tensor) *Float64 {
	if f, ok := tsr.(*Float64); ok {
		return f
	}
	f := NewFloat64(tsr.ShapeSizes()...)
	f.CopyFrom(tsr.AsValues())
	return f
}

// AsFloat32Tensor returns the tensor as a [Float32] tensor.
// If already is a Float32, it is returned as such.
// Otherwise, a new Float32 tensor is created and values are copied.
func AsFloat32Tensor(tsr Tensor) *Float32 {
	if f, ok := tsr.(*Float32); ok {
		return f
	}
	f := NewFloat32(tsr.ShapeSizes()...)
	f.CopyFrom(tsr.AsValues())
	return f
}

// AsStringTensor returns the tensor as a [String] tensor.
// If already is a String, it is returned as such.
// Otherwise, a new String tensor is created and values are copied.
// Use this function for interfacing with gonum or other apis that
// only operate on float64 types.
func AsStringTensor(tsr Tensor) *String {
	if f, ok := tsr.(*String); ok {
		return f
	}
	f := NewString(tsr.ShapeSizes()...)
	f.CopyFrom(tsr.AsValues())
	return f
}

// AsIntTensor returns the tensor as a [Int] tensor.
// If already is a Int, it is returned as such.
// Otherwise, a new Int tensor is created and values are copied.
// Use this function for interfacing with gonum or other apis that
// only operate on float64 types.
func AsIntTensor(tsr Tensor) *Int {
	if f, ok := tsr.(*Int); ok {
		return f
	}
	f := NewInt(tsr.ShapeSizes()...)
	f.CopyFrom(tsr.AsValues())
	return f
}
