// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"errors"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"cogentcore.org/core/base/fsx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewFromValues tests NewFromValues for each supported element type.
func TestNewFromValues(t *testing.T) {
	assert.Nil(t, NewFromValues())
	assert.Nil(t, NewFromValues(struct{}{}))
	cases := []struct {
		vals []any
		kind reflect.Kind
	}{
		{[]any{"a", "b"}, reflect.String},
		{[]any{true, false}, reflect.Bool},
		{[]any{1.0, 2.0}, reflect.Float64},
		{[]any{float32(1), float32(2)}, reflect.Float32},
		{[]any{1, 2}, reflect.Int},
		{[]any{int32(1), int32(2)}, reflect.Int32},
		{[]any{uint32(1), uint32(2)}, reflect.Uint32},
		{[]any{int64(1), int64(2)}, reflect.Int64},
		{[]any{byte(1), byte(2)}, reflect.Uint8},
	}
	for _, c := range cases {
		v := NewFromValues(c.vals...)
		require.NotNil(t, v, c.kind.String())
		assert.Equal(t, c.kind, v.DataType())
		assert.Equal(t, 2, v.Len())
	}
	assert.Equal(t, []float32{1, 2}, NewFloat32FromValues(1, 2).Values)
}

// TestCreateFull tests the constant-filled and random tensor constructors.
func TestCreateFull(t *testing.T) {
	assert.Equal(t, []int{3, 3, 3, 3}, NewIntFull(3, 2, 2).Values)
	assert.Equal(t, []string{"x", "x"}, NewStringFull("x", 2).Values)
	// large enough to use multiple threads
	big := NewIntFull(1, 1000)
	assert.Equal(t, 1000, big.Len())
	for _, v := range big.Values {
		assert.Equal(t, 1, v)
	}
	r := NewFloat64Rand(4, 5)
	assert.Equal(t, []int{4, 5}, r.ShapeSizes())
	for _, v := range r.Values {
		assert.True(t, v >= 0 && v < 1)
	}
}

// TestShapeFuncs tests Shape methods and the shape helper functions.
func TestShapeFuncs(t *testing.T) {
	sh := NewShape(2, 3, 4)
	assert.True(t, sh.IndexIsValid(1, 2, 3))
	assert.False(t, sh.IndexIsValid(1, 3, 0))
	assert.False(t, sh.IndexIsValid(-1, 0, 0))
	assert.False(t, sh.IndexIsValid(1, 2))
	assert.Equal(t, sh.IndexTo1D(1, 2), sh.Index2To1D(1, 2))
	assert.Equal(t, sh.IndexTo1D(1, 2, 3), sh.Index3To1D(1, 2, 3))
	assert.Equal(t, []int{2, 3, 4}, sh.SizesAsTensor().Values)
	assert.Equal(t, "[2 3 4]", sh.String())

	sh4 := NewShape(2, 3, 4, 5)
	assert.Equal(t, sh4.IndexTo1D(1, 2, 3, 4), sh4.Index4To1D(1, 2, 3, 4))
	assert.Equal(t, 119, sh4.Index4To1D(1, 2, 3, 4))
	sh5 := NewShape(2, 3, 4, 5, 6)
	assert.Equal(t, sh5.IndexTo1D(1, 2, 3, 4, 5), sh5.Index5To1D(1, 2, 3, 4, 5))
	assert.Equal(t, sh5.Len()-1, sh5.Index5To1D(1, 2, 3, 4, 5))
	for i := range sh5.Len() {
		assert.Equal(t, i, sh5.IndexTo1D(sh5.IndexFrom1D(i)...))
	}

	other := NewShape(2, 3, 4)
	assert.True(t, sh.IsEqual(other))
	assert.False(t, sh.IsEqual(NewShape(2, 3)))
	other.Strides = []int{1, 2, 6}
	assert.False(t, sh.IsEqual(other))

	st := &Shape{}
	st.SetShapeSizesFromTensor(NewIntFromValues(4, 5))
	assert.Equal(t, []int{4, 5}, st.Sizes)
	assert.Equal(t, 0, (&Shape{}).Len())
	r, c := (&Shape{}).RowCellSize()
	assert.Equal(t, 0, r)
	assert.Equal(t, 1, c)
	r, c = NewShape(0, 3, 2).RowCellSize()
	assert.Equal(t, 0, r)
	assert.Equal(t, 6, c)

	assert.Equal(t, []int{2, 3, 4, 5}, AddShapes(NewShape(2, 3), NewShape(4, 5)).Sizes)
	assert.Equal(t, []int{1}, CellsSize([]int{5}))
	assert.Equal(t, []int{3, 4}, CellsSize([]int{2, 3, 4}))

	assert.Nil(t, RowMajorStrides())
	assert.Equal(t, []int{12, 4, 1}, RowMajorStrides(2, 3, 4))
	assert.Equal(t, []int{0, 0}, RowMajorStrides(2, 0))
	assert.Equal(t, []int{1, 2, 6}, ColumnMajorStrides(2, 3, 4))
	assert.Equal(t, []int{1, 1}, ColumnMajorStrides(2, 0))

	// the caller's sizes must not be modified
	sizes := []int{0, 3}
	sh0 := NewShape(sizes...)
	assert.Equal(t, []int{0, 3}, sizes)
	assert.Equal(t, []int{0, 3}, sh0.Sizes)
	assert.Equal(t, []int{3, 1}, sh0.Strides)
}

