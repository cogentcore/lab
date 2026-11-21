// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// Groups generates indexes for each unique value in each of the given tensors.
// One can then use the resulting indexes for the [tensor.Rows] indexes to
// perform computations restricted to grouped subsets of data, as in the
// [GroupStats] function. See [GroupCombined] for function that makes a
// "Combined" Group that has a unique group for each _combination_ of
// the separate, independent groups created by this function.
// It creates subdirectories in a "Groups" directory within given [tensorfs],
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
func Groups(dir *tensorfs.Node, tsrs ...tensor.Tensor) error {
	gd := dir.Dir("Groups")
	for i, tsr := range tsrs {
		if tsr.DimSize(0) == 0 {
			continue
		}
		nm := metadata.Name(tsr)
		if nm == "" {
			nm = strconv.Itoa(i)
		}
		td := gd.Dir(nm)
		vals, groups := tensor.UniqueGroups(tsr)
		for gi, rows := range groups {
			it := tensorfs.Value[int](td, groupValueName(vals, gi), len(rows))
			for j, r := range rows {
				it.SetIntRow(r, j, 0)
			}
		}
	}
	return nil
}

// TableGroups runs [Groups] on the given columns from given [table.Table].
func TableGroups(dir *tensorfs.Node, dt *table.Table, columns ...string) error {
	dv := table.NewView(dt)
	// important for consistency across columns, to do full outer product sort first.
	dv.SortColumns(tensor.Ascending, tensor.StableSort, columns...)
	return Groups(dir, dv.ColumnList(columns...)...)
}

// GroupAll copies all indexes from the first given tensor,
// into an "All/All" tensor in the given [tensorfs], which can then
// be used with [GroupStats] to generate summary statistics across
// all the data. See [Groups] for more general documentation.
func GroupAll(dir *tensorfs.Node, tsrs ...tensor.Tensor) error {
	gd := dir.Dir("Groups")
	tsr := tensor.AsRows(tsrs[0])
	nr := tsr.NumRows()
	if nr == 0 {
		return nil
	}
	td := gd.Dir("All")
	it := tensorfs.Value[int](td, "All", nr)
	for j := range nr {
		it.SetIntRow(tsr.RowIndex(j), j, 0) // key to indirect through any existing indexes
	}
	return nil
}

// GroupCombined generates indexes for each unique combination of the
// values of the given tensors, as [Groups] does for each tensor separately,
// i.e., grouping by all of the tensors jointly (as groupby with a list of
// columns does in pandas), for the crosstabs and pivot tables of
// [GroupStats] results (see [tensor.Pivot]). The tensors must have the same
// rows, such as the columns of a table, and rows with a NaN value in any of
// them are left out (missing data). It creates one subdirectory in the
// "Groups" directory, named as the names of the tensors joined with "_",
// with an int tensor for each combination of values, in sorted order, named
// as the values joined with "_". [GroupStats] then makes a String column for
// each of the tensors, with its value of each combination, followed by the
// statistics of each combination.
func GroupCombined(dir *tensorfs.Node, tsrs ...tensor.Tensor) error {
	if len(tsrs) == 0 || tsrs[0].DimSize(0) == 0 {
		return nil
	}
	nt := len(tsrs)
	names := make([]string, nt)
	vals := make([]tensor.Values, nt)
	rowGroup := make([]map[int]int, nt) // the group of each value of each tensor, by its underlying row
	for i, tsr := range tsrs {
		names[i] = metadata.Name(tsr)
		if names[i] == "" {
			names[i] = strconv.Itoa(i)
		}
		var groups [][]int
		vals[i], groups = tensor.UniqueGroups(tsr)
		rowGroup[i] = map[int]int{}
		for gi, rows := range groups {
			for _, r := range rows {
				rowGroup[i][r] = gi
			}
		}
	}
	combos := map[string][]int{} // the rows of each combination, by its key of group indexes
	var keys [][]int
	rws := tensor.AsRows(tsrs[0])
	key := make([]int, nt)
	for ri := range rws.NumRows() {
		r := rws.RowIndex(ri)
		missing := false
		for i := range nt {
			gi, ok := rowGroup[i][r]
			if !ok {
				missing = true
				break
			}
			key[i] = gi
		}
		if missing {
			continue
		}
		ks := fmt.Sprint(key)
		if _, has := combos[ks]; !has {
			keys = append(keys, slices.Clone(key))
		}
		combos[ks] = append(combos[ks], r)
	}
	slices.SortFunc(keys, slices.Compare)
	td := dir.Dir("Groups").Dir(strings.Join(names, "_"))
	cvals := make([]string, nt)
	for _, k := range keys {
		for i, gi := range k {
			cvals[i] = groupValueName(vals[i], gi)
		}
		rows := combos[fmt.Sprint(k)]
		it := tensorfs.Value[int](td, strings.Join(cvals, "_"), len(rows))
		for j, r := range rows {
			it.SetIntRow(r, j, 0)
		}
		metadata.Set(it, "GroupNames", slices.Clone(names))
		metadata.Set(it, "GroupValues", slices.Clone(cvals))
	}
	return nil
}

