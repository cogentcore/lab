// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package matrix

import (
	"errors"
	"fmt"
	"math"

	errs "cogentcore.org/core/base/errors"
	"cogentcore.org/lab/tensor"
	"gonum.org/v1/gonum/mat"
)

// Diag returns a square 2D matrix with the values of the given 1D tensor
// on its diagonal and zeros elsewhere, or, for a 2D tensor, a 1D copy of the
// values of its diagonal (as diag does in NumPy). See [Diagonal] for a view
// of the diagonal of a matrix.
func Diag(a tensor.Tensor) *tensor.Float64 {
	if a.NumDims() >= 2 {
		return tensor.AsFloat64(tensor.Clone(Diagonal(a)))
	}
	n := a.Len()
	out := tensor.NewFloat64(n, n)
	for i := range n {
		out.Values[i*n+i] = a.Float1D(i)
	}
	return out
}

// Solve returns the solution x of the linear equations a x = b, for a 2D
// matrix a and a 1D vector or 2D matrix b (as linalg.solve does in NumPy).
// If a is not square, x is the least squares solution (as linalg.lstsq
// does in NumPy). Errors (e.g., for a singular matrix) are logged; see
// [SolveOut] for a version that returns them.
func Solve(a, b tensor.Tensor) *tensor.Float64 {
	return CallOut2(SolveOut, a, b)
}

// SolveOut computes the solution x of the linear equations a x = b, into
// the given output tensor, as [Solve] does, and returns any error.
func SolveOut(a, b tensor.Tensor, out *tensor.Float64) error {
	if err := matrix2D("matrix.Solve", a); err != nil {
		return err
	}
	vec := b.NumDims() == 1
	if vec {
		b = tensor.Reshape(b, b.Len(), 1)
	}
	if err := matrix2D("matrix.Solve", b); err != nil {
		return err
	}
	if a.DimSize(0) != b.DimSize(0) {
		return fmt.Errorf("matrix.Solve: the shapes %v and %v do not match: a and b must have the same number of rows: %w", a.ShapeSizes(), b.ShapeSizes(), mat.ErrShape)
	}
	ma, _ := NewMatrix(a)
	mb, _ := NewMatrix(b)
	var x mat.Dense
	err := x.Solve(ma, mb)
	CopyFromDense(out, &x)
	if vec {
		out.SetShapeSizes(out.Len())
	}
	return err
}

// PseudoInverse returns the Moore-Penrose pseudo-inverse of the given 2D
// matrix (as linalg.pinv does in NumPy), computed from its singular value
// decomposition, with the singular values that are smaller than 1e-15 times
// the largest one treated as zero.
func PseudoInverse(a tensor.Tensor) *tensor.Float64 {
	u, s, vh := SVDFactors(a)
	if s.Len() == 0 {
		return tensor.NewFloat64(0)
	}
	m, n := a.DimSize(0), a.DimSize(1)
	k := s.Len()
	tol := 1e-15 * s.Values[0]
	out := tensor.NewFloat64(n, m)
	for i := range n {
		for j := range m {
			sum := 0.0
			for l := range k {
				if s.Values[l] > tol {
					sum += vh.Values[l*n+i] * u.Values[j*k+l] / s.Values[l]
				}
			}
			out.Values[i*m+j] = sum
		}
	}
	return out
}

// SVDFactors returns the singular value decomposition of the given 2D
// matrix a of any shape (as linalg.svd does in NumPy, with
// full_matrices=False): a = u @ diag(s) @ vh, with the singular values s in
// descending order, u with a column for each of them, and vh with a row for
// each of them. See [SVD] for the decomposition of a symmetric matrix, as
// for principal components analysis. Errors are logged.
func SVDFactors(a tensor.Tensor) (u, s, vh *tensor.Float64) {
	u, s, vh = tensor.NewFloat64(0), tensor.NewFloat64(0), tensor.NewFloat64(0)
	if errs.Log(matrix2D("matrix.SVDFactors", a)) != nil {
		return
	}
	ma, _ := NewMatrix(a)
	var svd mat.SVD
	if !svd.Factorize(ma, mat.SVDThin) {
		errs.Log(errors.New("matrix.SVDFactors: the factorization failed"))
		return
	}
	var um, vm mat.Dense
	svd.UTo(&um)
	svd.VTo(&vm)
	CopyFromDense(u, &um)
	CopyFromDense(vh, mat.DenseCopyOf(vm.T()))
	s.SetShapeSizes(len(svd.Values(nil)))
	svd.Values(s.Values)
	return
}

