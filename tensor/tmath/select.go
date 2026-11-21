// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"cogentcore.org/core/base/errors"
	"cogentcore.org/lab/tensor"
)

// IsIn returns a [tensor.Bool] tensor with the shape of a, which is true
// for each value of a that is one of the given values (as isin does in
// NumPy), e.g., to select the rows of a table with some of the values of
// a column. Strings are compared as strings if either tensor has strings,
// and otherwise numbers as numbers (NaN is never in the values).
func IsIn(a, values tensor.Tensor) *tensor.Bool {
	out := tensor.NewBoolShape(a.Shape())
	n := a.Len()
	if a.IsString() || values.IsString() {
		set := map[string]bool{}
		for i := range values.Len() {
			set[values.String1D(i)] = true
		}
		for i := range n {
			out.SetBool1D(set[a.String1D(i)], i)
		}
		return out
	}
	set := map[float64]bool{}
	for i := range values.Len() {
		set[values.Float1D(i)] = true
	}
	for i := range n {
		out.SetBool1D(set[a.Float1D(i)], i)
	}
	return out
}

// Where returns a tensor with the values of a where cond is true, and of
// b where it is false (as where does in NumPy), with the shapes of the
// three aligned by broadcasting (e.g., a scalar b for all of the false
// values). The result has strings if a or b has strings, and otherwise
// float64 values.
func Where(cond, a, b tensor.Tensor) tensor.Tensor {
	_, _, abs, err := tensor.AlignShapes(a, b)
	if errors.Log(err) != nil {
		return nil
	}
	ab := tensor.NewFloat64(abs.Sizes...) // for the shape of a and b
	cs, _, os, err := tensor.AlignShapes(cond, ab)
	if errors.Log(err) != nil {
		return nil
	}
	as, bs, _, _ := tensor.AlignShapes(a, tensor.NewFloat64(os.Sizes...))
	bs, _, _, _ = tensor.AlignShapes(b, tensor.NewFloat64(os.Sizes...))
	str := a.IsString() || b.IsString()
	var out tensor.Values
	if str {
		out = tensor.NewString(os.Sizes...)
	} else {
		out = tensor.NewFloat64(os.Sizes...)
	}
	for i := range os.Len() {
		idx := os.IndexFrom1D(i)
		src, si := b, tensor.WrapIndex1D(bs, idx...)
		if cond.Float1D(tensor.WrapIndex1D(cs, idx...)) != 0 {
			src, si = a, tensor.WrapIndex1D(as, idx...)
		}
		if str {
			out.SetString1D(src.String1D(si), i)
		} else {
			out.SetFloat1D(src.Float1D(si), i)
		}
	}
	return out
}
