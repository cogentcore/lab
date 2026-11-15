// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package randx

import (
	"math"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSysRandSeeded tests that a seeded SysRand is reproducible and
// that every method returns values in its documented range.
func TestSysRandSeeded(t *testing.T) {
	a := NewSysRand(42)
	b := NewSysRand(42)
	assert.Equal(t, a.Int63(), b.Int63())
	assert.Equal(t, a.Uint32(), b.Uint32())
	assert.Equal(t, a.Uint64(), b.Uint64())
	assert.Equal(t, a.Int31(), b.Int31())
	assert.Equal(t, a.Int(), b.Int())
	assert.Equal(t, a.Float64(), b.Float64())
	assert.Equal(t, a.Float32(), b.Float32())
	assert.Equal(t, a.NormFloat64(), b.NormFloat64())
	assert.Equal(t, a.ExpFloat64(), b.ExpFloat64())
	assert.Equal(t, a.Perm(10), b.Perm(10))

	// Init and Seed both restart the sequence.
	a.Init(7)
	v1 := a.Int63()
	a.Seed(7)
	assert.Equal(t, v1, a.Int63())

	// Init on an empty SysRand creates the source.
	var c SysRand
	c.Init(7)
	require.NotNil(t, c.Rand)
	assert.Equal(t, v1, c.Int63())

	checkRanges(t, a)
}

// TestSysRandGlobal tests the methods of a SysRand that uses the
// global source (nil Rand), checking only their ranges.
func TestSysRandGlobal(t *testing.T) {
	g := NewGlobalRand()
	assert.Nil(t, g.Rand)
	checkRanges(t, g)
	g.Int63()
	g.Uint32()
	g.Uint64()
	g.Int31()
	g.Int()
	g.NormFloat64()
	g.ExpFloat64()
	assert.Len(t, g.Perm(5), 5)
	g.Init(3) // creates a source
	assert.NotNil(t, g.Rand)
}

// checkRanges checks the bounded generators of r.
func checkRanges(t *testing.T, r Rand) {
	for range 100 {
		assert.Less(t, r.Int63n(5), int64(5))
		assert.Less(t, r.Int31n(5), int32(5))
		n := r.Intn(5)
		assert.True(t, n >= 0 && n < 5)
		f := r.Float64()
		assert.True(t, f >= 0 && f < 1)
		f32 := r.Float32()
		assert.True(t, f32 >= 0 && f32 < 1)
	}
	p := r.Perm(6)
	sort.Ints(p)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, p)
	s := []int{0, 1, 2, 3}
	r.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	sort.Ints(s)
	assert.Equal(t, []int{0, 1, 2, 3}, s)
}

// TestInitSysRand tests that InitSysRand creates a Rand when nil and
// reseeds an existing one.
func TestInitSysRand(t *testing.T) {
	var r Rand
	InitSysRand(&r, 5)
	require.NotNil(t, r)
	v := r.Int63()
	InitSysRand(&r, 5)
	assert.Equal(t, v, r.Int63())
}

// TestBoolP tests BoolP and BoolP32 at the extreme probabilities and
// their approximate rate at 0.5.
func TestBoolP(t *testing.T) {
	r := NewSysRand(1)
	n := 0
	for range 1000 {
		assert.False(t, BoolP(0, r))
		assert.True(t, BoolP(1, r))
		assert.False(t, BoolP32(0, r))
		assert.True(t, BoolP32(1, r))
		if BoolP(0.5, r) {
			n++
		}
	}
	assert.InDelta(t, 500, n, 100)
	assert.False(t, BoolP(0))
	assert.True(t, BoolP32(1.1))
}

