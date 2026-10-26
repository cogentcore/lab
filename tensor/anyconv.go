// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"fmt"
	"iter"
	"reflect"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/reflectx"
)

// The To functions convert a value that is either a [Tensor] or a plain Go
// value (a number, bool, string, or a slice of them) to a given type. Goal
// math mode uses them wherever an expression can be either, as it only
// knows the syntax of the code, not the types of its variables. Their cost
// is a type switch for a value that already has the type.

// ToTensor returns the given value as a [Tensor]: a Tensor as itself,
// a number, bool or string as a scalar tensor, and a slice of them as a
// 1D tensor that wraps the slice values without copying them.
func ToTensor(v any) Tensor {
	switch x := v.(type) {
	case Tensor:
		return x
	case float64:
		return NewFloat64Scalar(x)
	case float32:
		return NewFloat32Scalar(x)
	case int:
		return NewIntScalar(x)
	case bool:
		return NewBoolFromValues(x)
	case string:
		return NewStringScalar(x)
	case []float64:
		return NewFloat64FromValues(x...)
	case []float32:
		return NewFloat32FromValues(x...)
	case []int:
		return NewIntFromValues(x...)
	case []bool:
		return NewBoolFromValues(x...)
	case []string:
		return NewStringFromValues(x...)
	case nil:
		panic("tensor.ToTensor: nil value")
	}
	if f, err := reflectx.ToFloat(v); err == nil {
		return NewFloat64Scalar(f)
	}
	panic(fmt.Sprintf("tensor.ToTensor: cannot convert value of type %T to a tensor", v))
}

// ToFloat64 returns the given value as a float64: the first value of a
// [Tensor], and any number or bool converted to float64.
func ToFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case Tensor:
		return AsFloat64Scalar(x)
	}
	f, err := reflectx.ToFloat(v)
	errors.Log(err)
	return f
}

// ToInt returns the given value as an int: the first value of a [Tensor],
// and any number or bool converted to int.
func ToInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case float64:
		return int(x)
	case Tensor:
		return AsIntScalar(x)
	}
	i, err := reflectx.ToInt(v)
	errors.Log(err)
	return int(i)
}

// ToBool returns the given value as a bool: the first value of a [Tensor]
// (true if non-zero for a numerical tensor), and any bool, or number
// (true if non-zero).
func ToBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case *Bool:
		return x.Len() > 0 && x.Value1D(0)
	case Tensor:
		return x.Len() > 0 && x.Float1D(0) != 0
	}
	b, err := reflectx.ToBool(v)
	errors.Log(err)
	return b
}

// ToString returns the given value as a string: the first value of a
// [Tensor], and any other value formatted as a string.
func ToString(v any) string {
	if x, ok := v.(Tensor); ok {
		return AsStringScalar(x)
	}
	return reflectx.ToString(v)
}

// ToInts returns the given value as a slice of ints, as used for the
// sizes of a shape: the values of a [Tensor], a slice of ints, and any
// number as a slice of one int.
func ToInts(v any) []int {
	switch x := v.(type) {
	case []int:
		return x
	case int:
		return []int{x}
	case Tensor:
		return AsIntSlice(x)
	}
	return []int{ToInt(v)}
}

// Iter returns an iterator over the given value, for the range loops of
// Goal math mode: over 0..n-1, with values of float64(i), for an int n,
// and over the indexes and values of a [Tensor] or a slice of numbers, with
// the values as float64.
func Iter(v any) iter.Seq2[int, float64] {
	switch x := v.(type) {
	case int:
		return func(yield func(int, float64) bool) {
			for i := range x {
				if !yield(i, float64(i)) {
					return
				}
			}
		}
	case []float64:
		return func(yield func(int, float64) bool) {
			for i, f := range x {
				if !yield(i, f) {
					return
				}
			}
		}
	}
	tsr := ToTensor(v)
	return func(yield func(int, float64) bool) {
		n := tsr.Len()
		for i := range n {
			if !yield(i, tsr.Float1D(i)) {
				return
			}
		}
	}
}

// SetVar sets the variable that the given pointer points to to the given
// value, converted to its type, for the assignments of Goal math mode,
// whose variables can be either tensors or plain Go values:
// a [Tensor] variable is set to the value as a Tensor (see [ToTensor]),
// and a number, bool or string variable to the value converted to its type.
func SetVar(ptr any, v any) {
	switch p := ptr.(type) {
	case *Tensor:
		*p = ToTensor(v)
		return
	case *Values:
		*p = ToTensor(v).AsValues()
		return
	case *float64:
		*p = ToFloat64(v)
		return
	case *int:
		*p = ToInt(v)
		return
	case *bool:
		*p = ToBool(v)
		return
	case *string:
		*p = ToString(v)
		return
	}
	// Other types are set by reflection, from a value that is assignable,
	// or by conversion for a number.
	pv := reflect.ValueOf(ptr)
	if pv.Kind() != reflect.Pointer || pv.IsNil() {
		panic(fmt.Sprintf("tensor.SetVar: %T is not a pointer to a variable", ptr))
	}
	dst := pv.Elem()
	val := reflect.ValueOf(v)
	switch {
	case val.IsValid() && val.Type().AssignableTo(dst.Type()):
		dst.Set(val)
	case reflectx.KindIsNumber(dst.Kind()):
		dst.Set(reflect.ValueOf(ToFloat64(v)).Convert(dst.Type()))
	default:
		panic(fmt.Sprintf("tensor.SetVar: cannot set a variable of type %s to a value of type %T", dst.Type(), v))
	}
}
