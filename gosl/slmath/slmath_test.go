// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"strconv"

	"cogentcore.org/core/base/tolassert"
	"cogentcore.org/core/math32"
)

// tol is the tolerance used for all float comparisons in these tests.
// The slmath functions are re-expressed versions of the math32 methods,
// so results can differ in the last bits but not more than this.
const tol = float32(1.0e-5)

// eqf asserts that got is about equal to want, using a tolerance that
// scales with the magnitude of the value, since float32 round-off in
// the re-expressed slmath computations grows with magnitude.
func eqf(t *testing.T, want, got float32, msg string) bool {
	t.Helper()
	return tolassert.EqualTol(t, want, got, tol*max(1, math32.Abs(want)), msg)
}

func eqb(t *testing.T, want, got bool, msg string) {
	t.Helper()
	if want != got {
		t.Errorf("%s: want %v, got %v", msg, want, got)
	}
}

func eq2(t *testing.T, want, got math32.Vector2, msg string) {
	t.Helper()
	eqf(t, want.X, got.X, msg+".X")
	eqf(t, want.Y, got.Y, msg+".Y")
}

func eq3(t *testing.T, want, got math32.Vector3, msg string) {
	t.Helper()
	eqf(t, want.X, got.X, msg+".X")
	eqf(t, want.Y, got.Y, msg+".Y")
	eqf(t, want.Z, got.Z, msg+".Z")
}

func eq4(t *testing.T, want, got math32.Vector4, msg string) {
	t.Helper()
	eqf(t, want.X, got.X, msg+".X")
	eqf(t, want.Y, got.Y, msg+".Y")
	eqf(t, want.Z, got.Z, msg+".Z")
	eqf(t, want.W, got.W, msg+".W")
}

func eqQ(t *testing.T, want, got math32.Quat, msg string) {
	t.Helper()
	eqf(t, want.X, got.X, msg+".X")
	eqf(t, want.Y, got.Y, msg+".Y")
	eqf(t, want.Z, got.Z, msg+".Z")
	eqf(t, want.W, got.W, msg+".W")
}

func eqM2(t *testing.T, want, got math32.Matrix2, msg string) {
	t.Helper()
	eqf(t, want.XX, got.XX, msg+".XX")
	eqf(t, want.YX, got.YX, msg+".YX")
	eqf(t, want.XY, got.XY, msg+".XY")
	eqf(t, want.YY, got.YY, msg+".YY")
	eqf(t, want.X0, got.X0, msg+".X0")
	eqf(t, want.Y0, got.Y0, msg+".Y0")
}

func eqM3(t *testing.T, want, got math32.Matrix3, msg string) {
	t.Helper()
	for i := range want {
		eqf(t, want[i], got[i], msg+"["+strconv.Itoa(i)+"]")
	}
}

func eqM4(t *testing.T, want, got math32.Matrix4, msg string) {
	t.Helper()
	for i := range want {
		eqf(t, want[i], got[i], msg+"["+strconv.Itoa(i)+"]")
	}
}

// Test input values, shared across the type-specific test files.
// They deliberately include zero, negative, fractional and
// large-magnitude values.

var testVec2s = []math32.Vector2{
	math32.Vec2(0, 0), math32.Vec2(1, 0), math32.Vec2(0, 1),
	math32.Vec2(1, 1), math32.Vec2(-1, 1), math32.Vec2(3, -4),
	math32.Vec2(0.5, -0.5), math32.Vec2(-2.25, 7.75),
	math32.Vec2(100, 0.001), math32.Vec2(-1.5, -2.5),
}

var testVec3s = []math32.Vector3{
	math32.Vec3(0, 0, 0), math32.Vec3(1, 0, 0), math32.Vec3(0, 1, 0),
	math32.Vec3(0, 0, 1), math32.Vec3(1, 1, 1), math32.Vec3(-1, 2, -3),
	math32.Vec3(3, -4, 12), math32.Vec3(0.5, -0.5, 0.25),
	math32.Vec3(-2.25, 7.75, -0.125), math32.Vec3(100, 0.001, -50),
}

