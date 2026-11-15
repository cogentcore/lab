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

// TestNewTypes tests that New and NewOfType create the right concrete
// type and data kind for every supported type, and panic otherwise.
func TestNewTypes(t *testing.T) {
	kinds := []reflect.Kind{reflect.String, reflect.Bool, reflect.Float64, reflect.Float32,
		reflect.Int, reflect.Int64, reflect.Uint64, reflect.Int32, reflect.Uint32, reflect.Uint8}
	for _, k := range kinds {
		v := NewOfType(k, 2, 3)
		assert.Equal(t, k, v.DataType(), k.String())
		assert.Equal(t, []int{2, 3}, v.ShapeSizes())
		assert.Equal(t, 6, v.Len())
	}
	assert.Panics(t, func() { NewOfType(reflect.Complex128, 2) })

	assert.Equal(t, reflect.String, New[string](2).DataType())
	assert.Equal(t, reflect.Bool, New[bool](2).DataType())
	assert.Equal(t, reflect.Float64, New[float64](2).DataType())
	assert.Equal(t, reflect.Float32, New[float32](2).DataType())
	assert.Equal(t, reflect.Int, New[int](2).DataType())
	assert.Equal(t, reflect.Int64, New[int64](2).DataType())
	assert.Equal(t, reflect.Uint64, New[uint64](2).DataType())
	assert.Equal(t, reflect.Int32, New[int32](2).DataType())
	assert.Equal(t, reflect.Uint32, New[uint32](2).DataType())
	assert.Equal(t, reflect.Uint8, New[byte](2).DataType())

	assert.Equal(t, reflect.Float32, NewFloat32(3).DataType())
	assert.Equal(t, reflect.Int32, NewInt32(3).DataType())
	assert.Equal(t, reflect.Uint32, NewUint32(3).DataType())
	assert.Equal(t, reflect.Uint8, NewByte(3).DataType())
	assert.Equal(t, int64(24), NewFloat64(3).Sizeof())
	assert.Equal(t, int64(12), NewInt32(3).Sizeof())
}

// TestShapeNames tests setting and getting the ShapeNames metadata.
func TestShapeNames(t *testing.T) {
	tsr := NewFloat64(2, 3)
	assert.Nil(t, ShapeNames(tsr.Metadata()))
	SetShapeNames(tsr.Metadata(), "rows", "cols")
	assert.Equal(t, []string{"rows", "cols"}, ShapeNames(tsr.Metadata()))
}

// TestBaseAccessors tests the generic n-dimensional Value, ValuePtr and
// Set accessors on a 5D tensor, which must all agree with IndexTo1D.
func TestBaseAccessors(t *testing.T) {
	tsr := NewFloat64(2, 3, 2, 2, 2)
	for i := range tsr.Len() {
		tsr.Set1D(float64(i), i)
	}
	sh := tsr.Shape()
	assert.Equal(t, float64(sh.IndexTo1D(1, 2)), tsr.Value(1, 2, 0, 0, 0))
	assert.Equal(t, float64(sh.IndexTo1D(1, 2, 1)), tsr.Value(1, 2, 1, 0, 0))
	assert.Equal(t, 7.0, tsr.Value1D(7))

	*tsr.ValuePtr(0, 1, 0, 0, 1) = -1
	assert.Equal(t, -1.0, tsr.Value(0, 1, 0, 0, 1))
	tsr.Set(-2, 1, 1, 1, 1, 1)
	assert.Equal(t, -2.0, tsr.Value5D(1, 1, 1, 1, 1))

	tsr.Set5D(5, 1, 2, 1, 1, 0)
	assert.Equal(t, 5.0, tsr.Value5D(1, 2, 1, 1, 0))
	*tsr.ValuePtr5D(1, 2, 1, 1, 0) = 55
	assert.Equal(t, 55.0, tsr.Value(1, 2, 1, 1, 0))

	t4 := NewInt(2, 3, 4, 5)
	t4.Set4D(4, 1, 2, 3, 4)
	assert.Equal(t, 4, t4.Value4D(1, 2, 3, 4))
	assert.Equal(t, 4, t4.Values[t4.Len()-1])
	*t4.ValuePtr4D(0, 0, 0, 1) = 9
	assert.Equal(t, 9, t4.Values[1])

	t3 := NewInt(2, 3, 4)
	t3.Set3D(3, 1, 0, 2)
	assert.Equal(t, 3, t3.Value3D(1, 0, 2))
	assert.Equal(t, 3, t3.Values[14])
	*t3.ValuePtr3D(0, 1, 0) = 8
	assert.Equal(t, 8, t3.Values[4])

	t2 := NewInt(3, 4)
	t2.Set2D(2, 2, 1)
	assert.Equal(t, 2, t2.Value2D(2, 1))
	assert.Equal(t, 2, t2.Values[9])
	*t2.ValuePtr2D(1, 1) = 7
	assert.Equal(t, 7, t2.Values[5])
	assert.Equal(t, "7", t2.StringValue(1, 1))
	assert.Equal(t, "7", t2.String1D(5))
	assert.Equal(t, "7", t2.StringRow(1, 1))
}

