// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"strconv"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/tensor"
	"cogentcore.org/core/tensor/datafs"
	"cogentcore.org/core/tensor/table"
)

// All returns a single "split" with all of the rows in given view
// useful for leveraging the aggregation management functions in splits
// func All(ix *table.Table) *table.Splits {
// 	spl := &table.Splits{}
// 	spl.Levels = []string{"All"}
// 	spl.New(ix.Table, []string{"All"}, ix.Indexes...)
// 	return spl
// }

// Groups generates indexes for each unique value in each of the given tensors.
// One can then use the resulting indexes for the [tensor.Indexed] indexes to
// perform computations restricted to grouped subsets of data, as in the
// [GroupStats] function. See [GroupCombined] for function that makes a
// "Combined" Group that has a unique group for each _combination_ of
// the separate, independent groups created by this function.
// It creates subdirectories in a "Groups" directory within given [datafs],
// for each tensor passed in here, using the metadata Name property for
// names (index if empty).
// Within each subdirectory there are int tensors for each unique 1D
// row-wise value of elements in the input tensor, named as the string
// representation of the value, where the int tensor contains a list of
// row-wise indexes corresponding to the source rows having that value.
// Note that these indexes are directly in terms of the underlying [Tensor] data
// rows, indirected through any existing indexes on the inputs, so that
// the results can be used directly as Indexes into the corresponding tensor data.
// Uses a stable sort on columns, so ordering of other dimensions is preserved.
func Groups(dir *datafs.Data, tsrs ...*tensor.Indexed) {
	gd, err := dir.RecycleDir("Groups")
	if errors.Log(err) != nil {
		return
	}
	makeIdxs := func(dir *datafs.Data, srt *tensor.Indexed, val string, start, r int) {
		n := r - start
		it := datafs.NewValue[int](dir, val, n)
		for j := range n {
			it.SetIntRow(srt.Indexes[start+j], j) // key to indirect through sort indexes
		}
	}

	for i, tsr := range tsrs {
		nr := tsr.NumRows()
		if nr == 0 {
			continue
		}
		nm := tsr.Tensor.Metadata().GetName()
		if nm == "" {
			nm = strconv.Itoa(i)
		}
		td, _ := gd.Mkdir(nm)
		srt := tsr.CloneIndexes()
		srt.SortStable(tensor.Ascending)
		start := 0
		if tsr.Tensor.IsString() {
			lastVal := srt.StringRow(0)
			for r := range nr {
				v := srt.StringRow(r)
				if v != lastVal {
					makeIdxs(td, srt, lastVal, start, r)
					start = r
					lastVal = v
				}
			}
			if start != nr-1 {
				makeIdxs(td, srt, lastVal, start, nr)
			}
		} else {
			lastVal := srt.FloatRow(0)
			for r := range nr {
				v := srt.FloatRow(r)
				if v != lastVal {
					makeIdxs(td, srt, tensor.Float64ToString(lastVal), start, r)
					start = r
					lastVal = v
				}
			}
			if start != nr-1 {
				makeIdxs(td, srt, tensor.Float64ToString(lastVal), start, nr)
			}
		}
	}
}

// TableGroups runs [Groups] on the given columns from given [table.Table].
func TableGroups(dir *datafs.Data, dt *table.Table, columns ...string) {
	dv := table.NewView(dt)
	// important for consistency across columns, to do full outer product sort first.
	dv.SortColumns(tensor.Ascending, tensor.Stable, columns...)
	Groups(dir, dv.ColumnList(columns...)...)
}

// todo: GroupCombined

// GroupStats computes the given stats function on the unique grouped indexes
// produced by the [Groups] function, in the given [datafs] directory,
// applied to each of the tensors passed here.
// It creates a "Stats" subdirectory in given directory, with
// subdirectories with the name of each value tensor (if it does not
// yet exist), and then creates a subdirectory within that
// for the statistic name.  Within that statistic directory, it creates
// a String tensor with the unique values of each source [Groups] tensor,
// and a aligned Float64 tensor with the statistics results for each such
// unique group value. See the README.md file for a diagram of the results.
func GroupStats(dir *datafs.Data, stat string, tsrs ...*tensor.Indexed) {
	gd, err := dir.RecycleDir("Groups")
	if errors.Log(err) != nil {
		return
	}
	sd, err := dir.RecycleDir("Stats")
	if errors.Log(err) != nil {
		return
	}
	stnm := StripPackage(stat)
	spl := strings.Split(stat, ".")
	if len(spl) == 2 {
		stnm = spl[1]
	}
	stout := tensor.NewFloat64Scalar(0)
	groups := gd.ItemsFunc(nil)
	for _, gp := range groups {
		gpnm := gp.Name()
		ggd, _ := gd.RecycleDir(gpnm)
		vals := ggd.ValuesFunc(nil)
		nv := len(vals)
		if nv == 0 {
			continue
		}
		sgd, _ := sd.RecycleDir(gpnm)
		gv := sgd.Item(gpnm)
		if gv == nil {
			gtsr := datafs.NewValue[string](sgd, gpnm, nv)
			for i, v := range vals {
				gtsr.SetStringRow(v.Tensor.Metadata().GetName(), i)
			}
		}
		for _, tsr := range tsrs {
			vd, _ := sgd.RecycleDir(tsr.Tensor.Metadata().GetName())
			sv := datafs.NewValue[float64](vd, stnm, nv)
			for i, v := range vals {
				idx := v.Tensor.(*tensor.Int).Values
				sg := tensor.NewIndexed(tsr.Tensor, idx)
				tensor.Call(stat, sg, stout)
				sv.SetFloatRow(stout.Float1D(0), i)
			}
		}
	}
}

// TableGroupStats runs [GroupStats] on the given columns from given [table.Table].
func TableGroupStats(dir *datafs.Data, stat string, dt *table.Table, columns ...string) {
	GroupStats(dir, stat, dt.ColumnList(columns...)...)
}

// GroupDescribe runs standard descriptive statistics on given tensor data
// using [GroupStats] function, with [DescriptiveStats] list of stats.
func GroupDescribe(dir *datafs.Data, tsrs ...*tensor.Indexed) {
	for _, st := range DescriptiveStats {
		GroupStats(dir, st.FuncName(), tsrs...)
	}
}

// TableGroupDescribe runs [GroupDescribe] on the given columns from given [table.Table].
func TableGroupDescribe(dir *datafs.Data, dt *table.Table, columns ...string) {
	GroupDescribe(dir, dt.ColumnList(columns...)...)
}