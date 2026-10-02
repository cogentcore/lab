// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
)

// TestMatrix4VsMath32 checks every Mat4 function against the
// equivalent [math32.Matrix4] method or constructor.
func TestMatrix4VsMath32(t *testing.T) {
	eqM4(t, *math32.Identity4(), Mat4Identity(), "Mat4Identity")

	for _, a := range testMatrix4s() {
		na := a
		eqM4(t, *na.Transpose(), Mat4Transpose(a), "Mat4Transpose")
		eqf(t, a.Determinant(), Mat4Determinant(a), "Mat4Determinant")
		eqf(t, a.GetMaxScaleOnAxis(), Mat4MaxScaleOnAxis(a), "Mat4MaxScaleOnAxis")
		eq3(t, a.Pos(), Mat4Pos(a), "Mat4Pos")
		eq3(t, math32.NewEulerAnglesFromMatrix(&a), Mat4ToEuler(a), "Mat4ToEuler")

		if a.Determinant() != 0 {
			inv, err := a.Inverse()
			if err == nil {
				eqM4(t, *inv, Mat4Inverse(a), "Mat4Inverse")
			}
		}

		er := *math32.Identity4()
		er.ExtractRotation(&a)
		eqM4(t, er, Mat4ExtractRotation(a), "Mat4ExtractRotation")

		for _, b := range testMatrix4s() {
			nb := b
			eqM4(t, *na.Mul(&nb), Mat4Mul(a, b), "Mat4Mul")
			cp := a
			cp.CopyPos(&nb)
			eqM4(t, cp, Mat4CopyPos(a, b), "Mat4CopyPos")
		}
		for _, s := range testScalars {
			ms := a
			ms.MulScalar(s)
			eqM4(t, ms, Mat4MulScalar(a, s), "Mat4MulScalar")
		}
		for _, v := range testVec3s {
			eq3(t, v.MulMatrix4(&a), Mat4MulVector3(a, v), "Mat4MulVector3")
			eq3(t, v.MulProjection(&a), Mat4MulProjection3(a, v), "Mat4MulProjection3")
			sp := a
			sp.SetPos(v)
			eqM4(t, sp, Mat4SetPos(a, v), "Mat4SetPos")
			sc := a
			sc.SetScaleCols(v)
			eqM4(t, sc, Mat4ScaleCols(a, v), "Mat4ScaleCols")
			for _, w := range testScalars {
				eq3(t, v.MulMatrix4AsVector4(&a, w), Mat4MulVector3AsVector4(a, v, w), "Mat4MulVector3AsVector4")
			}
		}
		for _, v := range testVec4s {
			eq4(t, v.MulMatrix4(&a), Mat4MulVector4(a, v), "Mat4MulVector4")
		}
	}

	for _, m3 := range testMatrix3s() {
		nm := *math32.Identity4()
		nm.SetFromMatrix3(&m3)
		eqM4(t, nm, Mat4FromMatrix3(m3), "Mat4FromMatrix3")
	}

	for _, s := range testScalars {
		rx := *math32.Identity4()
		rx.SetRotationX(s)
		eqM4(t, rx, Mat4RotateX(s), "Mat4RotateX")
		ry := *math32.Identity4()
		ry.SetRotationY(s)
		eqM4(t, ry, Mat4RotateY(s), "Mat4RotateY")
		rz := *math32.Identity4()
		rz.SetRotationZ(s)
		eqM4(t, rz, Mat4RotateZ(s), "Mat4RotateZ")

		for _, v := range testVec3s {
			ra := *math32.Identity4()
			ra.SetRotationAxis(&v, s)
			eqM4(t, ra, Mat4RotateAxis(v, s), "Mat4RotateAxis")
		}
	}

	for _, v := range testVec3s {
		tr := *math32.Identity4()
		tr.SetTranslation(v.X, v.Y, v.Z)
		eqM4(t, tr, Mat4Translate3D(v.X, v.Y, v.Z), "Mat4Translate3D")
		sc := *math32.Identity4()
		sc.SetScale(v.X, v.Y, v.Z)
		eqM4(t, sc, Mat4Scale3D(v.X, v.Y, v.Z), "Mat4Scale3D")
		re := *math32.Identity4()
		re.SetRotationFromEuler(v)
		eqM4(t, re, Mat4FromEuler(v), "Mat4FromEuler")
		bs := *math32.Identity4()
		bs.SetBasis(v, math32.Vec3(v.Z, v.X, v.Y), math32.Vec3(v.Y, v.Z, v.X))
		eqM4(t, bs, Mat4SetBasis(v, math32.Vec3(v.Z, v.X, v.Y), math32.Vec3(v.Y, v.Z, v.X)), "Mat4SetBasis")
	}

	for _, q := range testQuats {
		rq := *math32.Identity4()
		rq.SetRotationFromQuat(q)
		eqM4(t, rq, Mat4FromQuat(q), "Mat4FromQuat")
		for _, p := range testVec3s {
			for _, s := range testVec3s {
				xf := *math32.Identity4()
				xf.SetTransform(p, q, s)
				eqM4(t, xf, Mat4Transform(p, q, s), "Mat4Transform")
			}
		}
	}
}

