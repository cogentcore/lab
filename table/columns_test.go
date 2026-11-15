// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package table

import (
	"reflect"
	"testing"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestColumnAccess tests the column lookup methods by name and index.
func TestColumnAccess(t *testing.T) {
	dt := NewTestTable()
	assert.Equal(t, 3, dt.NumColumns())
	assert.Nil(t, dt.Column("Nope"))
	_, err := dt.ColumnTry("Nope")
	assert.Error(t, err)
	c, err := dt.ColumnTry("Int")
	require.NoError(t, err)
	assert.Equal(t, 2.0, c.FloatRow(2, 0))

	assert.Equal(t, "Flt64", dt.ColumnName(1))
	assert.Equal(t, 1, dt.ColumnIndex("Flt64"))
	assert.Equal(t, -1, dt.ColumnIndex("Nope"))
	assert.Equal(t, []int{2, 0}, dt.ColumnIndexList("Int", "Nope", "Str"))
	list := dt.ColumnList("Int", "Nope", "Str")
	require.Len(t, list, 2)
	assert.Equal(t, "Int", metadata.Name(list[0].(*tensor.Rows).Tensor))
	assert.Equal(t, "Str", metadata.Name(dt.ColumnByIndex(0).Tensor))
}

// TestInsertDeleteColumns tests inserting columns at an index (with
// the generic and Columns methods) and deleting them by name and index.
func TestInsertDeleteColumns(t *testing.T) {
	dt := NewTestTable()
	b := InsertColumn[bool](dt, "Bool", 1)
	assert.Equal(t, []int{3}, b.ShapeSizes())
	assert.Equal(t, "Bool", dt.ColumnName(1))
	assert.Equal(t, "Bool", metadata.Name(b))

	f := tensor.NewFloat32(1, 2)
	require.NoError(t, dt.InsertColumn(0, "F", f))
	assert.Equal(t, []int{3, 2}, f.ShapeSizes())
	assert.Equal(t, "F", dt.ColumnName(0))
	// duplicate names are an error, not a panic.
	assert.Error(t, dt.InsertColumn(0, "F", tensor.NewFloat32(1)))
	assert.Error(t, dt.AddColumn("F", tensor.NewFloat32(1)))
	assert.Equal(t, 5, dt.NumColumns())

	assert.True(t, dt.DeleteColumnName("Bool"))
	assert.False(t, dt.DeleteColumnName("Bool"))
	dt.DeleteColumnByIndex(0, 1)
	assert.Equal(t, []string{"Str", "Flt64", "Int"}, dt.Columns.Keys)

	dt.DeleteAll()
	assert.Equal(t, 0, dt.NumColumns())
}

// TestCloneAndView tests that Clone copies the data and indexes, and
// that NewView shares the data with its own indexes.
func TestCloneAndView(t *testing.T) {
	dt := NewTestTable()
	metadata.SetName(dt, "orig")
	dt.Indexes = []int{2, 0}
	cp := dt.Clone()
	assert.Equal(t, []int{2, 0}, cp.Indexes)
	assert.Equal(t, "orig", metadata.Name(cp))
	cp.Column("Int").SetFloatRow(99, 0, 0)
	assert.Equal(t, 2.0, dt.Column("Int").FloatRow(0, 0))

	vw := NewView(dt)
	assert.Equal(t, []int{2, 0}, vw.Indexes)
	vw.Indexes[0] = 1
	assert.Equal(t, 2, dt.Indexes[0]) // indexes are copied
	vw.Column("Int").SetFloatRow(77, 0, 0)
	assert.Equal(t, 77.0, dt.Columns.At("Int").Float1D(1)) // data is shared
	assert.Nil(t, NewView(New()).Indexes)

	var it Table
	it.Init()
	assert.NotNil(t, it.Columns)
	assert.NotNil(t, it.Metadata())
}

// TestSetNumRowsToMax tests that SetNumRowsToMax pads columns that
// were grown outside of the table.
func TestSetNumRowsToMax(t *testing.T) {
	dt := NewTestTable()
	dt.Columns.At("Int").SetNumRows(7)
	dt.SetNumRowsToMax()
	assert.Equal(t, 7, dt.NumRows())
	assert.Equal(t, 7, dt.Columns.At("Str").DimSize(0))

	// with indexes: new rows are added to the view, removed rows dropped.
	dt.IndexesNeeded()
	dt.SetNumRows(9)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8}, dt.Indexes)
	dt.SetNumRows(2)
	assert.Equal(t, []int{0, 1}, dt.Indexes)
	dt.AddRows(1)
	assert.Equal(t, 3, dt.NumRows())
}

