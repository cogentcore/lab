// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package table

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newIOTable returns a table with one column of each header type,
// including a 2x2 tensor cell column.
func newIOTable() *Table {
	dt := New()
	dt.AddStringColumn("Name")
	dt.AddFloat32Column("F32")
	dt.AddFloat64Column("F64")
	dt.AddIntColumn("Int")
	AddColumn[bool](dt, "Bool")
	dt.AddFloat32Column("Cell", 2, 2)
	dt.SetNumRows(3)
	for i := range 3 {
		dt.Column("Name").SetStringRow(string(rune('a'+i)), i, 0)
		dt.Column("F32").SetFloatRow(float64(i)+0.5, i, 0)
		dt.Column("F64").SetFloatRow(float64(i)*1.25, i, 0)
		dt.Column("Int").SetFloatRow(float64(10*i), i, 0)
		dt.Column("Bool").SetFloatRow(float64(i%2), i, 0)
		for c := range 4 {
			dt.Column("Cell").SetFloatRow(float64(i*4+c), i, c)
		}
	}
	return dt
}

// assertSameTable checks that two tables have the same columns, shapes
// and values.
func assertSameTable(t *testing.T, a, b *Table) {
	t.Helper()
	require.Equal(t, a.Columns.Keys, b.Columns.Keys)
	for i, ac := range a.Columns.Values {
		bc := b.Columns.Values[i]
		assert.Equal(t, ac.DataType(), bc.DataType(), a.ColumnName(i))
		assert.Equal(t, ac.ShapeSizes(), bc.ShapeSizes(), a.ColumnName(i))
		for j := range ac.Len() {
			assert.Equal(t, ac.String1D(j), bc.String1D(j), a.ColumnName(i))
		}
	}
}

// TestSaveOpenCSV tests that a table with every column type and a
// tensor cell column round-trips through SaveCSV and OpenCSV.
func TestSaveOpenCSV(t *testing.T) {
	dt := newIOTable()
	hdrs := dt.TableHeaders()
	assert.Equal(t, "^Bool", hdrs[4])
	assert.Equal(t, "%Cell[2:0,0]<2:2,2>", hdrs[5])
	assert.Equal(t, "%Cell[2:1,1]", hdrs[8])

	fn := filepath.Join(t.TempDir(), "t.tsv")
	require.NoError(t, dt.SaveCSV(fsx.Filename(fn), tensor.Tab, Headers))
	rt := New()
	require.NoError(t, rt.OpenCSV(fsx.Filename(fn), tensor.Tab))
	assertSameTable(t, dt, rt)

	// comma, through an fs.FS
	var b bytes.Buffer
	require.NoError(t, dt.WriteCSV(&b, tensor.Comma, Headers))
	fsys := fstest.MapFS{"t.csv": &fstest.MapFile{Data: b.Bytes()}}
	ft := New()
	require.NoError(t, ft.OpenFS(fsys, "t.csv", tensor.Comma))
	assertSameTable(t, dt, ft)

	assert.Error(t, ft.OpenFS(fsys, "missing.csv", tensor.Comma))
	assert.Error(t, ft.OpenCSV(fsx.Filename(filepath.Join(t.TempDir(), "missing")), tensor.Tab))
	assert.Error(t, dt.SaveCSV(fsx.Filename(filepath.Join(t.TempDir(), "no", "dir.tsv")), tensor.Tab, Headers))

	// String is the tab-separated form with headers.
	assert.True(t, strings.HasPrefix(dt.String(), "$Name\t%F32\t#F64\t|Int\t^Bool\t"))
}

// TestWriteCSVIndexedPrecision tests that WriteCSV writes rows in the
// indexed order, uses the Precision metadata for floats, and can omit
// headers.
func TestWriteCSVIndexedPrecision(t *testing.T) {
	dt := New()
	dt.AddStringColumn("S")
	dt.AddFloat64Column("F")
	dt.AddFloat64Column("C", 2)
	dt.SetNumRows(2)
	dt.Column("S").SetStringRow("x", 0, 0)
	dt.Column("S").SetStringRow("y", 1, 0)
	dt.Column("F").SetFloatRow(math.Pi, 0, 0)
	dt.Column("F").SetFloatRow(2, 1, 0)
	dt.Column("C").SetFloatRow(1.0/3, 1, 1)
	tensor.SetPrecision(dt, 3)
	dt.Indexes = []int{1, 0}

	var b bytes.Buffer
	require.NoError(t, dt.WriteCSV(&b, tensor.Comma, NoHeaders))
	assert.Equal(t, "y,2,0,0.333\nx,3.14,0,0\n", b.String())

	b.Reset()
	require.NoError(t, dt.WriteCSVRow(&b, 0, tensor.Space))
	assert.Equal(t, "x 3.14 0 0\n", b.String())
}

