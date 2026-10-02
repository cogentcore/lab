// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
)

// TestMatrix3VsMath32 checks every Mat3 function against the
// equivalent [math32.Matrix3] method or constructor.
func TestMatrix3VsMath32(t *testing.T) {
	eqM3(t, math32.Identity3(), Mat3Identity(), "Mat3Identity")

	for _, a := range testMatrix3s() {
		eqM3(t, a.Transpose(), Mat3Transpose(a), "Mat3Transpose")
		eqf(t, a.Determinant(), Mat3Determinant(a), "Mat3Determinant")
		if a.Determinant() != 0 {
			eqM3(t, a.Inverse(), Mat3Inverse(a), "Mat3Inverse")
		}

		for _, b := range testMatrix3s() {
			eqM3(t, a.Mul(b), Mat3Mul(a, b), "Mat3Mul")
		}
		for _, v := range testVec2s {
			eq2(t, a.MulVector(v), Mat3MulVector2(a, v), "Mat3MulVector2")
			eq2(t, a.MulPoint(v), Mat3MulPoint2(a, v), "Mat3MulPoint2")
		}
		for _, v := range testVec3s {
			eq3(t, a.MulVector3(v), Mat3MulVector3(a, v), "Mat3MulVector3")
			eq3(t, v.MulMatrix3(&a), Mat3MulVector3(a, v), "Vector3.MulMatrix3")
			na := a
			eqM3(t, *na.ScaleCols(v), Mat3ScaleCols(a, v), "Mat3ScaleCols")
		}
		for _, s := range testScalars {
			eqM3(t, a.MulScalar(s), Mat3MulScalar(a, s), "Mat3MulScalar")
		}
	}

	for _, m2 := range testMatrix2s() {
		eqM3(t, math32.Matrix3FromMatrix2(m2), Mat3FromMatrix2(m2), "Mat3FromMatrix2")
	}
	for _, m4 := range testMatrix4s() {
		eqM3(t, math32.Matrix3FromMatrix4(&m4), Mat3FromMatrix4(m4), "Mat3FromMatrix4")
		nm := math32.Matrix3{}
		if err := nm.SetNormalMatrix(&m4); err == nil {
			eqM3(t, nm, Mat3NormalMatrix(m4), "Mat3NormalMatrix")
		}
	}
	for _, x := range testScalars {
		for _, y := range testScalars {
			eqM3(t, math32.Matrix3Translate2D(x, y), Mat3Translate2D(x, y), "Mat3Translate2D")
			eqM3(t, math32.Matrix3Scale2D(x, y), Mat3Scale2D(x, y), "Mat3Scale2D")
		}
		eqM3(t, math32.Matrix3Rotate2D(x), Mat3Rotate2D(x), "Mat3Rotate2D")
	}
	for _, q := range testQuats {
		nm := math32.Matrix3{}
		nm.SetRotationFromQuat(q)
		eqM3(t, nm, QuatToMatrix3(q), "QuatToMatrix3")
	}
}

func TestMatrix3Transforms(t *testing.T) {
	v0 := math32.Vec2(0, 0)
	vx := math32.Vec2(1, 0)
	vy := math32.Vec2(0, 1)
	vxy := math32.Vec2(1, 1)

	eq2(t, vx, Mat3MulPoint2(Mat3Identity(), vx), "identity")
	eq2(t, vxy, Mat3MulPoint2(Mat3FromMatrix2(Mat2Translate2D(1, 1)), v0), "translate")
	eq2(t, vxy.MulScalar(2), Mat3MulPoint2(Mat3Scale2D(2, 2), vxy), "scale")
	eq2(t, vy, Mat3MulPoint2(Mat3Rotate2D(math32.DegToRad(90)), vx), "rotate 90")
	eq2(t, vx, Mat3MulPoint2(Mat3Rotate2D(math32.DegToRad(-90)), vy), "rotate -90")

	// 1,0 -> scale(2) = 2,0 -> rotate 90 = 0,2 -> trans 1,1 -> 1,3
	// multiplication order is the *reverse* of the "logical" order:
	xf := Mat3Mul(Mat3Mul(Mat3Translate2D(1, 1), Mat3Rotate2D(math32.DegToRad(90))), Mat3Scale2D(2, 2))
	eq2(t, math32.Vec2(1, 3), Mat3MulPoint2(xf, vx), "composed transform")
}

func TestMatrix3Inverse(t *testing.T) {
	m := math32.Mat3(1, 2, 3, 4, 5, 6, 7, 8, 10)
	eqM3(t, Mat3Identity(), Mat3Mul(m, Mat3Inverse(m)), "m * inverse is identity")
	eqM3(t, Mat3Identity(), Mat3Mul(Mat3Inverse(m), m), "inverse * m is identity")
	// singular matrix returns identity
	eqM3(t, Mat3Identity(), Mat3Inverse(math32.Mat3(1, 2, 3, 2, 4, 6, 3, 6, 9)), "singular")
	eqf(t, 0, Mat3Determinant(math32.Mat3(1, 2, 3, 2, 4, 6, 3, 6, 9)), "singular determinant")
}

func TestMatrix3Rotation(t *testing.T) {
	// rotating X by 90 degrees about Z gives Y
	q := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), math32.DegToRad(90))
	m := QuatToMatrix3(q)
	eq3(t, math32.Vec3(0, 1, 0), Mat3MulVector3(m, math32.Vec3(1, 0, 0)), "rotate X about Z")
	// a rotation matrix inverse is its transpose
	eqM3(t, Mat3Transpose(m), Mat3Inverse(m), "rotation inverse is transpose")
	eqf(t, 1, Mat3Determinant(m), "rotation determinant is 1")
}

// wgslMul3 computes the standard column-major matrix product m * n,
// which is what the WGSL mat3x3f * operator does. Both Go and WGSL store
// element (row r, col c) at flat index c*3+r, so this is directly
// comparable to the Go results.
func wgslMul3(m, n math32.Matrix3) math32.Matrix3 {
	var o math32.Matrix3
	for c := range 3 {
		for r := range 3 {
			var sum float32
			for k := range 3 {
				sum += m[k*3+r] * n[c*3+k]
			}
			o[c*3+r] = sum
		}
	}
	return o
}

// TestMatrix3MulWGSLOrder checks the operand order that gosl uses when it
// translates a [math32.Matrix3] Mul method call into the WGSL * operator.
// math32.Matrix3.Mul applies the argument's transform first, so a.Mul(b)
// is the standard product b * a, which is why gosl swaps the operands.
func TestMatrix3MulWGSLOrder(t *testing.T) {
	for _, a := range testMatrix3s() {
		for _, b := range testMatrix3s() {
			eqM3(t, Mat3Mul(a, b), wgslMul3(b, a), "a.Mul(b) is WGSL b*a")
		}
	}
}

// TestMatrix3MulVector3WGSL checks that Mat3MulVector3 is the standard
// matrix-vector product, which is what WGSL m * v does, so gosl can
// translate MulVector3 directly to the * operator.
func TestMatrix3MulVector3WGSL(t *testing.T) {
	for _, m := range testMatrix3s() {
		for _, v := range testVec3s {
			want := math32.Vec3(
				m[0]*v.X+m[3]*v.Y+m[6]*v.Z,
				m[1]*v.X+m[4]*v.Y+m[7]*v.Z,
				m[2]*v.X+m[5]*v.Y+m[8]*v.Z)
			eq3(t, want, Mat3MulVector3(m, v), "Mat3MulVector3 is WGSL m*v")
		}
	}
}