// TestNumberArith tests the in-place SetAdd/Sub/Mul/Div methods in all
// their dimensional variants.
func TestNumberArith(t *testing.T) {
	t1 := NewFloat64Full(10, 2)
	t1.SetAdd(2, 1)
	t1.SetSub(1, 0)
	assert.Equal(t, []float64{9, 12}, t1.Values)
	t1.SetMul(2, 1)
	t1.SetDiv(3, 0)
	assert.Equal(t, []float64{3, 24}, t1.Values)

	t2 := NewFloat64Full(12, 2, 2)
	t2.SetAdd2D(1, 0, 0)
	t2.SetSub2D(1, 0, 1)
	t2.SetMul2D(2, 1, 0)
	t2.SetDiv2D(2, 1, 1)
	assert.Equal(t, []float64{13, 11, 24, 6}, t2.Values)

	t3 := NewFloat64Full(12, 1, 2, 2)
	t3.SetAdd3D(1, 0, 0, 0)
	t3.SetSub3D(1, 0, 0, 1)
	t3.SetMul3D(2, 0, 1, 0)
	t3.SetDiv3D(2, 0, 1, 1)
	assert.Equal(t, []float64{13, 11, 24, 6}, t3.Values)

	t4 := NewFloat64Full(12, 1, 1, 2, 2)
	t4.SetAdd4D(1, 0, 0, 0, 0)
	t4.SetSub4D(1, 0, 0, 0, 1)
	t4.SetMul4D(2, 0, 0, 1, 0)
	t4.SetDiv4D(2, 0, 0, 1, 1)
	assert.Equal(t, []float64{13, 11, 24, 6}, t4.Values)

	t5 := NewFloat64Full(12, 1, 1, 1, 2, 2)
	t5.SetAdd5D(1, 0, 0, 0, 0, 0)
	t5.SetSub5D(1, 0, 0, 0, 0, 1)
	t5.SetMul5D(2, 0, 0, 0, 1, 0)
	t5.SetDiv5D(2, 0, 0, 0, 1, 1)
	assert.Equal(t, []float64{13, 11, 24, 6}, t5.Values)
}

