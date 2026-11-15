// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"math"
	"reflect"
	"testing"

	"cogentcore.org/core/base/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSeq returns a new Float64 tensor of given sizes with sequential values.
func newSeq(sizes ...int) *Float64 {
	tsr := NewFloat64(sizes...)
	for i := range tsr.Len() {
		tsr.Values[i] = float64(i)
	}
	return tsr
}

// TestRowsAccess tests the Rows view accessors with and without indexes,
// including 1D access with duplicated row indexes.
func TestRowsAccess(t *testing.T) {
	a := newSeq(3, 2)
	metadata.SetName(a, "a")
	rw := AsRows(a)
	assert.Same(t, rw, AsRows(rw))
	assert.False(t, rw.IsString())
	assert.Equal(t, reflect.Float64, rw.DataType())
	assert.Equal(t, "a [3 2]", rw.Label())
	assert.Equal(t, a.Metadata(), rw.Metadata())
	assert.Equal(t, a.String(), rw.String())
	assert.Same(t, a.Shape(), rw.Shape())
	assert.Equal(t, []int{3, 2}, rw.ShapeSizes())
	assert.Same(t, a, rw.AsValues())
	r, c := rw.RowCellSize()
	assert.Equal(t, 3, r)
	assert.Equal(t, 2, c)
	assert.Equal(t, 5.0, rw.FloatRow(2, 1))
	assert.Equal(t, "5", rw.StringRow(2, 1))
	assert.Equal(t, 5, rw.IntRow(2, 1))
	assert.Equal(t, 5.0, rw.Float1D(5))
	assert.Equal(t, 5, rw.Int1D(5))
	assert.Equal(t, "5", rw.String1D(5))
	rw.SetFloat1D(50, 5)
	rw.SetInt1D(40, 4)
	rw.SetString1D("30", 3)
	assert.Equal(t, []float64{30, 40, 50}, a.Values[3:])
	rw.SetFloat(1, 0, 0)
	rw.SetInt(2, 0, 1)
	rw.SetString("3", 1, 0)
	assert.Equal(t, []float64{1, 2, 3}, a.Values[:3])
	rw.Swap(0, 1) // no-op without indexes
	rw.SortIndexes()
	assert.Nil(t, rw.Indexes)

	// rows 2, 0, 0, 1: the view has more rows than the source
	rw.Indexes = []int{2, 0, 0, 1}
	assert.Equal(t, []int{4, 2}, rw.ShapeSizes())
	assert.Equal(t, []int{4, 2}, rw.Shape().Sizes)
	assert.Equal(t, 8, rw.Len())
	assert.Equal(t, 4, rw.DimSize(0))
	assert.Equal(t, 2, rw.DimSize(1))
	want := []float64{40, 50, 1, 2, 1, 2, 3, 30}
	for i, w := range want {
		assert.Equal(t, w, rw.Float1D(i), i)
		assert.Equal(t, int(w), rw.Int1D(i), i)
		assert.Equal(t, Float64ToString(w), rw.String1D(i), i)
	}
	assert.Equal(t, 3.0, rw.Float(3, 0))
	assert.Equal(t, 3, rw.Int(3, 0))
	assert.Equal(t, "3", rw.StringValue(3, 0))
	assert.Equal(t, 50.0, rw.FloatRow(0, 1))
	assert.Equal(t, 50, rw.IntRow(0, 1))
	assert.Equal(t, "50", rw.StringRow(0, 1))

	rw.SetFloat1D(-1, 7) // row 3 = source row 1, cell 1
	assert.Equal(t, -1.0, a.Value(1, 1))
	rw.SetInt1D(-2, 6)
	assert.Equal(t, -2.0, a.Value(1, 0))
	rw.SetString1D("-3", 1)
	assert.Equal(t, -3.0, a.Value(2, 1))
	rw.SetFloat(-4, 1, 0)
	rw.SetInt(-5, 2, 1)
	rw.SetString("-6", 0, 0)
	assert.Equal(t, []float64{-4, -5, -2, -1, -6, -3}, a.Values)
	rw.SetFloatRow(7, 3, 0)
	rw.SetIntRow(8, 3, 1)
	rw.SetStringRow("9", 0, 0)
	assert.Equal(t, []float64{-4, -5, 7, 8, 9, -3}, a.Values)

	vals := rw.AsValues()
	assert.Equal(t, []float64{9, -3, -4, -5, -4, -5, 7, 8}, vals.(*Float64).Values)
	assert.Equal(t, []float64{7, 8}, rw.RowTensor(3).(*Float64).Values)
	assert.Equal(t, []float64{8}, rw.SubSpace(3, 1).(*Float64).Values)
	assert.Nil(t, rw.SubSpace())
	rw.SetRowTensor(NewFloat64FromValues(0, 1), 0)
	assert.Equal(t, []float64{0, 1}, a.Values[4:])
}