// TestPChoose tests PChoose32 and PChoose64 with degenerate and
// spread-out probability tables.
func TestPChoose(t *testing.T) {
	r := NewSysRand(1)
	assert.Equal(t, 2, PChoose32([]float32{0, 0, 1}, r))
	assert.Equal(t, 1, PChoose64([]float64{0, 1, 0}, r))
	// probabilities that sum to less than 1 fall through to the last.
	assert.Equal(t, 1, PChoose64([]float64{0, 0}, r))
	assert.Equal(t, 1, PChoose32([]float32{0, 0}, r))
	counts := make([]int, 3)
	for range 3000 {
		counts[PChoose64([]float64{0.2, 0.3, 0.5}, r)]++
	}
	assert.InDelta(t, 600, counts[0], 120)
	assert.InDelta(t, 1500, counts[2], 150)
	assert.Equal(t, 0, PChoose32([]float32{1}))
	assert.Equal(t, 0, PChoose64([]float64{1}))
}

// TestPermute tests SequentialInts and that the Permute functions
// keep the same elements.
func TestPermute(t *testing.T) {
	r := NewSysRand(3)
	ins := make([]int, 20)
	SequentialInts(ins, 5)
	assert.Equal(t, 5, ins[0])
	assert.Equal(t, 24, ins[19])
	orig := slices.Clone(ins)
	PermuteInts(ins, r)
	assert.NotEqual(t, orig, ins)
	sort.Ints(ins)
	assert.Equal(t, orig, ins)
	PermuteInts(ins)
	sort.Ints(ins)
	assert.Equal(t, orig, ins)

	ss := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	PermuteStrings(ss, r)
	PermuteStrings(ss)
	slices.Sort(ss)
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f", "g", "h"}, ss)

	f32 := []float32{1, 2, 3, 4, 5}
	PermuteFloat32s(f32, r)
	PermuteFloat32s(f32)
	slices.Sort(f32)
	assert.Equal(t, []float32{1, 2, 3, 4, 5}, f32)

	f64 := []float64{1, 2, 3, 4, 5}
	PermuteFloat64s(f64, r)
	PermuteFloat64s(f64)
	slices.Sort(f64)
	assert.Equal(t, []float64{1, 2, 3, 4, 5}, f64)
}

// TestIntAndUniform tests the integer and uniform range generators.
func TestIntAndUniform(t *testing.T) {
	r := NewSysRand(9)
	for range 200 {
		v := IntZeroN(4, r)
		assert.True(t, v >= 0 && v < 4)
		v = IntMinMax(3, 6, r)
		assert.True(t, v >= 3 && v < 6)
		v = IntMeanRange(10, 2, r)
		assert.True(t, v >= 8 && v <= 12)
		f := ZeroOne(r)
		assert.True(t, f >= 0 && f < 1)
		f = UniformMinMax(-2, 2, r)
		assert.True(t, f >= -2 && f < 2)
		f = UniformMeanRange(5, 1, r)
		assert.True(t, f >= 4 && f < 6)
	}
	assert.Less(t, IntZeroN(3), int64(3))
	assert.Less(t, IntMinMax(1, 2), int64(2))
	assert.InDelta(t, 0, IntMeanRange(0, 1), 1)
	assert.Less(t, ZeroOne(), 1.0)
	assert.Less(t, UniformMinMax(0, 1), 1.0)
	assert.InDelta(t, 0, UniformMeanRange(0, 1), 1)
}

// TestRandParams tests Defaults, ShouldDisplay and Gen for every
// distribution.
func TestRandParams(t *testing.T) {
	var rp RandParams
	rp.Defaults()
	assert.Equal(t, 1.0, rp.Var)
	assert.Equal(t, 1.0, rp.Par)
	assert.False(t, rp.ShouldDisplay("Par"))
	assert.True(t, rp.ShouldDisplay("Mean"))
	rp.Dist = Gamma
	assert.True(t, rp.ShouldDisplay("Par"))

	r := NewSysRand(11)
	means := map[RandDists]float64{}
	for _, d := range RandDistsValues() {
		rp := RandParams{Dist: d, Mean: 2, Var: 0.5, Par: 4}
		if d == Binomial {
			rp.Var = 0.5 // p
			rp.Par = 10  // n
		}
		sum := 0.0
		const n = 4000
		for range n {
			v := rp.Gen(r)
			assert.False(t, math.IsNaN(v))
			sum += v
		}
		means[d] = sum / n
	}
	assert.InDelta(t, 2, means[Uniform], 0.05)
	assert.InDelta(t, 2+10*0.5, means[Binomial], 0.2)
	assert.InDelta(t, 2+0.5, means[Poisson], 0.1)
	assert.InDelta(t, 2+4/0.5, means[Gamma], 0.3) // mean = alpha/beta (rate)
	assert.InDelta(t, 2, means[Gaussian], 0.05)
	assert.InDelta(t, 2+0.5/(0.5+4), means[Beta], 0.05)
	assert.Equal(t, 2.0, means[Mean])

	rp = RandParams{Dist: Mean, Mean: 3}
	assert.Equal(t, 3.0, rp.Gen())
}

