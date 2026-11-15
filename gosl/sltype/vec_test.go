// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sltype

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInt32Vec4 tests the arithmetic of Int32Vec4.
func TestInt32Vec4(t *testing.T) {
	p := Int32Vec4{1, 2, 3, 4}
	q := Int32Vec4{10, -20, 30, -40}
	assert.Equal(t, Int32Vec4{11, -18, 33, -36}, p.Add(q))
	assert.Equal(t, Int32Vec4{9, -22, 27, -44}, q.Sub(p))
	assert.Equal(t, Int32Vec4{3, 6, 9, 12}, p.MulScalar(3))
	assert.Equal(t, Int32Vec4{3, -6, 10, -13}, q.DivScalar(3)) // truncates toward 0
}

// TestUint32Vec4SetFromVec2 tests setting a Uint32Vec4 from a Uint32Vec2,
// with Z = 0 and W = 1.
func TestUint32Vec4SetFromVec2(t *testing.T) {
	u := Uint32Vec4{9, 9, 9, 9}
	u.SetFromVec2(Uint32Vec2{5, 6})
	assert.Equal(t, Uint32Vec4{5, 6, 0, 1}, u)
}

// TestUint64LoHi tests the 64 bit arithmetic on lo, hi uint32 pairs.
func TestUint64LoHi(t *testing.T) {
	assert.Equal(t, uint64(math.MaxUint32)*math.MaxUint32, Uint32Mul64(math.MaxUint32, math.MaxUint32))
	assert.Equal(t, uint64(6), Uint32Mul64(2, 3))

	lh := Uint64ToLoHi(0x123456789abcdef0)
	assert.Equal(t, Uint32Vec2{X: 0x9abcdef0, Y: 0x12345678}, lh)
	assert.Equal(t, uint64(0x123456789abcdef0), Uint64FromLoHi(lh))

	// carries into the hi value
	ctr := Uint64Add32(Uint64FromLoHi(Uint32Vec2{X: 0xfffffffe, Y: 0}), 4)
	assert.Equal(t, Uint32Vec2{X: 2, Y: 1}, Uint64ToLoHi(ctr))
	ctr = Uint64Incr(Uint64Incr(Uint64FromLoHi(Uint32Vec2{X: 0xfffffffe, Y: 0})))
	assert.Equal(t, Uint32Vec2{X: 0, Y: 1}, Uint64ToLoHi(ctr))
}

// TestFloatConversions tests the known values of the float conversions,
// including the Vec2 versions.
func TestFloatConversions(t *testing.T) {
	assert.Equal(t, float32(1.16415321826934814453e-10), Uint32ToFloat32(0))
	assert.Equal(t, float32(0.5), Uint32ToFloat32(0x80000000))
	assert.Equal(t, float32(0.99999994), Uint32ToFloat32(math.MaxUint32)) // never 1
	assert.Less(t, Uint32ToFloat32(math.MaxUint32), float32(1))

	assert.Equal(t, float32(2.32830643653869628906e-10), Uint32ToFloat32Range11(0))
	assert.Equal(t, float32(-1), Uint32ToFloat32Range11(0x80000000))
	assert.Equal(t, float32(1), Uint32ToFloat32Range11(math.MaxInt32))
	assert.Equal(t, float32(-2.32830643653869628906e-10), Uint32ToFloat32Range11(math.MaxUint32))

	v := Uint32ToFloat32Vec2(Uint32Vec2{X: 0x80000000, Y: math.MaxUint32})
	assert.Equal(t, Float32Vec2{X: 0.5, Y: 0.99999994}, v)
	v = Uint32ToFloat32Range11Vec2(Uint32Vec2{X: 0x80000000, Y: math.MaxInt32})
	assert.Equal(t, Float32Vec2{X: -1, Y: 1}, v)
}
