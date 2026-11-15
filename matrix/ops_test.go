// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"testing"

	"cogentcore.org/core/base/tolassert"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/mat"
)

// sym2 returns a symmetric 2x2 matrix [[2, 1], [1, 2]], with eigenvalues 1 and 3.
func sym2() *tensor.Float64 {
	a := tensor.NewFloat64FromValues(2, 1, 1, 2)
	a.SetShapeSizes(2, 2)
	return a
}

// TestMatrixConstructors tests NewMatrix, NewSymmetric and NewDense,
// including their errors for string and non-2D tensors, and the
// mat.Matrix interface methods.
func TestMatrixConstructors(t *testing.T) {
	s := tensor.NewStringFromValues("a", "b")
	_, err := NewMatrix(s)
	assert.Error(t, err)
	_, err = NewSymmetric(s)
	assert.Error(t, err)
	assert.Error(t, StringCheck(s))

	v := tensor.NewFloat64(3)
	_, err = NewMatrix(v)
	assert.Error(t, err)
	_, err = NewSymmetric(v)
	assert.Error(t, err)
	_, err = NewDense(v)
	assert.Error(t, err)
	_, err = NewSymmetric(tensor.NewFloat64(2, 3))
	assert.Error(t, err)

	a := tensor.NewFloat64FromValues(1, 2, 3, 4, 5, 6)
	a.SetShapeSizes(2, 3)
	m, err := NewMatrix(a)
	require.NoError(t, err)
	r, c := m.Dims()
	assert.Equal(t, 2, r)
	assert.Equal(t, 3, c)
	assert.Equal(t, 6.0, m.At(1, 2))
	tr := m.T()
	r, c = tr.Dims()
	assert.Equal(t, 3, r)
	assert.Equal(t, 2, c)
	assert.Equal(t, 6.0, tr.At(2, 1))
	assert.Equal(t, 4.0, tr.At(0, 1))

	sy, err := NewSymmetric(sym2())
	require.NoError(t, err)
	assert.Equal(t, 2, sy.SymmetricDim())

	d, err := NewDense(a)
	require.NoError(t, err)
	d.Set(0, 0, 10) // a view onto the tensor data
	assert.Equal(t, 10.0, a.Float(0, 0))

	out := tensor.NewFloat64()
	CopyFromDense(out, mat.NewDense(2, 2, []float64{1, 2, 3, 4}))
	assert.Equal(t, []int{2, 2}, out.ShapeSizes())
	assert.Equal(t, []float64{1, 2, 3, 4}, out.Values)
}

// TestMulErrors tests the error cases of MulOut for mismatched shapes
// and string tensors.
func TestMulErrors(t *testing.T) {
	out := tensor.NewFloat64()
	a := tensor.NewFloat64(2, 3)
	assert.ErrorIs(t, MulOut(a, a, out), mat.ErrShape)
	assert.Error(t, MulOut(tensor.NewStringFromValues("a"), a, out))
	assert.Error(t, MulOut(a, tensor.NewStringFromValues("a"), out))
	assert.ErrorIs(t, MulOut(tensor.NewFloat64(2, 3, 3), a, out), mat.ErrShape)
	assert.ErrorIs(t, MulOut(a, tensor.NewFloat64(2, 2, 2), out), mat.ErrShape)
	assert.Error(t, MulOut(tensor.NewFloat64(2, 2, 2), tensor.NewFloat64(3, 2, 2), out))
	assert.ErrorIs(t, MulOut(tensor.NewFloat64(2, 2, 3), tensor.NewFloat64(2, 2, 2), out), mat.ErrShape)
}

// TestDetInverse tests Det, LogDet and InverseOut, including the error
// cases of InverseOut for non-square and singular matricies.
func TestDetInverse(t *testing.T) {
	a := sym2()
	assert.InDelta(t, 3.0, Det(a).Float1D(0), 1e-12)
	ld := LogDet(a)
	assert.InDelta(t, 1.0986122886681098, ld.Float1D(0), 1e-12) // log(3)
	assert.Equal(t, 1.0, ld.Float1D(1))

	v := tensor.NewFloat64(3)
	assert.Equal(t, 0.0, Det(v).Float1D(0))
	assert.Equal(t, 0.0, LogDet(v).Float1D(0))

	out := tensor.NewFloat64()
	assert.ErrorIs(t, InverseOut(v, out), mat.ErrShape)
	assert.ErrorIs(t, InverseOut(tensor.NewFloat64(2, 3), out), mat.ErrShape)
	assert.ErrorIs(t, InverseOut(tensor.NewFloat64(2, 2, 3), out), mat.ErrShape)
	assert.Error(t, InverseOut(tensor.NewStringFromValues("a"), out))
	assert.Error(t, InverseOut(tensor.NewFloat64(2, 2), out)) // singular

	// a list of one matrix is inverted as that matrix
	one := tensor.Reshape(a, 1, 2, 2)
	require.NoError(t, InverseOut(one, out))
	tolassert.EqualTolSlice(t, []float64{2.0 / 3, -1.0 / 3, -1.0 / 3, 2.0 / 3}, out.Values, 1e-12)

	// a list with a singular matrix
	two := tensor.NewFloat64(2, 2, 2)
	two.SetRowTensor(a, 0)
	assert.Error(t, InverseOut(two, out))
}

