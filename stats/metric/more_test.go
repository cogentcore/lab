// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import (
	"math"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// metricOutFuncs maps each metric to its output function.
var metricOutFuncs = map[Metrics]MetricOutFunc{
	MetricL2Norm:           L2NormOut,
	MetricSumSquares:       SumSquaresOut,
	MetricL1Norm:           L1NormOut,
	MetricHamming:          HammingOut,
	MetricL2NormBinTol:     L2NormBinTolOut,
	MetricSumSquaresBinTol: SumSquaresBinTolOut,
	MetricInvCosine:        InvCosineOut,
	MetricInvCorrelation:   InvCorrelationOut,
	MetricCrossEntropy:     CrossEntropyOut,
	MetricDotProduct:       DotProductOut,
	MetricCovariance:       CovarianceOut,
	MetricCorrelation:      CorrelationOut,
	MetricCosine:           CosineOut,
}

// makeTyped returns the given 2D values (rows x cells) as Float64,
// Float32, Int and Rows tensors; cells == 0 makes 1D tensors.
func makeTyped(rows, cells int, vals ...float64) []tensor.Tensor {
	shp := []int{rows}
	if cells > 0 {
		shp = append(shp, cells)
	}
	f64 := tensor.NewFloat64(shp...)
	f32 := tensor.NewFloat32(shp...)
	in := tensor.NewInt(shp...)
	for i, v := range vals {
		f64.SetFloat1D(v, i)
		f32.SetFloat1D(v, i)
		in.SetFloat1D(v, i)
	}
	return []tensor.Tensor{f64, f32, in, tensor.NewRows(f64)}
}

// TestMetricsKnownValues tests each metric on small integer-valued vectors
// against hand-computed values.
func TestMetricsKnownValues(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, 2, 0, 3)
	b := tensor.NewFloat64FromValues(2, 2, 1, 0)
	// means: a 1.5, b 1.25; deviations a: -.5 .5 -1.5 1.5; b: .75 .75 -.25 -1.25
	cov := (-0.375 + 0.375 + 0.375 - 1.875) / 4
	sda := math.Sqrt((0.25 + 0.25 + 2.25 + 2.25) / 4)
	sdb := math.Sqrt((0.5625 + 0.5625 + 0.0625 + 1.5625) / 4)
	cos := 6 / (math.Sqrt(14) * math.Sqrt(9))
	ce := -2*math.Log(0.999999) - math.Log(1-0.999999) - math.Log(0.000001)
	want := map[Metrics]float64{
		MetricL2Norm:           math.Sqrt(11),
		MetricSumSquares:       11,
		MetricL1Norm:           5,
		MetricHamming:          3,
		MetricL2NormBinTol:     math.Sqrt(11),
		MetricSumSquaresBinTol: 11,
		MetricDotProduct:       6,
		MetricCovariance:       cov,
		MetricCorrelation:      cov / (sda * sdb),
		MetricInvCorrelation:   1 - cov/(sda*sdb),
		MetricCosine:           cos,
		MetricInvCosine:        1 - cos,
		MetricCrossEntropy:     ce,
	}
	for m, w := range want {
		assert.InDelta(t, w, m.Call(a, b).Float1D(0), 1.0e-6, m.String())
	}
}

// TestMetricsTypes tests that every metric gives the same results for
// Float64, Float32, Int and Rows inputs in all combinations, for both
// 1D inputs and 2D inputs where each cell is computed separately.
func TestMetricsTypes(t *testing.T) {
	a0 := []float64{1, 2, 0, 3}
	b0 := []float64{2, 2, 1, 0}
	a1 := []float64{0, 1, 3, 1}
	b1 := []float64{1, 1, 2, 3}
	a2 := make([]float64, 8)
	b2 := make([]float64, 8)
	for i := range 4 {
		a2[2*i], a2[2*i+1] = a0[i], a1[i]
		b2[2*i], b2[2*i+1] = b0[i], b1[i]
	}
	as1 := makeTyped(4, 0, a0...)
	bs1 := makeTyped(4, 0, b0...)
	as2 := makeTyped(4, 2, a2...)
	bs2 := makeTyped(4, 2, b2...)
	ref1 := tensor.NewFloat64FromValues(a1...)
	refb1 := tensor.NewFloat64FromValues(b1...)

	for m := range MetricsN {
		fun := metricOutFuncs[m]
		require.NotNil(t, fun, m.String())
		w0 := m.Call(as1[0], bs1[0]).Float1D(0)
		w1 := m.Call(ref1, refb1).Float1D(0)
		for ai := range as1 {
			for bi := range bs1 {
				out := tensor.NewFloat64()
				require.NoError(t, fun(as1[ai], bs1[bi], out))
				assert.InDelta(t, w0, out.Float1D(0), 1.0e-6, "%s 1D %d %d", m, ai, bi)

				require.NoError(t, fun(as2[ai], bs2[bi], out))
				require.Equal(t, 2, out.Len())
				assert.InDelta(t, w0, out.Float1D(0), 1.0e-6, "%s 2D %d %d", m, ai, bi)
				assert.InDelta(t, w1, out.Float1D(1), 1.0e-6, "%s 2D %d %d", m, ai, bi)
			}
		}
	}
}

// TestMetricsErrors tests that every metric returns an error for
// inputs of different shapes, and handles empty inputs.
func TestMetricsErrors(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, 2, 3)
	b := tensor.NewFloat64FromValues(1, 2)
	for m, fun := range metricOutFuncs {
		out := tensor.NewFloat64()
		assert.Error(t, fun(a, b, out), m.String())
		e := tensor.NewFloat64(0)
		assert.NoError(t, fun(e, e, out), m.String())
	}
	_, err := SumSquaresOut64(a, b, tensor.NewFloat64())
	assert.Error(t, err)
}

