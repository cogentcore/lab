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

// MulSpatialTransforms computes the equivalent of matrix multiplication for
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
	if dim == 2 {
		nv.Z = val
	}
	if dim == 3 {
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

// QuatIdentity returns the identity quaternion (no rotation).
func QuatIdentity() math32.Quat {
	return math32.NewQuat(0, 0, 0, 1)
}

// QuatIsIdentity returns whether the quaternion is the identity rotation.
func QuatIsIdentity(q math32.Quat) bool {
	return q.X == 0 && q.Y == 0 && q.Z == 0 && q.W == 1
}

// QuatIsNil returns whether all the quaternion components are zero.
func QuatIsNil(q math32.Quat) bool {
	return q.X == 0 && q.Y == 0 && q.Z == 0 && q.W == 0
}

// QuatLengthSquared returns the length squared of this quaternion.
func QuatLengthSquared(q math32.Quat) float32 {
	return q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W
}

// QuatConjugate returns the conjugate of the quaternion.
func QuatConjugate(q math32.Quat) math32.Quat {
	return math32.NewQuat(-q.X, -q.Y, -q.Z, q.W)
}

// QuatNormalizeFast approximates normalizing the quaternion.
// Works best when the quaternion is already almost-normalized.
func QuatNormalizeFast(q math32.Quat) math32.Quat {
	f := (3 - QuatLengthSquared(q)) / 2
	if f == 0 {
		return QuatIdentity()
	}
	return QuatMulScalar(q, f)
}

// QuatFromEuler returns the quaternion for the given vector of
// euler angles for each axis, which are assumed to be in XYZ order.
func QuatFromEuler(euler math32.Vector3) math32.Quat {
	c1 := math32.Cos(euler.X / 2)
	c2 := math32.Cos(euler.Y / 2)
	c3 := math32.Cos(euler.Z / 2)
	s1 := math32.Sin(euler.X / 2)
	s2 := math32.Sin(euler.Y / 2)
	s3 := math32.Sin(euler.Z / 2)

	return math32.NewQuat(
		s1*c2*c3-c1*s2*s3,
		c1*s2*c3+s1*c2*s3,
		c1*c2*s3-s1*s2*c3,
		c1*c2*c3+s1*s2*s3)
}

// QuatToEuler returns the euler angles for the given quaternion.
func QuatToEuler(q math32.Quat) math32.Vector3 {
	// columns
	x := MulQuatVector(q, math32.Vec3(1, 0, 0))
	y := MulQuatVector(q, math32.Vec3(0, 1, 0))
	z := MulQuatVector(q, math32.Vec3(0, 0, 1))

	phi := math32.Atan2(z.Y, z.Z)
	sinp := -z.X
	var theta float32
	if math32.Abs(sinp) >= 1 {
		theta = 0.5 * SLPi * math32.Sign(sinp)
	} else {
		theta = math32.Asin(sinp)
	}
	psi := math32.Atan2(y.X, x.X)

	return math32.Vec3(-phi, -theta, -psi)
}

// EulerAnglesFromQuat returns the euler angles from the rotation matrix
// of the given quaternion. Note that this uses a different convention
// than [QuatToEuler].
func EulerAnglesFromQuat(q math32.Quat) math32.Vector3 {
	return Mat4ToEuler(Mat4FromQuat(q))
}

// QuatFromAxisAngle returns the quaternion for the rotation
// specified by the given axis and angle (in radians).
func QuatFromAxisAngle(axis math32.Vector3, angle float32) math32.Quat {
	ha := angle / 2
	s := math32.Sin(ha)
	return math32.NewQuat(axis.X*s, axis.Y*s, axis.Z*s, math32.Cos(ha))
}

// QuatToAxisAngle returns a [math32.Vector4] holding the axis (X, Y, Z)
// and angle (W) of the given quaternion, which is assumed to be normalized.
func QuatToAxisAngle(q math32.Quat) math32.Vector4 {
	// http://www.euclideanspace.com/maths/geometry/rotations/conversions/quaternionToAngle/index.htm
	qw := math32.Clamp(q.W, float32(-1), float32(1))
	w := 2 * math32.Acos(qw)
	s := math32.Sqrt(1 - qw*qw)
	if s < 0.0001 {
		return math32.Vec4(1, 0, 0, w)
	}
	return math32.Vec4(q.X/s, q.Y/s, q.Z/s, w)
}

// QuatFromMatrix4 returns the quaternion for the
// given pure rotation matrix.
func QuatFromMatrix4(m math32.Matrix4) math32.Quat {
	m11 := m[0]
	m12 := m[4]
	m13 := m[8]
	m21 := m[1]
	m22 := m[5]
	m23 := m[9]
	m31 := m[2]
	m32 := m[6]
	m33 := m[10]
	trace := m11 + m22 + m33

	var q math32.Quat
	var s float32
	if trace > 0 {
		s = 0.5 / math32.Sqrt(trace+1)
		q.W = 0.25 / s
		q.X = (m32 - m23) * s
		q.Y = (m13 - m31) * s
		q.Z = (m21 - m12) * s
	} else if m11 > m22 && m11 > m33 {
		s = 2 * math32.Sqrt(1+m11-m22-m33)
		q.W = (m32 - m23) / s
		q.X = 0.25 * s
		q.Y = (m12 + m21) / s
		q.Z = (m13 + m31) / s
	} else if m22 > m33 {
		s = 2 * math32.Sqrt(1+m22-m11-m33)
		q.W = (m13 - m31) / s
		q.X = (m12 + m21) / s
		q.Y = 0.25 * s
		q.Z = (m23 + m32) / s
	} else {
		s = 2 * math32.Sqrt(1+m33-m11-m22)
		q.W = (m21 - m12) / s
		q.X = (m13 + m31) / s
		q.Y = (m23 + m32) / s
		q.Z = 0.25 * s
	}
	return q
}

// QuatFromUnitVectors returns the quaternion for the rotation from
// vector vFrom to vTo. Both vectors must be normalized.
func QuatFromUnitVectors(vFrom, vTo math32.Vector3) math32.Quat {
	var v1 math32.Vector3
	r := Dot3(vFrom, vTo) + 1
	if r < 0.000001 {
		r = 0
		if math32.Abs(vFrom.X) > math32.Abs(vFrom.Z) {
			v1 = math32.Vec3(-vFrom.Y, vFrom.X, 0)
		} else {
			v1 = math32.Vec3(0, -vFrom.Z, vFrom.Y)
		}
	} else {
		v1 = Cross3(vFrom, vTo)
	}
	return QuatNormalize(math32.NewQuat(v1.X, v1.Y, v1.Z, r))
}

// QuatSlerp returns the spherically linear interpolation
// from quaternion q to o using t.
func QuatSlerp(q, o math32.Quat, t float32) math32.Quat {
	if t == 0 {
		return q
	}
	if t == 1 {
		return o
	}
	x := q.X
	y := q.Y
	z := q.Z
	w := q.W

	cosHalfTheta := w*o.W + x*o.X + y*o.Y + z*o.Z

	nq := o
	if cosHalfTheta < 0 {
		nq.X = -o.X
		nq.Y = -o.Y
		nq.Z = -o.Z
		nq.W = -o.W
		cosHalfTheta = -cosHalfTheta
	}
	if cosHalfTheta >= 1 {
		return q
	}

	var r math32.Quat
	sqrSinHalfTheta := 1 - cosHalfTheta*cosHalfTheta
	if sqrSinHalfTheta < 0.001 {
		s := 1 - t
		r.W = s*w + t*nq.W
		r.X = s*x + t*nq.X
		r.Y = s*y + t*nq.Y
		r.Z = s*z + t*nq.Z
		return QuatNormalize(r)
	}

	sinHalfTheta := math32.Sqrt(sqrSinHalfTheta)
	halfTheta := math32.Atan2(sinHalfTheta, cosHalfTheta)
	ratioA := math32.Sin((1-t)*halfTheta) / sinHalfTheta
	ratioB := math32.Sin(t*halfTheta) / sinHalfTheta

	r.W = w*ratioA + nq.W*ratioB
	r.X = x*ratioA + nq.X*ratioB
	r.Y = y*ratioA + nq.Y*ratioB
	r.Z = z*ratioA + nq.Z*ratioB
	return r
}

//gosl:end
