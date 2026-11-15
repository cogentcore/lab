// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"math"
	"testing"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// statCase is one statistic with its output function and expected value on
// the values 1, 2, 4, 5 (with NaN values ignored), and the factor that
// the result scales by when the data are doubled.
type statCase struct {
	stat  Stats
	fun   StatsOutFunc
	want  float64
	scale float64
	only1 bool // only 1D inputs are supported (quantiles)
}

func statCases() []statCase {
	v3 := 10.0 / 3.0
	return []statCase{
		{StatCount, CountOut, 4, 1, false},
		{StatSum, SumOut, 12, 2, false},
		{StatL1Norm, L1NormOut, 12, 2, false},
		{StatProd, ProdOut, 40, 16, false},
		{StatMin, MinOut, 1, 2, false},
		{StatMax, MaxOut, 5, 2, false},
		{StatMinAbs, MinAbsOut, 1, 2, false},
		{StatMaxAbs, MaxAbsOut, 5, 2, false},
		{StatMean, MeanOut, 3, 2, false},
		{StatVar, VarOut, v3, 4, false},
		{StatStd, StdOut, math.Sqrt(v3), 2, false},
		{StatSem, SemOut, math.Sqrt(v3) / 2, 2, false},
		{StatSumSq, SumSqOut, 46, 4, false},
		{StatL2Norm, L2NormOut, math.Sqrt(46), 2, false},
		{StatVarPop, VarPopOut, 2.5, 4, false},
		{StatStdPop, StdPopOut, math.Sqrt(2.5), 2, false},
		{StatSemPop, SemPopOut, math.Sqrt(2.5) / 2, 2, false},
		{StatMedian, MedianOut, 3, 2, true},
		{StatQ1, Q1Out, 1.75, 2, true},
		{StatQ3, Q3Out, 4.25, 2, true},
		{StatFirst, FirstOut, 1, 2, false},
		{StatFinal, FinalOut, 5, 2, false},
	}
}

// TestStatsKnownValues tests every statistic on known values, for
// Float64, Float32, Int and Rows inputs, in both 1D and 2D (cells) form,
// with NaN values skipped as missing.
func TestStatsKnownValues(t *testing.T) {
	nan := math.NaN()
	f64 := tensor.NewFloat64FromValues(1, 2, nan, 4, 5)
	f32 := tensor.NewFloat32FromValues(1, 2, float32(nan), 4, 5)
	ints := tensor.NewIntFromValues(1, 2, 4, 5)
	rows := tensor.NewRows(f64)

	// 2D: column 0 = x, column 1 = 2x
	x := []float64{1, 2, nan, 4, 5}
	c64 := tensor.NewFloat64(5, 2)
	c32 := tensor.NewFloat32(5, 2)
	ci := tensor.NewInt(4, 2)
	ri := 0
	for i, v := range x {
		c64.SetFloatRow(v, i, 0)
		c64.SetFloatRow(2*v, i, 1)
		c32.SetFloatRow(v, i, 0)
		c32.SetFloatRow(2*v, i, 1)
		if !math.IsNaN(v) {
			ci.SetFloatRow(v, ri, 0)
			ci.SetFloatRow(2*v, ri, 1)
			ri++
		}
	}
	tol := 1.0e-6
	for _, sc := range statCases() {
		name := sc.stat.String()
		assert.Equal(t, sc.stat.FuncName(), "stats."+name)
		assert.InDelta(t, sc.want, sc.stat.Call(f64).Float1D(0), tol, name)
		for _, in := range []tensor.Tensor{f64, f32, ints, rows} {
			out := tensor.NewFloat64()
			require.NoError(t, sc.fun(in, out), name)
			assert.InDelta(t, sc.want, out.Float1D(0), tol, name)
		}
		if sc.only1 {
			continue
		}
		for _, in := range []tensor.Tensor{c64, c32, ci, tensor.NewRows(c64)} {
			out := tensor.NewFloat64()
			require.NoError(t, sc.fun(in, out), name)
			assert.Equal(t, []int{2}, out.ShapeSizes(), name)
			assert.InDelta(t, sc.want, out.Float1D(0), tol, name)
			assert.InDelta(t, sc.scale*sc.want, out.Float1D(1), tol, name)
		}
	}
}