// TestConvert tests the copy and conversion functions in convert.go.
func TestConvert(t *testing.T) {
	a := NewFloat64FromValues(1, 2, 3, 4, 5, 6)
	a.SetShapeSizes(2, 3)

	c := Clone(a)
	c.SetFloat1D(-1, 0)
	assert.Equal(t, 1.0, a.Float1D(0))
	cv := CloneValues(a)
	assert.Equal(t, a.Values, cv.Values)

	f := Flatten(a)
	assert.Equal(t, []int{6}, f.ShapeSizes())
	f.SetFloat1D(-1, 0)
	assert.Equal(t, 1.0, a.Float1D(0))
	ms := NewMasked(a).Filter(func(tsr Tensor, idx int) bool { return tsr.Float1D(idx) > 4 })
	assert.Equal(t, []float64{5, 6}, Flatten(ms).(*Float64).Values)

	sq := NewFloat64(1, 3, 1, 2)
	assert.Equal(t, []int{3, 2}, Squeeze(sq).ShapeSizes())
	assert.Equal(t, a, Squeeze(a))

	assert.Equal(t, []int{6}, As1D(a).ShapeSizes())
	one := NewFloat64(3)
	assert.Equal(t, Tensor(one), As1D(one))

	assert.Equal(t, []float64{4, 5, 6}, AsFloat64Slice(Cells1D(a, 1)))
	sl := NewSliced(a)
	assert.Equal(t, []float64{4, 5, 6}, AsFloat64Slice(Cells1D(sl, 1)))

	_, err := MustBeValues(a)
	assert.NoError(t, err)
	_, err = MustBeValues(sl)
	assert.Error(t, err)
	assert.NoError(t, MustBeSameShape(a, NewInt(2, 3)))
	assert.Error(t, MustBeSameShape(a, NewInt(3, 2)))

	v := NewFloat64()
	SetShape(v, NewShape(4, 2))
	assert.Equal(t, []int{4, 2}, v.ShapeSizes())
	SetShapeSizesFromTensor(v, NewIntFromValues(2, 2))
	assert.Equal(t, []int{2, 2}, v.ShapeSizes())

	empty := NewFloat64()
	assert.Equal(t, 0.0, AsFloat64Scalar(empty))
	assert.Equal(t, 0, AsIntScalar(empty))
	assert.Equal(t, "", AsStringScalar(empty))
	assert.Nil(t, AsFloat64Slice(empty))
	assert.Nil(t, AsIntSlice(empty))
	assert.Nil(t, AsStringSlice(empty))
	assert.Equal(t, 2, AsIntScalar(NewIntScalar(2)))
	assert.Equal(t, "s", AsStringScalar(NewStringScalar("s")))
	assert.Equal(t, float32(1.5), NewFloat32Scalar(1.5).Values[0])
	assert.Equal(t, []string{"1", "2"}, AsStringSlice(NewIntFromValues(1, 2)))

	ai := NewIntFromValues(1, 2)
	assert.Same(t, ai, AsInt(ai))
	assert.Equal(t, []float64{1, 2}, AsFloat64(ai).Values)
	assert.Equal(t, []float32{1, 2}, AsFloat32(ai).Values)
	assert.Equal(t, []string{"1", "2"}, AsString(ai).Values)
	f64 := NewFloat64FromValues(1.5)
	assert.Same(t, f64, AsFloat64(f64))
	assert.Equal(t, []int{1}, AsInt(f64).Values)
	f32 := NewFloat32FromValues(1)
	assert.Same(t, f32, AsFloat32(f32))
	s := NewStringFromValues("a")
	assert.Same(t, s, AsString(s))
}

