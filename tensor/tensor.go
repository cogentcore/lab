// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

//go:generate core generate

import (
	"fmt"
	"reflect"

	"cogentcore.org/core/base/metadata"
)

// DataTypes are the primary tensor data types with specific support.
// Any numerical type can also be used.  bool is represented using an
// efficient bit slice.
type DataTypes interface {
	string | bool | float32 | float64 | int | int32 | byte
}

// MaxSprintLength is the default maximum length of a String() representation
// of a tensor, as generated by the Sprint function. Defaults to 1000.
var MaxSprintLength = 1000

// todo: add a conversion function to copy data from Column-Major to a tensor:
// It is also possible to use Column-Major order, which is used in R, Julia, and MATLAB
// where the inner-most index is first and outermost last.

// Tensor is the most general interface for n-dimensional tensors.
// Per C / Go / Python conventions, indexes are Row-Major, ordered from
// outer to inner left-to-right, so the inner-most is right-most.
// It is implemented for raw [Values] with direct integer indexing
// by the [Number], [String], and [Bool] types, covering the different
// concrete types specified by [DataTypes] (see [Values] for
// additional interface methods for raw value types).
// For float32 and float64 values, use NaN to indicate missing values,
// as all of the data analysis and plot packages skip NaNs.
// View Tensor types provide different ways of viewing a source tensor,
// including [Sliced] for arbitrary slices of dimension indexes,
// [Masked] for boolean masked access and setting of individual indexes,
// and [Indexed] for arbitrary indexes of values, organized into the
// shape of the indexes, not the original source data.
// The [Rows] view provides an optimized row-indexed view for [table.Table] data.
type Tensor interface {
	fmt.Stringer

	// Label satisfies the core.Labeler interface for a summary
	// description of the tensor, including metadata Name if set.
	Label() string

	// Metadata returns the metadata for this tensor, which can be used
	// to encode name, docs, shape dimension names, plotting options, etc.
	Metadata() *metadata.Data

	// Shape() returns a [Shape] representation of the tensor shape
	// (dimension sizes). For tensors that present a view onto another
	// tensor, this typically must be constructed.
	// In general, it is better to use the specific [Tensor.ShapeSizes],
	// [Tensor.ShapeInts], [Tensor.DimSize] etc as neeed.
	Shape() *Shape

	// ShapeSizes returns the sizes of each dimension as an int tensor.
	ShapeSizes() *Int

	// ShapeInts returns the sizes of each dimension as a slice of ints.
	// This is the preferred access for Go code.
	ShapeInts() []int

	// Len returns the total number of elements in the tensor,
	// i.e., the product of all shape dimensions.
	// Len must always be such that the 1D() accessors return
	// values using indexes from 0..Len()-1.
	Len() int

	// NumDims returns the total number of dimensions.
	NumDims() int

	// DimSize returns size of given dimension.
	DimSize(dim int) int

	// DataType returns the type of the data elements in the tensor.
	// Bool is returned for the Bool tensor type.
	DataType() reflect.Kind

	// IsString returns true if the data type is a String; otherwise it is numeric.
	IsString() bool

	// AsValues returns this tensor as raw [Values]. If it already is,
	// it is returned directly. If it is a View tensor, the view is
	// "rendered" into a fully contiguous and optimized [Values] representation
	// of that view, which will be faster to access for further processing,
	// and enables all the additional functionality provided by the [Values] interface.
	AsValues() Values

	/////////////////////  Floats

	// Float returns the value of given n-dimensional index (matching Shape) as a float64.
	Float(i ...int) float64

	// SetFloat sets the value of given n-dimensional index (matching Shape) as a float64.
	SetFloat(val float64, i ...int)

	// Float1D returns the value of given 1-dimensional index (0-Len()-1) as a float64.
	// This can be somewhat expensive in wrapper views ([Rows], [Sliced]), which
	// convert the flat index back into a full n-dimensional index and use that api.
	// [Tensor.FloatRowCell] is preferred.
	Float1D(i int) float64

	// SetFloat1D sets the value of given 1-dimensional index (0-Len()-1) as a float64.
	// This can be somewhat expensive in the commonly-used [Rows] view;
	// [Tensor.SetFloatRowCell] is preferred.
	SetFloat1D(val float64, i int)

	/////////////////////  Strings

	// StringValue returns the value of given n-dimensional index (matching Shape) as a string.
	// 'String' conflicts with [fmt.Stringer], so we have to use StringValue here.
	StringValue(i ...int) string

	// SetString sets the value of given n-dimensional index (matching Shape) as a string.
	SetString(val string, i ...int)

	// String1D returns the value of given 1-dimensional index (0-Len()-1) as a string.
	String1D(i int) string

	// SetString1D sets the value of given 1-dimensional index (0-Len()-1) as a string.
	SetString1D(val string, i int)

	/////////////////////  Ints

	// Int returns the value of given n-dimensional index (matching Shape) as a int.
	Int(i ...int) int

	// SetInt sets the value of given n-dimensional index (matching Shape) as a int.
	SetInt(val int, i ...int)

	// Int1D returns the value of given 1-dimensional index (0-Len()-1) as a int.
	Int1D(i int) int

	// SetInt1D sets the value of given 1-dimensional index (0-Len()-1) as a int.
	SetInt1D(val int, i int)
}
