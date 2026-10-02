// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// Polar2 returns a [math32.Vector2] from polar coordinates,
// with angle in radians CCW and radius the distance from (0,0).
func Polar2(angle, radius float32) math32.Vector2 {
	return math32.Vec2(radius*math32.Cos(angle), radius*math32.Sin(angle))
}

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

// Floor2 returns floor of this vector components.
func Floor2(v math32.Vector2) math32.Vector2 {
	return math32.Vec2(math32.Floor(v.X), math32.Floor(v.Y))
}

// Ceil2 returns ceil of this vector components.
func Ceil2(v math32.Vector2) math32.Vector2 {
	return math32.Vec2(math32.Ceil(v.X), math32.Ceil(v.Y))
}

// Round2 returns round of this vector components.
func Round2(v math32.Vector2) math32.Vector2 {
	return math32.Vec2(math32.Round(v.X), math32.Round(v.Y))
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

// ClampMagnitude2 clamps the magnitude of the components below given value.
func ClampMagnitude2(v math32.Vector2, mag float32) math32.Vector2 {
	r := v
	if r.X < -mag {
		r.X = -mag
	} else if r.X > mag {
		r.X = mag
	}
	if r.Y < -mag {
		r.Y = -mag
	} else if r.Y > mag {
		r.Y = mag
	}
	return r
}

// Normal2 returns this vector divided by its length (its unit vector),
// or the zero vector if the length is zero.
func Normal2(v math32.Vector2) math32.Vector2 {
	l := Length2(v)
	if l == 0 {
		return math32.Vec2(0, 0)
	}
	return v.DivScalar(l)
}

// Cross2 returns the cross product of this vector with other.
func Cross2(v, o math32.Vector2) float32 {
	return v.X*o.Y - v.Y*o.X
}

// DistanceTo2 returns the distance between these two vectors as points.
func DistanceTo2(v, o math32.Vector2) float32 {
	return math32.Sqrt(DistanceToSquared2(v, o))
}

// DistanceToSquared2 returns the squared distance between
// these two vectors as points.
func DistanceToSquared2(v, o math32.Vector2) float32 {
	dx := v.X - o.X
	dy := v.Y - o.Y
	return dx*dx + dy*dy
}

// CosTo2 returns the cosine (normalized dot product)
// between this vector and other.
func CosTo2(v, o math32.Vector2) float32 {
	return Dot2(v, o) / (Length2(v) * Length2(o))
}

// AngleTo2 returns the angle between this vector and other,
// in the range of -PI to PI (not 0 to 2 PI).
func AngleTo2(v, o math32.Vector2) float32 {
	ang := math32.Acos(math32.Clamp(CosTo2(v, o), float32(-1), float32(1)))
	cross := Cross2(v, o)
	if cross > 0 {
		ang = -ang
	}
	return ang
}

// Lerp2 returns vector with each component as the linear interpolated value of
// alpha between itself and the corresponding other component.
func Lerp2(v, o math32.Vector2, alpha float32) math32.Vector2 {
	return math32.Vec2(v.X+(o.X-v.X)*alpha, v.Y+(o.Y-v.Y)*alpha)
}

// InTriangle2 returns whether the vector is inside the specified triangle.
func InTriangle2(v, p0, p1, p2 math32.Vector2) bool {
	a := 0.5 * (-p1.Y*p2.X + p0.Y*(-p1.X+p2.X) + p0.X*(p1.Y-p2.Y) + p1.X*p2.Y)
	sign := float32(1)
	if a < 0 {
		sign = float32(-1)
	}
	s := (p0.Y*p2.X - p0.X*p2.Y + (p2.Y-p0.Y)*v.X + (p0.X-p2.X)*v.Y) * sign
	t := (p0.X*p1.Y - p0.Y*p1.X + (p0.Y-p1.Y)*v.X + (p1.X-p0.X)*v.Y) * sign
	return s >= 0 && t >= 0 && (s+t) < 2*a*sign
}

// Rot90CW2 rotates the line OP by 90 degrees clockwise.
func Rot90CW2(v math32.Vector2) math32.Vector2 {
	return math32.Vec2(v.Y, -v.X)
}

// Rot90CCW2 rotates the line OP by 90 degrees counter-clockwise.
func Rot90CCW2(v math32.Vector2) math32.Vector2 {
	return math32.Vec2(-v.Y, v.X)
}

// Rot2 rotates the line OP by phi radians counter-clockwise around p0.
func Rot2(v math32.Vector2, phi float32, p0 math32.Vector2) math32.Vector2 {
	sinphi := math32.Sin(phi)
	cosphi := math32.Cos(phi)
	return math32.Vec2(
		p0.X+cosphi*(v.X-p0.X)-sinphi*(v.Y-p0.Y),
		p0.Y+sinphi*(v.X-p0.X)+cosphi*(v.Y-p0.Y))
}

// WindowToNDC2 converts window (pixel) coordinates to
// normalized display coordinates (NDC), using given window size parameters.
// The Z depth coordinate (0-1) must be set manually or by reading
// from the framebuffer. flipY if true means flip the Y axis
// (top = 0 for windows vs. bottom = 0 for 3D coords).
func WindowToNDC2(v, size, off math32.Vector2, flipY bool) math32.Vector3 {
	var n math32.Vector3
	half := size.MulScalar(0.5)
	n.X = v.X - off.X
	n.Y = v.Y - off.Y
	if flipY {
		n.Y = size.Y - n.Y
	}
	n.X = n.X/half.X - 1
	n.Y = n.Y/half.Y - 1
	return n
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

// AddDim2 returns the vector with the given value added on the given dimension.
func AddDim2(v math32.Vector2, dim int32, val float32) math32.Vector2 {
	nv := v
	if dim == 0 {
		nv.X += val
	}
	if dim == 1 {
		nv.Y += val
	}
	return nv
}

// SubDim2 returns the vector with the given value subtracted on the given dimension.
func SubDim2(v math32.Vector2, dim int32, val float32) math32.Vector2 {
	nv := v
	if dim == 0 {
		nv.X -= val
	}
	if dim == 1 {
		nv.Y -= val
	}
	return nv
}

// MulDim2 returns the vector with the given dimension multiplied by the given value.
func MulDim2(v math32.Vector2, dim int32, val float32) math32.Vector2 {
	nv := v
	if dim == 0 {
		nv.X *= val
	}
	if dim == 1 {
		nv.Y *= val
	}
	return nv
}

// DivDim2 returns the vector with the given dimension divided by the given value.
func DivDim2(v math32.Vector2, dim int32, val float32) math32.Vector2 {
	nv := v
	if dim == 0 {
		nv.X /= val
	}
	if dim == 1 {
		nv.Y /= val
	}
	return nv
}

//gosl:end