// TestRowsOne tests 1D Rows views with duplicated indexes.
func TestRowsOne(t *testing.T) {
	a := NewFloat64FromValues(10, 20)
	rw := NewRows(a, 0, 0, 1)
	assert.Equal(t, []float64{10, 10, 20}, AsFloat64Slice(rw))
	rw.SetFloat1D(5, 2)
	assert.Equal(t, []float64{10, 5}, a.Values)
}

// TestRowsIndexes tests index-manipulating Rows methods: sorting,
// filtering, permuting, adding, inserting, and deleting rows.
func TestRowsIndexes(t *testing.T) {
	a := NewFloat64FromValues(3, math.NaN(), 1, 2, 1)
	rw := NewRows(a)
	assert.Equal(t, 0, rw.RowIndex(0))
	rw.ExcludeMissing()
	assert.Equal(t, []int{0, 2, 3, 4}, rw.Indexes)
	rw.SortStable(Ascending)
	assert.Equal(t, []int{2, 4, 3, 0}, rw.Indexes)
	rw.SortStable(Descending)
	assert.Equal(t, []int{0, 3, 2, 4}, rw.Indexes)
	rw.Sort(Ascending)
	assert.Equal(t, []float64{1, 1, 2, 3}, AsFloat64Slice(rw))
	rw.SortIndexes()
	assert.Equal(t, []int{0, 2, 3, 4}, rw.Indexes)
	rw.Swap(0, 1)
	assert.Equal(t, []int{2, 0, 3, 4}, rw.Indexes)
	rw.SortStableFunc(func(tsr Values, i, j int) int { return CompareAscending(i, j, Descending) })
	assert.Equal(t, []int{4, 3, 2, 0}, rw.Indexes)

	cl := rw.CloneIndexes()
	cl.Indexes[0] = 99
	assert.Equal(t, 4, rw.Indexes[0])
	cl.CopyIndexes(NewRows(a))
	assert.Nil(t, cl.Indexes)

	rw.Indexes = []int{0, 7, 2}
	rw.ValidIndexes()
	assert.Equal(t, []int{0, 2}, rw.Indexes)
	rw.SetTensor(a)
	assert.Nil(t, rw.Indexes)
	rw.ValidIndexes()
	assert.Nil(t, rw.Indexes)

	rw.Permuted()
	assert.Len(t, rw.Indexes, 5)
	rw.Permuted()
	rw.SortIndexes()
	assert.Equal(t, []int{0, 1, 2, 3, 4}, rw.Indexes)

	rw.DeleteRows(1, 2)
	assert.Equal(t, []int{0, 3, 4}, rw.Indexes)
	rw.InsertRows(1, 2)
	assert.Equal(t, []int{0, 5, 6, 3, 4}, rw.Indexes)
	assert.Equal(t, 7, a.Len())
	rw.AddRows(1)
	assert.Equal(t, []int{0, 5, 6, 3, 4, 7}, rw.Indexes)
	rw.AppendRowFloat(10)
	rw.AppendRowInt(11)
	rw.AppendRowString("12")
	rw.AppendRow(NewFloat64FromValues(13))
	assert.Equal(t, []int{0, 5, 6, 3, 4, 7, 8, 9, 10, 11}, rw.Indexes)
	assert.Equal(t, []float64{10, 11, 12, 13}, a.Values[8:])

	seq := NewRows(NewFloat64(2))
	seq.AddRows(2)
	assert.Nil(t, seq.Indexes)
	assert.Equal(t, 4, seq.NumRows())

	// empty tensors clear the indexes
	empty := NewRows(NewFloat64(0), 1, 2)
	empty.ExcludeMissing()
	assert.Nil(t, empty.Indexes)
	empty.Indexes = []int{1}
	empty.Permuted()
	assert.Nil(t, empty.Indexes)
	empty.Indexes = []int{1}
	empty.IndexesNeeded()
	assert.Nil(t, empty.Indexes)
	empty.Indexes = []int{1}
	empty.ValidIndexes()
	assert.Nil(t, empty.Indexes)
}