// TestStatsEmpty tests that stats on empty inputs do not panic
// and return zero-valued outputs.
func TestStatsEmpty(t *testing.T) {
	for _, sc := range statCases() {
		out := tensor.NewFloat64()
		in := tensor.NewFloat64(0)
		assert.NoError(t, sc.fun(in, out), sc.stat.String())
		if out.Len() > 0 {
			v := out.Float1D(0)
			assert.True(t, v == 0 || math.IsNaN(v), sc.stat.String())
		}
	}
	out := tensor.NewFloat64()
	assert.NoError(t, FirstOut(tensor.NewFloat64(0, 3), out))
	assert.Equal(t, 3, out.Len())
	assert.NoError(t, FinalOut(tensor.NewFloat64(0, 3), out))
	assert.Equal(t, 3, out.Len())
}

// TestQuantiles tests multiple quantiles at once, including out-of-range
// quantile values, and the errors for non-1D inputs.
func TestQuantiles(t *testing.T) {
	in := tensor.NewFloat64FromValues(5, 1, 4, 2, 3)
	qs := tensor.NewFloat64FromValues(-0.5, 0, 0.25, 0.5, 1, 1.5)
	out := Quantiles(in, qs)
	assert.Equal(t, []int{6}, out.ShapeSizes())
	assert.Equal(t, []float64{1, 1, 2, 3, 5, 5}, tensor.AsFloat64(out).Values)

	assert.Equal(t, 3.0, Median(in).Float1D(0))
	assert.Equal(t, 2.0, Q1(in).Float1D(0))
	assert.Equal(t, 4.0, Q3(in).Float1D(0))

	empty := tensor.NewFloat64FromValues(math.NaN())
	assert.Equal(t, 0.0, Median(empty).Float1D(0))

	o2 := tensor.NewFloat64()
	assert.Error(t, QuantilesOut(tensor.NewFloat64(2, 2), qs, o2))
	assert.Error(t, QuantilesOut(in, tensor.NewFloat64(2, 2), o2))
}

// TestStatsHelpers tests StripPackage, AsStatsFunc, and the generated
// Stats enum helpers.
func TestStatsHelpers(t *testing.T) {
	assert.Equal(t, "Mean", StripPackage("stats.Mean"))
	assert.Equal(t, "Mean", StripPackage("Mean"))
	assert.Equal(t, "C", StripPackage("a.b.C"))

	f, err := AsStatsFunc(Mean)
	require.NoError(t, err)
	assert.Equal(t, 2.0, f(tensor.NewFloat64FromValues(1, 3)).Float1D(0))
	_, err = AsStatsFunc(func() {})
	assert.Error(t, err)

	var s Stats
	require.NoError(t, s.SetString("Median"))
	assert.Equal(t, StatMedian, s)
	assert.Error(t, s.SetString("NotAStat"))
	assert.Len(t, StatsValues(), int(StatsN))
}

// TestNormFuncs tests ZScore, UnitNorm, Clamp and Binarize on known values.
func TestNormFuncs(t *testing.T) {
	in := tensor.NewFloat64FromValues(1, 2, 3, 4, 5)
	z := ZScore(in)
	sd := math.Sqrt(2.5)
	for i := range 5 {
		assert.InDelta(t, float64(i-2)/sd, z.Float1D(i), 1.0e-9)
	}
	u := UnitNorm(in)
	assert.Equal(t, []float64{0, 0.25, 0.5, 0.75, 1}, tensor.AsFloat64(u).Values)

	c := Clamp(in, tensor.NewFloat64Scalar(2), tensor.NewFloat64Scalar(4))
	assert.Equal(t, []float64{2, 2, 3, 4, 4}, tensor.AsFloat64(c).Values)

	b := Binarize(in, tensor.NewFloat64Scalar(3))
	assert.Equal(t, []float64{0, 0, 1, 1, 1}, tensor.AsFloat64(b).Values)
}

// TestGroupsSingletons tests that Groups includes a final group with a
// single row, and a tensor with only one row, for both string and
// numeric values.
func TestGroupsSingletons(t *testing.T) {
	dir, _ := tensorfs.NewDir("Root")
	str := tensor.NewStringFromValues("b", "a", "a")
	metadata.SetName(str, "Str")
	num := tensor.NewFloat64FromValues(2, 1, 1)
	metadata.SetName(num, "Num")
	one := tensor.NewFloat64FromValues(7)
	metadata.SetName(one, "One")
	require.NoError(t, Groups(dir, str, num, one, tensor.NewFloat64(0)))

	sg := dir.Dir("Groups/Str")
	vals := sg.ValuesFunc(nil)
	require.Len(t, vals, 2)
	assert.Equal(t, "a", metadata.Name(vals[0]))
	assert.Equal(t, []int{1, 2}, tensor.AsInt(vals[0]).Values)
	assert.Equal(t, "b", metadata.Name(vals[1]))
	assert.Equal(t, []int{0}, tensor.AsInt(vals[1]).Values)

	ng := dir.Dir("Groups/Num").ValuesFunc(nil)
	require.Len(t, ng, 2)
	assert.Equal(t, []int{1, 2}, tensor.AsInt(ng[0]).Values)
	assert.Equal(t, []int{0}, tensor.AsInt(ng[1]).Values)

	og := dir.Dir("Groups/One").ValuesFunc(nil)
	require.Len(t, og, 1)
	assert.Equal(t, []int{0}, tensor.AsInt(og[0]).Values)
}

