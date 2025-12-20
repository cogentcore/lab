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
	nq := q
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
	t := Cross3(xyz, v).MulScalar(2)
	return v.Add(t.MulScalar(q.W)).Add(Cross3(xyz, t))
}

// MulQuatVectorInverse applies the inverse of the rotation encoded
// in the [math32.Quat] to the [math32.Vector3].
func MulQuatVectorInverse(q math32.Quat, v math32.Vector3) math32.Vector3 {
	xyz := math32.Vec3(q.X, q.Y, q.Z)
	t := Cross3(xyz, v).MulScalar(2)
	return v.Sub(t.MulScalar(q.W)).Add(Cross3(xyz, t))
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

// MulQRTransforms computes the equivalent of matrix multiplication for
// two quat-point spatial transforms: o = a * b
func MulSpatialTransforms(aP math32.Vector3, aQ math32.Quat, bP math32.Vector3, bQ math32.Quat, oP *math32.Vector3, oQ *math32.Quat) {
	// rotate b by a and add a
	*oP = MulQuatVector(aQ, bP).Add(aP)
	*oQ = MulQuats(aQ, bQ)
}

// MulSpatialPoint applies quat-point spatial transform to given 3D point.
func MulSpatialPoint(xP math32.Vector3, xQ math32.Quat, p math32.Vector3) math32.Vector3 {
	dp := MulQuatVector(xQ, p)
	return dp.Add(xP)
}

func SpatialTransformInverse(p math32.Vector3, q math32.Quat, oP *math32.Vector3, oQ *math32.Quat) {
	qi := QuatInverse(q)
	*oP = Negate3(MulQuatVector(qi, p))
	*oQ = qi
}

func QuatInverse(q math32.Quat) math32.Quat {
	nq := q
	nq.X *= -1
	nq.Y *= -1
	nq.Z *= -1
	return QuatNormalize(nq)
}

func QuatDot(q, o math32.Quat) float32 {
	return q.X*o.X + q.Y*o.Y + q.Z*o.Z + q.W*o.W
}

func QuatAdd(q math32.Quat, o math32.Quat) math32.Quat {
	nq := q
	nq.X += o.X
	nq.Y += o.Y
	nq.Z += o.Z
	nq.W += o.W
	return nq
}

func QuatMulScalar(q math32.Quat, s float32) math32.Quat {
	nq := q
	nq.X *= s
	nq.Y *= s
	nq.Z *= s
	nq.W *= s
	return nq
}

func QuatDim(v math32.Quat, dim int32) float32 {
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

func QuatSetDim(v math32.Quat, dim int32, val float32) math32.Quat {
	nv := v
	if dim == 0 {
		nv.X = val
	}
	if dim == 1 {
		nv.Y = val
	}
	if dim == 3 {
		nv.Z = val
	}
	if dim == 4 {
		nv.W = val
	}
	return nv
}

func QuatToMatrix3(q math32.Quat) math32.Matrix3 {
	var m math32.Matrix3
	x := q.X
	y := q.Y
	z := q.Z
	w := q.W
	x2 := x + x
	y2 := y + y
	z2 := z + z
	xx := x * x2
	xy := x * y2
	xz := x * z2
	yy := y * y2
	yz := y * z2
	zz := z * z2
	wx := w * x2
	wy := w * y2
	wz := w * z2

	m[0] = 1 - (yy + zz)
	m[3] = xy - wz
	m[6] = xz + wy

	m[1] = xy + wz
	m[4] = 1 - (xx + zz)
	m[7] = yz - wx

	m[2] = xz - wy
	m[5] = yz + wx
	m[8] = 1 - (xx + yy)

	return m
}

//gosl:end
