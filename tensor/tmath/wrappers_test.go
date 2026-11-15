// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tmath

import (
	"math"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMathWrappers tests the allocating (non-Out) unary math functions,
// which return a new float64 tensor of the same shape as the input.
func TestMathWrappers(t *testing.T) {
	in := tensor.NewFloat64FromValues(0.1, 0.25, 0.5, 0.75, 0.9, 0.3)
	in.SetShapeSizes(2, 3)
	mfuncs := []func(float64) float64{math.Abs, math.Acos, math.Acosh, math.Asin, math.Asinh, math.Atan, math.Atanh, math.Cbrt, math.Ceil, math.Cos, math.Cosh, math.Erf, math.Erfc, math.Erfcinv, math.Erfinv, math.Exp, math.Exp2, math.Expm1, math.Floor, math.Gamma, math.J0, math.J1, math.Log, math.Log10, math.Log1p, math.Log2, math.Logb, math.Round, math.RoundToEven, math.Sin, math.Sinh, math.Sqrt, math.Tan, math.Tanh, math.Trunc, math.Y0, math.Y1}
	tfuncs := []func(tensor.Tensor) tensor.Values{Abs, Acos, Acosh, Asin, Asinh, Atan, Atanh, Cbrt, Ceil, Cos, Cosh, Erf, Erfc, Erfcinv, Erfinv, Exp, Exp2, Expm1, Floor, Gamma, J0, J1, Log, Log10, Log1p, Log2, Logb, Round, RoundToEven, Sin, Sinh, Sqrt, Tan, Tanh, Trunc, Y0, Y1}
	require.Equal(t, len(mfuncs), len(tfuncs))
	for fi, tf := range tfuncs {
		out := tf(in)
		assert.Equal(t, []int{2, 3}, out.ShapeSizes())
		for i := range in.Len() {
			testEqual(t, mfuncs[fi](in.Float1D(i)), out.Float1D(i))
		}
	}
}

// TestMathBinaryWrappers tests the allocating binary math functions,
// including broadcasting of a scalar second argument.
func TestMathBinaryWrappers(t *testing.T) {
	x := tensor.NewFloat64FromValues(-2, -0.5, 0.5, 3)
	y := tensor.NewFloat64Scalar(1.5)
	mfuncs := []func(a, b float64) float64{math.Atan2, math.Copysign, math.Dim, math.Hypot, math.Max, math.Min, math.Nextafter, math.Pow, math.Remainder}
	tfuncs := []func(a, b tensor.Tensor) tensor.Values{Atan2, Copysign, Dim, Hypot, Max, Min, Nextafter, Pow, Remainder}
	require.Equal(t, len(mfuncs), len(tfuncs))
	for fi, tf := range tfuncs {
		out := tf(x, y)
		assert.Equal(t, 4, out.Len())
		for i := range x.Len() {
			testEqual(t, mfuncs[fi](x.Float1D(i), 1.5), out.Float1D(i))
		}
	}
}

// TestOpsWrappers tests the allocating arithmetic operators and the
// assignment operators not covered elsewhere, for floats, ints and strings.
func TestOpsWrappers(t *testing.T) {
	a := tensor.NewFloat64FromValues(1, 2, 3, 4)
	b := tensor.NewFloat64FromValues(4, 3, 2, 1)
	assert.Equal(t, []float64{5, 5, 5, 5}, tensor.AsFloat64(Add(a, b)).Values)
	assert.Equal(t, []float64{-3, -1, 1, 3}, tensor.AsFloat64(Sub(a, b)).Values)
	assert.Equal(t, []float64{4, 6, 6, 4}, tensor.AsFloat64(Mul(a, b)).Values)
	assert.Equal(t, []float64{0.25, 2.0 / 3, 1.5, 4}, tensor.AsFloat64(Div(a, b)).Values)
	assert.Equal(t, []float64{1, 2, 1, 0}, tensor.AsFloat64(Mod(a, b)).Values)
	assert.Equal(t, []float64{-1, -2, -3, -4}, tensor.AsFloat64(Negate(a)).Values)

	// integer division is floating point
	ia := tensor.NewIntFromValues(1, 3)
	ib := tensor.NewIntFromValues(2, 2)
	dv := Div(ia, ib)
	assert.Equal(t, []float64{0.5, 1.5}, tensor.AsFloat64(dv).Values)

	// string concatenation
	sa := tensor.NewStringFromValues("a", "b")
	sb := tensor.NewStringFromValues("x", "y")
	assert.Equal(t, []string{"ax", "by"}, tensor.AsStringSlice(Add(sa, sb)))
	require.NoError(t, AddAssign(sa, sb))
	assert.Equal(t, []string{"ax", "by"}, tensor.AsStringSlice(sa))

	c := a.Clone()
	require.NoError(t, Assign(c, tensor.NewFloat64Scalar(7)))
	assert.Equal(t, []float64{7, 7, 7, 7}, tensor.AsFloat64(c).Values)
	require.NoError(t, ModAssign(c, tensor.NewFloat64Scalar(4)))
	assert.Equal(t, []float64{3, 3, 3, 3}, tensor.AsFloat64(c).Values)
	require.NoError(t, Inc(c))
	require.NoError(t, Dec(c))
	require.NoError(t, Dec(c))
	assert.Equal(t, []float64{2, 2, 2, 2}, tensor.AsFloat64(c).Values)
}

// TestBoolWrappers tests the allocating comparison and logical
// functions, for both numbers and strings.
func TestBoolWrappers(t *testing.T) {
	ar := tensor.NewIntFromValues(0, 1, 2, 3)
	sc := tensor.NewIntScalar(2)
	boolVals := func(b *tensor.Bool) []bool {
		r := make([]bool, b.Len())
		for i := range r {
			r[i] = b.Bool1D(i)
		}
		return r
	}
	assert.Equal(t, []bool{false, false, true, false}, boolVals(Equal(ar, sc)))
	assert.Equal(t, []bool{true, true, false, true}, boolVals(NotEqual(ar, sc)))
	assert.Equal(t, []bool{true, true, false, false}, boolVals(Less(ar, sc)))
	assert.Equal(t, []bool{false, false, false, true}, boolVals(Greater(ar, sc)))
	assert.Equal(t, []bool{true, true, true, false}, boolVals(LessEqual(ar, sc)))
	assert.Equal(t, []bool{false, false, true, true}, boolVals(GreaterEqual(ar, sc)))

	x := tensor.NewIntFromValues(0, 0, 1, 1)
	y := tensor.NewIntFromValues(0, 1, 0, 1)
	assert.Equal(t, []bool{false, true, true, true}, boolVals(Or(x, y)))
	assert.Equal(t, []bool{false, false, false, true}, boolVals(And(x, y)))
	assert.Equal(t, []bool{true, true, false, false}, boolVals(Not(x)))

	sa := tensor.NewStringFromValues("a", "b", "c")
	sb := tensor.NewStringScalar("b")
	assert.Equal(t, []bool{false, true, false}, boolVals(Equal(sa, sb)))
	assert.Equal(t, []bool{true, false, true}, boolVals(NotEqual(sa, sb)))
	assert.Equal(t, []bool{true, false, false}, boolVals(Less(sa, sb)))
	assert.Equal(t, []bool{false, false, true}, boolVals(Greater(sa, sb)))
	assert.Equal(t, []bool{true, true, false}, boolVals(LessEqual(sa, sb)))
	assert.Equal(t, []bool{false, true, true}, boolVals(GreaterEqual(sa, sb)))
}

// TestModAssignVarTypes tests ModAssignVar on float and tensor variables,
// and the error for a non-pointer argument.
func TestModAssignVarTypes(t *testing.T) {
	f := 7.5
	require.NoError(t, ModAssignVar(&f, 2))
	assert.Equal(t, 1.5, f)
	var tv tensor.Tensor = tensor.NewFloat64FromValues(5, 7)
	require.NoError(t, ModAssignVar(&tv, 3))
	assert.Equal(t, []float64{2, 1}, tensor.AsFloat64(tv).Values)
	assert.Error(t, ModAssignVar(f, 2))
	s := "x"
	assert.Error(t, ModAssignVar(&s, 2))
}
