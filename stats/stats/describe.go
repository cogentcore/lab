// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"strconv"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// DescriptiveStats are the standard descriptive stats used in Describe function.
// Cannot apply the final 3 sort-based stats to higher-dimensional data.
var DescriptiveStats = []Stats{StatCount, StatMean, StatStd, StatSem, StatMin, StatQ1, StatMedian, StatQ3, StatMax}

// Describe adds standard descriptive statistics for given tensor
// to the given [tensorfs] directory, adding a directory for each tensor
// and result tensor stats for each result.
// This is an easy way to provide a comprehensive description of data.
// The [DescriptiveStats] list is: [Count], [Mean], [Std], [Sem],
// [Min], [Q1], [Median], [Q3], [Max]
func Describe(dir *tensorfs.Node, tsrs ...tensor.Tensor) {
	dd := dir.Dir("Describe")
	for i, tsr := range tsrs {
		nr := tsr.DimSize(0)
		if nr == 0 {
			continue
		}
		nm := metadata.Name(tsr)
		if nm == "" {
			nm = strconv.Itoa(i)
		}
		td := dd.Dir(nm)
		for _, st := range DescriptiveStats {
			stnm := st.String()
			sv := tensorfs.Scalar[float64](td, stnm)
			stout := st.Call(tsr)
			sv.CopyFrom(stout)
		}
	}
}

// DescribeTable runs [Describe] on given columns in table.
func DescribeTable(dir *tensorfs.Node, dt *table.Table, columns ...string) {
	Describe(dir, dt.ColumnList(columns...)...)
}

// DescribeTableAll runs [Describe] on all numeric columns in given table.
func DescribeTableAll(dir *tensorfs.Node, dt *table.Table) {
	var cols []string
	for i, cl := range dt.Columns.Values {
		if !cl.IsString() {
			cols = append(cols, dt.ColumnName(i))
		}
	}
	Describe(dir, dt.ColumnList(cols...)...)
}
