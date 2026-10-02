// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
	"github.com/stretchr/testify/assert"
)

// TestQuatVsMath32 checks every Quat function against the
// equivalent [math32.Quat] method or constructor.
func TestQuatVsMath32(t *testing.T) {
	eqQ(t, math32.NewQuatIdentity(), QuatIdentity(), "QuatIdentity")

	for _, q := range testQuats {
		eqf(t, q.Length(), QuatLength(q), "QuatLength")
		eqf(t, q.LengthSq(), QuatLengthSquared(q), "QuatLengthSquared")
		eqQ(t, q.Conjugate(), QuatConjugate(q), "QuatConjugate")
		eqQ(t, q.Inverse(), QuatInverse(q), "QuatInverse")
		eqb(t, q.IsIdentity(), QuatIsIdentity(q), "QuatIsIdentity")
		eqb(t, q.IsNil(), QuatIsNil(q), "QuatIsNil")
		eq3(t, q.ToEuler(), QuatToEuler(q), "QuatToEuler")
		eq4(t, q.ToAxisAngle(), QuatToAxisAngle(q), "QuatToAxisAngle")

		nq := q
		nq.Normalize()
		eqQ(t, nq, QuatNormalize(q), "QuatNormalize")

		nf := q
		nf.NormalizeFast()
		eqQ(t, nf, QuatNormalizeFast(q), "QuatNormalizeFast")

		ev := math32.Vector3{}
		ev.SetEulerAnglesFromQuat(q)
		eq3(t, ev, EulerAnglesFromQuat(q), "EulerAnglesFromQuat")

		for d := int32(0); d < 4; d++ {
			eqf(t, qdim(q, d), QuatDim(q, d), "QuatDim")
			for _, s := range testScalars {
				eqQ(t, qsetdim(q, d, s), QuatSetDim(q, d, s), "QuatSetDim")
			}
		}

		for _, v := range testVec3s {
			eq3(t, q.MulVector(v), MulQuatVector(q, v), "MulQuatVector")
			eq3(t, q.MulVectorInverse(v), MulQuatVectorInverse(q, v), "MulQuatVectorInverse")
		}
		for _, s := range testScalars {
			eqQ(t, q.MulScalar(s), QuatMulScalar(q, s), "QuatMulScalar")
		}
		for _, o := range testQuats {
			eqf(t, q.Dot(o), QuatDot(q, o), "QuatDot")
			eqQ(t, q.Add(o), QuatAdd(q, o), "QuatAdd")
			eqQ(t, q.Mul(o), MulQuats(q, o), "MulQuats")
			for _, s := range testScalars {
				sq := q
				sq.Slerp(o, s)
				eqQ(t, sq, QuatSlerp(q, o, s), "QuatSlerp")
			}
		}
	}

	for _, v := range testVec3s {
		eqQ(t, math32.NewQuatEuler(v), QuatFromEuler(v), "QuatFromEuler")
		for _, s := range testScalars {
			eqQ(t, math32.NewQuatAxisAngle(v, s), QuatFromAxisAngle(v, s), "QuatFromAxisAngle")
		}
		for _, o := range testVec3s {
			nq := math32.Quat{}
			nq.SetFromUnitVectors(Normal3(v), Normal3(o))
			eqQ(t, nq, QuatFromUnitVectors(Normal3(v), Normal3(o)), "QuatFromUnitVectors")
		}
	}

	for _, m := range testMatrix4s() {
		// SetFromRotationMatrix requires a pure rotation matrix
		rm := Mat4ExtractRotation(m)
		nq := math32.Quat{}
		nq.SetFromRotationMatrix(&rm)
		eqQ(t, nq, QuatFromMatrix4(rm), "QuatFromMatrix4")
	}
}

// qdim / qsetdim are the reference implementations of QuatDim / QuatSetDim.
func qdim(q math32.Quat, dim int32) float32 {
	switch dim {
	case 0:
		return q.X
	case 1:
		return q.Y
	case 2:
		return q.Z
	}
	return q.W
}

func qsetdim(q math32.Quat, dim int32, val float32) math32.Quat {
	switch dim {
	case 0:
		q.X = val
	case 1:
		q.Y = val
	case 2:
		q.Z = val
	case 3:
		q.W = val
	}
	return q
}

