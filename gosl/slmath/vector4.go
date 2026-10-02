// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// Vec4FromVec3 returns a [math32.Vector4] from the given [math32.Vector3]
// and the given W component.
func Vec4FromVec3(v math32.Vector3, w float32) math32.Vector4 {
	return math32.Vec4(v.X, v.Y, v.Z, w)
}

// DivSafe4 divides v by o elementwise, only where o != 0
func DivSafe4(v math32.Vector4, o math32.Vector4) math32.Vector4 {
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
	if o.W != 0 {
		nv.W /= o.W
	}
	return nv
}

func Negate4(v math32.Vector4) math32.Vector4 {
	return math32.Vec4(-v.X, -v.Y, -v.Z, -v.W)
}

// Length4 returns the length (magnitude) of this vector.
func Length4(v math32.Vector4) float32 {
	return math32.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)
}

// LengthSquared4 returns the length squared of this vector.
func LengthSquared4(v math32.Vector4) float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W
}

func Dot4(v, o math32.Vector4) float32 {
	return v.X*o.X + v.Y*o.Y + v.Z*o.Z + v.W*o.W
}

// Max4 returns max of this vector components vs. other vector.
func Max4(v, o math32.Vector4) math32.Vector4 {
	return math32.Vec4(max(v.X, o.X), max(v.Y, o.Y), max(v.Z, o.Z), max(v.W, o.W))
}

// Min4 returns min of this vector components vs. other vector.
func Min4(v, o math32.Vector4) math32.Vector4 {
	return math32.Vec4(min(v.X, o.X), min(v.Y, o.Y), min(v.Z, o.Z), min(v.W, o.W))
}

// Abs4 returns abs of this vector components.
func Abs4(v math32.Vector4) math32.Vector4 {
	return math32.Vec4(math32.Abs(v.X), math32.Abs(v.Y), math32.Abs(v.Z), math32.Abs(v.W))
}

// Floor4 returns floor of this vector components.
func Floor4(v math32.Vector4) math32.Vector4 {
	return math32.Vec4(math32.Floor(v.X), math32.Floor(v.Y), math32.Floor(v.Z), math32.Floor(v.W))
}

// Ceil4 returns ceil of this vector components.
func Ceil4(v math32.Vector4) math32.Vector4 {
	return math32.Vec4(math32.Ceil(v.X), math32.Ceil(v.Y), math32.Ceil(v.Z), math32.Ceil(v.W))
}

// Round4 returns round of this vector components.
func Round4(v math32.Vector4) math32.Vector4 {
	return math32.Vec4(math32.Round(v.X), math32.Round(v.Y), math32.Round(v.Z), math32.Round(v.W))
}

func Clamp4(v, min, max math32.Vector4) math32.Vector4 {
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
	if r.W < min.W {
		r.W = min.W
	} else if r.W > max.W {
		r.W = max.W
	}
	return r
}

// ClampMagnitude4 clamps the magnitude of the components below given value.
func ClampMagnitude4(v math32.Vector4, mag float32) math32.Vector4 {
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
	if r.W < -mag {
		r.W = -mag
	} else if r.W > mag {
		r.W = mag
	}
	return r
}

// Normal4 returns this vector divided by its length (its unit vector),
// or the zero vector if the length is zero.
func Normal4(v math32.Vector4) math32.Vector4 {
	l := Length4(v)
	if l == 0 {
		return math32.Vec4(0, 0, 0, 0)
	}
	return v.DivScalar(l)
}

// Lerp4 returns vector with each component as the linear interpolated value of
// alpha between itself and the corresponding other component.
func Lerp4(v, o math32.Vector4, alpha float32) math32.Vector4 {
	return math32.Vec4(v.X+(o.X-v.X)*alpha, v.Y+(o.Y-v.Y)*alpha,
		v.Z+(o.Z-v.Z)*alpha, v.W+(o.W-v.W)*alpha)
}

// PerspDiv4 returns the 3-vector of normalized display coordinates (NDC)
// from the given 4-vector, by dividing by the 4th W component.
func PerspDiv4(v math32.Vector4) math32.Vector3 {
	return math32.Vec3(v.X/v.W, v.Y/v.W, v.Z/v.W)
}

func Dim4(v math32.Vector4, dim int32) float32 {
	if dim == 0 {
		return v.X
	}
	if dim == 1 {
		return v.Y
	}
	if dim == 2 {
		return v.Z
	}
	return v.W
}

func SetDim4(v math32.Vector4, dim int32, val float32) math32.Vector4 {
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
	if dim == 3 {
		nv.W = val
	}
	return nv
}

//gosl:end
