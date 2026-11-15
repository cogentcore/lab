// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slrand

import (
	"testing"

	"cogentcore.org/lab/gosl/sltype"
	"github.com/stretchr/testify/assert"
)

// TestKAT is the Known Answer Test for values from the DEShawREsearch reference impl
func TestKAT(t *testing.T) {
	kats := []struct {
		ctr sltype.Uint32Vec2
		key uint32
		res sltype.Uint32Vec2
	}{{sltype.Uint32Vec2{X: 0, Y: 0}, 0, sltype.Uint32Vec2{X: 0xff1dae59, Y: 0x6cd10df2}},
		{sltype.Uint32Vec2{X: 0xffffffff, Y: 0xffffffff}, 0xffffffff, sltype.Uint32Vec2{X: 0x2c3f628b, Y: 0xab4fd7ad}},
		{sltype.Uint32Vec2{X: 0x243f6a88, Y: 0x85a308d3}, 0x13198a2e, sltype.Uint32Vec2{X: 0xdd7ce038, Y: 0xf62a4c12}}}

	for _, tv := range kats {
		r := Philox2x32(sltype.Uint64FromLoHi(tv.ctr), tv.key)
		if r != tv.res {
			t.Errorf("ctr: %v  key: %d != result: %v -- got: %v", tv.ctr, tv.key, tv.res, r)
		}
	}
}

// TestRand tests known values of Float32, Float32Range11 and Float32Norm.
func TestRand(t *testing.T) {
	trgs := [][]float32{
		{0.9965466, -0.84816515, 1.2053814},
		{0.86274576, -0.25368667, -0.13640901},
		{0.018057441, -0.5596625, 0.61585814},
		{0.010364834, 0.38940117, -1.4226444},
		{0.75196105, 0.57544005, -0.68210393},
		{0.23327535, 0.4237375, 0.2814367},
		{0.50003797, 0.82759297, -0.5035681},
		{0.6322405, -0.21457514, 0.74390095},
		{0.59605914, 0.9313429, -0.06783484},
		{0.57019144, -0.32633832, 0.7762274},
	}

	var counter uint64
	for i := range uint32(10) {
		f, f11, fn := Float32(counter, i, 0), Float32Range11(counter, i, 1), Float32Norm(counter, i, 2)
		// fmt.Printf("{%g, %g, %g},\n", f, f11, fn)
		assert.Equal(t, trgs[i][0], f)
		assert.Equal(t, trgs[i][1], f11)
		assert.Equal(t, trgs[i][2], fn)
	}
}

// TestIntn tests that Uint32N stays in range.
func TestIntn(t *testing.T) {
	var counter uint64
	n := uint32(20)
	for i := range uint32(1000) {
		r := Uint32N(counter, i, 0, n)
		if r >= n {
			t.Errorf("r >= n: %d\n", r)
		}
		// fmt.Printf("%d\t%d\n", i, r)
	}
}
