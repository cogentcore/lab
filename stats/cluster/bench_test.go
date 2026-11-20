// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"fmt"
	"math/rand"
	"testing"

	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/tensor"
)

// randDistMatrix returns the L2 distance matrix of n random patterns.
func randDistMatrix(n int) tensor.Tensor {
	rng := rand.New(rand.NewSource(1))
	pats := tensor.NewFloat64(n, 20)
	for i := range pats.Len() {
		pats.SetFloat1D(rng.NormFloat64(), i)
	}
	return metric.Matrix(metric.L2Norm, pats)
}

func BenchmarkCluster(b *testing.B) {
	for _, n := range []int{50, 200, 1000} {
		dmat := randDistMatrix(n)
		for m := range MetricsN {
			b.Run(fmt.Sprintf("n=%d/%s", n, m), func(b *testing.B) {
				for b.Loop() {
					Cluster(m, dmat, nil)
				}
			})
		}
	}
}