// TestRowsStrings tests sorting and filtering string Rows.
func TestRowsStrings(t *testing.T) {
	s := NewStringFromValues("pear", "Apple", "fig")
	rw := NewRows(s)
	assert.True(t, rw.IsString())
	rw.Sort(Ascending)
	assert.Equal(t, []string{"Apple", "fig", "pear"}, AsStringSlice(rw))
	rw.SortStable(Descending)
	assert.Equal(t, []string{"pear", "fig", "Apple"}, AsStringSlice(rw))
	rw.FilterString("apple", StringMatch{IgnoreCase: true, Exclude: true})
	assert.Equal(t, []string{"pear", "fig"}, AsStringSlice(rw))
}

// TestSlicedMore tests the Sliced view constructors, accessors, and index
// manipulation methods.
func TestSlicedMore(t *testing.T) {
	a := newSeq(3, 4)
	sl := AsSliced(a)
	assert.Same(t, sl, AsSliced(sl))
	sl.Indexes[0] = []int{2, 0}
	sl.Indexes[1] = []int{3, 1}
	assert.Equal(t, []int{2, 2}, sl.ShapeSizes())
	assert.Equal(t, []int{2, 1}, sl.SourceIndexesFrom1D(1))
	assert.Equal(t, 9.0, sl.Float(0, 1))
	sl.SetFloat(-9, 0, 1)
	assert.Equal(t, -9.0, a.Value(2, 1))
	sl.SetInt(-3, 1, 0)
	assert.Equal(t, -3.0, a.Value(0, 3))
	assert.Equal(t, -3, sl.Int(1, 0))
	sl.SetString("-1", 1, 1)
	assert.Equal(t, "-1", sl.StringValue(1, 1))
	assert.Equal(t, []float64{11, -9, -3, -1}, AsFloat64Slice(sl.AsValues()))

	iv := NewSliced(NewIntFromValues(5, 6, 7), []int{2, 0})
	assert.Equal(t, []int{7, 5}, iv.AsValues().(*Int).Values)

	sl.SortIndexes(0)
	assert.Equal(t, []int{0, 2}, sl.Indexes[0])
	sl.Sequential()
	assert.Nil(t, sl.Indexes[0])
	sl.SortIndexes(0) // no-op on nil
	assert.Nil(t, sl.Indexes[0])

	sl.SortStableFunc(1, func(tsr Tensor, i, j int) int { return CompareAscending(i, j, Descending) })
	assert.Equal(t, []int{3, 2, 1, 0}, sl.Indexes[1])
	sl.Filter(1, func(tsr Tensor, dim, idx int) bool { return idx%2 == 0 })
	assert.Equal(t, []int{2, 0}, sl.Indexes[1])

	sl.Permuted(0)
	assert.ElementsMatch(t, []int{0, 1, 2}, sl.Indexes[0])
	sl.Permuted(0)
	assert.ElementsMatch(t, []int{0, 1, 2}, sl.Indexes[0])

	b := newSeq(2)
	sl.SetTensor(b)
	assert.Len(t, sl.Indexes, 1)
	assert.Nil(t, sl.Indexes[0])

	assert.Equal(t, NewSlice(1, 5, 2), Slice{Start: 1, Stop: 5, Step: 2})
}

// TestAnySlice tests that AnySlice dispatches to the right view type.
func TestAnySlice(t *testing.T) {
	a := newSeq(3, 3)
	assert.Same(t, a, AnySlice(a).(*Float64))
	mask := NewBool(3, 3)
	mask.Set(true, 1, 1)
	ms, ok := AnySlice(a, mask).(*Masked)
	require.True(t, ok)
	assert.Equal(t, []float64{4}, AsFloat64Slice(ms.AsValues()))

	idx := NewIntFromValues(0, 1, 2, 2)
	idx.SetShapeSizes(2, 2)
	ix, ok := AnySlice(a, idx).(*Indexed)
	require.True(t, ok)
	assert.Equal(t, []float64{1, 8}, AsFloat64Slice(ix))

	rs := AnySlice(a, 1)
	assert.Equal(t, []float64{3, 4, 5}, AsFloat64Slice(rs))
	assert.Panics(t, func() { AnySlice(a, 5) })
	assert.Equal(t, []float64{8}, AsFloat64Slice(AnySlice(a, -1, -1)))
}

