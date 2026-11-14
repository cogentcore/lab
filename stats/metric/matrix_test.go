// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"cogentcore.org/core/base/tolassert"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestCovarIris(t *testing.T) {
	// note: these results are verified against this example:
	// https://plot.ly/ipython-notebooks/principal-component-analysis/

	dt := table.New()
	dt.AddFloat64Column("data", 4)
	dt.AddStringColumn("class")
	err := dt.OpenCSV("testdata/iris.data", tensor.Comma)
	if err != nil {
		t.Error(err)
	}
	data := dt.Column("data")
	covar := tensor.NewFloat64()
	err = CovarianceMatrixOut(Correlation, data, covar)
	assert.NoError(t, err)
	// fmt.Printf("covar: %s\n", covar.String())
	// tensor.SaveCSV(covar, "testdata/iris-covar.tsv", tensor.Tab)

	cv := []float64{1, -0.10936924995064935, 0.8717541573048719, 0.8179536333691635,
		-0.10936924995064935, 1, -0.4205160964011548, -0.3565440896138057,
		0.8717541573048719, -0.4205160964011548, 1, 0.9627570970509667,
		0.8179536333691635, -0.3565440896138057, 0.9627570970509667, 1}

	tolassert.EqualTolSlice(t, cv, covar.Values, 1.0e-8)
}

func runBenchCovar(b *testing.B, n int, thread bool) {
	if thread {
		tensor.ThreadingThreshold = 1
	} else {
		tensor.ThreadingThreshold = 100_000_000
	}
	nrows := 10
	av := tensor.AsFloat64(tensor.Reshape(tensor.NewIntRange(1, nrows*n*n+1), nrows, n, n))
	ov := tensor.NewFloat64(nrows, n, n)
	b.ResetTimer()
	for range b.N {
		CovarianceMatrixOut(Correlation, av, ov)
	}
}

// to run this benchmark, do:
// go test -bench BenchmarkCovar -count 10 >bench.txt
// go install golang.org/x/perf/cmd/benchstat@latest
// benchstat -row /n -col .name bench.txt

var ns = []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 40}

func BenchmarkCovarThreaded(b *testing.B) {
	for _, n := range ns {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			runBenchCovar(b, n, true)
		})
	}
}

func BenchmarkCovarSingle(b *testing.B) {
	for _, n := range ns {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			runBenchCovar(b, n, false)
		})
	}
}

// covarMatrixRef computes the covariance matrix by calling fun on each
// pair of explicitly extracted cell column vectors, as a reference.
// Like [CovarianceMatrixOut], it computes the lower triangle fun(i, j), i >= j,
// and mirrors it, because [Covariance] is asymmetric with NaN values
// (it normalizes by the non-NaN count of its first argument).
func covarMatrixRef(fun MetricFunc, in *tensor.Float64) []float64 {
	rows, cells := in.Shape().RowCellSize()
	col := func(c int) *tensor.Float64 {
		v := tensor.NewFloat64(rows)
		for r := range rows {
			v.Values[r] = in.Values[r*cells+c]
		}
		return v
	}
	res := make([]float64, cells*cells)
	for i := range cells {
		for j := 0; j <= i; j++ {
			v := fun(col(i), col(j)).Float1D(0)
			res[i*cells+j] = v
			res[j*cells+i] = v
		}
	}
	return res
}

func covarTestInput(rows int, nan bool) *tensor.Float64 {
	in := tensor.NewFloat64(rows, 3, 4)
	rnd := rand.New(rand.NewPCG(1, 2))
	for i := range in.Len() {
		v := rnd.Float64()
		if i%5 == 0 {
			v = 0
		}
		in.Values[i] = v
	}
	_, cells := in.Shape().RowCellSize()
	for r := range rows { // constant (zero variance) cell 7
		in.Values[r*cells+7] = 0.5
	}
	if nan {
		in.Values[3*cells+2] = math.NaN()
	}
	return in
}

func TestCovarianceMatrixFuncs(t *testing.T) {
	for _, nan := range []bool{false, true} {
		in := covarTestInput(40, nan)
		for _, fn := range []struct {
			name string
			fun  MetricFunc
		}{{"Covariance", Covariance}, {"Correlation", Correlation}, {"InvCosine", InvCosine}} {
			t.Run(fmt.Sprintf("%s nan=%v", fn.name, nan), func(t *testing.T) {
				out := tensor.NewFloat64()
				assert.NoError(t, CovarianceMatrixOut(fn.fun, in, out))
				tolassert.EqualTolSlice(t, covarMatrixRef(fn.fun, in), out.Values, 1.0e-12)
			})
		}
	}
}

func BenchmarkCovarianceMatrixPCA(b *testing.B) {
	// typical hidden layer PCA: 512 trials x 20x20 units
	in := tensor.NewFloat64(512, 20, 20)
	rnd := rand.New(rand.NewPCG(1, 2))
	for i := range in.Len() {
		in.Values[i] = rnd.Float64()
	}
	out := tensor.NewFloat64()
	for b.Loop() {
		CovarianceMatrixOut(Covariance, in, out)
	}
}
