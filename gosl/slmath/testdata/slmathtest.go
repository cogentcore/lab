// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slmathtest exercises every slmath function through gosl, so that
// the generated WGSL is checked against the golden file and validated by
// the WGSL compiler. It is not meant to compute anything meaningful.
package slmathtest

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/gosl/slmath"
)

//gosl:start
//gosl:import "cogentcore.org/lab/gosl/slmath"

// Values is one set of test values, as a struct so that it can be
// used directly as a gosl variable.
type Values struct {
	Val  float32
	pad  float32
	pad1 float32
	pad2 float32
}

//gosl:vars
var (
	// Params are the read-only parameters for the computation.
	//
	//gosl:read-only
	Params []Values

	// Data is the data on which the computation operates.
	Data []Values
)

// Vector2Test calls every slmath Vector2 function.
func Vector2Test(v, o math32.Vector2, s float32, d int32) float32 {
	r := slmath.Polar2(s, s)
	r = r.Add(slmath.DivSafe2(v, o))
	r = r.Add(slmath.Negate2(v))
	r = r.Add(slmath.Max2(v, o))
	r = r.Add(slmath.Min2(v, o))
	r = r.Add(slmath.Abs2(v))
	r = r.Add(slmath.Floor2(v))
	r = r.Add(slmath.Ceil2(v))
	r = r.Add(slmath.Round2(v))
	r = r.Add(slmath.Clamp2(v, o, o))
	r = r.Add(slmath.ClampMagnitude2(v, s))
	r = r.Add(slmath.Normal2(v))
	r = r.Add(slmath.Lerp2(v, o, s))
	r = r.Add(slmath.Rot90CW2(v))
	r = r.Add(slmath.Rot90CCW2(v))
	r = r.Add(slmath.Rot2(v, s, o))
	r = r.Add(slmath.SetDim2(v, d, s))
	r = r.Add(slmath.AddDim2(v, d, s))
	r = r.Add(slmath.SubDim2(v, d, s))
	r = r.Add(slmath.MulDim2(v, d, s))
	r = r.Add(slmath.DivDim2(v, d, s))

	sum := slmath.Length2(r)
	sum += slmath.LengthSquared2(r)
	sum += slmath.Dot2(v, o)
	sum += slmath.Cross2(v, o)
	sum += slmath.DistanceTo2(v, o)
	sum += slmath.DistanceToSquared2(v, o)
	sum += slmath.CosTo2(v, o)
	sum += slmath.AngleTo2(v, o)
	sum += slmath.Dim2(v, d)
	if slmath.InTriangle2(v, o, v, o) {
		sum += 1
	}
	nd := slmath.WindowToNDC2(v, o, v, true)
	return sum + nd.X + nd.Y + nd.Z
}

// Vector3Test calls every slmath Vector3 function.
func Vector3Test(v, o math32.Vector3, s float32, d int32) float32 {
	r := slmath.DivSafe3(v, o)
	r = r.Add(slmath.Negate3(v))
	r = r.Add(slmath.Max3(v, o))
	r = r.Add(slmath.Min3(v, o))
	r = r.Add(slmath.Abs3(v))
	r = r.Add(slmath.Floor3(v))
	r = r.Add(slmath.Ceil3(v))
	r = r.Add(slmath.Round3(v))
	r = r.Add(slmath.Clamp3(v, o, o))
	r = r.Add(slmath.ClampMagnitude3(v, s))
	r = r.Add(slmath.Normal3(v))
	r = r.Add(slmath.Cross3(v, o))
	r = r.Add(slmath.Lerp3(v, o, s))
	r = r.Add(slmath.ProjectOnVector3(v, o))
	r = r.Add(slmath.ProjectOnPlane3(v, o))
	r = r.Add(slmath.Reflect3(v, o))
	r = r.Add(slmath.SetDim3(v, d, s))
	r = r.Add(slmath.NDCToWindow3(v, math32.Vec2(v.X, v.Y), math32.Vec2(o.X, o.Y), s, s, true))
	r = r.Add(slmath.Vec3FromVec4(math32.Vec4(v.X, v.Y, v.Z, s)))

	sum := slmath.Length3(r)
	sum += slmath.LengthSquared3(r)
	sum += slmath.Dot3(v, o)
	sum += slmath.DistanceTo3(v, o)
	sum += slmath.DistanceToSquared3(v, o)
	sum += slmath.CosTo3(v, o)
	sum += slmath.AngleTo3(v, o)
	sum += slmath.Dim3(v, d)
	return sum
}