// TestMaskedMore tests the Masked view accessors, conversions and
// source index reporting.
func TestMaskedMore(t *testing.T) {
	a := newSeq(2, 3)
	ms := AsMasked(a)
	assert.Same(t, ms, AsMasked(ms))
	assert.Equal(t, []int{2, 3}, ms.ShapeSizes())
	assert.Equal(t, 3, ms.DimSize(1))
	ms.Filter(func(tsr Tensor, idx int) bool { return int(tsr.Float1D(idx))%2 == 1 })
	assert.True(t, math.IsNaN(ms.Float(0, 0)))
	assert.Equal(t, 1.0, ms.Float(0, 1))
	assert.Equal(t, 0, ms.Int(0, 0))
	assert.Equal(t, 3, ms.Int(1, 0))
	assert.Equal(t, "", ms.StringValue(0, 0))
	assert.Equal(t, "5", ms.StringValue(1, 2))
	assert.Equal(t, 0, ms.Int1D(0))
	assert.Equal(t, 1, ms.Int1D(1))
	assert.Equal(t, "", ms.String1D(0))
	assert.Equal(t, "3", ms.String1D(3))

	ms.SetFloat(-1, 0, 0) // masked: ignored
	ms.SetFloat(-1, 0, 1)
	ms.SetFloat1D(-2, 2)
	ms.SetFloat1D(-3, 3)
	ms.SetInt(-4, 1, 1)
	ms.SetInt(-5, 1, 2)
	ms.SetInt1D(-6, 0)
	ms.SetString("-7", 0, 2)
	ms.SetString1D("-8", 4)
	ms.SetString1D("-9", 5)
	assert.Equal(t, []float64{0, -1, 2, -3, 4, -9}, a.Values)
	assert.Equal(t, []int{0, 1, 1, 0, 1, 2}, ms.SourceIndexes(true).Values)
	assert.Equal(t, []int{3, 2}, ms.SourceIndexes(false).ShapeSizes())

	iv := NewMasked(NewIntFromValues(1, 2, 3))
	iv.Mask.Set1D(false, 1)
	assert.Equal(t, []int{1, 3}, iv.AsValues().(*Int).Values)
	sv := NewMasked(NewStringFromValues("a", "b"))
	sv.Mask.Set1D(false, 0)
	assert.Equal(t, []string{"b"}, sv.AsValues().(*String).Values)

	// mismatched mask shapes are reset to all true
	m2 := Mask(a, NewBool(4))
	assert.Equal(t, []int{2, 3}, m2.(*Masked).Mask.ShapeSizes())
	assert.True(t, m2.(*Masked).Mask.Bool1D(5))
	assert.Same(t, a, Mask(a, NewFloat64(2, 3)).(*Float64))
	ms.SetTensor(newSeq(4))
	assert.Equal(t, []int{4}, ms.Mask.ShapeSizes())
	ms.Mask = nil
	ms.SyncShape()
	assert.Equal(t, 4, ms.Mask.Len())
}

