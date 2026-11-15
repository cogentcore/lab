// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package table

import (
	"math"
	"slices"
	"strings"
	"testing"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSortTable returns a table with a Name string column, a Val float
// column and a 2-cell Cell column, with 5 rows in a known order.
func newSortTable() *Table {
	dt := New("sort")
	dt.AddStringColumn("Name")
	dt.AddFloat64Column("Val")
	dt.AddIntColumn("Cell", 2)
	dt.SetNumRows(5)
	names := []string{"c", "a", "e", "b", "d"}
	vals := []float64{3, 1, 3, 2, 0}
	for i := range 5 {
		dt.Column("Name").SetStringRow(names[i], i, 0)
		dt.Column("Val").SetFloatRow(vals[i], i, 0)
		dt.Column("Cell").SetFloatRow(float64(10*i), i, 0)
		dt.Column("Cell").SetFloatRow(float64(10*i+1), i, 1)
	}
	return dt
}

// columnStrings returns the string values of the named column in the
// current indexed order.
func columnStrings(dt *Table, name string) []string {
	cl := dt.Column(name)
	s := make([]string, dt.NumRows())
	for i := range s {
		s[i] = cl.StringRow(i, 0)
	}
	return s
}

// TestSortColumns tests sorting by one and by multiple columns, both
// ascending and descending, stable and unstable, including strings.
func TestSortColumns(t *testing.T) {
	dt := newSortTable()
	require.NoError(t, dt.SortColumn("Name", tensor.Ascending))
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, columnStrings(dt, "Name"))
	assert.Error(t, dt.SortColumn("Nope", tensor.Ascending))

	dt.Sequential()
	dt.SortColumns(tensor.Descending, tensor.StableSort, "Val", "Name")
	assert.Equal(t, []string{"e", "c", "b", "a", "d"}, columnStrings(dt, "Name"))

	dt.Sequential()
	dt.SortColumns(tensor.Ascending, false, "Val", "Name")
	assert.Equal(t, []string{"d", "a", "b", "c", "e"}, columnStrings(dt, "Name"))
	// the cells follow their rows.
	assert.Equal(t, 41.0, dt.Column("Cell").FloatRow(0, 1))

	dt.SortStableFunc(func(dt *Table, i, j int) int {
		return strings.Compare(dt.Columns.At("Name").String1D(j), dt.Columns.At("Name").String1D(i))
	})
	assert.Equal(t, []string{"e", "d", "c", "b", "a"}, columnStrings(dt, "Name"))

	dt.SortIndexes()
	assert.Equal(t, []int{0, 1, 2, 3, 4}, dt.Indexes)
	dt.Sequential()
	dt.SortIndexes() // nil: no-op
	assert.Nil(t, dt.Indexes)
}

// TestFilter tests Filter, FilterString, DeleteRows, Swap and New,
// which flattens the indexed view into a new table.
func TestFilter(t *testing.T) {
	dt := newSortTable()
	dt.Filter(func(dt *Table, row int) bool {
		return dt.Columns.At("Val").Float1D(row) >= 2
	})
	assert.Equal(t, []int{0, 2, 3}, dt.Indexes)
	assert.Equal(t, 3, dt.NumRows())
	assert.Equal(t, 2, dt.RowIndex(1))
	assert.NoError(t, dt.IsValidRow(2))
	assert.Error(t, dt.IsValidRow(3))
	assert.Error(t, dt.IsValidRow(-1))

	dt.Swap(0, 2)
	assert.Equal(t, []int{3, 2, 0}, dt.Indexes)

	nt := dt.New()
	assert.Nil(t, nt.Indexes)
	assert.Equal(t, 3, nt.Columns.Rows)
	assert.Equal(t, []string{"b", "e", "c"}, columnStrings(nt, "Name"))
	assert.Equal(t, 31.0, nt.Column("Cell").FloatRow(0, 1))
	assert.Equal(t, "sort", metadata.Name(nt))

	dt.DeleteRows(0, 2)
	assert.Equal(t, []int{0}, dt.Indexes)

	dt.Sequential()
	require.NoError(t, dt.FilterString("Name", "B", tensor.StringMatch{IgnoreCase: true}))
	assert.Equal(t, []string{"b"}, columnStrings(dt, "Name"))
	assert.Error(t, dt.FilterString("Nope", "x", tensor.StringMatch{}))

	// New with an empty view and with no view.
	dt.Filter(func(dt *Table, row int) bool { return false })
	assert.Equal(t, 0, dt.New().NumRows())
	dt.Sequential()
	assert.Equal(t, 5, dt.New().NumRows())
}

// TestFilterNA tests NAInRow and FilterNA in its any and all modes.
func TestFilterNA(t *testing.T) {
	dt := New()
	dt.AddStringColumn("S")
	dt.AddFloat64Column("A")
	dt.AddFloat32Column("B")
	dt.SetNumRows(3)
	nan := math.NaN()
	dt.Column("A").SetFloatRow(nan, 1, 0)
	dt.Column("A").SetFloatRow(nan, 2, 0)
	dt.Column("B").SetFloatRow(nan, 2, 0)

	n, tot := dt.NAInRow(2)
	assert.Equal(t, 2, n)
	assert.Equal(t, 2, tot) // string column is not counted

	dt.FilterNA(AllNA)
	assert.Equal(t, []int{0, 1}, dt.Indexes)
	dt.Sequential().FilterNA(AnyNA)
	assert.Equal(t, []int{0}, dt.Indexes)
}

// TestPermutedValid tests Permuted (with and without existing indexes),
// ValidIndexes and IndexesFromTensor.
func TestPermutedValid(t *testing.T) {
	dt := newSortTable()
	dt.Permuted()
	idx := slices.Clone(dt.Indexes)
	slices.Sort(idx)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, idx)
	dt.Permuted()
	assert.Len(t, dt.Indexes, 5)

	// shrink the underlying columns: invalid indexes are dropped.
	dt.Columns.SetNumRows(3)
	dt.ValidIndexes()
	idx = slices.Clone(dt.Indexes)
	slices.Sort(idx)
	assert.Equal(t, []int{0, 1, 2}, idx)

	dt.IndexesFromTensor(tensor.NewRows(dt.Columns.Values[0], 2, 1))
	assert.Equal(t, []int{2, 1}, dt.Indexes)

	empty := New()
	empty.AddFloat64Column("x")
	empty.Indexes = []int{}
	assert.Nil(t, empty.Permuted().Indexes)
	empty.Indexes = []int{}
	assert.Nil(t, empty.ValidIndexes().Indexes)
}

// TestInsertRowsFiltered tests InsertRows on a filtered view, where
// the view has fewer indexes than the table has rows.
func TestInsertRowsFiltered(t *testing.T) {
	dt := newSortTable()
	dt.Filter(func(dt *Table, row int) bool { return row%2 == 0 })
	assert.Equal(t, []int{0, 2, 4}, dt.Indexes)
	dt.InsertRows(1, 1)
	assert.Equal(t, []int{0, 5, 2, 4}, dt.Indexes)
	assert.Equal(t, 6, dt.Columns.Rows)
	dt.InsertRows(4, 2)
	assert.Equal(t, []int{0, 5, 2, 4, 6, 7}, dt.Indexes)
}
