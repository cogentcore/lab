// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// DivSafe3 divides v by o elementwise, only where o != 0
func DivSafe3(v math32.Vector3, o math32.Vector3) math32.Vector3 {
	nv := v
	if o.X != 0 {
		nv.X /= o.X
	}
	if o.Y != 0 {
		nv.Y /= o.Y
	}
	if o.Z != 0 {
		nv.Z /= o.Z
	}
	return nv
}

func Negate3(v math32.Vector3) math32.Vector3 {
	return math32.Vec3(-v.X, -v.Y, -v.Z)
}

// Length3 returns the length (magnitude) of this vector.
func Length3(v math32.Vector3) float32 {
	return math32.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// LengthSquared3 returns the length squared of this vector.
func LengthSquared3(v math32.Vector3) float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func Dot3(v, o math32.Vector3) float32 {
	return v.X*o.X + v.Y*o.Y + v.Z*o.Z
}

// Max3 returns max of this vector components vs. other vector.
func Max3(v, o math32.Vector3) math32.Vector3 {
	return math32.Vec3(max(v.X, o.X), max(v.Y, o.Y), max(v.Z, o.Z))
}

// Min3 returns min of this vector components vs. other vector.
func Min3(v, o math32.Vector3) math32.Vector3 {
	return math32.Vec3(min(v.X, o.X), min(v.Y, o.Y), min(v.Z, o.Z))
}

// Abs3 returns abs of this vector components.
func Abs3(v math32.Vector3) math32.Vector3 {
	return math32.Vec3(math32.Abs(v.X), math32.Abs(v.Y), math32.Abs(v.Z))
}

func Clamp3(v, min, max math32.Vector3) math32.Vector3 {
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
	if r.Z < min.Z {
		r.Z = min.Z
	} else if r.Z > max.Z {
		r.Z = max.Z
	}
	return r
}

// ClampMagnitude3 clamps the magnitude of the components below given value.
func ClampMagnitude3(v math32.Vector3, mag float32) math32.Vector3 {
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
	if r.Z < -mag {
		r.Z = -mag
	} else if r.Z > mag {
		r.Z = mag
	}
	return r
}

// Normal3 returns this vector divided by its length (its unit vector),
// or the zero vector if the length is zero.
func Normal3(v math32.Vector3) math32.Vector3 {
	l := Length3(v)
	if l == 0 {
		return math32.Vec3(0, 0, 0)
	}
	return v.DivScalar(l)
}

// Cross3 returns the cross product of this vector with other.
func Cross3(v, o math32.Vector3) math32.Vector3 {
	return math32.Vec3(v.Y*o.Z-v.Z*o.Y, v.Z*o.X-v.X*o.Z, v.X*o.Y-v.Y*o.X)
}

// Floor3 returns floor of this vector components.
func Floor3(v math32.Vector3) math32.Vector3 {
	return math32.Vec3(math32.Floor(v.X), math32.Floor(v.Y), math32.Floor(v.Z))
}

// Ceil3 returns ceil of this vector components.
func Ceil3(v math32.Vector3) math32.Vector3 {
	return math32.Vec3(math32.Ceil(v.X), math32.Ceil(v.Y), math32.Ceil(v.Z))
}

// Round3 returns round of this vector components.
func Round3(v math32.Vector3) math32.Vector3 {
	return math32.Vec3(math32.Round(v.X), math32.Round(v.Y), math32.Round(v.Z))
}

// DistanceTo3 returns the distance between these two vectors as points.
func DistanceTo3(v, o math32.Vector3) float32 {
	return math32.Sqrt(DistanceToSquared3(v, o))
}

// DistanceToSquared3 returns the squared distance between
// these two vectors as points.
func DistanceToSquared3(v, o math32.Vector3) float32 {
	dx := v.X - o.X
	dy := v.Y - o.Y
	dz := v.Z - o.Z
	return dx*dx + dy*dy + dz*dz
}

// Lerp3 returns vector with each component as the linear interpolated value of
// alpha between itself and the corresponding other component.
func Lerp3(v, o math32.Vector3, alpha float32) math32.Vector3 {
	return math32.Vec3(v.X+(o.X-v.X)*alpha, v.Y+(o.Y-v.Y)*alpha, v.Z+(o.Z-v.Z)*alpha)
}

// CosTo3 returns the cosine (normalized dot product)
// between this vector and other.
func CosTo3(v, o math32.Vector3) float32 {
	return Dot3(v, o) / (Length3(v) * Length3(o))
}

// AngleTo3 returns the angle between this vector and other,
// in the range of -PI to PI (not 0 to 2 PI).
func AngleTo3(v, o math32.Vector3) float32 {
	ang := math32.Acos(math32.Clamp(CosTo3(v, o), float32(-1), float32(1)))
	cross := Cross3(v, o)
	ax := math32.Abs(cross.X)
	ay := math32.Abs(cross.Y)
	az := math32.Abs(cross.Z)
	if az >= ay && az >= ax {
		if cross.Z > 0 {
			ang = -ang
		}
	} else if ay >= az && ay >= ax {
		if cross.Y > 0 {
			ang = -ang
		}
	} else if ax >= az && ax >= ay {
		if cross.X > 0 {
			ang = -ang
		}
	}
	return ang
}

// ProjectOnVector3 returns vector projected on other vector.
func ProjectOnVector3(v, o math32.Vector3) math32.Vector3 {
	on := Normal3(o)
	return on.MulScalar(Dot3(v, on))
}

// ProjectOnPlane3 returns vector projected on the plane
// specified by the given normal vector.
func ProjectOnPlane3(v, planeNormal math32.Vector3) math32.Vector3 {
	return v.Sub(ProjectOnVector3(v, planeNormal))
}

// Reflect3 returns vector reflected relative to the normal vector,
// which is assumed to be already normalized.
func Reflect3(v, normal math32.Vector3) math32.Vector3 {
	return v.Sub(normal.MulScalar(2 * Dot3(v, normal)))
}

// NDCToWindow3 converts normalized display coordinates (NDC) to window
// (pixel) coordinates, using given window size parameters.
// near, far are 0, 1 by default (glDepthRange defaults).
// flipY if true means flip the Y axis
// (top = 0 for windows vs. bottom = 0 for 3D coords).
func NDCToWindow3(v math32.Vector3, size, off math32.Vector2, near, far float32, flipY bool) math32.Vector3 {
	var w math32.Vector3
	half := size.MulScalar(0.5)
	w.X = half.X*v.X + half.X
	w.Y = half.Y*v.Y + half.Y
	w.Z = 0.5*(far-near)*v.Z + 0.5*(far+near)
	if flipY {
		w.Y = size.Y - w.Y
	}
	w.X += off.X
	w.Y += off.Y
	return w
}

// Vec3FromVec4 returns a [math32.Vector3] from the X, Y, Z
// components of the given [math32.Vector4].
func Vec3FromVec4(v math32.Vector4) math32.Vector3 {
	return math32.Vec3(v.X, v.Y, v.Z)
}

func Dim3(v math32.Vector3, dim int32) float32 {
	if dim == 0 {
		return v.X
	}
	if dim == 1 {
		return v.Y
	}
	return v.Z
}

func SetDim3(v math32.Vector3, dim int32, val float32) math32.Vector3 {
	nv := v
	if dim == 0 {
		nv.X = val
	}
	if dim == 1 {
		nv.Y = val
	}
	if dim == 2 {
		nv.Z = val
	}
	return nv
}

//gosl:end
