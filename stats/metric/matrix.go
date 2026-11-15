// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import (
	"math"
	"reflect"

	"cogentcore.org/lab/matrix"
	"cogentcore.org/lab/tensor"
)

// MatrixOut computes the rows x rows square distance / similarity matrix
// between the patterns for each row of the given higher dimensional input tensor,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns (cells). Use [tensor.NewRowCellsView]
// to organize data into the desired split between a 1D outermost Row dimension
// and the remaining Cells dimension.
// The metric function must have the [MetricFunc] signature.
// The results fill in the elements of the output matrix, which is symmetric,
// and only the lower triangular part is computed, with results copied
// to the upper triangular region, for maximum efficiency.
func MatrixOut(fun any, in tensor.Tensor, out tensor.Values) error {
	mfun, err := AsMetricFunc(fun)
	if err != nil {
		return err
	}
	rows, cells := in.Shape().RowCellSize()
	if rows == 0 || cells == 0 {
		return nil
	}
	out.SetShapeSizes(rows, rows)
	coords := matrix.TriLIndicies(rows)
	nc := coords.DimSize(0)
	// note: flops estimating 3 per item on average -- different for different metrics.
	tensor.VectorizeThreaded(cells*3, func(tsr ...tensor.Tensor) int { return nc },
		func(idx int, tsr ...tensor.Tensor) {
			cx := coords.Int(idx, 0)
			cy := coords.Int(idx, 1)
			sa := tensor.Cells1D(tsr[0], cx)
			sb := tensor.Cells1D(tsr[0], cy)
			mout := mfun(sa, sb)
			tsr[1].SetFloat(mout.Float1D(0), cx, cy)
		}, in, out)
	for idx := range nc { // copy to upper
		cx := coords.Int(idx, 0)
		cy := coords.Int(idx, 1)
		if cx == cy { // exclude diag
			continue
		}
		out.SetFloat(out.Float(cx, cy), cy, cx)
	}
	return nil
}

// Matrix computes the rows x rows square distance / similarity matrix
// between the patterns for each row of the given higher dimensional input tensor,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns (cells). Use [tensor.NewRowCellsView]
// to organize data into the desired split between a 1D outermost Row dimension
// and the remaining Cells dimension.
// The metric function must have the [MetricFunc] signature.
// The results fill in the elements of the output matrix, which is symmetric,
// and only the lower triangular part is computed, with results copied
// to the upper triangular region, for maximum efficiency.
func Matrix(fun any, in tensor.Tensor) tensor.Values {
	return tensor.CallOut1Gen1(MatrixOut, fun, in)
}

// CrossMatrixOut computes the distance / similarity matrix between
// two different sets of patterns in the two input tensors, where
// the patterns are in the sub-space cells of the tensors,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns that the given distance metric
// function is applied to, with the results filling in the cells of the output matrix.
// The metric function must have the [MetricFunc] signature.
// The rows of the output matrix are the rows of the first input tensor,
// and the columns of the output are the rows of the second input tensor.
func CrossMatrixOut(fun any, a, b tensor.Tensor, out tensor.Values) error {
	mfun, err := AsMetricFunc(fun)
	if err != nil {
		return err
	}
	arows, acells := a.Shape().RowCellSize()
	if arows == 0 || acells == 0 {
		return nil
	}
	brows, bcells := b.Shape().RowCellSize()
	if brows == 0 || bcells == 0 {
		return nil
	}
	out.SetShapeSizes(arows, brows)
	// note: flops estimating 3 per item on average -- different for different metrics.
	flops := min(acells, bcells) * 3
	nc := arows * brows
	tensor.VectorizeThreaded(flops, func(tsr ...tensor.Tensor) int { return nc },
		func(idx int, tsr ...tensor.Tensor) {
			ar := idx / brows
			br := idx % brows
			sa := tensor.Cells1D(tsr[0], ar)
			sb := tensor.Cells1D(tsr[1], br)
			mout := mfun(sa, sb)
			tsr[2].SetFloat(mout.Float1D(0), ar, br)
		}, a, b, out)
	return nil
}

// CrossMatrix computes the distance / similarity matrix between
// two different sets of patterns in the two input tensors, where
// the patterns are in the sub-space cells of the tensors,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns that the given distance metric
// function is applied to, with the results filling in the cells of the output matrix.
// The metric function must have the [MetricFunc] signature.
// The rows of the output matrix are the rows of the first input tensor,
// and the columns of the output are the rows of the second input tensor.
func CrossMatrix(fun any, a, b tensor.Tensor) tensor.Values {
	return tensor.CallOut2Gen1(CrossMatrixOut, fun, a, b)
}