// TestNumberRowsAndStrings tests row accessors, appending rows, and
// string conversions on Number tensors, including negative 1D indexes.
func TestNumberRowsAndStrings(t *testing.T) {
	tsr := NewInt(0, 2)
	tsr.AppendRowInt(1, 2)
	tsr.AppendRowFloat(3.7, 4)
	tsr.AppendRowString("5", "x")
	assert.Equal(t, []int{3, 2}, tsr.ShapeSizes())
	assert.Equal(t, []int{1, 2, 3, 4, 5, 0}, tsr.Values)
	assert.Equal(t, 3, tsr.IntRow(1, 0))
	assert.Equal(t, 4.0, tsr.FloatRow(1, 1))
	tsr.SetIntRow(9, 2, 1)
	assert.Equal(t, 9, tsr.Int1D(-1))
	tsr.SetInt1D(8, -1)
	assert.Equal(t, 8, tsr.Int(2, 1))
	tsr.SetInt(7, 2, 1)
	assert.Equal(t, 7.0, tsr.Float1D(-1))
	tsr.SetFloat1D(6, -1)
	assert.Equal(t, 6.0, tsr.Float(2, 1))
	tsr.SetStringRow("11", 0, 0)
	tsr.SetStringRow("bad", 0, 1)
	assert.Equal(t, []int{11, 2}, tsr.Values[:2])
	tsr.SetString1D("12", 1)
	tsr.SetString("13", 1, 0)
	assert.Equal(t, []int{11, 12, 13}, tsr.Values[:3])

	// 0-dimensional tensors become 1D when appended to.
	z := &Float64{}
	z.AppendRowFloat(1)
	assert.Equal(t, []int{1}, z.ShapeSizes())
	z = &Float64{}
	z.AppendRowInt(2)
	assert.Equal(t, []float64{2}, z.Values)
	z = &Float64{}
	z.AppendRowString("3")
	assert.Equal(t, []float64{3}, z.Values)
	z = &Float64{}
	z.SetNumRows(2)
	assert.Equal(t, []int{2}, z.ShapeSizes())

	rt := NewFloat64(2, 3)
	for i := range 6 {
		rt.Values[i] = float64(i)
	}
	assert.Equal(t, []float64{3, 4, 5}, rt.RowTensor(1).(*Float64).Values)
	assert.Nil(t, rt.SubSpace(0, 0, 0))
	assert.Equal(t, []float64{4}, rt.SubSpace(1, 1).(*Float64).Values)
	rt.SetRowTensor(NewFloat64FromValues(-1, -2, -3, -4), 0)
	assert.Equal(t, []float64{-1, -2, -3, 3, 4, 5}, rt.Values)
	rt.AppendRow(NewIntFromValues(7, 8, 9))
	assert.Equal(t, []int{3, 3}, rt.ShapeSizes())
	assert.Equal(t, []float64{7, 8, 9}, rt.Values[6:])
	z = &Float64{}
	z.AppendRow(NewFloat64FromValues(1))
	assert.Equal(t, []float64{1}, z.Values)
}

// TestNumberCopy tests Clone, CopyFrom, AppendFrom and CopyCellsFrom on
// Number tensors, both from the same type and from other types.
func TestNumberCopy(t *testing.T) {
	a := NewFloat64FromValues(1.5, 2.5, 3.5)
	c := a.Clone().(*Float64)
	c.Values[0] = 0
	assert.Equal(t, 1.5, a.Values[0])

	i := NewInt(3)
	i.CopyFrom(a)
	assert.Equal(t, []int{1, 2, 3}, i.Values)
	f := NewFloat32(2)
	f.CopyFrom(a)
	assert.Equal(t, []float32{1.5, 2.5}, f.Values)
	same := NewFloat64(3)
	same.CopyFrom(a)
	assert.Equal(t, a.Values, same.Values)

	app := NewFloat64FromValues(1)
	app.AppendFrom(a)
	assert.Equal(t, []float64{1, 1.5, 2.5, 3.5}, app.Values)
	app.AppendFrom(NewIntFromValues(4, 5))
	assert.Equal(t, []float64{1, 1.5, 2.5, 3.5, 4, 5}, app.Values)
	// mismatched cell sizes are an error and leave the tensor unchanged
	app.AppendFrom(NewFloat64(2, 2))
	assert.Equal(t, 6, app.Len())

	cc := NewFloat64(4)
	cc.CopyCellsFrom(a, 1, 1, 2)
	assert.Equal(t, []float64{0, 2.5, 3.5, 0}, cc.Values)
	cc.CopyCellsFrom(NewIntFromValues(7, 8), 0, 0, 2)
	assert.Equal(t, []float64{7, 8, 3.5, 0}, cc.Values)
	cc.SetZeros()
	assert.Equal(t, []float64{0, 0, 0, 0}, cc.Values)
}

// TestBytesRoundTrip tests that Bytes and SetFromBytes round-trip for
// Number, String and Bool tensors.
func TestBytesRoundTrip(t *testing.T) {
	f := NewFloat64FromValues(1, -2, math.Pi)
	f2 := NewFloat64(3)
	f2.SetFromBytes(f.Bytes())
	assert.Equal(t, f.Values, f2.Values)

	s := NewStringFromValues("a", "", "hello world")
	s2 := NewString(3)
	s2.SetFromBytes(s.Bytes())
	assert.Equal(t, s.Values, s2.Values)

	b := NewBoolFromValues(true, false, true, true, false, false, true, false, true)
	b2 := NewBool(9)
	b2.SetFromBytes(b.Bytes())
	for i := range 9 {
		assert.Equal(t, b.Value1D(i), b2.Value1D(i))
	}
}

