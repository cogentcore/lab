// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// MulQuat returns vector multiplied by specified quaternion and
// then by the quaternion inverse.
// It basically applies the rotation encoded in the quaternion to this vector.
func MulQuat(v math32.Vector3, q math32.Quat) math32.Vector3 {
	// calculate quat * vector
	ix := q.W*v.X + q.Y*v.Z - q.Z*v.Y
	iy := q.W*v.Y + q.Z*v.X - q.X*v.Z
	iz := q.W*v.Z + q.X*v.Y - q.Y*v.X
	iw := -q.X*v.X - q.Y*v.Y - q.Z*v.Z
	// calculate result * inverse quat
	return math32.Vec3(ix*q.W+iw*-q.X+iy*-q.Z-iz*-q.Y,
		iy*q.W+iw*-q.Y+iz*-q.X-ix*-q.Z,
		iz*q.W+iw*-q.Z+ix*-q.Y-iy*-q.X)
}

// MulQuats set this quaternion to the multiplication of a by b.
func MulQuats(a, b math32.Quat) math32.Quat {
	// from http://www.euclideanspace.com/maths/algebra/realNormedAlgebra/quaternions/code/index.htm
	var q math32.Quat
	q.X = a.X*b.W + a.W*b.X + a.Y*b.Z - a.Z*b.Y
	q.Y = a.Y*b.W + a.W*b.Y + a.Z*b.X - a.X*b.Z
	q.Z = a.Z*b.W + a.W*b.Z + a.X*b.Y - a.Y*b.X
	q.W = a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z
	return q
}

//gosl:end
