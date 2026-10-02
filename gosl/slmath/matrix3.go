// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// Note: [math32.Matrix3] is stored column-wise, so element i of the Go
// array is column i/3, row i%3, which is exactly how the WGSL mat3x3f
// that it maps to is indexed. [math32.Mat3] takes its 9 values in that
// same column-wise order, matching the WGSL mat3x3f constructor.
// All of the Mul functions multiply the vector or point on the right,
// so a chain of Mat3Mul calls applies its transforms right to left.

// Mat3Identity returns the identity [math32.Matrix3] matrix.
func Mat3Identity() math32.Matrix3 {
	return math32.Mat3(
		1, 0, 0,
		0, 1, 0,
		0, 0, 1)
}

// Mat3FromMatrix2 returns a [math32.Matrix3] from a [math32.Matrix2],
// as a standard homogeneous 2D affine matrix with the translation
// in the third column.
func Mat3FromMatrix2(m math32.Matrix2) math32.Matrix3 {
	return math32.Mat3(
		m.XX, m.YX, 0,
		m.XY, m.YY, 0,
		m.X0, m.Y0, 1)
}

// Mat3FromMatrix4 returns a [math32.Matrix3] from the upper-left 3x3
// of a [math32.Matrix4].
func Mat3FromMatrix4(m math32.Matrix4) math32.Matrix3 {
	return math32.Mat3(
		m[0], m[1], m[2],
		m[4], m[5], m[6],
		m[8], m[9], m[10])
}

// Mat3Translate2D returns a [math32.Matrix3] 2D matrix
// with the given translations.
func Mat3Translate2D(x, y float32) math32.Matrix3 {
	return Mat3FromMatrix2(Mat2Translate2D(x, y))
}

// Mat3Scale2D returns a [math32.Matrix3] 2D matrix
// with the given scaling factors.
func Mat3Scale2D(x, y float32) math32.Matrix3 {
	return Mat3FromMatrix2(Mat2Scale2D(x, y))
}

// Mat3Rotate2D returns a [math32.Matrix3] 2D matrix
// with the given rotation, specified in radians.
func Mat3Rotate2D(angle float32) math32.Matrix3 {
	return Mat3FromMatrix2(Mat2Rotate2D(angle))
}

// Mat3Mul returns the standard matrix multiplication a * b: element
// (row i, column j) of the result is row i of a dotted with column j of b.
// This is the same as [math32.Matrix3.Mul] and the WGSL mat3x3f * operator.
func Mat3Mul(a, b math32.Matrix3) math32.Matrix3 {
	return math32.Mat3(
		a[0]*b[0]+a[3]*b[1]+a[6]*b[2],
		a[1]*b[0]+a[4]*b[1]+a[7]*b[2],
		a[2]*b[0]+a[5]*b[1]+a[8]*b[2],

		a[0]*b[3]+a[3]*b[4]+a[6]*b[5],
		a[1]*b[3]+a[4]*b[4]+a[7]*b[5],
		a[2]*b[3]+a[5]*b[4]+a[8]*b[5],

		a[0]*b[6]+a[3]*b[7]+a[6]*b[8],
		a[1]*b[6]+a[4]*b[7]+a[7]*b[8],
		a[2]*b[6]+a[5]*b[7]+a[8]*b[8])
}

// Mat3MulScalar returns the matrix with each element
// multiplied by the given scalar.
func Mat3MulScalar(m math32.Matrix3, s float32) math32.Matrix3 {
	return math32.Mat3(
		m[0]*s, m[1]*s, m[2]*s,
		m[3]*s, m[4]*s, m[5]*s,
		m[6]*s, m[7]*s, m[8]*s)
}

// Mat3MulVector2 multiplies the [math32.Vector2] as a vector,
// without adding translations. This is for directional vectors, not points.
func Mat3MulVector2(a math32.Matrix3, v math32.Vector2) math32.Vector2 {
	return math32.Vec2(a[0]*v.X+a[3]*v.Y, a[1]*v.X+a[4]*v.Y)
}

// Mat3MulPoint2 multiplies the [math32.Vector2] as a point,
// including adding translations.
func Mat3MulPoint2(a math32.Matrix3, v math32.Vector2) math32.Vector2 {
	return math32.Vec2(a[0]*v.X+a[3]*v.Y+a[6], a[1]*v.X+a[4]*v.Y+a[7])
}

// Mat3MulVector3 multiplies the [math32.Vector3] on the right,
// as a standard matrix multiply.
func Mat3MulVector3(a math32.Matrix3, v math32.Vector3) math32.Vector3 {
	return math32.Vec3(
		a[0]*v.X+a[3]*v.Y+a[6]*v.Z,
		a[1]*v.X+a[4]*v.Y+a[7]*v.Z,
		a[2]*v.X+a[5]*v.Y+a[8]*v.Z)
}

// Mat3Determinant returns the determinant of the matrix.
func Mat3Determinant(m math32.Matrix3) float32 {
	return m[0]*m[4]*m[8] -
		m[0]*m[5]*m[7] -
		m[1]*m[3]*m[8] +
		m[1]*m[5]*m[6] +
		m[2]*m[3]*m[7] -
		m[2]*m[4]*m[6]
}

// Mat3Inverse returns the inverse of the matrix, or the identity matrix
// if it cannot be inverted (i.e., the determinant is 0).
func Mat3Inverse(m math32.Matrix3) math32.Matrix3 {
	n11 := m[0]
	n21 := m[1]
	n31 := m[2]
	n12 := m[3]
	n22 := m[4]
	n32 := m[5]
	n13 := m[6]
	n23 := m[7]
	n33 := m[8]

	t11 := n33*n22 - n32*n23
	t12 := n32*n13 - n33*n12
	t13 := n23*n12 - n22*n13

	det := n11*t11 + n21*t12 + n31*t13
	if det == 0 {
		return Mat3Identity()
	}
	di := 1 / det

	return math32.Mat3(
		t11*di,
		(n31*n23-n33*n21)*di,
		(n32*n21-n31*n22)*di,

		t12*di,
		(n33*n11-n31*n13)*di,
		(n31*n12-n32*n11)*di,

		t13*di,
		(n21*n13-n23*n11)*di,
		(n22*n11-n21*n12)*di)
}

// Mat3Transpose returns the transpose of the matrix.
func Mat3Transpose(m math32.Matrix3) math32.Matrix3 {
	return math32.Mat3(
		m[0], m[3], m[6],
		m[1], m[4], m[7],
		m[2], m[5], m[8])
}

// Mat3ScaleCols returns the matrix with its columns multiplied by the
// corresponding vector components. This can be used to multiply the matrix
// by a diagonal matrix stored as a vector of its diagonal components.
func Mat3ScaleCols(m math32.Matrix3, v math32.Vector3) math32.Matrix3 {
	return math32.Mat3(
		m[0]*v.X, m[1]*v.X, m[2]*v.X,
		m[3]*v.Y, m[4]*v.Y, m[5]*v.Y,
		m[6]*v.Z, m[7]*v.Z, m[8]*v.Z)
}

// Mat3NormalMatrix returns the matrix that transforms normal vectors
// from the given matrix, which is used to transform the vertices
// (e.g., a ModelView matrix).
func Mat3NormalMatrix(m math32.Matrix4) math32.Matrix3 {
	return Mat3Transpose(Mat3Inverse(Mat3FromMatrix4(m)))
}

//gosl:end