// Cholesky returns the lower triangular matrix L of the Cholesky
// decomposition of the given symmetric positive definite 2D matrix a,
// such that a = L @ L.T (as linalg.cholesky does in NumPy). Errors, such
// as for a matrix that is not positive definite, are logged.
func Cholesky(a tensor.Tensor) *tensor.Float64 {
	out := tensor.NewFloat64(0)
	if errs.Log(matrix2D("matrix.Cholesky", a)) != nil {
		return out
	}
	sy, err := NewSymmetric(a)
	if errs.Log(err) != nil {
		return out
	}
	var ch mat.Cholesky
	if !ch.Factorize(sy) {
		errs.Log(errors.New("matrix.Cholesky: the matrix is not positive definite"))
		return out
	}
	var l mat.TriDense
	ch.LTo(&l)
	CopyFromDense(out, mat.DenseCopyOf(&l))
	return out
}

// QR returns the QR decomposition of the given 2D matrix a, such that
// a = q @ r, with q orthogonal and r upper triangular (as linalg.qr does
// in NumPy, in its "complete" mode).
func QR(a tensor.Tensor) (q, r *tensor.Float64) {
	q, r = tensor.NewFloat64(0), tensor.NewFloat64(0)
	if errs.Log(matrix2D("matrix.QR", a)) != nil {
		return
	}
	ma, _ := NewMatrix(a)
	var qr mat.QR
	qr.Factorize(ma)
	var qm, rm mat.Dense
	qr.QTo(&qm)
	qr.RTo(&rm)
	CopyFromDense(q, &qm)
	CopyFromDense(r, &rm)
	return
}

// LU returns the LU decomposition with partial pivoting of the given
// square 2D matrix a, such that a = p @ l @ u, with the permutation matrix
// p, the lower triangular l with ones on its diagonal, and the upper
// triangular u (as scipy.linalg.lu does).
func LU(a tensor.Tensor) (p, l, u *tensor.Float64) {
	p, l, u = tensor.NewFloat64(0), tensor.NewFloat64(0), tensor.NewFloat64(0)
	if errs.Log(matrix2D("matrix.LU", a)) != nil {
		return
	}
	n := a.DimSize(0)
	if n != a.DimSize(1) {
		errs.Log(squareError("matrix.LU", a))
		return
	}
	ma, _ := NewMatrix(a)
	var lu mat.LU
	lu.Factorize(ma)
	var lm, um mat.TriDense
	lu.LTo(&lm)
	lu.UTo(&um)
	CopyFromDense(l, mat.DenseCopyOf(&lm))
	CopyFromDense(u, mat.DenseCopyOf(&um))
	p.SetShapeSizes(n, n)
	for i, r := range lu.RowPivots(nil) {
		p.Values[i*n+r] = 1
	}
	return
}

// Rank returns the rank of the given 2D matrix (as linalg.matrix_rank
// does in NumPy): the number of its singular values that are greater than
// the largest one times the largest dimension times the float64 epsilon.
func Rank(a tensor.Tensor) *tensor.Int {
	_, s, _ := SVDFactors(a)
	rank := 0
	if s.Len() > 0 {
		tol := s.Values[0] * float64(max(a.DimSize(0), a.DimSize(1))) * 0x1p-52
		for _, v := range s.Values {
			if v > tol {
				rank++
			}
		}
	}
	return tensor.NewIntScalar(rank)
}

// Norm returns the L2 (Euclidean, or Frobenius for a matrix) norm of all
// of the values of the given tensor (as linalg.norm does in NumPy).
func Norm(a tensor.Tensor) *tensor.Float64 {
	sum := 0.0
	for i := range a.Len() {
		v := a.Float1D(i)
		sum += v * v
	}
	return tensor.NewFloat64Scalar(math.Sqrt(sum))
}

// matrix2D returns an error if the given tensor is not a 2D matrix of
// numbers, for the given function.
func matrix2D(fun string, a tensor.Tensor) error {
	if err := StringCheck(a); err != nil {
		return err
	}
	if a.NumDims() != 2 {
		return fmt.Errorf("%s: the shape %v is not that of a 2D matrix: %w", fun, a.ShapeSizes(), mat.ErrShape)
	}
	return nil
}