// TableGroupCombined runs [GroupCombined] on the given columns from the
// given [table.Table].
func TableGroupCombined(dir *tensorfs.Node, dt *table.Table, columns ...string) error {
	return GroupCombined(dir, dt.ColumnList(columns...)...)
}

// groupValueName returns the name of the given value of the given unique
// values of a group: the string, or the number formatted as a string.
func groupValueName(vals tensor.Values, i int) string {
	if vals.IsString() {
		return vals.String1D(i)
	}
	return tensor.Float64ToString(vals.Float1D(i))
}

// GroupStats computes the given stats function on the unique grouped indexes
// produced by the [Groups] function, in the given [tensorfs] directory,
// applied to each of the tensors passed here.
// It creates a "Stats" subdirectory in given directory, with a
// subdirectory for each source [Groups] tensor, with subdirectories
// with the name of each value tensor (if they do not yet exist).
// Within each value directory, it creates a String tensor with the
// unique values of the source [Groups] tensor, named as it, as the
// first column, and an aligned Float64 tensor with the statistics results
// for each such unique group value, named as the statistic, so that each
// value directory is a table of its statistics for each group, which can
// be read and plotted directly. See the README.md file for a diagram of
// the results, and [GroupStatsDirTable] for a table of all the values.
// The stats are computed on float64 values, converting other types of data.
func GroupStats(dir *tensorfs.Node, stat Stats, tsrs ...tensor.Tensor) error {
	gd := dir.Dir("Groups")
	sd := dir.Dir("Stats")
	stnm := StripPackage(stat.String())
	groups, _ := gd.Nodes()
	for _, gp := range groups {
		gpnm := gp.Name()
		ggd := gd.Dir(gpnm)
		vals := ggd.ValuesFunc(nil)
		nv := len(vals)
		if nv == 0 {
			continue
		}
		sgd := sd.Dir(gpnm)
		for _, tsr := range tsrs {
			vd := sgd.Dir(metadata.Name(tsr))
			addGroupLabels(vd, gpnm, vals)
			sv := tensorfs.Value[float64](vd, stnm, nv)
			ftsr := tensor.AsFloat64(tsr.AsValues()) // not truncated for int data
			for i, v := range vals {
				idx := tensor.AsIntSlice(v)
				sg := tensor.NewRows(ftsr, idx...)
				stout := stat.Call(sg)
				sv.SetFloatRow(stout.Float1D(0), i, 0)
			}
		}
	}
	return nil
}

// addGroupLabels adds the String tensors of the values of the given groups
// to the given directory of the statistics of a value, if not there yet:
// one named as the group, or, for the groups of [GroupCombined], one for
// each of its tensors, named as it.
func addGroupLabels(vd *tensorfs.Node, gpnm string, vals []tensor.Tensor) {
	nv := len(vals)
	names, err := metadata.Get[[]string](vals[0], "GroupNames")
	if err != nil {
		if vd.Node(gpnm) == nil {
			gtsr := tensorfs.Value[string](vd, gpnm, nv)
			for i, v := range vals {
				gtsr.SetStringRow(metadata.Name(v), i, 0)
			}
			metadata.Set(gtsr, "GroupLabel", true)
		}
		return
	}
	for li, lnm := range names {
		if vd.Node(lnm) != nil {
			continue
		}
		gtsr := tensorfs.Value[string](vd, lnm, nv)
		for i, v := range vals {
			if cv, err := metadata.Get[[]string](v, "GroupValues"); err == nil && li < len(cv) {
				gtsr.SetStringRow(cv[li], i, 0)
			}
		}
		metadata.Set(gtsr, "GroupLabel", true)
	}
}