// TestMetricsInfNaN tests the handling of infinite differences and
// NaN values, which are skipped.
func TestMetricsInfNaN(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, math.Inf(1))
	b := tensor.NewFloat64FromValues(0, 0)
	assert.True(t, math.IsInf(L2Norm(a, b).Float1D(0), 1))
	assert.True(t, math.IsInf(SumSquares(a, b).Float1D(0), 1))
	assert.True(t, math.IsInf(L2NormBinTol(a, b).Float1D(0), 1))
	assert.True(t, math.IsInf(SumSquaresBinTol(a, b).Float1D(0), 1))

	n := tensor.NewFloat64FromValues(1, math.NaN(), 3)
	z := tensor.NewFloat64FromValues(0, 5, 0)
	assert.Equal(t, 10.0, SumSquares(n, z).Float1D(0))
	assert.Equal(t, 4.0, L1Norm(n, z).Float1D(0))
	assert.Equal(t, 2.0, Hamming(n, z).Float1D(0))
	assert.Equal(t, 0.0, DotProduct(n, z).Float1D(0))

	// differences below 0.5 are ignored with binary tolerance
	c := tensor.NewFloat64FromValues(0.4, 1, 0)
	d := tensor.NewFloat64FromValues(0, 0, 0)
	assert.Equal(t, 1.0, SumSquaresBinTol(c, d).Float1D(0))
	assert.Equal(t, 1.0, L2NormBinTol(c, d).Float1D(0))

	// zero vectors have zero cosine and correlation
	assert.Equal(t, 0.0, Cosine(d, d).Float1D(0))
	assert.Equal(t, 0.0, Correlation(d, d).Float1D(0))
	assert.Equal(t, 1.0, InvCosine(d, d).Float1D(0))
	assert.Equal(t, 1.0, InvCorrelation(d, d).Float1D(0))
}

// TestMetricsHelpers tests Increasing, AsMetricFunc, AsMetricOutFunc and
// the generated Metrics enum helpers.
func TestMetricsHelpers(t *testing.T) {
	assert.True(t, MetricL2Norm.Increasing())
	assert.True(t, MetricCrossEntropy.Increasing())
	assert.False(t, MetricDotProduct.Increasing())
	assert.False(t, MetricCosine.Increasing())

	_, err := AsMetricFunc(L2Norm)
	assert.NoError(t, err)
	_, err = AsMetricFunc(L2NormOut)
	assert.Error(t, err)
	_, err = AsMetricOutFunc(L2NormOut)
	assert.NoError(t, err)
	_, err = AsMetricOutFunc(L2Norm)
	assert.Error(t, err)

	var m Metrics
	require.NoError(t, m.SetString("Cosine"))
	assert.Equal(t, MetricCosine, m)
	assert.Equal(t, "metric.Cosine", m.FuncName())
	assert.Error(t, m.SetString("NotAMetric"))
	assert.Len(t, MetricsValues(), int(MetricsN))
}

// TestMatrixFuncs tests Matrix, CrossMatrix and CovarianceMatrix on small
// known inputs, and their error and empty cases.
func TestMatrixFuncs(t *testing.T) {
	in := tensor.NewFloat64(3, 2)
	copy(in.Values, []float64{0, 0, 3, 4, 0, 1})
	m := Matrix(L2Norm, in)
	assert.Equal(t, []int{3, 3}, m.ShapeSizes())
	exp := []float64{0, 5, 1, 5, 0, math.Sqrt(18), 1, math.Sqrt(18), 0}
	for i, v := range exp {
		assert.InDelta(t, v, m.Float1D(i), 1.0e-9)
	}

	b := tensor.NewFloat64(2, 2)
	copy(b.Values, []float64{0, 0, 1, 1})
	cm := CrossMatrix(L1Norm, in, b)
	assert.Equal(t, []int{3, 2}, cm.ShapeSizes())
	assert.Equal(t, []float64{0, 2, 7, 5, 1, 1}, tensor.AsFloat64(cm).Values)

	cv := CovarianceMatrix(Covariance, in)
	assert.Equal(t, []int{2, 2}, cv.ShapeSizes())
	assert.InDelta(t, 2.0, cv.Float(0, 0), 1.0e-9)            // var of 0,3,0
	assert.InDelta(t, cv.Float(0, 1), cv.Float(1, 0), 1.0e-9) // symmetric

	out := tensor.NewFloat64()
	assert.Error(t, MatrixOut(1, in, out))
	assert.Error(t, CrossMatrixOut(1, in, b, out))
	assert.Error(t, CovarianceMatrixOut(1, in, out))
	assert.NoError(t, MatrixOut(L2Norm, tensor.NewFloat64(0, 2), out))
	assert.NoError(t, CrossMatrixOut(L2Norm, tensor.NewFloat64(0, 2), b, out))
	assert.NoError(t, CrossMatrixOut(L2Norm, in, tensor.NewFloat64(0, 2), out))
}

// TestClosestRow tests ClosestRow finding the nearest row of a vocabulary.
func TestClosestRow(t *testing.T) {
	vocab := tensor.NewFloat64(3, 2)
	copy(vocab.Values, []float64{0, 0, 5, 5, 1, 1})
	probe := tensor.NewFloat64FromValues(4, 4)
	cr := ClosestRow(L2Norm, probe, vocab)
	assert.Equal(t, 1.0, cr.Float1D(0))
	assert.InDelta(t, math.Sqrt(2), cr.Float1D(1), 1.0e-9)

	assert.Error(t, ClosestRowOut(1, probe, vocab, tensor.NewFloat64()))
}
