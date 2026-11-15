// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slvec

import (
	"testing"
	"unsafe"

	"cogentcore.org/core/math32"
	"github.com/stretchr/testify/assert"
)

// TestSizes tests that the vector types are padded to 16 bytes for the GPU.
func TestSizes(t *testing.T) {
	assert.Equal(t, uintptr(16), unsafe.Sizeof(Vector2{}))
	assert.Equal(t, uintptr(16), unsafe.Sizeof(Vector2i{}))
	assert.Equal(t, uintptr(16), unsafe.Sizeof(Vector3{}))
	assert.Equal(t, uintptr(16), unsafe.Sizeof(Vector3i{}))
}

// TestVector2 tests setting and getting a Vector2.
func TestVector2(t *testing.T) {
	var v Vector2
	v.Set(1, 2)
	assert.Equal(t, math32.Vec2(1, 2), v.V())
	v.SetV(math32.Vec2(3, -4))
	assert.Equal(t, float32(3), v.X)
	assert.Equal(t, float32(-4), v.Y)
	assert.Equal(t, math32.Vec2(3, -4), v.V())
}

// TestVector2i tests setting and getting a Vector2i.
func TestVector2i(t *testing.T) {
	var v Vector2i
	v.Set(1, 2)
	assert.Equal(t, math32.Vec2i(1, 2), v.V())
	v.SetV(math32.Vec2i(3, -4))
	assert.Equal(t, int32(3), v.X)
	assert.Equal(t, int32(-4), v.Y)
	assert.Equal(t, math32.Vec2i(3, -4), v.V())
}

// TestVector3 tests setting and getting a Vector3.
func TestVector3(t *testing.T) {
	var v Vector3
	v.Set(1, 2, 3)
	assert.Equal(t, math32.Vec3(1, 2, 3), v.V())
	v.SetV(math32.Vec3(-1, 5, 7))
	assert.Equal(t, Vector3{X: -1, Y: 5, Z: 7}, v)
	assert.Equal(t, math32.Vec3(-1, 5, 7), v.V())
}

// TestVector3i tests setting and getting a Vector3i.
func TestVector3i(t *testing.T) {
	var v Vector3i
	v.Set(1, 2, 3)
	assert.Equal(t, math32.Vec3i(1, 2, 3), v.V())
	v.SetV(math32.Vec3i(-1, 5, 7))
	assert.Equal(t, Vector3i{X: -1, Y: 5, Z: 7}, v)
	assert.Equal(t, math32.Vec3i(-1, 5, 7), v.V())
}
