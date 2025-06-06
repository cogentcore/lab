// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"fmt"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestOps(t *testing.T) {
	scalar := tensor.NewFloat64Scalar(-5.5)
	scb := scalar.Clone()
	scb.SetFloat1D(-4.0, 0)
	scout := scalar.Clone()

	vals := []float64{-1.507556722888818, -1.2060453783110545, -0.9045340337332908, -0.6030226891555273, -0.3015113445777635, 0.1, 0.3015113445777635, 0.603022689155527, 0.904534033733291, 1.2060453783110545, 1.507556722888818, .3}

	oned := tensor.NewNumberFromValues(vals...)
	oneout := oned.Clone()

	cell2d := tensor.NewFloat32(5, 12)
	_, cells := cell2d.Shape().RowCellSize()
	assert.Equal(t, cells, 12)
	tensor.VectorizeThreaded(1, tensor.NFirstLen, func(idx int, tsr ...tensor.Tensor) {
		ci := idx % cells
		cell2d.SetFloat1D(oned.Float1D(ci), idx)
	}, cell2d)
	// cell2d.DeleteRows(3, 1)
	cellout := cell2d.Clone()
	_ = cellout

	AddOut(scalar, scb, scout)
	assert.Equal(t, -5.5+-4, scout.Float1D(0))

	AddOut(scalar, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, v+-5.5, oneout.Float1D(i))
	}

	AddOut(oned, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, v+v, oneout.Float1D(i))
	}

	AddOut(cell2d, oned, cellout)
	for ri := range 5 {
		for i, v := range vals {
			assert.InDelta(t, v+v, cellout.FloatRow(ri, i), 1.0e-6)
		}
	}

	SubOut(scalar, scb, scout)
	assert.Equal(t, -5.5 - -4, scout.Float1D(0))

	SubOut(scb, scalar, scout)
	assert.Equal(t, -4 - -5.5, scout.Float1D(0))

	SubOut(scalar, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, -5.5-v, oneout.Float1D(i))
	}

	SubOut(oned, scalar, oneout)
	for i, v := range vals {
		assert.Equal(t, v - -5.5, oneout.Float1D(i))
	}

	SubOut(oned, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, v-v, oneout.Float1D(i))
	}

	SubOut(cell2d, oned, cellout)
	for ri := range 5 {
		for i, v := range vals {
			assert.InDelta(t, v-v, cellout.FloatRow(ri, i), 1.0e-6)
		}
	}

	MulOut(scalar, scb, scout)
	assert.Equal(t, -5.5*-4, scout.Float1D(0))

	MulOut(scalar, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, v*-5.5, oneout.Float1D(i))
	}

	MulOut(oned, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, v*v, oneout.Float1D(i))
	}

	MulOut(cell2d, oned, cellout)
	for ri := range 5 {
		for i, v := range vals {
			assert.InDelta(t, v*v, cellout.FloatRow(ri, i), 1.0e-6)
		}
	}

	DivOut(scalar, scb, scout)
	assert.Equal(t, -5.5/-4, scout.Float1D(0))

	DivOut(scb, scalar, scout)
	assert.Equal(t, -4/-5.5, scout.Float1D(0))

	DivOut(scalar, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, -5.5/v, oneout.Float1D(i))
	}

	DivOut(oned, scalar, oneout)
	for i, v := range vals {
		assert.Equal(t, v/-5.5, oneout.Float1D(i))
	}

	DivOut(oned, oned, oneout)
	for i, v := range vals {
		assert.Equal(t, v/v, oneout.Float1D(i))
	}

	DivOut(cell2d, oned, cellout)
	for ri := range 5 {
		for i, v := range vals {
			assert.InDelta(t, v/v, cellout.FloatRow(ri, i), 1.0e-6)
		}
	}

	onedc := tensor.Clone(oned)
	AddAssign(onedc, scalar)
	for i, v := range vals {
		assert.Equal(t, v+-5.5, onedc.Float1D(i))
	}

	SubAssign(onedc, scalar)
	for i, v := range vals {
		assert.InDelta(t, v, onedc.Float1D(i), 1.0e-8)
	}

	MulAssign(onedc, scalar)
	for i, v := range vals {
		assert.InDelta(t, v*-5.5, onedc.Float1D(i), 1.0e-7)
	}

	DivAssign(onedc, scalar)
	for i, v := range vals {
		assert.InDelta(t, v, onedc.Float1D(i), 1.0e-7)
	}

	Inc(onedc)
	for i, v := range vals {
		assert.InDelta(t, v+1, onedc.Float1D(i), 1.0e-7)
	}

	Dec(onedc)
	for i, v := range vals {
		assert.InDelta(t, v, onedc.Float1D(i), 1.0e-7)
	}
}

func runBenchMult(b *testing.B, n int, thread bool) {
	if thread {
		tensor.ThreadingThreshold = 1
	} else {
		tensor.ThreadingThreshold = 100_000_000
	}
	av := tensor.NewFloat64(n)
	bv := tensor.NewFloat64(n)
	ov := tensor.NewFloat64(n)
	for i := range n {
		av.SetFloat1D(1.0/float64(n), i)
		bv.SetFloat1D(1.0/float64(n), i)
	}
	b.ResetTimer()
	for range b.N {
		MulOut(av, bv, ov)
	}
}

// to run this benchmark, do:
// go test -bench BenchmarkMult -count 10 >bench.txt
// go install golang.org/x/perf/cmd/benchstat@latest
// benchstat -row /n -col .name bench.txt

// goos: darwin
// goarch: arm64
// pkg: cogentcore.org/lab/tensor/tmath
//         │ MultThreaded │              MultSingle              │
//         │    sec/op    │   sec/op     vs base                 │
// 10        3656.5n ± 0%   699.8n ± 0%   -80.86% (p=0.000 n=10)
// 100        7.288µ ± 1%   4.510µ ± 0%   -38.12% (p=0.000 n=10)
// 200        9.813µ ± 1%   8.761µ ± 0%   -10.72% (p=0.000 n=10)
// 300        12.06µ ± 2%   13.04µ ± 0%    +8.12% (p=0.000 n=10)
// 400        14.53µ ± 3%   17.13µ ± 1%   +17.88% (p=0.000 n=10)
// 500        16.65µ ± 2%   21.35µ ± 1%   +28.19% (p=0.000 n=10)
// 600        18.74µ ± 2%   25.68µ ± 0%   +37.00% (p=0.000 n=10)
// 700        20.83µ ± 3%   29.94µ ± 0%   +43.74% (p=0.000 n=10)
// 800        22.33µ ± 1%   34.11µ ± 0%   +52.72% (p=0.000 n=10)
// 900        24.13µ ± 2%   38.23µ ± 0%   +58.44% (p=0.000 n=10)
// 1000       26.25µ ± 1%   42.41µ ± 0%   +61.55% (p=0.000 n=10)
// 10000      127.2µ ± 1%   424.1µ ± 1%  +233.37% (p=0.000 n=10)
// geomean    16.88µ        19.11µ        +13.21%

var ns = []int{10, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000, 10_000}

func BenchmarkMultThreaded(b *testing.B) {
	for _, n := range ns {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			runBenchMult(b, n, true)
		})
	}
}

func BenchmarkMultSingle(b *testing.B) {
	for _, n := range ns {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			runBenchMult(b, n, false)
		})
	}
}
