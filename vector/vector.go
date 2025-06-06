// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package vector provides standard vector math functions that
// always operate on 1D views of tensor inputs regardless of the original
// vector shape.
package vector

import (
	"math"

	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensor/tmath"
)

// Mul multiplies two vectors element-wise, using a 1D vector
// view of the two vectors, returning the output 1D vector.
func Mul(a, b tensor.Tensor) tensor.Values {
	return tensor.CallOut2Float64(MulOut, a, b)
}

// MulOut multiplies two vectors element-wise, using a 1D vector
// view of the two vectors, filling in values in the output 1D vector.
func MulOut(a, b tensor.Tensor, out tensor.Values) error {
	return tmath.MulOut(tensor.As1D(a), tensor.As1D(b), out)
}

// Sum returns the sum of all values in the tensor, as a scalar.
func Sum(a tensor.Tensor) tensor.Values {
	n := a.Len()
	sum := 0.0
	tensor.Vectorize(func(tsr ...tensor.Tensor) int { return n },
		func(idx int, tsr ...tensor.Tensor) {
			sum += tsr[0].Float1D(idx)
		}, a)
	return tensor.NewFloat64Scalar(sum)
}

// Dot performs the vector dot product: the [Sum] of the [Mul] product
// of the two tensors, returning a scalar value. Also known as the inner product.
func Dot(a, b tensor.Tensor) tensor.Values {
	return Sum(Mul(a, b))
}

// L2Norm returns the length of the vector as the L2 Norm:
// square root of the sum of squared values of the vector, as a scalar.
// This is the Sqrt of the [Dot] product of the vector with itself.
func L2Norm(a tensor.Tensor) tensor.Values {
	dot := Dot(a, a).Float1D(0)
	return tensor.NewFloat64Scalar(math.Sqrt(dot))
}

// L1Norm returns the length of the vector as the L1 Norm:
// sum of the absolute values of the tensor, as a scalar.
func L1Norm(a tensor.Tensor) tensor.Values {
	n := a.Len()
	sum := 0.0
	tensor.Vectorize(func(tsr ...tensor.Tensor) int { return n },
		func(idx int, tsr ...tensor.Tensor) {
			sum += math.Abs(tsr[0].Float1D(idx))
		}, a)
	return tensor.NewFloat64Scalar(sum)
}
