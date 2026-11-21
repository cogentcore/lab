// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

// assertNear asserts that the values of the given tensors are within 1e-9.
func assertNear(t *testing.T, want, got tensor.Tensor) {
	t.Helper()
	assert.Equal(t, want.ShapeSizes(), got.ShapeSizes())
	assert.InDeltaSlice(t, tensor.AsFloat64Slice(want), tensor.AsFloat64Slice(got), 1e-9)
}

func TestDiag(t *testing.T) {
	d := Diag(tensor.NewFloat64FromValues(1, 2))
	assert.Equal(t, []float64{1, 0, 0, 2}, d.Values)
	assert.Equal(t, []float64{1, 2}, Diag(d).Values)
}

func TestSolve(t *testing.T) {
	a := tensor.NewFloat64FromValues(2, 1, 1, 3)
	a.SetShapeSizes(2, 2)
	x := Solve(a, tensor.NewFloat64FromValues(3, 5))
	assertNear(t, tensor.NewFloat64FromValues(0.8, 1.4), x)

	// least squares: the line through (0, 1), (1, 3), (2, 5)
	l := tensor.NewFloat64FromValues(1, 0, 1, 1, 1, 2)
	l.SetShapeSizes(3, 2)
	assertNear(t, tensor.NewFloat64FromValues(1, 2), Solve(l, tensor.NewFloat64FromValues(1, 3, 5)))
	assert.Error(t, SolveOut(a, tensor.NewFloat64FromValues(1, 2, 3), tensor.NewFloat64()))
}

func TestDecompositions(t *testing.T) {
	a := tensor.NewFloat64FromValues(4, 2, 2, 3, 1, 5)
	a.SetShapeSizes(3, 2)
	u, s, vh := SVDFactors(a)
	assert.Equal(t, []int{3, 2}, u.ShapeSizes())
	assert.Greater(t, s.Values[0], s.Values[1])
	assertNear(t, a, Mul(Mul(u, Diag(s)), vh))
	assertNear(t, Identity(2), Mul(PseudoInverse(a), a))
	assert.Equal(t, 2, Rank(a).Int1D(0))
	assert.Equal(t, 1, Rank(tensor.Reshape(tensor.NewFloat64FromValues(1, 2, 2, 4), 2, 2)).Int1D(0))

	q, r := QR(a)
	assertNear(t, a, Mul(q, r))

	sp := tensor.NewFloat64FromValues(4, 2, 2, 3)
	sp.SetShapeSizes(2, 2)
	l := Cholesky(sp)
	assert.Equal(t, 0.0, l.Value(0, 1), "lower triangular")
	assertNear(t, sp, Mul(l, tensor.Transpose(l)))
	assert.InDelta(t, 5.0, Norm(tensor.NewFloat64FromValues(3, 4)).Float1D(0), 1e-12)
}
