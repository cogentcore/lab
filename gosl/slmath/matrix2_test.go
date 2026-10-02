// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
	"github.com/stretchr/testify/assert"
)

// TestMatrix2VsMath32 checks every Mat2 function against the
// equivalent [math32.Matrix2] method or constructor.
func TestMatrix2VsMath32(t *testing.T) {
	eqM2(t, math32.Identity2(), Mat2Identity(), "Mat2Identity")

	for _, a := range testMatrix2s() {
		eqb(t, a.IsIdentity(), Mat2IsIdentity(a), "Mat2IsIdentity")
		eqM2(t, a.Transpose(), Mat2Transpose(a), "Mat2Transpose")
		eqf(t, a.Det(), Mat2Det(a), "Mat2Det")
		eqf(t, a.ExtractRot(), Mat2ExtractRot(a), "Mat2ExtractRot")
		tx, ty := a.Pos()
		eq2(t, math32.Vec2(tx, ty), Mat2Pos(a), "Mat2Pos")
		if a.Det() != 0 {
			eqM2(t, a.Inverse(), Mat2Inverse(a), "Mat2Inverse")
		}

		for _, b := range testMatrix2s() {
			eqM2(t, a.Mul(b), Mat2Mul(a, b), "Mat2Mul")
		}
		for _, v := range testVec2s {
			eq2(t, a.MulVector(v), Mat2MulVector(a, v), "Mat2MulVector")
			eq2(t, a.MulPoint(v), Mat2MulPoint(a, v), "Mat2MulPoint")
		}
		for _, x := range testScalars {
			for _, y := range testScalars {
				eqM2(t, a.Translate(x, y), Mat2Translate(a, x, y), "Mat2Translate")
				eqM2(t, a.Scale(x, y), Mat2Scale(a, x, y), "Mat2Scale")
				eqM2(t, a.Shear(x, y), Mat2Shear(a, x, y), "Mat2Shear")
				eqM2(t, a.Skew(x, y), Mat2Skew(a, x, y), "Mat2Skew")
				eqM2(t, a.RotateAbout(0.7, x, y), Mat2RotateAbout(a, 0.7, x, y), "Mat2RotateAbout")
				eqM2(t, a.ScaleAbout(2, 3, x, y), Mat2ScaleAbout(a, 2, 3, x, y), "Mat2ScaleAbout")
			}
			eqM2(t, a.Rotate(x), Mat2Rotate(a, x), "Mat2Rotate")
		}
	}

	for _, x := range testScalars {
		for _, y := range testScalars {
			eqM2(t, math32.Translate2D(x, y), Mat2Translate2D(x, y), "Mat2Translate2D")
			eqM2(t, math32.Scale2D(x, y), Mat2Scale2D(x, y), "Mat2Scale2D")
			eqM2(t, math32.Shear2D(x, y), Mat2Shear2D(x, y), "Mat2Shear2D")
			eqM2(t, math32.Skew2D(x, y), Mat2Skew2D(x, y), "Mat2Skew2D")
		}
		eqM2(t, math32.Rotate2D(x), Mat2Rotate2D(x), "Mat2Rotate2D")
		for _, p := range testVec2s {
			eqM2(t, math32.Rotate2DAround(x, p), Mat2Rotate2DAround(x, p), "Mat2Rotate2DAround")
		}
	}
}

func TestMatrix2Transforms(t *testing.T) {
	v0 := math32.Vec2(0, 0)
	vx := math32.Vec2(1, 0)
	vy := math32.Vec2(0, 1)
	vxy := math32.Vec2(1, 1)

	eq2(t, vx, Mat2MulPoint(Mat2Identity(), vx), "identity")
	eq2(t, vxy, Mat2MulPoint(Mat2Translate2D(1, 1), v0), "translate")
	eq2(t, vxy.MulScalar(2), Mat2MulPoint(Mat2Scale2D(2, 2), vxy), "scale")

	eq2(t, vy, Mat2MulPoint(Mat2Rotate2D(math32.DegToRad(90)), vx), "rotate 90")
	eq2(t, vx, Mat2MulPoint(Mat2Rotate2D(math32.DegToRad(-90)), vy), "rotate -90")
	eq2(t, Normal2(vxy), Mat2MulPoint(Mat2Rotate2D(math32.DegToRad(45)), vx), "rotate 45")

	// 1,0 -> scale(2) = 2,0 -> rotate 90 = 0,2 -> trans 1,1 -> 1,3
	// multiplication order is the *reverse* of the "logical" order:
	xf := Mat2Mul(Mat2Mul(Mat2Translate2D(1, 1), Mat2Rotate2D(math32.DegToRad(90))), Mat2Scale2D(2, 2))
	eq2(t, math32.Vec2(1, 3), Mat2MulPoint(xf, vx), "composed transform")

	// inverse round trip
	eq2(t, vx, Mat2MulPoint(Mat2Inverse(xf), Mat2MulPoint(xf, vx)), "inverse round trip")
	eqM2(t, Mat2Identity(), Mat2Mul(xf, Mat2Inverse(xf)), "m * inverse is identity")

	assert.True(t, Mat2IsIdentity(Mat2Identity()))
	assert.False(t, Mat2IsIdentity(Mat2Translate2D(1, 0)))
}

func TestMatrix2SingularInverse(t *testing.T) {
	// unlike math32.Matrix2.Inverse, which divides by zero, a singular
	// matrix returns the identity, as in Mat3Inverse and Mat4Inverse.
	eqM2(t, Mat2Identity(), Mat2Inverse(Mat2Scale2D(0, 0)), "singular Mat2Inverse")
}

// TestMat2Set checks that Mat2Set assigns each value to the field that
// math32 names, which is what fixes the WGSL mat2x3f element order:
// gosl maps XX,XY,X0 to column 0 and YX,YY,Y0 to column 1.
func TestMat2Set(t *testing.T) {
	m := Mat2Set(1, 2, 3, 4, 5, 6)
	eqM2(t, math32.Matrix2{XX: 1, YX: 2, XY: 3, YY: 4, X0: 5, Y0: 6}, m, "Mat2Set")
	eqf(t, 1, m.XX, "XX")
	eqf(t, 2, m.YX, "YX")
	eqf(t, 3, m.XY, "XY")
	eqf(t, 4, m.YY, "YY")
	eqf(t, 5, m.X0, "X0")
	eqf(t, 6, m.Y0, "Y0")
}
