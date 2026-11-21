// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import "math"

// Pivot reshapes values given in long form, one per row with the row and
// column labels of each, such as the columns of a table of the statistics of
// combined groups (stats.GroupCombined), into a matrix, as pivot does in
// pandas. The rows of the matrix are the sorted unique values of rows
// ([Unique]), and its columns those of cols, which are returned as the
// labels of the matrix. Each pair of row and column labels should occur
// once (the last value is used otherwise), and the elements of the matrix
// without a value are NaN (missing data).
func Pivot(rows, cols, values Tensor) (matrix *Float64, rowLabels, colLabels Values) {
	rowLabels = Unique(rows)
	colLabels = Unique(cols)
	nr, nc := rowLabels.Len(), colLabels.Len()
	matrix = NewFloat64(nr, nc)
	for i := range matrix.Values {
		matrix.Values[i] = math.NaN()
	}
	ri := labelIndexes(rowLabels)
	ci := labelIndexes(colLabels)
	n := min(rows.Len(), cols.Len(), values.Len())
	for i := range n {
		r, rok := ri[rows.String1D(i)]
		c, cok := ci[cols.String1D(i)]
		if rok && cok {
			matrix.Values[r*nc+c] = values.Float1D(i)
		}
	}
	return
}

// labelIndexes returns the index of each of the given labels by its string
// value.
func labelIndexes(labels Values) map[string]int {
	idx := make(map[string]int, labels.Len())
	for i := range labels.Len() {
		idx[labels.String1D(i)] = i
	}
	return idx
}