// TestEigSVD tests EigSym, SVD and SVDValues on a single 2x2 symmetric
// matrix, on a list of two of them, and on a list of one, and their
// shape and string errors.
func TestEigSVD(t *testing.T) {
	a := sym2()
	vecs, vals := EigSym(a)
	tolassert.EqualTolSlice(t, []float64{1, 3}, vals.Values, 1e-12)
	assert.Equal(t, []int{2, 2}, vecs.ShapeSizes())
	// the last column is the eigenvector of the largest eigenvalue: (1,1)/sqrt(2)
	assert.InDelta(t, vecs.Float(0, 1), vecs.Float(1, 1), 1e-12)

	vecs, vals = SVD(a)
	tolassert.EqualTolSlice(t, []float64{3, 1}, vals.Values, 1e-12)
	assert.InDelta(t, vecs.Float(0, 0), vecs.Float(1, 0), 1e-12)

	vals = SVDValues(a)
	tolassert.EqualTolSlice(t, []float64{3, 1}, vals.Values, 1e-12)

	two := tensor.NewFloat64(2, 2, 2)
	two.SetRowTensor(a, 0)
	two.SetRowTensor(a, 1)
	vecs, vals = EigSym(two)
	assert.Equal(t, []int{2, 2, 2}, vecs.ShapeSizes())
	tolassert.EqualTolSlice(t, []float64{1, 3, 1, 3}, vals.Values, 1e-12)
	vecs, vals = SVD(two)
	assert.Equal(t, []int{2, 2, 2}, vecs.ShapeSizes())
	tolassert.EqualTolSlice(t, []float64{3, 1, 3, 1}, vals.Values, 1e-12)
	vals = SVDValues(two)
	tolassert.EqualTolSlice(t, []float64{3, 1, 3, 1}, vals.Values, 1e-12)

	// a list of one matrix is decomposed as that matrix
	one := tensor.Reshape(a, 1, 2, 2)
	vecs, vals = tensor.NewFloat64(), tensor.NewFloat64()
	require.NoError(t, EigSymOut(one, vecs, vals))
	tolassert.EqualTolSlice(t, []float64{1, 3}, vals.Values, 1e-12)
	require.NoError(t, SVDOut(one, vecs, vals))
	tolassert.EqualTolSlice(t, []float64{3, 1}, vals.Values, 1e-12)
	require.NoError(t, SVDValuesOut(one, vals))
	tolassert.EqualTolSlice(t, []float64{3, 1}, vals.Values, 1e-12)

	s := tensor.NewStringFromValues("a")
	v := tensor.NewFloat64(3)
	ns := tensor.NewFloat64(2, 3)
	ns3 := tensor.NewFloat64(2, 2, 3)
	for _, tsr := range []tensor.Tensor{v, ns, ns3} {
		assert.ErrorIs(t, EigSymOut(tsr, vecs, vals), mat.ErrShape)
		assert.ErrorIs(t, SVDOut(tsr, vecs, vals), mat.ErrShape)
		assert.ErrorIs(t, SVDValuesOut(tsr, vals), mat.ErrShape)
		assert.ErrorIs(t, EigOut(tsr, vecs, vals), mat.ErrShape)
	}
	assert.Error(t, EigSymOut(s, vecs, vals))
	assert.Error(t, SVDOut(s, vecs, vals))
	assert.Error(t, SVDValuesOut(s, vals))
	assert.Error(t, EigOut(s, vecs, vals))
}

// TestEig tests the shapes produced by the non-symmetric Eig, which does
// not yet fill in its (complex valued) results.
func TestEig(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, 2, 3, 4)
	a.SetShapeSizes(2, 2)
	vecs, vals := Eig(a)
	assert.Equal(t, []int{2, 2}, vecs.ShapeSizes())
	assert.Equal(t, []int{2}, vals.ShapeSizes())

	two := tensor.NewFloat64(2, 2, 2)
	two.SetRowTensor(a, 0)
	two.SetRowTensor(a, 1)
	vecs, vals = Eig(two)
	assert.Equal(t, []int{2, 2, 2}, vecs.ShapeSizes())
	assert.Equal(t, []int{2, 2}, vals.ShapeSizes())

	require.NoError(t, EigOut(tensor.Reshape(a, 1, 2, 2), vecs, vals))
	assert.Equal(t, []int{2, 2}, vecs.ShapeSizes())
}

// TestProjectOnMatrixColumn tests projecting rows of a 2D tensor onto
// a column of a matrix.
func TestProjectOnMatrixColumn(t *testing.T) {
	mtx := tensor.NewFloat64FromValues(1, 2, 3, 4)
	mtx.SetShapeSizes(2, 2)
	vec := tensor.NewFloat64FromValues(1, 1, 2, 0)
	vec.SetShapeSizes(2, 2)
	out := ProjectOnMatrixColumn(mtx, vec, tensor.NewIntScalar(1))
	// column 1 is (2, 4): rows project to 1*2+1*4=6 and 2*2+0*4=4
	assert.Equal(t, []float64{6, 4}, tensor.AsFloat64(out).Values)
}

// TestIndicesViews tests Diagonal, TriLView and TriUView with offsets,
// and that they return nil for a non-2D tensor.
func TestIndicesViews(t *testing.T) {
	a := tensor.AsFloat64(tensor.Reshape(tensor.NewIntRange(9), 3, 3))
	assert.Equal(t, []float64{0, 4, 8}, tensor.AsFloat64(Diagonal(a)).Values)
	assert.Equal(t, []float64{1, 5}, tensor.AsFloat64(Diagonal(a, 1)).Values)
	assert.Equal(t, []float64{3, 7}, tensor.AsFloat64(Diagonal(a, -1)).Values)
	assert.Equal(t, 6, TriLView(a).Len())
	assert.Equal(t, 3, TriLView(a, -1).Len())
	assert.Equal(t, 6, TriUView(a).Len())
	assert.Equal(t, 3, TriUView(a, 1).Len())
	v := tensor.NewFloat64(3)
	assert.Nil(t, Diagonal(v))
	assert.Nil(t, TriLView(v))
	assert.Nil(t, TriUView(v))
}