// Vector4Test calls every slmath Vector4 function.
func Vector4Test(v, o math32.Vector4, s float32, d int32) float32 {
	r := slmath.DivSafe4(v, o)
	r = r.Add(slmath.Negate4(v))
	r = r.Add(slmath.Max4(v, o))
	r = r.Add(slmath.Min4(v, o))
	r = r.Add(slmath.Abs4(v))
	r = r.Add(slmath.Floor4(v))
	r = r.Add(slmath.Ceil4(v))
	r = r.Add(slmath.Round4(v))
	r = r.Add(slmath.Clamp4(v, o, o))
	r = r.Add(slmath.ClampMagnitude4(v, s))
	r = r.Add(slmath.Normal4(v))
	r = r.Add(slmath.Lerp4(v, o, s))
	r = r.Add(slmath.SetDim4(v, d, s))
	r = r.Add(slmath.Vec4FromVec3(math32.Vec3(v.X, v.Y, v.Z), s))

	sum := slmath.Length4(r)
	sum += slmath.LengthSquared4(r)
	sum += slmath.Dot4(v, o)
	sum += slmath.Dim4(v, d)
	pd := slmath.PerspDiv4(r)
	return sum + pd.X + pd.Y + pd.Z
}

// Matrix2Test calls every slmath Matrix2 function.
func Matrix2Test(v math32.Vector2, s float32) float32 {
	m := slmath.Mat2Set(s, s, s, s, s, s)
	m = slmath.Mat2Mul(m, slmath.Mat2Identity())
	m = slmath.Mat2Mul(m, slmath.Mat2Translate2D(v.X, v.Y))
	m = slmath.Mat2Mul(m, slmath.Mat2Scale2D(v.X, v.Y))
	m = slmath.Mat2Mul(m, slmath.Mat2Rotate2D(s))
	m = slmath.Mat2Mul(m, slmath.Mat2Rotate2DAround(s, v))
	m = slmath.Mat2Mul(m, slmath.Mat2Shear2D(v.X, v.Y))
	m = slmath.Mat2Mul(m, slmath.Mat2Skew2D(v.X, v.Y))
	m = slmath.Mat2Translate(m, v.X, v.Y)
	m = slmath.Mat2Scale(m, v.X, v.Y)
	m = slmath.Mat2ScaleAbout(m, v.X, v.Y, v.X, v.Y)
	m = slmath.Mat2Rotate(m, s)
	m = slmath.Mat2RotateAbout(m, s, v.X, v.Y)
	m = slmath.Mat2Shear(m, v.X, v.Y)
	m = slmath.Mat2Skew(m, v.X, v.Y)
	m = slmath.Mat2Transpose(m)
	m = slmath.Mat2Inverse(m)

	r := slmath.Mat2MulVector(m, v)
	r = r.Add(slmath.Mat2MulPoint(m, v))
	r = r.Add(slmath.Mat2Pos(m))

	sum := slmath.Mat2Det(m) + slmath.Mat2ExtractRot(m) + r.X + r.Y
	if slmath.Mat2IsIdentity(m) {
		sum += 1
	}
	return sum
}

