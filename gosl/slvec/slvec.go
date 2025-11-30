// Copyright 2025 Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slvec

import "cogentcore.org/core/math32"

//gosl:start

// Vector2 is a 2D vector/point with X and Y components.
// with padding values so it works in a GPU struct. Use the
// V() method to get a math32.Vector2 that supports standard
// math operations. Cannot use those math ops in gosl GPU
// code at this point, unfortunately.
type Vector2 struct {
	X float32
	Y float32

	pad, pad1 float32
}

func (v *Vector2) V() math32.Vector2 {
	return math32.Vec2(v.X, v.Y)
}

func (v *Vector2) Set(x, y float32) {
	v.X = x
	v.Y = y
}

func (v *Vector2) SetV(mv math32.Vector2) {
	v.X = mv.X
	v.Y = mv.Y
}

// Vector2i is a 2D vector/point with X and Y integer components.
// with padding values so it works in a GPU struct. Use the
// V() method to get a math32.Vector2i that supports standard
// math operations. Cannot use those math ops in gosl GPU
// code at this point, unfortunately.
type Vector2i struct {
	X int32
	Y int32

	pad, pad1 int32
}

func (v *Vector2i) V() math32.Vector2i {
	return math32.Vec2i(int(v.X), int(v.Y))
}

func (v *Vector2i) Set(x, y int) {
	v.X = int32(x)
	v.Y = int32(y)
}

func (v *Vector2i) SetV(mv math32.Vector2i) {
	v.X = mv.X
	v.Y = mv.Y
}

//gosl:end