// TestCloneRowsCopies tests that Clone and Flatten of a Rows view with no
// indexes return a copy and do not modify the source tensor.
func TestCloneRowsCopies(t *testing.T) {
	a := NewFloat64FromValues(1, 2, 3, 4)
	a.SetShapeSizes(2, 2)
	rw := NewRows(a)
	c := Clone(rw)
	c.SetFloat1D(-1, 0)
	assert.Equal(t, 1.0, a.Float1D(0))

	f := Flatten(rw)
	assert.Equal(t, []int{4}, f.ShapeSizes())
	assert.Equal(t, []int{2, 2}, a.ShapeSizes())
}

// TestRangeContains tests Range and the Contains functions.
func TestRangeContains(t *testing.T) {
	v := NewFloat64FromValues(3, math.NaN(), -2, 7, 1)
	mn, mx, mni, mxi := Range(v)
	assert.Equal(t, -2.0, mn)
	assert.Equal(t, 7.0, mx)
	assert.Equal(t, 2, mni)
	assert.Equal(t, 3, mxi)
	_, _, mni, mxi = Range(NewFloat64FromValues(math.NaN()))
	assert.Equal(t, -1, mni)
	assert.Equal(t, -1, mxi)

	assert.True(t, ContainsFloat(v, NewFloat64FromValues(9, 7)))
	assert.False(t, ContainsFloat(v, NewFloat64FromValues(9)))
	assert.False(t, ContainsFloat(v, NewFloat64()))
	iv := NewIntFromValues(1, 2, 3)
	assert.True(t, ContainsInt(iv, NewIntFromValues(3)))
	assert.False(t, ContainsInt(iv, NewIntFromValues(4)))
	assert.False(t, ContainsInt(iv, NewInt()))
	sv := NewStringFromValues("Apple", "banana")
	assert.True(t, ContainsString(sv, NewStringFromValues("banana"), StringMatch{}))
	assert.False(t, ContainsString(sv, NewStringFromValues("apple"), StringMatch{}))
	assert.True(t, ContainsString(sv, NewStringFromValues("apple"), StringMatch{IgnoreCase: true}))
	assert.True(t, ContainsString(sv, NewStringFromValues("nan"), StringMatch{Contains: true}))
	assert.True(t, ContainsString(sv, NewStringFromValues("APP"), StringMatch{Contains: true, IgnoreCase: true}))
	assert.True(t, ContainsString(sv, NewStringFromValues("Apple"), StringMatch{Exclude: true}))
	assert.False(t, ContainsString(sv, NewString(), StringMatch{}))
}

// TestCopyFromLargerShape tests copying the overlapping region from a
// larger tensor, for numeric and string tensors.
func TestCopyFromLargerShape(t *testing.T) {
	big := NewFloat64(3, 4)
	for i := range 12 {
		big.Values[i] = float64(i)
	}
	small := NewFloat64(2, 2)
	CopyFromLargerShape(small, big)
	assert.Equal(t, []float64{0, 1, 4, 5}, small.Values)

	bigs := NewString(3, 3)
	for i := range 9 {
		bigs.Values[i] = string(rune('a' + i))
	}
	smalls := NewString(2, 2)
	CopyFromLargerShape(smalls, bigs)
	assert.Equal(t, []string{"a", "b", "d", "e"}, smalls.Values)
}

