// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package convolve

import (
	"math/rand"
	"testing"
)

func BenchmarkSlice64(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	src := make([]float64, 10000)
	for i := range src {
		src[i] = rng.NormFloat64()
	}
	kern := GaussianKernel64(10, 3)
	var dest []float64
	for b.Loop() {
		Slice64(&dest, src, kern)
	}
}