// TestAppendRowsIndexed tests AppendRows on a table with indexes, and
// with a source that is missing one of the columns.
func TestAppendRowsIndexed(t *testing.T) {
	dt := NewTestTable()
	dt.Indexes = []int{2, 1}
	src := New()
	src.AddIntColumn("Int")
	src.SetNumRows(2)
	src.Column("Int").SetFloatRow(5, 1, 0)
	dt.AppendRows(src)
	assert.Equal(t, 5, dt.Columns.Rows)
	assert.Equal(t, []int{2, 1, 3, 4}, dt.Indexes)
	assert.Equal(t, 5.0, dt.Column("Int").FloatRow(3, 0))
	assert.Equal(t, 5, dt.Columns.At("Str").DimSize(0))
}

// TestInsertKeyColumns tests that InsertKeyColumns adds constant
// columns at the start of a copy of the table.
func TestInsertKeyColumns(t *testing.T) {
	dt := NewTestTable()
	kt := dt.InsertKeyColumns("Run", "r1", "Cond", "A")
	assert.Equal(t, []string{"Cond", "Run", "Str", "Flt64", "Int"}, kt.Columns.Keys)
	assert.Equal(t, "r1", kt.Column("Run").StringRow(2, 0))
	assert.Equal(t, "A", kt.Column("Cond").StringRow(0, 0))
	assert.Equal(t, 3, dt.NumColumns()) // original unchanged

	assert.Same(t, dt, dt.InsertKeyColumns("odd")) // odd args: no-op
}

// TestConfigFromTable tests configuring columns from a Name, Type
// format table, including an unknown type error.
func TestConfigFromTable(t *testing.T) {
	ft := New()
	ft.AddStringColumn("Name")
	ft.AddStringColumn("Type")
	types := [][2]string{{"s", "string"}, {"b", "Bool"}, {"f32", "float32"},
		{"f64", "float64"}, {"i", "int"}, {"i32", "int32"}, {"u8", "byte"}, {"x", "complex"}}
	ft.SetNumRows(len(types))
	for i, ty := range types {
		ft.Column("Name").SetStringRow(ty[0], i, 0)
		ft.Column("Type").SetStringRow(ty[1], i, 0)
	}
	dt := New()
	err := dt.ConfigFromTable(ft)
	assert.ErrorContains(t, err, "complex")
	kinds := []reflect.Kind{reflect.String, reflect.Bool, reflect.Float32,
		reflect.Float64, reflect.Int, reflect.Int32, reflect.Uint8, reflect.Float64}
	require.Equal(t, len(kinds), dt.NumColumns())
	for i, k := range kinds {
		assert.Equal(t, k, dt.Columns.Values[i].DataType(), dt.ColumnName(i))
	}
}

// TestColumnsClone tests Columns.Clone and the Rows of an empty
// Columns set by its first column.
func TestColumnsClone(t *testing.T) {
	cl := NewColumns()
	require.NoError(t, cl.AddColumn("a", tensor.NewFloat64(4)))
	assert.Equal(t, 4, cl.Rows)
	cl.UpdateRows()
	cp := cl.Clone()
	assert.Equal(t, 4, cp.Rows)
	assert.Equal(t, []string{"a"}, cp.Keys)
	cp.Values[0].SetFloat1D(1, 0)
	assert.Equal(t, 0.0, cl.Values[0].Float1D(0))
	NewColumns().UpdateRows() // empty: no-op
}