// TestStringTensor tests the String tensor numeric conversions, rows,
// and copy functions.
func TestStringTensor(t *testing.T) {
	s := NewString(2, 2)
	assert.True(t, s.IsString())
	assert.Equal(t, s, s.AsValues())
	s.SetFloat(1.5, 0, 0)
	s.SetInt(2, 0, 1)
	s.SetFloat1D(3, 2)
	s.SetInt1D(4, -1)
	assert.Equal(t, []string{"1.5", "2", "3", "4"}, s.Values)
	assert.Equal(t, 1.5, s.Float(0, 0))
	assert.Equal(t, 2, s.Int(0, 1))
	assert.Equal(t, 3.0, s.Float1D(2))
	assert.Equal(t, 4, s.Int1D(-1))
	assert.Equal(t, 3.0, s.FloatRow(1, 0))
	assert.Equal(t, 4, s.IntRow(1, 1))
	s.SetFloatRow(5, 1, 0)
	s.SetIntRow(6, 1, 1)
	s.SetStringRow("x", 0, 0)
	assert.Equal(t, []string{"x", "2", "5", "6"}, s.Values)
	assert.Equal(t, 0.0, s.Float1D(0))
	assert.Equal(t, 0, s.Int1D(0))
	assert.Equal(t, 0.0, StringToFloat64("nope"))
	assert.Equal(t, "0.25", Float64ToString(0.25))

	s.AppendRowString("a", "b", "c")
	s.AppendRowFloat(1)
	s.AppendRowInt(2, 3)
	assert.Equal(t, []int{5, 2}, s.ShapeSizes())
	assert.Equal(t, []string{"a", "b", "1", "", "2", "3"}, s.Values[4:])
	z := &String{}
	z.AppendRowString("q")
	assert.Equal(t, []string{"q"}, z.Values)
	z = &String{}
	z.AppendRowFloat(1)
	assert.Equal(t, []string{"1"}, z.Values)
	z = &String{}
	z.AppendRowInt(1)
	assert.Equal(t, []string{"1"}, z.Values)
	z = &String{}
	z.AppendRow(NewStringFromValues("r"))
	assert.Equal(t, []string{"r"}, z.Values)

	c := s.Clone().(*String)
	c.Values[0] = "changed"
	assert.Equal(t, "x", s.Values[0])

	n := NewString(3)
	n.CopyFrom(NewFloat64FromValues(1, 2.5))
	assert.Equal(t, []string{"1", "2.5", ""}, n.Values)
	n.CopyFrom(NewStringFromValues("a", "b", "c"))
	assert.Equal(t, []string{"a", "b", "c"}, n.Values)
	n.AppendFrom(NewStringFromValues("d"))
	n.AppendFrom(NewIntFromValues(5))
	assert.Equal(t, []string{"a", "b", "c", "d", "5"}, n.Values)
	n.AppendFrom(NewString(1, 2))
	assert.Equal(t, 5, n.Len())
	n.CopyCellsFrom(NewStringFromValues("p", "q"), 0, 1, 1)
	n.CopyCellsFrom(NewIntFromValues(8, 9), 1, 0, 2)
	assert.Equal(t, []string{"q", "8", "9", "d", "5"}, n.Values)

	r := NewString(2, 2)
	r.SetRowTensor(NewStringFromValues("a", "b", "c"), 1)
	assert.Equal(t, []string{"", "", "a", "b"}, r.Values)
	assert.Equal(t, []string{"a", "b"}, r.RowTensor(1).(*String).Values)
	r.AppendRow(NewStringFromValues("e", "f"))
	assert.Equal(t, []string{"e", "f"}, r.Values[4:])
	r.SetZeros()
	assert.Equal(t, []string{"", "", "", "", "", ""}, r.Values)
}

