// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"math/rand"
	"testing"

	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// randTensor returns a tensor of the given shape with gaussian random values.
func randTensor(sizes ...int) *tensor.Float64 {
	rng := rand.New(rand.NewSource(1))
	t := tensor.NewFloat64(sizes...)
	for i := range t.Len() {
		t.SetFloat1D(rng.NormFloat64(), i)
	}
	return t
}

func BenchmarkQuantiles(b *testing.B) {
	qs := tensor.NewFloat64FromValues(0.25, 0.5, 0.75)
	for _, n := range []int{1000, 100000} {
		in := randTensor(n)
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			for b.Loop() {
				Quantiles(in, qs)
			}
		})
	}
}

// BenchmarkCells computes stats on rows x cells tensors, which compute
// a stat for each cell across the rows.
func BenchmarkCells(b *testing.B) {
	in := randTensor(1000, 100)
	for _, st := range []Stats{StatMean, StatStd, StatMedian} {
		b.Run(st.String(), func(b *testing.B) {
			for b.Loop() {
				st.Call(in)
			}
		})
	}
}

func BenchmarkNorm(b *testing.B) {
	in := randTensor(10000)
	b.Run("ZScore", func(b *testing.B) {
		for b.Loop() {
			ZScore(in)
		}
	})
	b.Run("UnitNorm", func(b *testing.B) {
		for b.Loop() {
			UnitNorm(in)
		}
	})
}

// groupTable returns a table with n rows, a string group column with ng
// unique values, and a float value column.
func groupTable(n, ng int) *table.Table {
	dt := table.New()
	gp := dt.AddStringColumn("Group")
	val := dt.AddFloat64Column("Value")
	dt.SetNumRows(n)
	rng := rand.New(rand.NewSource(1))
	for i := range n {
		gp.SetString1D(fmt.Sprintf("g%d", rng.Intn(ng)), i)
		val.SetFloat1D(rng.NormFloat64(), i)
	}
	return dt
}

func BenchmarkGroups(b *testing.B) {
	for _, ng := range []int{10, 1000} {
		dt := groupTable(10000, ng)
		b.Run(fmt.Sprintf("groups=%d", ng), func(b *testing.B) {
			for b.Loop() {
				dir, _ := tensorfs.NewDir("Bench")
				TableGroups(dir, dt, "Group")
			}
		})
		b.Run(fmt.Sprintf("stats-groups=%d", ng), func(b *testing.B) {
			dir, _ := tensorfs.NewDir("Bench")
			TableGroups(dir, dt, "Group")
			for b.Loop() {
				TableGroupStats(dir, StatMean, dt, "Value")
			}
		})
	}
}