// TestBinomialBranches tests the three BinomialGen algorithms (direct,
// Poisson rejection and Cauchy rejection), including p > 0.5.
func TestBinomialBranches(t *testing.T) {
	r := NewSysRand(5)
	cases := []struct{ n, p float64 }{
		{10, 0.7},    // direct, flipped p
		{100, 0.005}, // am < 1: Poisson rejection
		{100, 0.995}, // same, flipped
		{100, 0.3},   // Cauchy rejection
		{100, 0.8},   // same, flipped
	}
	for _, c := range cases {
		sum := 0.0
		const ns = 4000
		for range ns {
			v := BinomialGen(c.n, c.p, r)
			assert.True(t, v >= 0 && v <= c.n, "n=%g p=%g v=%g", c.n, c.p, v)
			sum += v
		}
		mean := c.n * c.p
		sd := math.Sqrt(c.n * c.p * (1 - c.p))
		assert.InDelta(t, mean, sum/ns, 5*sd/math.Sqrt(ns)+0.01, "n=%g p=%g", c.n, c.p)
	}
}

// TestGammaBranches tests the GammaGen algorithms for alpha == 1,
// small alpha, alpha < 1, and the panics on invalid parameters.
func TestGammaBranches(t *testing.T) {
	r := NewSysRand(5)
	for _, a := range []float64{1, 0.1, 0.5, 3} {
		sum := 0.0
		const ns = 8000
		for range ns {
			v := GammaGen(a, 2, r)
			assert.True(t, v >= 0)
			sum += v
		}
		assert.InDelta(t, a/2, sum/ns, 0.05, "alpha=%g", a)
	}
	assert.Panics(t, func() { GammaGen(1, 0, r) })
	assert.Panics(t, func() { GammaGen(0, 1, r) })
	assert.Greater(t, GammaGen(2, 1), 0.0)
	assert.Greater(t, PoissonGen(20, r), -1.0)
}

// TestSeeds tests Seeds Init, Set and NewSeeds.
func TestSeeds(t *testing.T) {
	var s Seeds
	s.Init(3)
	assert.Equal(t, Seeds{1, 2, 3}, s)
	r := NewSysRand(100)
	s.Set(1, r)
	v := r.Int63()
	assert.Equal(t, NewSysRand(2).Int63(), v)
	s.NewSeeds()
	assert.Equal(t, s[0]+1, s[1])
	assert.Equal(t, s[0]+2, s[2])
}

// TestRandDistsEnum tests the generated enum methods of RandDists.
func TestRandDistsEnum(t *testing.T) {
	assert.Equal(t, "Gaussian", Gaussian.String())
	var d RandDists
	require.NoError(t, d.SetString("Beta"))
	assert.Equal(t, Beta, d)
	assert.Error(t, d.SetString("Nope"))
	assert.Equal(t, int64(Beta), d.Int64())
	d.SetInt64(int64(Poisson))
	assert.Equal(t, Poisson, d)
	assert.NotEmpty(t, Gamma.Desc())
	assert.Len(t, d.Values(), len(RandDistsValues()))
	b, err := Uniform.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "Uniform", string(b))
	require.NoError(t, d.UnmarshalText([]byte("Mean")))
	assert.Equal(t, Mean, d)
}
