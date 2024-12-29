// Copyright (c) 2019, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patterns

import (
	"fmt"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/lab/tensor"
)

// Mix mixes patterns from different tensors into a combined set of patterns,
// over the outermost row dimension (i.e., each source is a list of patterns over rows).
// The source tensors must have the same cell size, and the existing shape of the destination
// will be used if compatible, otherwise reshaped with linear list of sub-tensors.
// Each source list wraps around if shorter than the total number of rows specified.
func Mix(dest tensor.Values, rows int, srcs ...tensor.Values) error {
	var cells int
	for i, src := range srcs {
		_, c := src.Shape().RowCellSize()
		if i == 0 {
			cells = c
		} else {
			if c != cells {
				err := errors.Log(fmt.Errorf("MixPatterns: cells size of source number %d, %d != first source: %d", i, c, cells))
				return err
			}
		}
	}

	totlen := len(srcs) * cells * rows
	if dest.Len() != totlen {
		_, dcells := dest.Shape().RowCellSize()
		if dcells == cells*len(srcs) {
			dest.SetNumRows(rows)
		} else {
			sz := append([]int{rows}, len(srcs), cells)
			dest.SetShapeSizes(sz...)
		}
	}

	dtype := dest.DataType()
	for i, src := range srcs {
		si := i * cells
		srows := src.DimSize(0)
		for row := range rows {
			srow := row % srows
			for ci := range cells {
				switch {
				case reflectx.KindIsFloat(dtype):
					dest.SetFloatRow(src.FloatRow(srow, ci), row, si+ci)
				}
			}
		}
	}
	return nil
}
