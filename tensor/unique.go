// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import "math"

// UniqueGroups returns the distinct values of the first cell of each row
// of the given tensor (e.g., a table column), sorted in ascending order,
// as a 1D tensor of its type, along with the row indexes that have each
// value, in the order of the rows. The row indexes are those of the
// underlying data, through any indexes of a [Rows] view. NaN values are
// missing data, which are left out. This is the basis of [Unique] and of
// grouping by value (e.g., stats.Groups).
func UniqueGroups(tsr Tensor) (Values, [][]int) {
	nr := tsr.DimSize(0)
	if tsr.NumDims() == 0 || nr == 0 {
		return NewOfType(tsr.DataType(), 0), nil
	}
	srt := AsRows(tsr).CloneIndexes()
	srt.SortStable(Ascending)
	var groups [][]int
	var starts []int // the sorted row of the first value of each group
	str := tsr.IsString()
	for r := range nr {
		if !str && math.IsNaN(srt.FloatRow(r, 0)) {
			continue
		}
		n := len(starts)
		same := false
		if n > 0 {
			if str {
				same = srt.StringRow(r, 0) == srt.StringRow(starts[n-1], 0)
			} else {
				same = srt.FloatRow(r, 0) == srt.FloatRow(starts[n-1], 0)
			}
		}
		if !same {
			starts = append(starts, r)
			groups = append(groups, nil)
		}
		groups[len(groups)-1] = append(groups[len(groups)-1], srt.Indexes[r])
	}
	vals := NewOfType(tsr.DataType(), len(starts))
	for i, r := range starts {
		if str {
			vals.SetString1D(srt.StringRow(r, 0), i)
		} else {
			vals.SetFloat1D(srt.FloatRow(r, 0), i)
		}
	}
	return vals, groups
}

// Unique returns the distinct values of the given tensor, sorted in
// ascending order, as a 1D tensor of its type (as unique does in NumPy).
// All of the values of a higher-dimensional tensor are used, and NaN
// values are missing data, which are left out.
func Unique(tsr Tensor) Values {
	vals, _ := UniqueGroups(As1D(tsr))
	return vals
}

// UniqueCounts returns the number of times that each of the distinct
// values of the given tensor occurs, in the order of [Unique] (as the
// counts of unique_counts in NumPy).
func UniqueCounts(tsr Tensor) *Int {
	_, groups := UniqueGroups(As1D(tsr))
	counts := NewInt(len(groups))
	for i, g := range groups {
		counts.Values[i] = len(g)
	}
	return counts
}
