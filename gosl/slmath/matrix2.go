// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// Note: matrices are always constructed by assigning to the fields of a
// var, never with a composite literal, because the WGSL mat2x3f that
// [math32.Matrix2] maps to orders its values by column, which is not the
// order of the Go struct fields.

// Mat2Set returns a [math32.Matrix2] with the given element values.
func Mat2Set(xx, yx, xy, yy, x0, y0 float32) math32.Matrix2 {
	var m math32.Matrix2
	m.XX = xx
	m.YX = yx
	m.XY = xy
	m.YY = yy
	m.X0 = x0
	m.Y0 = y0
	return m
}

// Mat2Identity returns the identity [math32.Matrix2] matrix.
func Mat2Identity() math32.Matrix2 {
	return Mat2Set(1, 0, 0, 1, 0, 0)
}

// Mat2IsIdentity returns whether the matrix is the identity matrix.
func Mat2IsIdentity(m math32.Matrix2) bool {
	return m.XX == 1 && m.YX == 0 && m.XY == 0 && m.YY == 1 && m.X0 == 0 && m.Y0 == 0
}

// Mat2Translate2D returns a [math32.Matrix2] with the given translations.
func Mat2Translate2D(x, y float32) math32.Matrix2 {
	return Mat2Set(1, 0, 0, 1, x, y)
}

// Mat2Scale2D returns a [math32.Matrix2] with the given scaling factors.
func Mat2Scale2D(x, y float32) math32.Matrix2 {
	return Mat2Set(x, 0, 0, y, 0, 0)
}

// Mat2Rotate2D returns a [math32.Matrix2] with the given rotation,
// specified in radians. This uses the standard graphics convention where
// increasing Y goes _down_ instead of up, in contrast with the mathematical
// coordinate system where Y is up.
func Mat2Rotate2D(angle float32) math32.Matrix2 {
	s := math32.Sin(angle)
	c := math32.Cos(angle)
	return Mat2Set(c, s, -s, c, 0, 0)
}

// Mat2Shear2D returns a [math32.Matrix2] with the given shearing.
func Mat2Shear2D(x, y float32) math32.Matrix2 {
	return Mat2Set(1, y, x, 1, 0, 0)
}

// Mat2Skew2D returns a [math32.Matrix2] with the given skewing.
func Mat2Skew2D(x, y float32) math32.Matrix2 {
	return Mat2Set(1, math32.Tan(y), math32.Tan(x), 1, 0, 0)
}

// Mat2Mul returns a * b.
func Mat2Mul(a, b math32.Matrix2) math32.Matrix2 {
	return Mat2Set(
		a.XX*b.XX+a.XY*b.YX,
		a.YX*b.XX+a.YY*b.YX,
		a.XX*b.XY+a.XY*b.YY,
		a.YX*b.XY+a.YY*b.YY,
		a.XX*b.X0+a.XY*b.Y0+a.X0,
		a.YX*b.X0+a.YY*b.Y0+a.Y0)
}

// Mat2MulVector multiplies the [math32.Vector2] as a vector,
// without adding translations. This is for directional vectors, not points.
func Mat2MulVector(a math32.Matrix2, v math32.Vector2) math32.Vector2 {
	return math32.Vec2(a.XX*v.X+a.XY*v.Y, a.YX*v.X+a.YY*v.Y)
}

// Mat2MulPoint multiplies the [math32.Vector2] as a point,
// including adding translations.
func Mat2MulPoint(a math32.Matrix2, v math32.Vector2) math32.Vector2 {
	return math32.Vec2(a.XX*v.X+a.XY*v.Y+a.X0, a.YX*v.X+a.YY*v.Y+a.Y0)
}

// Mat2Translate adds a translation to the given matrix.
func Mat2Translate(a math32.Matrix2, x, y float32) math32.Matrix2 {
	return Mat2Mul(a, Mat2Translate2D(x, y))
}

// Mat2Scale adds a scaling to the given matrix.
func Mat2Scale(a math32.Matrix2, x, y float32) math32.Matrix2 {
	return Mat2Mul(a, Mat2Scale2D(x, y))
}

// Mat2ScaleAbout adds a scaling transformation about (x,y) in sx and sy.
// When scale is negative it will flip those axes.
func Mat2ScaleAbout(a math32.Matrix2, sx, sy, x, y float32) math32.Matrix2 {
	return Mat2Translate(Mat2Scale(Mat2Translate(a, x, y), sx, sy), -x, -y)
}

// Mat2Rotate adds a rotation to the given matrix, in radians.
func Mat2Rotate(a math32.Matrix2, angle float32) math32.Matrix2 {
	return Mat2Mul(a, Mat2Rotate2D(angle))
}

// Mat2RotateAbout adds a rotation transformation about (x,y),
// with rot in radians counter-clockwise.
func Mat2RotateAbout(a math32.Matrix2, rot, x, y float32) math32.Matrix2 {
	return Mat2Translate(Mat2Rotate(Mat2Translate(a, x, y), rot), -x, -y)
}

// Mat2Rotate2DAround returns a [math32.Matrix2] with the given rotation,
// specified in radians, around the given offset point that is
// translated to and from.
func Mat2Rotate2DAround(angle float32, pos math32.Vector2) math32.Matrix2 {
	return Mat2Translate(Mat2Rotate(Mat2Translate(Mat2Identity(), pos.X, pos.Y), angle), -pos.X, -pos.Y)
}

// Mat2Shear adds a shearing to the given matrix.
func Mat2Shear(a math32.Matrix2, x, y float32) math32.Matrix2 {
	return Mat2Mul(a, Mat2Shear2D(x, y))
}

// Mat2Skew adds a skewing to the given matrix.
func Mat2Skew(a math32.Matrix2, x, y float32) math32.Matrix2 {
	return Mat2Mul(a, Mat2Skew2D(x, y))
}

// Mat2ExtractRot does a simple extraction of the rotation
// for a single rotation.
func Mat2ExtractRot(a math32.Matrix2) float32 {
	return math32.Atan2(-a.XY, a.XX)
}

// Mat2Pos returns the translation values, X0, Y0, as a [math32.Vector2].
func Mat2Pos(a math32.Matrix2) math32.Vector2 {
	return math32.Vec2(a.X0, a.Y0)
}

// Mat2Transpose returns the transpose of the matrix.
func Mat2Transpose(a math32.Matrix2) math32.Matrix2 {
	return Mat2Set(a.XX, a.XY, a.YX, a.YY, a.X0, a.Y0)
}

// Mat2Det returns the determinant of the matrix.
func Mat2Det(a math32.Matrix2) float32 {
	return a.XX*a.YY - a.XY*a.YX
}

// Mat2Inverse returns the inverse of the matrix, for inverting transforms.
func Mat2Inverse(a math32.Matrix2) math32.Matrix2 {
	det := Mat2Det(a)
	if det == 0 {
		return Mat2Identity()
	}
	di := 1 / det
	return Mat2Set(
		a.YY*di,
		-a.YX*di,
		-a.XY*di,
		a.XX*di,
		(a.Y0*a.XY-a.YY*a.X0)*di,
		(a.X0*a.YX-a.XX*a.Y0)*di)
}

//gosl:end