func TestQuatValues(t *testing.T) {
	q := QuatFromEuler(math32.Vec3(0.1, 0.2, 0.3))
	eqQ(t, math32.NewQuat(0.034270797, 0.10602052, 0.14357218, 0.9833474), q, "QuatFromEuler")
	// note: QuatToEuler and EulerAnglesFromQuat use a different euler
	// convention than QuatFromEuler, so they do not round trip with it,
	// but they do agree with each other.
	eq3(t, QuatToEuler(q), EulerAnglesFromQuat(q), "euler conventions agree")

	ax := QuatFromAxisAngle(math32.Vec3(1, 0, 0), math32.Pi/2)
	eqQ(t, math32.NewQuat(0.7071068, 0, 0, 0.7071068), ax, "QuatFromAxisAngle")
	eq4(t, math32.Vec4(1, 0, 0, math32.Pi/2), QuatToAxisAngle(ax), "QuatToAxisAngle")

	// rotating X by 90 degrees about Z gives Y
	rz := QuatFromAxisAngle(math32.Vec3(0, 0, 1), math32.Pi/2)
	eq3(t, math32.Vec3(0, 1, 0), MulQuatVector(rz, math32.Vec3(1, 0, 0)), "MulQuatVector")
	eq3(t, math32.Vec3(1, 0, 0), MulQuatVectorInverse(rz, math32.Vec3(0, 1, 0)), "MulQuatVectorInverse")
	eq3(t, math32.Vec3(1, 0, 0), MulQuatVector(QuatInverse(rz), math32.Vec3(0, 1, 0)), "QuatInverse")

	eqf(t, 1, QuatLength(QuatNormalize(math32.NewQuat(1, 2, 3, 4))), "QuatNormalize is unit")
	eqQ(t, QuatIdentity(), QuatNormalize(math32.NewQuat(0, 0, 0, 0)), "QuatNormalize of zero")
	assert.True(t, QuatIsIdentity(QuatIdentity()))
	assert.True(t, QuatIsNil(math32.NewQuat(0, 0, 0, 0)))

	// slerp endpoints
	a := QuatFromAxisAngle(math32.Vec3(0, 0, 1), 0)
	b := QuatFromAxisAngle(math32.Vec3(0, 0, 1), math32.Pi/2)
	eqQ(t, a, QuatSlerp(a, b, 0), "slerp at 0")
	eqQ(t, b, QuatSlerp(a, b, 1), "slerp at 1")
	eqQ(t, QuatFromAxisAngle(math32.Vec3(0, 0, 1), math32.Pi/4), QuatSlerp(a, b, 0.5), "slerp at 0.5")
}

func TestQuatMatrix3(t *testing.T) {
	for _, q := range testQuats {
		nq := QuatNormalize(q)
		m3 := QuatToMatrix3(nq)
		for _, v := range testVec3s {
			eq3(t, MulQuatVector(nq, v), Mat3MulVector3(m3, v), "QuatToMatrix3 matches MulQuatVector")
		}
		m4 := Mat4FromQuat(nq)
		for _, v := range testVec3s {
			eq3(t, MulQuatVector(nq, v), Mat4MulVector3(m4, v), "Mat4FromQuat matches MulQuatVector")
		}
	}
}

func TestSpatialTransforms(t *testing.T) {
	aP := math32.Vec3(1, 2, 3)
	aQ := QuatFromAxisAngle(math32.Vec3(0, 0, 1), math32.Pi/2)
	bP := math32.Vec3(-2, 0.5, 4)
	bQ := QuatFromAxisAngle(math32.Vec3(1, 0, 0), math32.Pi/3)

	var oP math32.Vector3
	var oQ math32.Quat
	MulSpatialTransforms(aP, aQ, bP, bQ, &oP, &oQ)
	eq3(t, MulQuatVector(aQ, bP).Add(aP), oP, "MulSpatialTransforms pos")
	eqQ(t, MulQuats(aQ, bQ), oQ, "MulSpatialTransforms quat")

	for _, p := range testVec3s {
		// applying a then b composed is the same as applying the composite
		eq3(t, MulSpatialPoint(aP, aQ, MulSpatialPoint(bP, bQ, p)), MulSpatialPoint(oP, oQ, p), "composite spatial transform")
	}

	var iP math32.Vector3
	var iQ math32.Quat
	SpatialTransformInverse(aP, aQ, &iP, &iQ)
	for _, p := range testVec3s {
		eq3(t, p, MulSpatialPoint(iP, iQ, MulSpatialPoint(aP, aQ, p)), "spatial transform inverse round trip")
	}
}