func TestMatrix4Transforms(t *testing.T) {
	vx := math32.Vec3(1, 0, 0)
	vy := math32.Vec3(0, 1, 0)
	vz := math32.Vec3(0, 0, 1)

	eq3(t, vx, Mat4MulVector3(Mat4Identity(), vx), "identity")
	eq3(t, math32.Vec3(2, -2, 4), Mat4MulVector3(Mat4Translate3D(1, -2, 4), vx), "translate")
	eq3(t, math32.Vec3(2, 0, 0), Mat4MulVector3(Mat4Scale3D(2, 3, 4), vx), "scale")

	eq3(t, Negate3(vz), Mat4MulVector3(Mat4RotateY(math32.DegToRad(90)), vx), "rotate Y 90")
	eq3(t, vy, Mat4MulVector3(Mat4RotateZ(math32.DegToRad(90)), vx), "rotate Z 90")
	eq3(t, vz, Mat4MulVector3(Mat4RotateX(math32.DegToRad(90)), vy), "rotate X 90")
	eq3(t, vy, Mat4MulVector3(Mat4RotateAxis(vz, math32.DegToRad(90)), vx), "rotate axis Z 90")

	// multiplication order is the *reverse* of the "logical" order:
	// 1,0,0 -> scale(2) = 2,0,0 -> rotate Z 90 = 0,2,0 -> trans 1,1,1 -> 1,3,1
	xf := Mat4Mul(Mat4Mul(Mat4Translate3D(1, 1, 1), Mat4RotateZ(math32.DegToRad(90))), Mat4Scale3D(2, 2, 2))
	eq3(t, math32.Vec3(1, 3, 1), Mat4MulVector3(xf, vx), "composed transform")

	eqM4(t, Mat4Identity(), Mat4Mul(xf, Mat4Inverse(xf)), "m * inverse is identity")
	eq3(t, vx, Mat4MulVector3(Mat4Inverse(xf), Mat4MulVector3(xf, vx)), "inverse round trip")
}

func TestMatrix4Inverse(t *testing.T) {
	// singular matrix returns identity
	eqM4(t, Mat4Identity(), Mat4Inverse(Mat4Scale3D(0, 0, 0)), "singular Mat4Inverse")
	eqf(t, 0, Mat4Determinant(Mat4Scale3D(0, 0, 0)), "singular determinant")
	eqf(t, 24, Mat4Determinant(Mat4Scale3D(2, 3, 4)), "scale determinant")
}

func TestMatrix4ScaleCols(t *testing.T) {
	// note: math32.Matrix4.ScaleCols returns a zero matrix because it does
	// not copy the source first; Mat4ScaleCols does the intended thing,
	// matching math32.Matrix4.SetScaleCols and math32.Matrix3.ScaleCols.
	m := Mat4ScaleCols(Mat4Identity(), math32.Vec3(2, 3, 4))
	eqM4(t, Mat4Scale3D(2, 3, 4), m, "Mat4ScaleCols of identity is a scale matrix")
}

func TestMatrix4Transform(t *testing.T) {
	pos := math32.Vec3(1, 2, 3)
	q := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), math32.DegToRad(90))
	scale := math32.Vec3(2, 2, 2)
	m := Mat4Transform(pos, q, scale)
	// 1,0,0 -> scale 2 -> 2,0,0 -> rotate Z 90 -> 0,2,0 -> + pos -> 1,4,3
	eq3(t, math32.Vec3(1, 4, 3), Mat4MulVector3(m, math32.Vec3(1, 0, 0)), "transform")
	eq3(t, pos, Mat4Pos(m), "Mat4Pos")
	eqM4(t, Mat4FromQuat(q), Mat4ExtractRotation(m), "Mat4ExtractRotation")
}

// wgslMul4 computes the standard column-major matrix product m * n,
// which is what the WGSL mat4x4f * operator does.
func wgslMul4(m, n math32.Matrix4) math32.Matrix4 {
	var o math32.Matrix4
	for c := range 4 {
		for r := range 4 {
			var sum float32
			for k := range 4 {
				sum += m[k*4+r] * n[c*4+k]
			}
			o[c*4+r] = sum
		}
	}
	return o
}

// TestMatrix4MulWGSLOrder checks the operand order that gosl uses when it
// translates a [math32.Matrix4] Mul method call into the WGSL * operator.
// Unlike Matrix3, math32.Matrix4.Mul is already the standard product a * b,
// so gosl keeps the operands in order.
func TestMatrix4MulWGSLOrder(t *testing.T) {
	for _, a := range testMatrix4s() {
		for _, b := range testMatrix4s() {
			eqM4(t, Mat4Mul(a, b), wgslMul4(a, b), "a.Mul(b) is WGSL a*b")
		}
	}
}

// TestMat4Set checks that Mat4Set takes its values row by row and stores
// them column-wise, exactly as [math32.Matrix4.Set] does. The flat index
// of element (row r, col c) is c*4+r, which is what lets gosl translate
// a flat Go index k into the WGSL mat4x4f index [k/4][k%4].
func TestMat4Set(t *testing.T) {
	var want math32.Matrix4
	want.Set(
		11, 12, 13, 14,
		21, 22, 23, 24,
		31, 32, 33, 34,
		41, 42, 43, 44)
	got := Mat4Set(
		11, 12, 13, 14,
		21, 22, 23, 24,
		31, 32, 33, 34,
		41, 42, 43, 44)
	eqM4(t, want, got, "Mat4Set")
	for r := range 4 {
		for c := range 4 {
			eqf(t, float32((r+1)*10+c+1), got[c*4+r], "element by column-wise flat index")
		}
	}
}
