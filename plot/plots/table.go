// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

// Table is an interface for tabular data for plotting,
// with columns of values.
type Table interface {
	// number of columns of data
	NumColumns() int

	//	name of given column
	ColumnName(i int) string

	//	number of rows of data
	NumRows() int

	// PlotData returns the data value at given column and row
	PlotData(column, row int) float32
}

func TableColumnByIndex(tab Table, name string) int {
	for i := range tab.NumColumns() {
		if tab.ColumnName(i) == name {
			return i
		}
	}
	return -1
}

// // TableXYer is an interface for providing XY access to Table data
// type TableXYer struct {
// 	Table Table
//
// 	// the indexes of the tensor columns to use for the X and Y data, respectively
// 	XColumn, YColumn int
// }
//
// func NewTableXYer(tab Table, xcolumn, ycolumn int) *TableXYer {
// 	txy := &TableXYer{Table: tab, XColumn: xcolumn, YColumn: ycolumn}
// 	return txy
// }
//
// func (dt *TableXYer) Len() int {
// 	return dt.Table.NumRows()
// }
//
// func (dt *TableXYer) XY(i int) (x, y float32) {
// 	return dt.Table.PlotData(dt.XColumn, i), dt.Table.PlotData(dt.YColumn, i)
// }
//
// // AddTableLine adds XY Line with given x, y columns from given tabular data.
// func AddTableLine(plt *plot.Plot, tab Table, xcolumn, ycolumn int) *XY {
// 	txy := NewTableXYer(tab, xcolumn, ycolumn)
// 	ln := NewLine(txy)
// 	if ln == nil {
// 		return nil
// 	}
// 	plt.Add(ln)
// 	return ln
// }