// TestGroupAll tests GroupAll followed by GroupStats over all rows,
// including a Rows view with indexes.
func TestGroupAll(t *testing.T) {
	dir, _ := tensorfs.NewDir("Root")
	vals := tensor.NewFloat64FromValues(10, 20, 30, 40)
	metadata.SetName(vals, "Value")
	view := tensor.NewRows(vals, 3, 1)
	require.NoError(t, GroupAll(dir, view))
	all := dir.Dir("Groups/All").ValuesFunc(nil)
	require.Len(t, all, 1)
	assert.Equal(t, []int{3, 1}, tensor.AsInt(all[0]).Values)

	require.NoError(t, GroupStats(dir, StatMean, vals))
	mean := dir.Dir("Stats/All/Value").Node("Mean")
	require.NotNil(t, mean)
	assert.Equal(t, 30.0, mean.Tensor.Float1D(0))

	d2, _ := tensorfs.NewDir("Empty")
	assert.NoError(t, GroupAll(d2, tensor.NewFloat64(0)))
}

// TestDescribe tests Describe, DescribeTable and DescribeTableAll.
func TestDescribe(t *testing.T) {
	dt := table.New()
	dt.AddStringColumn("Name")
	dt.AddFloat64Column("X")
	dt.AddIntColumn("Y")
	dt.SetNumRows(4)
	for i := range 4 {
		dt.Column("Name").SetStringRow("n", i, 0)
		dt.Column("X").SetFloatRow(float64(i+1), i, 0)
		dt.Column("Y").SetFloatRow(float64(2*(i+1)), i, 0)
	}
	dir, _ := tensorfs.NewDir("Root")
	cols := DescribeTableAll(dir, dt)
	assert.Equal(t, []string{"X", "Y"}, cols)
	xd := dir.Dir("Describe/X")
	assert.Equal(t, 4.0, xd.Node("Count").Tensor.Float1D(0))
	assert.Equal(t, 2.5, xd.Node("Mean").Tensor.Float1D(0))
	assert.Equal(t, 1.0, xd.Node("Min").Tensor.Float1D(0))
	assert.Equal(t, 4.0, xd.Node("Max").Tensor.Float1D(0))
	assert.Equal(t, 2.5, xd.Node("Median").Tensor.Float1D(0))
	yd := dir.Dir("Describe/Y")
	assert.Equal(t, 5.0, yd.Node("Mean").Tensor.Float1D(0)) // int not truncated

	d2, _ := tensorfs.NewDir("Root2")
	anon := tensor.NewFloat64FromValues(1, 3)
	Describe(d2, anon, tensor.NewFloat64(0))
	assert.Equal(t, 2.0, d2.Dir("Describe/0").Node("Mean").Tensor.Float1D(0))
	assert.Nil(t, d2.Dir("Describe").Node("1"))
}

// TestMeanTables tests MeanTables with tables of different lengths,
// where the longest table is not the first one.
func TestMeanTables(t *testing.T) {
	mk := func(vals ...float64) *table.Table {
		dt := table.New()
		dt.AddStringColumn("Name")
		dt.AddFloat64Column("X")
		dt.SetNumRows(len(vals))
		for i, v := range vals {
			dt.Column("Name").SetStringRow("r", i, 0)
			dt.Column("X").SetFloatRow(v, i, 0)
		}
		return dt
	}
	assert.Nil(t, MeanTables(nil))
	assert.Nil(t, MeanTables([]*table.Table{mk()}))

	a := mk(1, 2)
	b := mk(3, 4, 6)
	mt := MeanTables([]*table.Table{a, b})
	require.NotNil(t, mt)
	assert.Equal(t, 3, mt.NumRows())
	x := mt.Column("X")
	assert.Equal(t, 2.0, x.Float1D(0))
	assert.Equal(t, 3.0, x.Float1D(1))
	assert.Equal(t, 6.0, x.Float1D(2))
	// inputs are not modified
	assert.Equal(t, 3.0, b.Column("X").Float1D(0))
}
