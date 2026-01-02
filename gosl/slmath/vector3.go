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

// Normal3 returns this vector divided by its length (its unit vector).
func Normal3(v math32.Vector3) math32.Vector3 {
	return v.DivScalar(Length3(v))
}

// Cross3 returns the cross product of this vector with other.
func Cross3(v, o math32.Vector3) math32.Vector3 {
	return math32.Vec3(v.Y*o.Z-v.Z*o.Y, v.Z*o.X-v.X*o.Z, v.X*o.Y-v.Y*o.X)
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
