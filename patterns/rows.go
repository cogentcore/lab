// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patterns

//go:generate core generate -add-types

import (
	"fmt"
	"strconv"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// NOnInTensor returns the number of bits active in given tensor
func NOnInTensor(trow tensor.Values) int {
	return stats.Sum(trow).Int1D(0)
}

// PctActInTensor returns the percent activity in given tensor (NOn / size)
func PctActInTensor(trow tensor.Values) float32 {
	return float32(NOnInTensor(trow)) / float32(trow.Len())
}

// Note: AppendFrom can be used to concatenate tensors.

// NameRows sets strings as prefix + row number with given number
// of leading zeros.
func NameRows(tsr tensor.Values, prefix string, nzeros int) {
	ft := fmt.Sprintf("%s%%0%dd", prefix, nzeros)
	rows := tsr.DimSize(0)
	for i := range rows {
		tsr.SetString1D(fmt.Sprintf(ft, i), i)
	}
}

// Shuffle returns a [tensor.Rows] view of the given source tensor
// with the outer row-wise dimension randomly shuffled (permuted).
func Shuffle(src tensor.Values) *tensor.Rows {
	idx := RandSource.Perm(src.DimSize(0))
	return tensor.NewRows(src, idx...)
}

// ReplicateRows adds nCopies rows of the source tensor pattern into
// the destination tensor. The destination shape is set to ensure
// it can contain the results, preserving any existing rows of data.
func ReplicateRows(dest, src tensor.Values, nCopies int) {
	curRows := 0
	if dest.NumDims() > 0 {
		curRows = dest.DimSize(0)
	}
	totRows := curRows + nCopies
	dshp := append([]int{totRows}, src.Shape().Sizes...)
	dest.SetShapeSizes(dshp...)
	for rw := range nCopies {
		dest.SetRowTensor(src, curRows+rw)
	}
}

// SplitRows splits a source tensor into a set of tensors in the given
// tensorfs directory, with the given list of names, splitting at given
// rows. There should be 1 more name than rows. If names are omitted then
// the source name + incrementing counter will be used.
func SplitRows(dir *tensorfs.Node, src tensor.Values, names []string, rows ...int) error {
	hasNames := len(names) != 0
	if hasNames && len(names) != len(rows)+1 {
		err := errors.Log(fmt.Errorf("patterns.SplitRows: must pass one more name than number of rows to split on"))
		return err
	}
	all := append(rows, src.DimSize(0)) // final row
	srcName := metadata.Name(src)
	srcShape := src.ShapeSizes()

	dtype := src.DataType()
	prev := 0
	for i, cur := range all {
		if prev >= cur {
			err := errors.Log(fmt.Errorf("patterns.SplitRows: rows must increase progressively"))
			return err
		}
		name := ""
		switch {
		case hasNames:
			name = names[i]
		case len(srcName) > 0:
			name = fmt.Sprintf("%s_%d", srcName, i)
		default:
			name = strconv.Itoa(i)
		}
		nrows := cur - prev
		srcShape[0] = nrows
		spl := tensorfs.ValueType(dir, name, dtype, srcShape...)
		for rw := range nrows {
			spl.SubSpace(rw).CopyFrom(src.SubSpace(prev + rw))
		}
		prev = cur
	}
	return nil
}

// AddVocabDrift adds a row-by-row drifting pool to the vocabulary,
// starting from the given row in existing vocabulary item
// (which becomes starting row in this one -- drift starts in second row).
// The current row patterns are generated by taking the previous row
// pattern and flipping pctDrift percent of active bits (min of 1 bit).
// func AddVocabDrift(mp Vocab, name string, rows int, pctDrift float32, copyFrom string, copyRow int) (tensor.Values, error) {
// 	cp, err := mp.ByName(copyFrom)
// 	if err != nil {
// 		return nil, err
// 	}
// 	tsr := &tensor.Float32{}
// 	cpshp := cp.Shape().Sizes
// 	cpshp[0] = rows
// 	tsr.SetShapeSizes(cpshp...)
// 	mp[name] = tsr
// 	cprow := cp.SubSpace(copyRow).(tensor.Values)
// 	trow := tsr.SubSpace(0)
// 	trow.CopyFrom(cprow)
// 	nOn := NOnInTensor(cprow)
// 	rmdr := 0.0                               // remainder carryover in drift
// 	drift := float64(nOn) * float64(pctDrift) // precise fractional amount of drift
// 	for i := 1; i < rows; i++ {
// 		srow := tsr.SubSpace(i - 1)
// 		trow := tsr.SubSpace(i)
// 		trow.CopyFrom(srow)
// 		curDrift := math.Round(drift + rmdr) // integer amount
// 		nDrift := int(curDrift)
// 		if nDrift > 0 {
// 			FlipBits(trow, nDrift, nDrift, 1, 0)
// 		}
// 		rmdr += drift - curDrift // accumulate remainder
// 	}
// 	return tsr, nil
// }