// CovarianceMatrixOut generates the cells x cells square covariance matrix
// for all per-row cells of the given higher dimensional input tensor,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns (cells).
// Each value in the resulting matrix represents the extent to which the
// value of a given cell covaries across the rows of the tensor with the
// value of another cell.
// Uses the given metric function, typically [Covariance] or [Correlation],
// The metric function must have the [MetricFunc] signature.
// Use Covariance if vars have similar overall scaling,
// which is typical in neural network models, and use
// Correlation if they are on very different scales, because it effectively rescales).
// The resulting matrix can be used as the input to PCA or SVD eigenvalue decomposition.
//
// The input data is first copied into contiguous per-cell vectors, so the
// metric function operates on plain 1D tensors. [Covariance] and [Correlation]
// on data without NaN values are computed directly from per-cell means
// computed once, which is much faster than calling the metric per pair.
func CovarianceMatrixOut(fun any, in tensor.Tensor, out tensor.Values) error {
	mfun, err := AsMetricFunc(fun)
	if err != nil {
		return err
	}
	rows, cells := in.Shape().RowCellSize()
	if rows == 0 || cells == 0 {
		return nil
	}
	out.SetShapeSizes(cells, cells)

	// cols has the data transposed into one contiguous vector per cell,
	// read once from the input in row-major order.
	cols := tensor.NewFloat64(cells, rows)
	hasNaN := false
	for r := range rows {
		for c := range cells {
			v := in.Float1D(r*cells + c)
			if math.IsNaN(v) {
				hasNaN = true
			}
			cols.Values[c*rows+r] = v
		}
	}

	if !hasNaN {
		fp := reflect.ValueOf(mfun).Pointer()
		switch fp {
		case reflect.ValueOf(Covariance).Pointer():
			covarianceMatrixDirect(cols.Values, rows, cells, false, out)
			return nil
		case reflect.ValueOf(Correlation).Pointer():
			covarianceMatrixDirect(cols.Values, rows, cells, true, out)
			return nil
		}
	}

	coords := matrix.TriLIndicies(cells)
	nc := coords.DimSize(0)
	// note: flops estimating 3 per item on average -- different for different metrics.
	tensor.VectorizeThreaded(rows*3, func(tsr ...tensor.Tensor) int { return nc },
		func(idx int, tsr ...tensor.Tensor) {
			cx := coords.Int(idx, 0)
			cy := coords.Int(idx, 1)
			av := tensor.Cells1D(cols, cx)
			bv := tensor.Cells1D(cols, cy)
			mout := mfun(av, bv)
			tsr[0].SetFloat(mout.Float1D(0), cx, cy)
		}, out)
	for idx := range nc { // copy to upper
		cx := coords.Int(idx, 0)
		cy := coords.Int(idx, 1)
		if cx == cy { // exclude diag
			continue
		}
		out.SetFloat(out.Float(cx, cy), cy, cx)
	}
	return nil
}

// covarianceMatrixDirect computes the covariance matrix (or the correlation
// matrix if corr) of the per-cell vectors in cols ([cells][rows], without NaNs)
// into out, giving the same values as [Covariance] or [Correlation] per pair.
// Centers each cell vector once, then computes the lower triangle of
// dot products, threaded over output rows.
func covarianceMatrixDirect(cols []float64, rows, cells int, corr bool, out tensor.Values) {
	for c := range cells {
		x := cols[c*rows : (c+1)*rows]
		mean := 0.0
		for _, v := range x {
			mean += v
		}
		mean /= float64(rows)
		for r := range x {
			x[r] -= mean
		}
	}
	res := make([]float64, cells*cells)
	n := float64(rows)
	tensor.VectorizeThreaded(rows*cells/2, func(tsr ...tensor.Tensor) int { return cells },
		func(i int, tsr ...tensor.Tensor) {
			xi := cols[i*rows : (i+1)*rows]
			for j := 0; j <= i; j++ {
				xj := cols[j*rows : (j+1)*rows]
				s := 0.0
				for r, v := range xi {
					s += v * xj[r]
				}
				res[i*cells+j] = s / n
			}
		}, out)
	if corr {
		for i := range cells {
			for j := range i {
				vp := math.Sqrt(res[i*cells+i] * res[j*cells+j])
				if vp > 0 {
					res[i*cells+j] /= vp
				}
			}
		}
		for i := range cells { // after off-diagonals, which use the variances
			if res[i*cells+i] > 0 {
				res[i*cells+i] = 1
			}
		}
	}
	for i := range cells {
		for j := 0; j <= i; j++ {
			v := res[i*cells+j]
			out.SetFloat1D(v, i*cells+j)
			out.SetFloat1D(v, j*cells+i)
		}
	}
}

// CovarianceMatrix generates the cells x cells square covariance matrix
// for all per-row cells of the given higher dimensional input tensor,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns (cells).
// Each value in the resulting matrix represents the extent to which the
// value of a given cell covaries across the rows of the tensor with the
// value of another cell.
// Uses the given metric function, typically [Covariance] or [Correlation],
// The metric function must have the [MetricFunc] signature.
// Use Covariance if vars have similar overall scaling,
// which is typical in neural network models, and use
// Correlation if they are on very different scales, because it effectively rescales).
// The resulting matrix can be used as the input to PCA or SVD eigenvalue decomposition.
func CovarianceMatrix(fun any, in tensor.Tensor) tensor.Values {
	return tensor.CallOut1Gen1(CovarianceMatrixOut, fun, in)
}