// TestVectorize tests the threaded and unthreaded vectorize functions
// and the N* length helper functions.
func TestVectorize(t *testing.T) {
	out := NewFloat64(1000)
	VectorizeOnThreads(3, NFirstLen, func(idx int, tsr ...Tensor) {
		tsr[0].SetFloat1D(float64(idx), idx)
	}, out)
	for i := range 1000 {
		assert.Equal(t, float64(i), out.Values[i])
	}
	VectorizeOnThreads(2, NFirstLen, func(idx int, tsr ...Tensor) { t.Fatal("not called") }, NewFloat64())
	VectorizeThreaded(-1, NFirstLen, func(idx int, tsr ...Tensor) {
		tsr[0].SetFloat1D(1, idx)
	}, out)
	assert.Equal(t, 1.0, out.Values[999])
	VectorizeThreaded(1, NFirstLen, func(idx int, tsr ...Tensor) { t.Fatal("not called") }, NewFloat64())

	nt := NumThreads
	NumThreads = 3
	assert.Equal(t, 3, DefaultNumThreads())
	NumThreads = nt
	assert.Greater(t, DefaultNumThreads(), 0)

	assert.Equal(t, 0, NFirstRows())
	assert.Equal(t, 4, NFirstRows(NewFloat64(4, 2)))
	assert.Equal(t, 0, NFirstLen())
	assert.Equal(t, 0, NMinLen(2))
	assert.Equal(t, 3, NMinLen(2, NewFloat64(5), NewFloat64(3), NewFloat64(1)))
	assert.Equal(t, 1, NMinLen(3, NewFloat64(5), NewFloat64(3), NewFloat64(1)))
}

// TestCallOut tests the CallOut* helpers, FloatPromoteType, and Func.String.
func TestCallOut(t *testing.T) {
	a := NewIntFromValues(1, 2)
	b := NewFloat32FromValues(3, 4)
	c := NewFloat64FromValues(5, 6)
	assert.Equal(t, reflect.Int, FloatPromoteType(a, a))
	assert.Equal(t, reflect.Float32, FloatPromoteType(a, b))
	assert.Equal(t, reflect.Float64, FloatPromoteType(a, b, c))
	assert.Equal(t, reflect.Float64, FloatPromoteType(c, b))

	copy1 := func(x Tensor, out Values) error {
		SetShapeFrom(out, x)
		out.CopyFrom(x.AsValues())
		return nil
	}
	assert.Equal(t, reflect.Int, CallOut1(copy1, a).DataType())
	assert.Equal(t, []float64{1, 2}, CallOut1Float64(copy1, a).(*Float64).Values)
	add2 := func(x, y Tensor, out Values) error {
		return FloatBinaryFuncOut(1, func(a, b float64) float64 { return a + b }, x, y, out)
	}
	assert.Equal(t, []float32{4, 6}, CallOut2(add2, a, b).(*Float32).Values)
	assert.Equal(t, []float64{4, 6}, CallOut2Float64(add2, a, b).(*Float64).Values)
	add3 := func(x, y, z Tensor, out Values) error {
		if err := add2(x, y, out); err != nil {
			return err
		}
		return add2(out, z, out)
	}
	assert.Equal(t, []float64{9, 12}, CallOut3(add3, a, b, c).(*Float64).Values)
	gt := func(x, y Tensor, out *Bool) error {
		return BoolFloatsFuncOut(func(a, b float64) bool { return a > b }, x, y, out)
	}
	assert.Equal(t, []bool{false, false}, boolValues(CallOut2Bool(gt, a, b)))
	scale := func(g float64, x Tensor, out Values) error {
		return FloatFuncOut(1, func(v float64) float64 { return v * g }, x, out)
	}
	assert.Equal(t, []float64{10, 12}, CallOut1Gen1(scale, 2.0, c).(*Float64).Values)
	affine := func(g, h float64, x Tensor, out Values) error {
		return FloatFuncOut(1, func(v float64) float64 { return v*g + h }, x, out)
	}
	assert.Equal(t, []float64{11, 13}, CallOut1Gen2(affine, 2.0, 1.0, c).(*Float64).Values)
	sum2 := func(g float64, x, y Tensor, out Values) error {
		return FloatBinaryFuncOut(1, func(a, b float64) float64 { return g * (a + b) }, x, y, out)
	}
	assert.Equal(t, []float64{12, 16}, CallOut2Gen1(sum2, 2.0, a, c).(*Float64).Values)
	assert.Equal(t, []float32{4, 6}, FloatBinaryFunc(1, func(a, b float64) float64 { return a + b }, a, b).(*Float32).Values)

	fn, err := NewFunc("Add", add2)
	require.NoError(t, err)
	assert.Equal(t, "Add(tensor.Tensor, tensor.Tensor, tensor.Values)", fn.String())
	vf, _ := NewFunc("V", func(a any, ix ...int) {})
	assert.Equal(t, "V(any, ...[]int)", vf.String())
	assert.True(t, vf.Args[1].IsInt)
	assert.True(t, vf.Args[1].IsVariadic)
	nf, _ := NewFunc("N", func() {})
	assert.Equal(t, "N()", nf.String())
}

