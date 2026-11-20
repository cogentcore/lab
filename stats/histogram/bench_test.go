// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package histogram

import (
	"math/rand"
	"testing"
)

func BenchmarkF64(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	vals := make([]float64, 100000)
	for i := range vals {
		vals[i] = rng.NormFloat64()
	}
	var hist []float64
	for b.Loop() {
		F64(&hist, vals, 50, -4, 4)
	}
}
