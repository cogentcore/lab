// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

func AddScalar3(v math32.Vector3, s float32) math32.Vector3 {
	return math32.Vec3(v.X+s, v.Y+s, v.Z+s)
}

func SubScalar3(v math32.Vector3, s float32) math32.Vector3 {
	return math32.Vec3(v.X-s, v.Y-s, v.Z-s)
}

func MulScalar3(v math32.Vector3, s float32) math32.Vector3 {
	return math32.Vec3(v.X*s, v.Y*s, v.Z*s)
}

func DivScalar3(v math32.Vector3, s float32) math32.Vector3 {
	return math32.Vec3(v.X/s, v.Y/s, v.Z/s)
}

// Length3 returns the length (magnitude) of this vector.
func Length3(v math32.Vector3) float32 {
	return math32.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Normal3 returns this vector divided by its length (its unit vector).
func Normal3(v math32.Vector3) math32.Vector3 {
	return DivScalar3(v, Length3(v))
}

// Cross3 returns the cross product of this vector with other.
func Cross3(v, o math32.Vector3) math32.Vector3 {
	return math32.Vec3(v.Y*o.Z-v.Z*o.Y, v.Z*o.X-v.X*o.Z, v.X*o.Y-v.Y*o.X)
}

//gosl:end