// TestReadCSVExistingColumns tests reading data without headers into a
// table with existing columns, including NaN for missing numbers and
// short records.
func TestReadCSVExistingColumns(t *testing.T) {
	dt := New()
	dt.AddStringColumn("S")
	dt.AddFloat64Column("F")
	dt.AddIntColumn("I")
	data := "a\t1.5\t2\nb\t\t3\nc\tNaN\t-Inf\n"
	require.NoError(t, dt.ReadCSV(strings.NewReader(data), tensor.Tab))
	assert.Equal(t, 3, dt.NumRows())
	assert.Equal(t, "b", dt.Column("S").StringRow(1, 0))
	assert.Equal(t, 1.5, dt.Column("F").FloatRow(0, 0))
	assert.True(t, math.IsNaN(dt.Column("F").FloatRow(1, 0)))
	assert.True(t, math.IsNaN(dt.Column("F").FloatRow(2, 0)))
	assert.Equal(t, 3.0, dt.Column("I").FloatRow(1, 0))

	// a data row marker is skipped; empty input is not an error.
	dt.ReadCSVRow([]string{"_D:", "z", "9"}, 0)
	assert.Equal(t, "z", dt.Column("S").StringRow(0, 0))
	assert.Equal(t, 9.0, dt.Column("F").FloatRow(0, 0))
	require.NoError(t, dt.ReadCSV(strings.NewReader(""), tensor.Tab))
	assert.Equal(t, 3, dt.NumRows())

	// malformed CSV is an error.
	assert.Error(t, dt.ReadCSV(strings.NewReader("a,\"b\n"), tensor.Comma))
}

// TestConfigFromDataValues tests type inference for plain headers,
// including an empty header name and strings appearing after numbers
// or before them.
func TestConfigFromDataValues(t *testing.T) {
	data := "Int,Float,,Str,Mixed,Late\n1,1.5,1,x,1,1\n2,2,2,y,2.5,2\n3,3,3,z,3,3\n4,4,4,w,4,x\n"
	dt := New()
	require.NoError(t, dt.ReadCSV(strings.NewReader(data), tensor.Comma))
	assert.Equal(t, []string{"Int", "Float", "col_2", "Str", "Mixed", "Late"}, dt.Columns.Keys)
	kinds := []reflect.Kind{reflect.Int, reflect.Float64, reflect.Int, reflect.String, reflect.Float64, reflect.String}
	for i, k := range kinds {
		assert.Equal(t, k, dt.Columns.Values[i].DataType(), dt.ColumnName(i))
	}
	assert.Equal(t, 4, dt.NumRows())
	assert.Equal(t, 2.5, dt.Column("Mixed").FloatRow(1, 0))

	// a string value first stays string even if numbers follow.
	dt = New()
	require.NoError(t, dt.ReadCSV(strings.NewReader("S\nabc\n1\n2\n3\n"), tensor.Comma))
	assert.Equal(t, reflect.String, dt.Columns.Values[0].DataType())
	assert.Equal(t, "abc", dt.Column("S").StringRow(0, 0))

	assert.Equal(t, reflect.Float64, InferDataType("1e5"))
	assert.Equal(t, reflect.Float64, InferDataType("1.0"))
	assert.Equal(t, reflect.Int, InferDataType("-7"))
	assert.Equal(t, reflect.String, InferDataType("1.2.3"))
}

// TestHeaderHelpers tests the header character and shape parsing
// helpers.
func TestHeaderHelpers(t *testing.T) {
	assert.Equal(t, byte('|'), TableHeaderChar(reflect.Uint8))
	assert.Equal(t, byte('$'), TableHeaderChar(reflect.String))
	k, nm := TableColumnType("Plain")
	assert.Equal(t, reflect.String, k)
	assert.Equal(t, "Plain", nm)
	assert.Equal(t, []int{3, 4, 5}, ShapeFromString("3:3,4,5"))
	assert.True(t, DetectTableHeaders([]string{"_H:", "x"}))
	assert.True(t, DetectTableHeaders([]string{"", "$x"}))
	assert.False(t, DetectTableHeaders([]string{"$x", "y"}))
}

// TestCleanCatTSV tests that CleanCatTSV removes repeated headers from
// a concatenated file and sorts by the given column.
func TestCleanCatTSV(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "cat.tsv")
	data := "$Name\t|Val\nb\t2\na\t1\n$Name\t|Val\nc\t0\n"
	require.NoError(t, os.WriteFile(fn, []byte(data), 0666))
	require.NoError(t, CleanCatTSV(fn, "Val"))
	out, err := os.ReadFile(fn)
	require.NoError(t, err)
	assert.Equal(t, "$Name\t|Val\nc\t0\na\t1\nb\t2\n", string(out))

	assert.Error(t, CleanCatTSV(filepath.Join(t.TempDir(), "missing.tsv")))
}
