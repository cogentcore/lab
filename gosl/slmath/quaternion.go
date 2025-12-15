// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// QuatLength returns the length of this quaternion.
func QuatLength(q math32.Quat) float32 {
	return math32.Sqrt(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)
}

// QuatNormalize normalizes the quaternion.
func QuatNormalize(q math32.Quat) math32.Quat {
	var nq math32.Quat
	l := QuatLength(q)
	if l == 0 {
		nq.X = 0
		nq.Y = 0
		nq.Z = 0
		nq.W = 1
	} else {
		l = 1 / l
		nq.X *= l
		nq.Y *= l
		nq.Z *= l
		nq.W *= l
	}
	return nq
}

// MulQuatVector applies the rotation encoded in the [math32.Quat]
// to the [math32.Vector3].
func MulQuatVector(q math32.Quat, v math32.Vector3) math32.Vector3 {
	xyz := math32.Vec3(q.X, q.Y, q.Z)
	t := MulScalar3(Cross3(xyz, v), 2)
	return v.Add(MulScalar3(t, q.W)).Add(Cross3(xyz, t))
}

// MulQuatVectorInverse applies the inverse of the rotation encoded
// in the [math32.Quat] to the [math32.Vector3].
func MulQuatVectorInverse(q math32.Quat, v math32.Vector3) math32.Vector3 {
	xyz := math32.Vec3(q.X, q.Y, q.Z)
	t := MulScalar3(Cross3(xyz, v), 2)
	return v.Sub(MulScalar3(t, q.W)).Add(Cross3(xyz, t))
}

// MulQuats returns multiplication of a by b quaternions.
func MulQuats(a, b math32.Quat) math32.Quat {
	// from http://www.euclideanspace.com/maths/algebra/realNormedAlgebra/quaternions/code/index.htm
	var q math32.Quat
	q.X = a.X*b.W + a.W*b.X + a.Y*b.Z - a.Z*b.Y
	q.Y = a.Y*b.W + a.W*b.Y + a.Z*b.X - a.X*b.Z
	q.Z = a.Z*b.W + a.W*b.Z + a.X*b.Y - a.Y*b.X
	q.W = a.W*b.W - a.X*b.X - a.Y*b.Y - a.Z*b.Z
	return q
}

// MulQPTransforms computes the equivalent of matrix multiplication for
// two quat-point spatial transforms: o = a * b
func MulQPTransforms(aP math32.Vector3, aQ math32.Quat, bP math32.Vector3, bQ math32.Quat, oP *math32.Vector3, oQ *math32.Quat) {
	// rotate b by a and add a
	br := MulQuatVector(aQ, bP)
	*oP = br.Add(aP)
	*oQ = MulQuats(aQ, bQ)
}

// MulQPPoint applies quat-point spatial transform to given 3D point.
func MulQPPoint(xP math32.Vector3, xQ math32.Quat, p math32.Vector3) math32.Vector3 {
	dp := MulQuatVector(xQ, p)
	return dp.Add(xP)
}

//gosl:end
