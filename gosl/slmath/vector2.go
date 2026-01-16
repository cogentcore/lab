// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// DivSafe2 divides v by o elementwise, only where o != 0
func DivSafe2(v math32.Vector2, o math32.Vector2) math32.Vector2 {
	nv := v
	if o.X != 0 {
		nv.X /= o.X
	}
	if o.Y != 0 {
		nv.Y /= o.Y
	}
	return nv
}

func Negate2(v math32.Vector2) math32.Vector2 {
	return math32.Vec2(-v.X, -v.Y)
}

// Length2 returns the length (magnitude) of this vector.
func Length2(v math32.Vector2) float32 {
	return math32.Sqrt(v.X*v.X + v.Y*v.Y)
}

// LengthSquared2 returns the length squared of this vector.
func LengthSquared2(v math32.Vector2) float32 {
	return v.X*v.X + v.Y*v.Y
}

func Dot2(v, o math32.Vector2) float32 {
	return v.X*o.X + v.Y*o.Y
}

// Max2 returns max of this vector components vs. other vector.
func Max2(v, o math32.Vector2) math32.Vector2 {
	return math32.Vec2(max(v.X, o.X), max(v.Y, o.Y))
}

// Min2 returns min of this vector components vs. other vector.
func Min2(v, o math32.Vector2) math32.Vector2 {
	return math32.Vec2(min(v.X, o.X), min(v.Y, o.Y))
}

// Abs2 returns abs of this vector components.
func Abs2(v math32.Vector2) math32.Vector2 {
	return math32.Vec2(math32.Abs(v.X), math32.Abs(v.Y))
}

func Clamp2(v, min, max math32.Vector2) math32.Vector2 {
	r := v
	if r.X < min.X {
		r.X = min.X
	} else if r.X > max.X {
		r.X = max.X
	}
	if r.Y < min.Y {
		r.Y = min.Y
	} else if r.Y > max.Y {
		r.Y = max.Y
	}
	return r
}

// Normal2 returns this vector divided by its length (its unit vector).
func Normal2(v math32.Vector2) math32.Vector2 {
	return v.DivScalar(Length2(v))
}

// Cross2 returns the cross product of this vector with other.
func Cross2(v, o math32.Vector2) float32 {
	return v.X*o.Y - v.Y*o.X
}

func Dim2(v math32.Vector2, dim int32) float32 {
	if dim == 0 {
		return v.X
	}
	if dim == 1 {
		return v.Y
	}
	return 0
}

func SetDim2(v math32.Vector2, dim int32, val float32) math32.Vector2 {
	nv := v
	if dim == 0 {
		nv.X = val
	}
	if dim == 1 {
		nv.Y = val
	}
	return nv
}

//gosl:end