// TestCalc tests SetCalcFunc and Calc, including the missing-function error.
func TestCalc(t *testing.T) {
	a := NewFloat64(2)
	assert.Error(t, Calc(a))
	SetCalcFunc(a, func() error {
		a.SetFloat1D(42, 0)
		return nil
	})
	assert.NoError(t, Calc(a))
	assert.Equal(t, 42.0, a.Values[0])
	SetCalcFunc(a, func() error { return errors.New("fail") })
	assert.EqualError(t, Calc(a), "fail")
}

// TestAlignFuncs tests the shape alignment and broadcasting function
// helpers in align.go, including their error paths.
func TestAlignFuncs(t *testing.T) {
	// size-1 dimensions are broadcast: index 4 on dim 0 wraps to 0
	assert.Equal(t, 2, WrapIndex1D(NewShape(1, 3), 4, 2))
	_, _, err := AlignForAssign(NewFloat64(2, 3), NewFloat64(3))
	assert.NoError(t, err)
	_, _, err = AlignForAssign(NewFloat64(3), NewFloat64(2, 3))
	assert.Error(t, err)
	_, _, _, err = AlignShapes(NewFloat64(2, 3), NewFloat64(4))
	assert.Error(t, err)

	assert.Equal(t, []int{6, 4}, SplitAtInnerDims(NewFloat64(2, 3, 4), 1))
	assert.Equal(t, []int{2, 3, 4}, SplitAtInnerDims(NewFloat64(2, 3, 4), 2))
	assert.Nil(t, SplitAtInnerDims(NewFloat64(2), 2))

	a := NewFloat64FromValues(1, 2, 3, 4, 5, 6)
	a.SetShapeSizes(2, 3)
	b := NewFloat64FromValues(10, 20, 30)
	require.NoError(t, FloatAssignFunc(func(x, y float64) float64 { return x + y }, a, b))
	assert.Equal(t, []float64{11, 22, 33, 14, 25, 36}, a.Values)
	assert.Error(t, FloatAssignFunc(func(x, y float64) float64 { return x }, b, a))

	s := NewStringFromValues("a", "b", "c", "d")
	s.SetShapeSizes(2, 2)
	require.NoError(t, StringAssignFunc(func(x, y string) string { return x + y }, s, NewStringFromValues("1", "2")))
	assert.Equal(t, []string{"a1", "b2", "c1", "d2"}, s.Values)
	assert.Error(t, StringAssignFunc(func(x, y string) string { return x }, NewString(2), NewString(3)))

	sb := StringBinaryFunc(func(x, y string) string { return x + y }, NewStringFromValues("x", "y"), NewStringFromValues("!"))
	assert.Equal(t, []string{"x!", "y!"}, AsStringSlice(sb))
	assert.Error(t, StringBinaryFuncOut(func(x, y string) string { return x }, NewString(2), NewString(3), NewString()))
	assert.Error(t, FloatBinaryFuncOut(1, func(x, y float64) float64 { return x }, NewFloat64(2), NewFloat64(3), NewFloat64()))

	ff := FloatFunc(1, func(v float64) float64 { return -v }, NewFloat64FromValues(1, 2))
	assert.Equal(t, []float64{-1, -2}, ff.(*Float64).Values)
	fs := NewFloat64(3)
	assert.NoError(t, FloatSetFunc(1, func(idx int) float64 { return float64(idx * 2) }, fs))
	assert.Equal(t, []float64{0, 2, 4}, fs.Values)

	bs := BoolStringsFunc(func(x, y string) bool { return x == y }, NewStringFromValues("a", "b"), NewStringFromValues("b"))
	assert.Equal(t, []bool{false, true}, boolValues(bs))
	assert.Error(t, BoolStringsFuncOut(func(x, y string) bool { return true }, NewString(2), NewString(3), NewBool()))
	bf := BoolFloatsFunc(func(x, y float64) bool { return x < y }, NewFloat64FromValues(1, 3), NewFloat64FromValues(2))
	assert.Equal(t, []bool{true, false}, boolValues(bf))
	assert.Error(t, BoolFloatsFuncOut(func(x, y float64) bool { return true }, NewFloat64(2), NewFloat64(3), NewBool()))
	bi := BoolIntsFunc(func(x, y int) bool { return x == y }, NewIntFromValues(1, 2, 3), NewIntFromValues(2))
	assert.Equal(t, []bool{false, true, false}, boolValues(bi))
	assert.Error(t, BoolIntsFuncOut(func(x, y int) bool { return true }, NewInt(2), NewInt(3), NewBool()))
}