// TableGroupStats runs [GroupStats] using standard [Stats]
// on the given columns from given [table.Table].
func TableGroupStats(dir *tensorfs.Node, stat Stats, dt *table.Table, columns ...string) error {
	return GroupStats(dir, stat, dt.ColumnList(columns...)...)
}

// GroupDescribe runs standard descriptive statistics on given tensor data
// using [GroupStats] function, with [DescriptiveStats] list of stats.
func GroupDescribe(dir *tensorfs.Node, tsrs ...tensor.Tensor) error {
	for _, st := range DescriptiveStats {
		err := GroupStats(dir, st, tsrs...)
		if err != nil {
			return err
		}
	}
	return nil
}

// TableGroupDescribe runs [GroupDescribe] on the given columns from given [table.Table].
func TableGroupDescribe(dir *tensorfs.Node, dt *table.Table, columns ...string) error {
	return GroupDescribe(dir, dt.ColumnList(columns...)...)
}

// GroupStatsAsTable returns the results from [GroupStats] in given directory
// as a [table.Table], with the columns of [GroupStatsDirTable] for each group.
func GroupStatsAsTable(dir *tensorfs.Node) *table.Table {
	dt := table.New("Stats")
	gds, _ := dir.Dir("Stats").Nodes()
	for _, gd := range gds {
		addGroupStats(dt, gd)
	}
	return dt
}

// GroupStatsDirTable returns the results from [GroupStats] for one group,
// in the given directory of the group within the "Stats" directory, as a
// [table.Table], with the group values as the first column, named as the
// group, followed by a Value/Stat column for each statistic of each value.
func GroupStatsDirTable(gdir *tensorfs.Node) *table.Table {
	dt := table.New(gdir.Name())
	addGroupStats(dt, gdir)
	return dt
}

// addGroupStats adds the columns of [GroupStatsDirTable] for the given
// group directory to the given table.
func addGroupStats(dt *table.Table, gdir *tensorfs.Node) {
	gpnm := gdir.Name()
	vds, _ := gdir.Nodes()
	for _, vd := range vds {
		if !vd.IsDir() {
			continue
		}
		sts, _ := vd.Nodes()
		for _, st := range sts {
			if st.IsDir() || st.Tensor == nil || st.Tensor.NumDims() == 0 {
				continue
			}
			nm := vd.Name() + "/" + st.Name()
			if isLabel, _ := metadata.Get[bool](st.Tensor, "GroupLabel"); isLabel || st.Name() == gpnm {
				if dt.Columns.IndexByKey(st.Name()) >= 0 {
					continue // only once
				}
				nm = st.Name()
			}
			if rows := st.Tensor.DimSize(0); dt.Columns.Rows < rows {
				dt.SetNumRows(rows)
			}
			dt.AddColumn(nm, st.Tensor.AsValues())
		}
	}
}

// GroupStatsAsTableNoStatName returns the results from [GroupStats]
// in given directory as a [table.Table], as [GroupStatsAsTable] does.
// Column names are updated to not include the stat name, if there is only
// one statistic such that the resulting name will still be unique.
// Otherwise, column names are Value/Stat.
func GroupStatsAsTableNoStatName(dir *tensorfs.Node) *table.Table {
	dt := GroupStatsAsTable(dir)
	cols := make(map[string]string)
	for _, nm := range dt.Columns.Keys {
		vn := nm
		si := strings.Index(nm, "/")
		if si > 0 {
			vn = nm[:si]
		}
		if _, exists := cols[vn]; exists {
			continue
		}
		cols[vn] = nm
	}
	for k, v := range cols {
		ci := dt.Columns.IndexByKey(v)
		dt.Columns.RenameIndex(ci, k)
	}
	return dt
}
