// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"fmt"
	"math"
	"reflect"

	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/lab/tensor"
)

// The AssignVar functions do the assignment operators of Goal math mode
// on a variable that can be either a [tensor.Tensor] or a plain Go number
// (or string for +=), given a pointer to it, as Goal only knows the syntax
// of the code, not the types of its variables. The values of a tensor
// variable are updated in place, as by [AddAssign] etc, and a Go variable
// is set to the result of the operator on its value and b, which is
// converted to its type (see [tensor.ToFloat64]).

// AddAssignVar does += on the variable that ptr points to.
func AddAssignVar(ptr any, b any) error {
	if p, ok := ptr.(*string); ok {
		*p += tensor.ToString(b)
		return nil
	}
	return assignVar(ptr, b, AddAssign, func(a, b float64) float64 { return a + b })
}

// SubAssignVar does -= on the variable that ptr points to.
func SubAssignVar(ptr any, b any) error {
	return assignVar(ptr, b, SubAssign, func(a, b float64) float64 { return a - b })
}

// MulAssignVar does *= on the variable that ptr points to.
func MulAssignVar(ptr any, b any) error {
	return assignVar(ptr, b, MulAssign, func(a, b float64) float64 { return a * b })
}

// DivAssignVar does /= on the variable that ptr points to.
// An int variable is divided as ints, as in Go.
func DivAssignVar(ptr any, b any) error {
	if p, ok := ptr.(*int); ok {
		*p /= tensor.ToInt(b)
		return nil
	}
	return assignVar(ptr, b, DivAssign, func(a, b float64) float64 { return a / b })
}

// ModAssignVar does %= on the variable that ptr points to.
func ModAssignVar(ptr any, b any) error {
	if p, ok := ptr.(*int); ok {
		*p %= tensor.ToInt(b)
		return nil
	}
	return assignVar(ptr, b, ModAssign, func(a, b float64) float64 { return math.Mod(a, b) })
}

// assignVar does an assignment operator on the variable that ptr points
// to: with the tensor function for a tensor variable, and with the float64
// function for a Go number variable.
func assignVar(ptr any, b any, tfun func(a, b tensor.Tensor) error, ffun func(a, b float64) float64) error {
	switch p := ptr.(type) {
	case *tensor.Tensor:
		return tfun(*p, tensor.ToTensor(b))
	case *tensor.Values:
		return tfun(*p, tensor.ToTensor(b))
	case *float64:
		*p = ffun(*p, tensor.ToFloat64(b))
		return nil
	case *int:
		*p = int(ffun(float64(*p), tensor.ToFloat64(b)))
		return nil
	}
	// Other types are a tensor or a number, by reflection.
	pv := reflect.ValueOf(ptr)
	if pv.Kind() != reflect.Pointer || pv.IsNil() {
		return fmt.Errorf("tmath: %T is not a pointer to a variable", ptr)
	}
	dst := pv.Elem()
	if t, ok := dst.Interface().(tensor.Tensor); ok {
		return tfun(t, tensor.ToTensor(b))
	}
	if !reflectx.KindIsNumber(dst.Kind()) {
		return fmt.Errorf("tmath: cannot do an assignment operator on a variable of type %s", dst.Type())
	}
	a, _ := reflectx.ToFloat(dst.Interface())
	dst.Set(reflect.ValueOf(ffun(a, tensor.ToFloat64(b))).Convert(dst.Type()))
	return nil
}