// TestCSV tests writing and reading tensors as CSV / TSV, via files,
// an fs.FS, and in memory, including precision metadata.
func TestCSV(t *testing.T) {
	assert.Equal(t, '\t', Tab.Rune())
	assert.Equal(t, ',', Comma.Rune())
	assert.Equal(t, ' ', Space.Rune())
	assert.Equal(t, '\t', Detect.Rune())

	a := NewFloat64FromValues(1.23456, 2, 3, 4, 5, 6)
	a.SetShapeSizes(3, 2)
	var sb strings.Builder
	require.NoError(t, WriteCSV(a, &sb, Comma))
	assert.Equal(t, "1.23456,2\n3,4\n5,6\n", sb.String())
	SetPrecision(a, 2)
	p, err := Precision(a)
	assert.NoError(t, err)
	assert.Equal(t, 2, p)
	sb.Reset()
	require.NoError(t, WriteCSV(a, &sb, Tab))
	assert.Equal(t, "1.2\t2\n3\t4\n5\t6\n", sb.String())

	b := NewFloat64(3, 2)
	require.NoError(t, ReadCSV(b, strings.NewReader("1,2\n3,4\n5,6\n7,8\n"), Comma))
	assert.Equal(t, []float64{1, 2, 3, 4, 5, 6}, b.Values)
	assert.NoError(t, ReadCSV(b, strings.NewReader(""), Comma))
	assert.Error(t, ReadCSV(b, strings.NewReader("1,2\n3\n"), Comma))

	s := NewStringFromValues("a b", "c")
	sb.Reset()
	require.NoError(t, WriteCSV(s, &sb, Space))
	assert.Equal(t, "\"a b\"\nc\n", sb.String())

	dir := t.TempDir()
	fn := dir + "/a.tsv"
	require.NoError(t, SaveCSV(a, fsx.Filename(fn), Tab))
	c := NewFloat64(3, 2)
	require.NoError(t, OpenCSV(c, fsx.Filename(fn), Tab))
	assert.Equal(t, []float64{1.2, 2, 3, 4, 5, 6}, c.Values)
	assert.Error(t, OpenCSV(c, fsx.Filename(dir+"/missing.tsv"), Tab))
	assert.Error(t, SaveCSV(a, fsx.Filename(dir+"/nodir/a.tsv"), Tab))

	d := NewFloat64(3, 2)
	require.NoError(t, OpenFS(d, os.DirFS(dir), "a.tsv", Tab))
	assert.Equal(t, c.Values, d.Values)
	assert.Error(t, OpenFS(d, os.DirFS(dir), "missing.tsv", Tab))
}

// TestEnums tests the generated enum methods for Delims and SlicesMagic.
func TestEnums(t *testing.T) {
	var d Delims
	require.NoError(t, d.SetString("Comma"))
	assert.Equal(t, Comma, d)
	assert.Equal(t, "Comma", d.String())
	assert.Error(t, d.SetString("bogus"))
	d.SetInt64(2)
	assert.Equal(t, int64(2), d.Int64())
	assert.NotEmpty(t, d.Desc())
	assert.Len(t, DelimsValues(), 4)
	assert.Len(t, d.Values(), 4)
	txt, err := d.MarshalText()
	require.NoError(t, err)
	var d2 Delims
	require.NoError(t, d2.UnmarshalText(txt))
	assert.Equal(t, d, d2)

	var m SlicesMagic
	require.NoError(t, m.SetString("Ellipsis"))
	assert.Equal(t, Ellipsis, m)
	assert.Equal(t, "Ellipsis", m.String())
	m.SetInt64(int64(NewAxis))
	assert.Equal(t, int64(NewAxis), m.Int64())
	assert.NotEmpty(t, m.Desc())
	assert.Len(t, SlicesMagicValues(), 3)
	assert.Len(t, m.Values(), 3)
	txt, err = m.MarshalText()
	require.NoError(t, err)
	var m2 SlicesMagic
	require.NoError(t, m2.UnmarshalText(txt))
	assert.Equal(t, m, m2)
}