// TestBoolTensor tests the Bool tensor accessors and conversions.
func TestBoolTensor(t *testing.T) {
	b := NewBool(2, 3)
	assert.Equal(t, reflect.Bool, b.DataType())
	assert.False(t, b.IsString())
	assert.Equal(t, b, b.AsValues())
	assert.Equal(t, 3, b.DimSize(1))
	assert.Equal(t, 2, b.NumDims())
	assert.Equal(t, []int{2, 3}, b.ShapeSizes())
	assert.Equal(t, int64(len(b.Values)), b.Sizeof())
	metadata.SetName(b, "flags")
	assert.Equal(t, "flags [2 3]", b.Label())
	assert.Contains(t, b.String(), "flags")

	b.Set(true, 0, 1)
	assert.True(t, b.Value(0, 1))
	assert.True(t, b.Bool(0, 1))
	b.SetBool(true, 1, 2)
	assert.True(t, b.Bool1D(5))
	b.Set1D(false, 5)
	assert.False(t, b.Value1D(5))
	b.SetBool1D(true, 5)

	b.SetFloat(0, 0, 1)
	assert.Equal(t, 0.0, b.Float(0, 1))
	b.SetFloat1D(2, 0)
	assert.Equal(t, 1.0, b.Float1D(0))
	b.SetFloatRow(1, 1, 0)
	assert.Equal(t, 1.0, b.FloatRow(1, 0))
	b.SetInt(1, 0, 2)
	assert.Equal(t, 1, b.Int(0, 2))
	b.SetInt1D(0, 2)
	assert.Equal(t, 0, b.Int1D(2))
	b.SetIntRow(1, 1, 1)
	assert.Equal(t, 1, b.IntRow(1, 1))

	b.SetString("true", 0, 1)
	assert.Equal(t, "true", b.StringValue(0, 1))
	b.SetString1D("false", 1)
	assert.Equal(t, "false", b.String1D(1))
	b.SetStringRow("true", 1, 2)
	assert.Equal(t, "true", b.StringRow(1, 2))
	b.SetString("notabool", 1, 2) // ignored
	assert.True(t, b.Bool(1, 2))

	assert.True(t, Float64ToBool(0.5))
	assert.False(t, Float64ToBool(0))
	assert.True(t, IntToBool(3))
	assert.Equal(t, 1, BoolToInt(true))
	assert.Equal(t, 0.0, BoolToFloat64(false))

	// row-based functions are not supported by Bool
	assert.Nil(t, b.SubSpace(0))
	assert.Nil(t, b.RowTensor(0))
	b.SetRowTensor(NewBool(3), 0)
	b.AppendRow(NewBool(3))
	b.AppendRowFloat(1)
	b.AppendRowInt(1)
	b.AppendRowString("true")
	assert.Equal(t, 6, b.Len())

	c := b.Clone().(*Bool)
	c.SetZeros()
	assert.True(t, b.Bool1D(5))
	assert.False(t, c.Bool1D(5))
	c.SetTrue()
	assert.True(t, c.Bool1D(0))

	n := NewBool(3)
	n.CopyFrom(NewFloat64FromValues(1, 0, 2))
	assert.Equal(t, []bool{true, false, true}, []bool{n.Bool1D(0), n.Bool1D(1), n.Bool1D(2)})
	m := NewBool(3)
	m.CopyFrom(n)
	assert.Equal(t, []bool{true, false, true}, []bool{m.Bool1D(0), m.Bool1D(1), m.Bool1D(2)})

	cc := NewBool(4)
	cc.CopyCellsFrom(n, 1, 0, 3)
	assert.Equal(t, []bool{false, true, false, true}, boolValues(cc))
	cc.CopyCellsFrom(NewIntFromValues(1), 0, 0, 1)
	assert.True(t, cc.Bool1D(0))

	nr := NewBool(2, 2)
	nr.SetNumRows(3)
	assert.Equal(t, 6, nr.Len())
}

// TestBoolAppendFrom tests appending to a Bool tensor from Bool and
// numeric tensors, at bit offsets that are not byte aligned.
func TestBoolAppendFrom(t *testing.T) {
	a := NewBoolFromValues(true, false, true)
	a.AppendFrom(NewBoolFromValues(false, true))
	require.Equal(t, 5, a.Len())
	assert.Equal(t, []bool{true, false, true, false, true}, boolValues(a))
	a.AppendFrom(NewFloat64FromValues(1, 0))
	assert.Equal(t, []bool{true, false, true, false, true, true, false}, boolValues(a))
	// copying from a Bool tensor of a different length
	c := NewBoolFromValues(false, false, false, false, false, false, false, false, false, false)
	c.CopyFrom(a)
	assert.Equal(t, []bool{true, false, true, false, true, true, false, false, false, false}, boolValues(c))
	a.AppendFrom(NewBool(2, 2))
	assert.Equal(t, 7, a.Len())
}

func boolValues(b *Bool) []bool {
	n := b.Len()
	v := make([]bool, n)
	for i := range n {
		v[i] = b.Bool1D(i)
	}
	return v
}
