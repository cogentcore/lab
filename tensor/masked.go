// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"math"
	"reflect"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/base/reflectx"
)

// Masked is a wrapper around another [Tensor] that provides a
// bit-masked view onto the Tensor defined by a [Bool] [Values]
// tensor with a matching shape. If the bool mask has a 'false'
// then the corresponding value cannot be set and Float access returns
// NaN indicating missing data.
// To produce a new [Values] tensor with only the 'true' cases,
// (i.e., the copy function of numpy), call [Masked.AsValues].
type Masked struct { //types:add

	// Tensor that we are a masked view onto.
	Tensor Tensor

	// Bool tensor with same shape as source tensor, providing mask.
	Mask *Bool
}

// NewMasked returns a new [Masked] view of given tensor,
// with given [Bool] mask values.
func NewMasked(tsr Tensor, mask ...*Bool) *Masked {
	ms := &Masked{Tensor: tsr}
	if len(mask) == 1 {
		ms.Mask = mask[0]
		ms.SyncShape()
	} else {
		ms.Mask = NewBoolShape(tsr.Shape())
		ms.Mask.SetTrue()
	}
	return ms
}

// AsMasked returns the tensor as a [Masked] view.
// If it already is one, then it is returned, otherwise it is wrapped
// with an initially transparent mask.
func AsMasked(tsr Tensor) *Masked {
	if ms, ok := tsr.(*Masked); ok {
		return ms
	}
	return NewMasked(tsr)
}

// SetTensor sets as indexes into given tensor with sequential initial indexes.
func (ms *Masked) SetTensor(tsr Tensor) {
	ms.Tensor = tsr
	ms.SyncShape()
}

// SyncShape ensures that [Masked.Mask] shape is the same as source tensor.
func (ms *Masked) SyncShape() {
	if ms.Mask == nil {
		ms.Mask = NewBoolShape(ms.Tensor.Shape())
		return
	}
	SetShapeFrom(ms.Mask, ms.Tensor)
}

// Label satisfies the core.Labeler interface for a summary description of the tensor.
func (ms *Masked) Label() string {
	return label(ms.Metadata().Name(), ms.Shape())
}

// String satisfies the fmt.Stringer interface for string of tensor data.
func (ms *Masked) String() string { return sprint(ms, 0) }

// Metadata returns the metadata for this tensor, which can be used
// to encode plotting options, etc.
func (ms *Masked) Metadata() *metadata.Data { return ms.Tensor.Metadata() }

func (ms *Masked) IsString() bool { return ms.Tensor.IsString() }

func (ms *Masked) DataType() reflect.Kind { return ms.Tensor.DataType() }

func (ms *Masked) ShapeSizes() []int { return ms.Tensor.ShapeSizes() }

func (ms *Masked) Shape() *Shape { return ms.Tensor.Shape() }

// Len returns the total number of elements in our view of the tensor.
func (ms *Masked) Len() int { return ms.Tensor.Len() }

// NumDims returns the total number of dimensions.
func (ms *Masked) NumDims() int { return ms.Tensor.NumDims() }

// DimSize returns the effective view size of given dimension.
func (ms *Masked) DimSize(dim int) int { return ms.Tensor.DimSize(dim) }

// AsValues returns a copy of this tensor as raw [Values].
// This "renders" the Masked view into a fully contiguous
// and optimized memory representation of that view.
// Because the masking pattern is unpredictable, only a 1D shape is possible.
func (ms *Masked) AsValues() Values {
	dt := ms.Tensor.DataType()
	n := ms.Len()
	switch {
	case ms.Tensor.IsString():
		vals := make([]string, 0, n)
		for i := range n {
			if !ms.Mask.Bool1D(i) {
				continue
			}
			vals = append(vals, ms.Tensor.String1D(i))
		}
		return NewStringFromSlice(vals...)
	case reflectx.KindIsFloat(dt):
		vals := make([]float64, 0, n)
		for i := range n {
			if !ms.Mask.Bool1D(i) {
				continue
			}
			vals = append(vals, ms.Tensor.Float1D(i))
		}
		return NewFloat64FromSlice(vals...)
	default:
		vals := make([]int, 0, n)
		for i := range n {
			if !ms.Mask.Bool1D(i) {
				continue
			}
			vals = append(vals, ms.Tensor.Int1D(i))
		}
		return NewIntFromSlice(vals...)
	}
}

///////////////////////////////////////////////
// Masked access

/////////////////////  Floats

func (ms *Masked) Float(i ...int) float64 {
	if !ms.Mask.Bool(i...) {
		return math.NaN()
	}
	return ms.Tensor.Float(i...)
}

func (ms *Masked) SetFloat(val float64, i ...int) {
	if !ms.Mask.Bool(i...) {
		return
	}
	ms.Tensor.SetFloat(val, i...)
}

func (ms *Masked) Float1D(i int) float64 {
	if !ms.Mask.Bool1D(i) {
		return math.NaN()
	}
	return ms.Tensor.Float1D(i)
}

func (ms *Masked) SetFloat1D(val float64, i int) {
	if !ms.Mask.Bool1D(i) {
		return
	}
	ms.Tensor.SetFloat1D(val, i)
}

/////////////////////  Strings

func (ms *Masked) StringValue(i ...int) string {
	if !ms.Mask.Bool(i...) {
		return ""
	}
	return ms.Tensor.StringValue(i...)
}

func (ms *Masked) SetString(val string, i ...int) {
	if !ms.Mask.Bool(i...) {
		return
	}
	ms.Tensor.SetString(val, i...)
}

func (ms *Masked) String1D(i int) string {
	if !ms.Mask.Bool1D(i) {
		return ""
	}
	return ms.Tensor.String1D(i)
}

func (ms *Masked) SetString1D(val string, i int) {
	if !ms.Mask.Bool1D(i) {
		return
	}
	ms.Tensor.SetString1D(val, i)
}

/////////////////////  Ints

func (ms *Masked) Int(i ...int) int {
	if !ms.Mask.Bool(i...) {
		return 0
	}
	return ms.Tensor.Int(i...)
}

func (ms *Masked) SetInt(val int, i ...int) {
	if !ms.Mask.Bool(i...) {
		return
	}
	ms.Tensor.SetInt(val, i...)
}

func (ms *Masked) Int1D(i int) int {
	if !ms.Mask.Bool1D(i) {
		return 0
	}
	return ms.Tensor.Int1D(i)
}

// SetInt1D is somewhat expensive if indexes are set, because it needs to convert
// the flat index back into a full n-dimensional index and then use that api.
func (ms *Masked) SetInt1D(val int, i int) {
	if !ms.Mask.Bool1D(i) {
		return
	}
	ms.Tensor.SetInt1D(val, i)
}

// Filter sets the mask values using given Filter function.
// The filter function gets the 1D index into the source tensor.
func (ms *Masked) Filter(filterer func(tsr Tensor, idx int) bool) {
	n := ms.Tensor.Len()
	for i := range n {
		ms.Mask.SetBool1D(filterer(ms.Tensor, i), i)
	}
}

// check for interface impl
var _ Tensor = (*Masked)(nil)