// TestIndexedMore tests the Indexed view accessors.
func TestIndexedMore(t *testing.T) {
	a := newSeq(3, 3)
	idx := NewIntFromValues(0, 1, 2, 2, 1, 0, 2, 0)
	idx.SetShapeSizes(2, 2, 2)
	ix := NewIndexed(a, idx)
	assert.Same(t, ix, AsIndexed(ix))
	assert.Nil(t, AsIndexed(a))
	assert.Equal(t, 2, ix.NumDims())
	assert.Equal(t, 2, ix.DimSize(1))
	assert.Equal(t, []int{2, 0}, ix.SourceIndexes(1, 1))
	assert.Equal(t, 6.0, ix.Float(1, 1))
	assert.Equal(t, 8, ix.Int(0, 1))
	assert.Equal(t, "3", ix.StringValue(1, 0))
	assert.Equal(t, 6.0, ix.Float1D(3))
	assert.Equal(t, 8, ix.Int1D(1))
	assert.Equal(t, "3", ix.String1D(2))
	assert.Equal(t, []float64{1, 8, 3, 6}, ix.AsValues().(*Float64).Values)

	ix.SetFloat(-1, 0, 0)
	ix.SetInt(-2, 0, 1)
	ix.SetString("-3", 1, 0)
	ix.SetFloat1D(-4, 3)
	assert.Equal(t, []float64{-1, -2, -3, -4}, AsFloat64Slice(ix))
	ix.SetInt1D(-5, 0)
	ix.SetString1D("-6", 1)
	assert.Equal(t, -5.0, a.Value(0, 1))
	assert.Equal(t, -6.0, a.Value(2, 2))

	sx := NewIndexed(NewStringFromValues("a", "b", "c"), NewIntFromValues(2, 0))
	sx.Indexes.SetShapeSizes(2, 1)
	assert.Equal(t, []string{"c", "a"}, sx.AsValues().(*String).Values)
	nx := NewIndexed(NewIntFromValues(4, 5), NewIntFromValues(1))
	nx.Indexes.SetShapeSizes(1, 1)
	assert.Equal(t, []int{5}, nx.AsValues().(*Int).Values)
	nx.SetTensor(NewIntFromValues(6, 7))
	assert.Equal(t, 7, nx.Int1D(0))
}

// TestReshapedMore tests Reshape, NewRowCellsView, Transpose and the
// Reshaped accessors and error paths.
func TestReshapedMore(t *testing.T) {
	a := newSeq(2, 3)
	assert.Same(t, a, Reshape(a).(*Float64))
	assert.Equal(t, []int{6}, Reshape(a, -1).ShapeSizes())
	assert.Equal(t, []int{3, 2}, Reshape(a, 3, -1).ShapeSizes())
	rs := &Reshaped{Tensor: a}
	assert.Error(t, rs.SetShapeSizes(4, -1))
	assert.Error(t, rs.SetShapeSizes(4))
	sc := &Reshaped{Tensor: NewFloat64FromValues(1)}
	assert.NoError(t, sc.SetShapeSizes(-1))
	assert.Equal(t, []int{1}, sc.ShapeSizes())
	em := &Reshaped{Tensor: NewFloat64(0)}
	assert.NoError(t, em.SetShapeSizes(5))

	r3 := newSeq(2, 3, 4)
	rc := NewRowCellsView(r3, 2)
	assert.Equal(t, []int{6, 4}, rc.ShapeSizes())
	assert.Same(t, rc, AsReshaped(rc))
	ar := AsReshaped(a)
	assert.Equal(t, []int{2, 3}, ar.ShapeSizes())

	r := NewReshaped(a, 3, 2)
	assert.Equal(t, 5, r.Int(2, 1))
	assert.Equal(t, "5", r.StringValue(2, 1))
	assert.Equal(t, 4, r.Int1D(4))
	assert.Equal(t, "4", r.String1D(4))
	r.SetFloat1D(-1, 0)
	r.SetInt(-2, 0, 1)
	r.SetInt1D(-3, 2)
	r.SetString("-4", 1, 1)
	r.SetString1D("-5", 4)
	assert.Equal(t, []float64{-1, -2, -3, -4, -5, 5}, a.Values)

	tr := Transpose(a)
	assert.Equal(t, []int{3, 2}, tr.ShapeSizes())
	assert.Equal(t, -5.0, tr.Float(1, 1))
	assert.Equal(t, -2.0, tr.Float(1, 0))
	assert.Equal(t, []float64{-1, -4, -2, -5, -3, 5}, AsFloat64Slice(tr.AsValues()))

	s := NewStringFromValues("a", "b", "c", "d")
	s.SetShapeSizes(2, 2)
	assert.Equal(t, []string{"a", "c", "b", "d"}, Transpose(s).AsValues().(*String).Values)
}

// TestProjection2DAccess tests the Projection2D string and set functions.
func TestProjection2DAccess(t *testing.T) {
	a := NewString(2, 2, 2)
	Projection2DSetString(a, OnedRow, 3, 1, "x")
	assert.Equal(t, "x", a.Value(1, 1, 1))
	assert.Equal(t, "x", Projection2DString(a, OnedRow, 3, 1))
	f := NewFloat64(2, 3)
	Projection2DSet(f, OnedRow, 1, 2, 7)
	assert.Equal(t, 7.0, f.Value(1, 2))
	assert.Equal(t, 7.0, Projection2DValue(f, OnedRow, 1, 2))
}