// Matrix3Test calls every slmath Matrix3 function.
func Matrix3Test(v2 math32.Vector2, v3 math32.Vector3, q math32.Quat, s float32) float32 {
	m := slmath.Mat3Identity()
	m = slmath.Mat3Mul(m, slmath.Mat3FromMatrix2(slmath.Mat2Rotate2D(s)))
	m = slmath.Mat3Mul(m, slmath.Mat3FromMatrix4(slmath.Mat4Identity()))
	m = slmath.Mat3Mul(m, slmath.Mat3Translate2D(v2.X, v2.Y))
	m = slmath.Mat3Mul(m, slmath.Mat3Scale2D(v2.X, v2.Y))
	m = slmath.Mat3Mul(m, slmath.Mat3Rotate2D(s))
	m = slmath.Mat3Mul(m, slmath.QuatToMatrix3(q))
	m = slmath.Mat3MulScalar(m, s)
	m = slmath.Mat3Transpose(m)
	m = slmath.Mat3Inverse(m)
	m = slmath.Mat3ScaleCols(m, v3)
	m = slmath.Mat3NormalMatrix(slmath.Mat4Identity())

	r2 := slmath.Mat3MulVector2(m, v2)
	r2 = r2.Add(slmath.Mat3MulPoint2(m, v2))
	r3 := slmath.Mat3MulVector3(m, v3)

	return slmath.Mat3Determinant(m) + r2.X + r2.Y + r3.X + r3.Y + r3.Z
}

// Matrix4Test calls every slmath Matrix4 function.
func Matrix4Test(v3 math32.Vector3, v4 math32.Vector4, q math32.Quat, s float32) float32 {
	m := slmath.Mat4Set(s, s, s, s, s, s, s, s, s, s, s, s, s, s, s, s)
	m = slmath.Mat4Mul(m, slmath.Mat4Identity())
	m = slmath.Mat4Mul(m, slmath.Mat4FromMatrix3(slmath.Mat3Identity()))
	m = slmath.Mat4Mul(m, slmath.Mat4Translate3D(v3.X, v3.Y, v3.Z))
	m = slmath.Mat4Mul(m, slmath.Mat4Scale3D(v3.X, v3.Y, v3.Z))
	m = slmath.Mat4Mul(m, slmath.Mat4RotateX(s))
	m = slmath.Mat4Mul(m, slmath.Mat4RotateY(s))
	m = slmath.Mat4Mul(m, slmath.Mat4RotateZ(s))
	m = slmath.Mat4Mul(m, slmath.Mat4RotateAxis(v3, s))
	m = slmath.Mat4Mul(m, slmath.Mat4SetBasis(v3, v3, v3))
	m = slmath.Mat4Mul(m, slmath.Mat4FromQuat(q))
	m = slmath.Mat4Mul(m, slmath.Mat4FromEuler(v3))
	m = slmath.Mat4Mul(m, slmath.Mat4Transform(v3, q, v3))
	m = slmath.Mat4MulScalar(m, s)
	m = slmath.Mat4Transpose(m)
	m = slmath.Mat4Inverse(m)
	m = slmath.Mat4ScaleCols(m, v3)
	m = slmath.Mat4SetPos(m, v3)
	m = slmath.Mat4CopyPos(m, slmath.Mat4Identity())
	m = slmath.Mat4ExtractRotation(m)

	r3 := slmath.Mat4MulVector3(m, v3)
	r3 = r3.Add(slmath.Mat4MulVector3AsVector4(m, v3, s))
	r3 = r3.Add(slmath.Mat4MulProjection3(m, v3))
	r3 = r3.Add(slmath.Mat4Pos(m))
	r3 = r3.Add(slmath.Mat4ToEuler(m))
	r4 := slmath.Mat4MulVector4(m, v4)

	sum := slmath.Mat4Determinant(m) + slmath.Mat4MaxScaleOnAxis(m)
	return sum + r3.X + r3.Y + r3.Z + r4.X + r4.Y + r4.Z + r4.W
}

