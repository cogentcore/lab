// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metric

import (
	"cogentcore.org/core/math32/vecint"
	"cogentcore.org/core/tensor"
)

func init() {
	tensor.AddFunc("metric.Matrix", Matrix, 1, tensor.StringFirstArg)
	tensor.AddFunc("metric.CrossMatrix", CrossMatrix, 1, tensor.StringFirstArg)
}

// Matrix computes the rows x rows square distance / similarity matrix
// between the patterns for each row of the given higher dimensional input tensor,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns (cells).
// The metric function registered in tensor Funcs can be passed as Metrics.String().
// The results fill in the elements of the output matrix, which is symmetric,
// and only the lower triangular part is computed, with results copied
// to the upper triangular region, for maximum efficiency.
func Matrix(funcName string, in, out *tensor.Indexed) {
	rows, cells := in.RowCellSize()
	if rows == 0 || cells == 0 {
		return
	}
	out.Tensor.SetShape(rows, rows)
	mout := tensor.NewFloatScalar(0.0)
	coords := TriangularLIndicies(rows)
	nc := len(coords)
	// note: flops estimating 3 per item on average -- different for different metrics.
	tensor.VectorizeThreaded(cells*3, func(tsr ...*tensor.Indexed) int { return nc },
		func(idx int, tsr ...*tensor.Indexed) {
			c := coords[idx]
			sa := tsr[0].Cells1D(c.X)
			sb := tsr[0].Cells1D(c.Y)
			tensor.Call(funcName, sa, sb, mout)
			tsr[1].SetFloat(mout.Tensor.Float1D(0), c.X, c.Y)
		}, in, out)
	for _, c := range coords { // copy to upper
		if c.X == c.Y { // exclude diag
			continue
		}
		out.Tensor.SetFloat(out.Tensor.Float(c.X, c.Y), c.Y, c.X)
	}
}

// CrossMatrix computes the distance / similarity matrix between
// two different sets of patterns in the two input tensors, where
// the patterns are in the sub-space cells of the tensors,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns that the given distance metric
// function is applied to, with the results filling in the cells of the output matrix.
// The metric function registered in tensor Funcs can be passed as Metrics.String().
// The rows of the output matrix are the rows of the first input tensor,
// and the columns of the output are the rows of the second input tensor.
func CrossMatrix(funcName string, a, b, out *tensor.Indexed) {
	arows, acells := a.RowCellSize()
	if arows == 0 || acells == 0 {
		return
	}
	brows, bcells := b.RowCellSize()
	if brows == 0 || bcells == 0 {
		return
	}
	out.Tensor.SetShape(arows, brows)
	mout := tensor.NewFloatScalar(0.0)
	// note: flops estimating 3 per item on average -- different for different metrics.
	flops := min(acells, bcells) * 3
	nc := arows * brows
	tensor.VectorizeThreaded(flops, func(tsr ...*tensor.Indexed) int { return nc },
		func(idx int, tsr ...*tensor.Indexed) {
			ar := idx / brows
			br := idx % brows
			sa := tsr[0].Cells1D(ar)
			sb := tsr[1].Cells1D(br)
			tensor.Call(funcName, sa, sb, mout)
			tsr[2].SetFloat(mout.Tensor.Float1D(0), ar, br)
		}, a, b, out)
}

// CovarMatrix generates the cells x cells square covariance matrix
// for all per-row cells of the given higher dimensional input tensor,
// which must have at least 2 dimensions: the outermost rows,
// and within that, 1+dimensional patterns (cells).
// Each value in the resulting matrix represents the extent to which the
// value of a given cell covaries across the rows of the tensor with the
// value of another cell.
// Uses the given metric function, typically [Covariance] or [Correlation],
// which must be registered in tensor Funcs, and can be passed as Metrics.String().
// Use Covariance if vars have similar overall scaling,
// which is typical in neural network models, and use
// Correlation if they are on very different scales, because it effectively rescales).
// The resulting matrix can be used as the input to PCA or SVD eigenvalue decomposition.
func CovarMatrix(funcName string, in, out *tensor.Indexed) {
	rows, cells := in.RowCellSize()
	if rows == 0 || cells == 0 {
		return
	}
	mout := tensor.NewFloatScalar(0.0)
	out.Tensor.SetShape(cells, cells)
	av := tensor.NewIndexed(tensor.NewFloat64(rows))
	bv := tensor.NewIndexed(tensor.NewFloat64(rows))

	coords := TriangularLIndicies(cells)
	nc := len(coords)
	// note: flops estimating 3 per item on average -- different for different metrics.
	tensor.VectorizeThreaded(rows*3, func(tsr ...*tensor.Indexed) int { return nc },
		func(idx int, tsr ...*tensor.Indexed) {
			// c := coords[idx]
			// todo: vectorize: only get new data if diff!
			for ai := 0; ai < cells; ai++ {
				// todo: extract data from given cell into av
				// TableColumnRowsVec(av, ix, col, ai)
				for bi := 0; bi <= ai; bi++ { // lower diag
					// TableColumnRowsVec(bv, ix, col, bi)
					// likewise
					tensor.Call(funcName, av, bv, mout)
					tsr[1].SetFloat(mout.Tensor.Float1D(0), ai, bi)
				}
			}
		}, in, out)
	for _, c := range coords { // copy to upper
		if c.X == c.Y { // exclude diag
			continue
		}
		out.Tensor.SetFloat(out.Tensor.Float(c.X, c.Y), c.Y, c.X)
	}
	// if nm, has := ix.Table.MetaData["name"]; has {
	// 	cmat.SetMetaData("name", nm+"_"+column)
	// } else {
	// 	cmat.SetMetaData("name", column)
	// }
	// if ds, has := ix.Table.MetaData["desc"]; has {
	// 	cmat.SetMetaData("desc", ds)
	// }
	// return nil
}

////////////////////////////////////////////
// 	Triangular square matrix functions

// TODO: move these somewhere more appropriate

// note: this answer gives an index into the upper triangular
// https://math.stackexchange.com/questions/2134011/conversion-of-upper-triangle-linear-index-from-index-on-symmetrical-array
// return TriangularN(n) - ((n-c)*(n-c-1))/2 + r - c - 1 <- this works for lower excluding diag
// return (n * (n - 1) / 2) - ((n-r)*(n-r-1))/2 + c <- this works for upper including diag
// but I wasn't able to get an equation for r, c back from index, for this "including diagonal"
// https://stackoverflow.com/questions/27086195/linear-index-upper-triangular-matrix?rq=3
// python just iterates manually and returns a list
// https://github.com/numpy/numpy/blob/v2.1.0/numpy/lib/_twodim_base_impl.py#L902-L985

// TriangularN returns the number of elements in the triangular region
// of a square matrix of given size, where the triangle includes the
// n elements along the diagonal.
func TriangularN(n int) int {
	return n + (n*(n-1))/2
}

// TriangularLIndicies returns the list of r, c indexes (as X, Y coordinates)
// for the lower triangular portion of a square matrix of size n,
// including the diagonal.
func TriangularLIndicies(n int) []vecint.Vector2i {
	trin := TriangularN(n)
	coords := make([]vecint.Vector2i, trin)
	i := 0
	for r := range n {
		for c := range n {
			if c <= r {
				coords[i] = vecint.Vector2i{X: r, Y: c}
				i++
			}
		}
	}
	return coords
}
