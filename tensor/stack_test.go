// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcatenate(t *testing.T) {
	a := NewIntFromValues(0, 1, 2, 3)
	a.SetShapeSizes(2, 2)
	b := NewIntFromValues(4, 5)
	b.SetShapeSizes(1, 2)
	c := Concatenate(0, a, b)
	assert.Equal(t, []int{3, 2}, c.ShapeSizes())
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, AsIntSlice(c))
	_, isInt := c.(*Int)
	assert.True(t, isInt, "the type of the tensors")

	c = Concatenate(-1, a, a)
	assert.Equal(t, []int{2, 4}, c.ShapeSizes())
	assert.Equal(t, []int{0, 1, 0, 1, 2, 3, 2, 3}, AsIntSlice(c))
	assert.Panics(t, func() { Concatenate(1, a, b) }, "sizes do not match")

	f := Concatenate(0, NewIntFromValues(1), NewFloat64FromValues(1.5))
	assert.Equal(t, []float64{1, 1.5}, f.(*Float64).Values, "mixed types are float64")
	s := Concatenate(0, NewIntFromValues(1), NewStringFromValues("x"))
	assert.Equal(t, []string{"1", "x"}, s.(*String).Values)
}

func TestStacks(t *testing.T) {
	x := NewIntFromValues(1, 2)
	y := NewIntFromValues(3, 4)
	st := Stack(0, x, y)
	assert.Equal(t, []int{2, 2}, st.ShapeSizes())
	assert.Equal(t, []int{1, 2, 3, 4}, AsIntSlice(st))
	st = Stack(1, x, y)
	assert.Equal(t, []int{1, 3, 2, 4}, AsIntSlice(st))

	assert.Equal(t, []int{2, 2}, VStack(x, y).ShapeSizes())
	assert.Equal(t, []int{1, 2, 3, 4}, AsIntSlice(VStack(x, y)))
	assert.Equal(t, []int{4}, HStack(x, y).ShapeSizes())
	cs := ColumnStack(x, y)
	assert.Equal(t, []int{2, 2}, cs.ShapeSizes())
	assert.Equal(t, []int{1, 3, 2, 4}, AsIntSlice(cs))
	hs := HStack(cs, cs)
	assert.Equal(t, []int{2, 4}, hs.ShapeSizes())
	assert.Equal(t, []int{1, 3, 1, 3, 2, 4, 2, 4}, AsIntSlice(hs))
}

func TestTile(t *testing.T) {
	x := NewIntFromValues(1, 2)
	tl := Tile(x, 2)
	assert.Equal(t, []int{1, 2, 1, 2}, AsIntSlice(tl))
	tl = Tile(x, 2, 2)
	assert.Equal(t, []int{2, 4}, tl.ShapeSizes())
	assert.Equal(t, []int{1, 2, 1, 2, 1, 2, 1, 2}, AsIntSlice(tl))
	m := NewIntFromValues(1, 2, 3, 4)
	m.SetShapeSizes(2, 2)
	tl = Tile(m, 2)
	assert.Equal(t, []int{2, 4}, tl.ShapeSizes(), "repeats of the last dimension")
	assert.Equal(t, []int{1, 2, 1, 2, 3, 4, 3, 4}, AsIntSlice(tl))
}

func TestMeshgrid(t *testing.T) {
	xs, ys := Meshgrid(NewFloat64FromValues(1, 2, 3), NewFloat64FromValues(10, 20))
	assert.Equal(t, []int{2, 3}, xs.ShapeSizes())
	assert.Equal(t, []float64{1, 2, 3, 1, 2, 3}, AsFloat64Slice(xs))
	assert.Equal(t, []float64{10, 10, 10, 20, 20, 20}, AsFloat64Slice(ys))
}

func TestSorted(t *testing.T) {
	m := NewFloat64FromValues(3, math.NaN(), 1, 9, 7, 8)
	m.SetShapeSizes(2, 3)
	assert.Equal(t, []int{2, 0, 1, 1, 2, 0}, ArgSort(m).Values, "NaN last")
	s := Sorted(m)
	assert.Equal(t, []float64{1, 3}, AsFloat64Slice(s)[:2])
	assert.True(t, math.IsNaN(s.Float1D(2)))
	assert.Equal(t, []float64{7, 8, 9}, AsFloat64Slice(s)[3:])
	s = Sorted(m, 0) // each column
	sv := AsFloat64Slice(s)
	assert.Equal(t, []float64{3, 7, 1, 9}, sv[:4])
	assert.True(t, math.IsNaN(sv[4]))
	assert.Equal(t, 8.0, sv[5])

	str := NewStringFromValues("b", "c", "a")
	assert.Equal(t, []string{"a", "b", "c"}, Sorted(str).(*String).Values)
}

func TestArgWhere(t *testing.T) {
	b := NewBoolShape(NewShape(2, 2))
	b.SetBool1D(true, 1)
	b.SetBool1D(true, 2)
	aw := ArgWhere(b)
	assert.Equal(t, []int{2, 2}, aw.ShapeSizes())
	assert.Equal(t, []int{0, 1, 1, 0}, aw.Values)
	aw = ArgWhere(NewFloat64FromValues(0, 3, 0, 5))
	assert.Equal(t, []int{2, 1}, aw.ShapeSizes())
	assert.Equal(t, []int{1, 3}, aw.Values)
}

func TestResliceIndexes(t *testing.T) {
	a := NewIntFromValues(0, 1, 2, 3, 4, 5)
	a.SetShapeSizes(2, 3)
	rs := Reslice(a, FullAxis, NewIntFromValues(2, 0))
	assert.Equal(t, []int{2, 2}, rs.ShapeSizes())
	assert.Equal(t, []int{2, 0, 5, 3}, AsIntSlice(rs))
	mask := NewBoolShape(NewShape(3))
	mask.SetBool1D(true, 1)
	rs = Reslice(a, FullAxis, mask)
	assert.Equal(t, []int{2, 1}, rs.ShapeSizes())
	assert.Equal(t, []int{1, 4}, AsIntSlice(rs))
	rs = Reslice(a, NewIntFromValues(1), FullAxis)
	assert.Equal(t, []int{3}, rs.ShapeSizes(), "one value is an index")
	assert.Equal(t, []int{3, 4, 5}, AsIntSlice(rs))
	assert.Equal(t, []int{5}, AsIntSlice(AnySlice(a, NewIntFromValues(1, 2))), "an index of each dimension")
}

func TestPivot(t *testing.T) {
	rows := NewStringFromValues("a", "a", "b")
	cols := NewIntFromValues(2000, 1990, 1990)
	vals := NewFloat64FromValues(1, 2, 3)
	m, rl, cl := Pivot(rows, cols, vals)
	assert.Equal(t, []string{"a", "b"}, rl.(*String).Values)
	assert.Equal(t, []int{1990, 2000}, cl.(*Int).Values)
	assert.Equal(t, []int{2, 2}, m.ShapeSizes())
	assert.Equal(t, []float64{2, 1, 3}, m.Values[:3])
	assert.True(t, math.IsNaN(m.Values[3]), "no value")
}
