// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import (
	"fmt"
	"math/rand"
	"testing"

	"cogentcore.org/lab/tensor"
)

// randPats returns rows x cells gaussian random patterns.
func randPats(rows, cells int) *tensor.Float64 {
	rng := rand.New(rand.NewSource(1))
	t := tensor.NewFloat64(rows, cells)
	for i := range t.Len() {
		t.SetFloat1D(rng.NormFloat64(), i)
	}
	return t
}

func BenchmarkMatrix(b *testing.B) {
	for _, rows := range []int{100, 400} {
		in := randPats(rows, 100)
		b.Run(fmt.Sprintf("rows=%d", rows), func(b *testing.B) {
			for b.Loop() {
				Matrix(L2Norm, in)
			}
		})
	}
}

func BenchmarkCrossMatrix(b *testing.B) {
	a, c := randPats(200, 100), randPats(300, 100)
	for b.Loop() {
		CrossMatrix(L2Norm, a, c)
	}
}

func BenchmarkClosestRow(b *testing.B) {
	vocab := randPats(1000, 100)
	probe := tensor.Cells1D(randPats(1, 100), 0)
	for b.Loop() {
		ClosestRow(L2Norm, probe, vocab)
	}
}