// QuatTest calls every slmath Quat function.
func QuatTest(q, o math32.Quat, v math32.Vector3, s float32, d int32) float32 {
	r := slmath.QuatIdentity()
	r = slmath.QuatAdd(r, slmath.QuatNormalize(q))
	r = slmath.QuatAdd(r, slmath.QuatNormalizeFast(q))
	r = slmath.QuatAdd(r, slmath.QuatConjugate(q))
	r = slmath.QuatAdd(r, slmath.QuatInverse(q))
	r = slmath.QuatAdd(r, slmath.MulQuats(q, o))
	r = slmath.QuatAdd(r, slmath.QuatMulScalar(q, s))
	r = slmath.QuatAdd(r, slmath.QuatSetDim(q, d, s))
	r = slmath.QuatAdd(r, slmath.QuatFromEuler(v))
	r = slmath.QuatAdd(r, slmath.QuatFromAxisAngle(v, s))
	r = slmath.QuatAdd(r, slmath.QuatFromMatrix4(slmath.Mat4Identity()))
	r = slmath.QuatAdd(r, slmath.QuatFromUnitVectors(v, v))
	r = slmath.QuatAdd(r, slmath.QuatSlerp(q, o, s))

	e := slmath.QuatToEuler(r)
	e = e.Add(slmath.EulerAnglesFromQuat(r))
	e = e.Add(slmath.MulQuatVector(r, v))
	e = e.Add(slmath.MulQuatVectorInverse(r, v))
	aa := slmath.QuatToAxisAngle(r)

	sum := slmath.QuatLength(r) + slmath.QuatLengthSquared(r)
	sum += slmath.QuatDot(q, o) + slmath.QuatDim(q, d)
	if slmath.QuatIsIdentity(r) {
		sum += 1
	}
	if slmath.QuatIsNil(r) {
		sum += 1
	}
	return sum + e.X + e.Y + e.Z + aa.X + aa.Y + aa.Z + aa.W
}

// SpatialTest calls the slmath spatial transform functions.
func SpatialTest(aP math32.Vector3, aQ math32.Quat, bP math32.Vector3, bQ math32.Quat) float32 {
	var oP math32.Vector3
	var oQ math32.Quat
	slmath.MulSpatialTransforms(aP, aQ, bP, bQ, &oP, &oQ)
	var iP math32.Vector3
	var iQ math32.Quat
	slmath.SpatialTransformInverse(oP, oQ, &iP, &iQ)
	p := slmath.MulSpatialPoint(iP, iQ, bP)
	return p.X + p.Y + p.Z + slmath.MinAngleDiff(aP.X, bP.X)
}

// MatrixOpsTest exercises the matrix translations that gosl does directly,
// rather than through an slmath function: the operator form of the Matrix3
// methods, and indexing a matrix with a non-constant index.
// The [math32.Matrix4] methods all take pointers, which do not translate,
// so the slmath functions must be used for those.
func MatrixOpsTest(v3 math32.Vector3, s float32, d int32) float32 {
	m3 := slmath.Mat3Rotate2D(s)
	o3 := slmath.Mat3Scale2D(s, s)
	// Matrix3.Mul applies the argument first, so it is b*a in WGSL.
	r3 := m3.Mul(o3)
	r3 = r3.MulScalar(s)
	mv := r3.MulVector3(v3)

	r4 := slmath.Mat4MulScalar(slmath.Mat4Mul(slmath.Mat4RotateX(s), slmath.Mat4Scale3D(s, s, s)), s)

	// non-constant index: divided at runtime into [col][row].
	sum := r3[d] + r4[d]
	return sum + mv.X + mv.Y + mv.Z
}

// SLMathTest is the kernel that reaches every slmath function.
func SLMathTest(i uint32) { //gosl:kernel
	s := Data[i].Val + Params[0].Val
	d := int32(i) % 2
	v2 := math32.Vec2(s, s+1)
	o2 := math32.Vec2(s-1, s+2)
	v3 := math32.Vec3(s, s+1, s+2)
	o3 := math32.Vec3(s-1, s+2, s-3)
	v4 := math32.Vec4(s, s+1, s+2, s+3)
	o4 := math32.Vec4(s-1, s+2, s-3, s+4)
	q := math32.NewQuat(s, s+1, s+2, s+3)
	p := math32.NewQuat(s-1, s+2, s-3, s+4)

	sum := Vector2Test(v2, o2, s, d)
	sum += Vector3Test(v3, o3, s, d)
	sum += Vector4Test(v4, o4, s, d)
	sum += Matrix2Test(v2, s)
	sum += Matrix3Test(v2, v3, q, s)
	sum += Matrix4Test(v3, v4, q, s)
	sum += QuatTest(q, p, v3, s, d)
	sum += SpatialTest(v3, q, o3, p)
	sum += MatrixOpsTest(v3, s, d)
	Data[i].Val = sum
}

//gosl:end