var testVec4s = []math32.Vector4{
	math32.Vec4(0, 0, 0, 0), math32.Vec4(1, 0, 0, 0), math32.Vec4(0, 1, 0, 0),
	math32.Vec4(0, 0, 1, 0), math32.Vec4(0, 0, 0, 1), math32.Vec4(1, 1, 1, 1),
	math32.Vec4(-1, 2, -3, 4), math32.Vec4(3, -4, 12, -84),
	math32.Vec4(0.5, -0.5, 0.25, -0.125), math32.Vec4(100, 0.001, -50, 2),
}

var testQuats = []math32.Quat{
	math32.NewQuat(0, 0, 0, 1),
	math32.NewQuat(0.7071068, 0, 0, 0.7071068),
	math32.NewQuat(0, 0.7071068, 0, 0.7071068),
	math32.NewQuat(0, 0, 0.7071068, 0.7071068),
	math32.NewQuat(0.034270797, 0.10602052, 0.14357218, 0.9833474),
	math32.NewQuat(-0.5, 0.5, -0.5, 0.5),
	math32.NewQuat(0.1, 0.2, 0.3, 0.4),
}

var testScalars = []float32{0, 1, -1, 0.5, -2.5, 3.75}

// testMatrix2s returns a set of Matrix2 test values.
func testMatrix2s() []math32.Matrix2 {
	return []math32.Matrix2{
		math32.Identity2(),
		math32.Translate2D(3, -4),
		math32.Scale2D(2, 0.5),
		math32.Rotate2D(math32.DegToRad(30)),
		math32.Shear2D(0.3, -0.7),
		math32.Skew2D(0.2, 0.4),
		math32.Identity2().Translate(1, 2).Rotate(math32.DegToRad(45)).Scale(2, 3),
		math32.Matrix2{XX: 1, YX: 2, XY: 3, YY: 4, X0: 5, Y0: 6},
	}
}

// testMatrix3s returns a set of Matrix3 test values.
func testMatrix3s() []math32.Matrix3 {
	rq := math32.Matrix3{}
	rq.SetRotationFromQuat(math32.NewQuatEuler(math32.Vec3(0.3, -0.5, 0.9)))
	return []math32.Matrix3{
		math32.Identity3(),
		math32.Matrix3Translate2D(3, -4),
		math32.Matrix3Scale2D(2, 0.5),
		math32.Matrix3Rotate2D(math32.DegToRad(30)),
		rq,
		math32.Mat3(1, 2, 3, 4, 5, 6, 7, 8, 10),
		math32.Mat3(2, 0, 0, 0, 3, 0, 0, 0, 4),
	}
}

// testMatrix4s returns a set of Matrix4 test values.
func testMatrix4s() []math32.Matrix4 {
	id := *math32.Identity4()
	tr := *math32.Identity4()
	tr.SetTranslation(1, -2, 3)
	sc := *math32.Identity4()
	sc.SetScale(2, 0.5, -3)
	rx := *math32.Identity4()
	rx.SetRotationX(math32.DegToRad(30))
	ry := *math32.Identity4()
	ry.SetRotationY(math32.DegToRad(-70))
	rz := *math32.Identity4()
	rz.SetRotationZ(math32.DegToRad(125))
	rq := *math32.Identity4()
	rq.SetRotationFromQuat(math32.NewQuatEuler(math32.Vec3(0.3, -0.5, 0.9)))
	xf := *math32.Identity4()
	xf.SetTransform(math32.Vec3(1, 2, 3), math32.NewQuatEuler(math32.Vec3(0.1, 0.2, 0.3)), math32.Vec3(2, 3, 4))
	var gen math32.Matrix4
	gen.Set(
		1, 2, 3, 4,
		5, 7, 6, 8,
		9, 10, 12, 11,
		13, 15, 14, 17)
	return []math32.Matrix4{id, tr, sc, rx, ry, rz, rq, xf, gen}
}
