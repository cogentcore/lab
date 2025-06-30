// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"reflect"

	"cogentcore.org/lab/table"
)

// MeanTables returns a [table.Table] with the mean values across all float
// columns of the input tables, which must have the same columns but not
// necessarily the same number of rows.
func MeanTables(dts []*table.Table) *table.Table {
	nt := len(dts)
	if nt == 0 {
		return nil
	}
	maxRows := 0
	var maxdt *table.Table
	for _, dt := range dts {
		nr := dt.NumRows()
		if nr > maxRows {
			maxRows = nr
			maxdt = dt
		}
	}
	if maxRows == 0 {
		return nil
	}
	ot := maxdt.Clone()

	// N samples per row
	rns := make([]int, maxRows)
	for _, dt := range dts {
		dnr := dt.NumRows()
		mx := min(dnr, maxRows)
		for ri := 0; ri < mx; ri++ {
			rns[ri]++
		}
	}
	for ci := range ot.Columns.Values {
		cl := ot.ColumnByIndex(ci)
		if cl.DataType() != reflect.Float32 && cl.DataType() != reflect.Float64 {
			continue
		}
		_, cells := cl.RowCellSize()
		for di, dt := range dts {
			if di == 0 {
				continue
			}
			dc := dt.ColumnByIndex(ci)
			dnr := dt.NumRows()
			mx := min(dnr, maxRows)
			for ri := 0; ri < mx; ri++ {
				for j := 0; j < cells; j++ {
					cv := cl.FloatRow(ri, j)
					cv += dc.FloatRow(ri, j)
					cl.SetFloatRow(cv, ri, j)
				}
			}
		}
		for ri := 0; ri < maxRows; ri++ {
			for j := 0; j < cells; j++ {
				cv := cl.FloatRow(ri, j)
				if rns[ri] > 0 {
					cv /= float64(rns[ri])
					cl.SetFloatRow(cv, ri, j)
				}
			}
		}
	}
	return ot
}
