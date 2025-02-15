// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patterns

import (
	"cogentcore.org/lab/base/randx"
	"cogentcore.org/lab/tensor"
)

// FlipBits turns nOff bits that are currently On to Off and
// nOn bits that are currently Off to On, using permuted lists.
func FlipBits(tsr tensor.Values, nOff, nOn int, onVal, offVal float64) {
	ln := tsr.Len()
	if ln == 0 {
		return
	}
	var ons, offs []int
	for i := range ln {
		vl := tsr.Float1D(i)
		if vl == offVal {
			offs = append(offs, i)
		} else {
			ons = append(ons, i)
		}
	}
	randx.PermuteInts(ons, RandSource)
	randx.PermuteInts(offs, RandSource)
	if nOff > len(ons) {
		nOff = len(ons)
	}
	if nOn > len(offs) {
		nOn = len(offs)
	}
	for i := range nOff {
		tsr.SetFloat1D(offVal, ons[i])
	}
	for i := range nOn {
		tsr.SetFloat1D(onVal, offs[i])
	}
}

// FlipBitsRows turns nOff bits that are currently On to Off and
// nOn bits that are currently Off to On, using permuted lists.
// Iterates over the outer-most tensor dimension as rows.
func FlipBitsRows(tsr tensor.Values, nOff, nOn int, onVal, offVal float64) {
	rows, _ := tsr.Shape().RowCellSize()
	for i := range rows {
		trow := tsr.SubSpace(i)
		FlipBits(trow, nOff, nOn, onVal, offVal)
	}
}
