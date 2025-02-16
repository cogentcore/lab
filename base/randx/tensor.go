// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package randx

import (
	"cogentcore.org/lab/tensor"
)

// Tensor makes a new random tensor of the given size using the given random number parameters.
func Tensor(rp *RandParams, sizes ...int) *tensor.Float64 {
	tsr := tensor.NewFloat64(sizes...)
	tensor.FloatSetFunc(20, func(idx int) float64 { // TODO: is the right number of flops?
		return rp.Gen()
	}, tsr)
	return tsr
}
