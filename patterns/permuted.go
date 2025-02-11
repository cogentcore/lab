// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patterns

import (
	"fmt"
	"math"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/lab/base/randx"
	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/tensor"
)

// NFromPct returns the number of bits for given pct (proportion 0-1),
// relative to total n: just int(math.Round(pct * n))
func NFromPct(pct float64, n int) int {
	return int(math.Round(pct * float64(n)))
}

// PermutedBinary sets the given tensor to contain nOn onVal values and the
// remainder are offVal values, using a permuted order of tensor elements (i.e.,
// randomly shuffled or permuted).
func PermutedBinary(tsr tensor.Values, nOn int, onVal, offVal float64) {
	ln := tsr.Len()
	if ln == 0 {
		return
	}
	pord := RandSource.Perm(ln)
	for i := range ln {
		if i < nOn {
			tsr.SetFloat1D(onVal, pord[i])
		} else {
			tsr.SetFloat1D(offVal, pord[i])
		}
	}
}

// PermutedBinaryRows uses the [tensor.RowMajor] view of a tensor as a column of rows
// as in a [table.Table], setting each row to contain nOn onVal values with the
// remainder being offVal values, using a permuted order of tensor elements
// (i.e., randomly shuffled or permuted). See also [PermutedBinaryMinDiff].
func PermutedBinaryRows(tsr tensor.Values, nOn int, onVal, offVal float64) {
	rows, cells := tsr.Shape().RowCellSize()
	if rows == 0 || cells == 0 {
		return
	}
	pord := RandSource.Perm(cells)
	for rw := range rows {
		stidx := rw * cells
		for i := 0; i < cells; i++ {
			if i < nOn {
				tsr.SetFloat1D(onVal, stidx+pord[i])
			} else {
				tsr.SetFloat1D(offVal, stidx+pord[i])
			}
		}
		randx.PermuteInts(pord, RandSource)
	}
}

// MinDiffPrintIterations set this to true to see the iteration stats for
// PermutedBinaryMinDiff -- for large, long-running cases.
var MinDiffPrintIterations = false

// PermutedBinaryMinDiff uses the [tensor.RowMajor] view of a tensor as a column of rows
// as in a [table.Table], setting each row to contain nOn onVal values, with the
// remainder being offVal values, using a permuted order of tensor elements
// (i.e., randomly shuffled or permuted). This version (see also [PermutedBinaryRows])
// ensures that all patterns have at least a given minimum distance
// from each other, expressed using minDiff = number of bits that must be different
// (can't be > nOn). If the mindiff constraint cannot be met within 100 iterations,
// an error is returned and automatically logged.
func PermutedBinaryMinDiff(tsr tensor.Values, nOn int, onVal, offVal float64, minDiff int) error {
	rows, cells := tsr.Shape().RowCellSize()
	if rows == 0 || cells == 0 {
		return errors.New("empty tensor")
	}
	pord := RandSource.Perm(cells)
	iters := 100
	nunder := make([]int, rows) // per row
	fails := 0
	for itr := range iters {
		for rw := range rows {
			if itr > 0 && nunder[rw] == 0 {
				continue
			}
			stidx := rw * cells
			for i := range cells {
				if i < nOn {
					tsr.SetFloat1D(onVal, stidx+pord[i])
				} else {
					tsr.SetFloat1D(offVal, stidx+pord[i])
				}
			}
			randx.PermuteInts(pord, RandSource)
		}
		for i := range nunder {
			nunder[i] = 0
		}
		nbad := 0
		mxnun := 0
		for r1 := range rows {
			r1v := tsr.SubSpace(r1)
			for r2 := r1 + 1; r2 < rows; r2++ {
				r2v := tsr.SubSpace(r2)
				dst := metric.Hamming(tensor.As1D(r1v), tensor.As1D(r2v)).Float1D(0)
				df := int(math.Round(float64(.5 * dst)))
				if df < minDiff {
					nunder[r1]++
					mxnun = max(nunder[r1])
					nunder[r2]++
					mxnun = max(nunder[r2])
					nbad++
				}
			}
		}
		if nbad == 0 {
			break
		}
		fails++
		if MinDiffPrintIterations {
			fmt.Printf("PermutedBinaryMinDiff: Itr: %d  NBad: %d  MaxN: %d\n", itr, nbad, mxnun)
		}
	}
	if fails == iters {
		err := errors.Log(fmt.Errorf("PermutedBinaryMinDiff: minimum difference of: %d was not met: %d times, rows: %d", minDiff, fails, rows))
		return err
	}
	return nil
}
