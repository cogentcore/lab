// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import "cogentcore.org/core/math32"

//gosl:start

// Note: [math32.Matrix4] is stored column-wise, so element i of the Go
// array is column i/4, row i%4, which is exactly how the WGSL mat4x4f
// that it maps to is indexed. Matrices are always built by assigning to
// the elements of a var, never with a composite literal, because the
// WGSL constructor takes its values column-wise.

// Mat4Set returns a [math32.Matrix4] with the given element values,
// given row by row, as in [math32.Matrix4.Set].
func Mat4Set(n11, n12, n13, n14, n21, n22, n23, n24, n31, n32, n33, n34, n41, n42, n43, n44 float32) math32.Matrix4 {
	var m math32.Matrix4
	m[0] = n11
	m[4] = n12
	m[8] = n13
	m[12] = n14
	m[1] = n21
	m[5] = n22
	m[9] = n23
	m[13] = n24
	m[2] = n31
	m[6] = n32
	m[10] = n33
	m[14] = n34
	m[3] = n41
	m[7] = n42
	m[11] = n43
	m[15] = n44
	return m
}

// Mat4Identity returns the identity [math32.Matrix4] matrix.
func Mat4Identity() math32.Matrix4 {
	return Mat4Set(
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1)
}

// Mat4FromMatrix3 returns a [math32.Matrix4] with the upper-left 3x3
// set from the given [math32.Matrix3], and the rest set to identity.
func Mat4FromMatrix3(s math32.Matrix3) math32.Matrix4 {
	return Mat4Set(
		s[0], s[3], s[6], 0,
		s[1], s[4], s[7], 0,
		s[2], s[5], s[8], 0,
		0, 0, 0, 1)
}

// Mat4Mul returns the matrix multiplication a * b.
func Mat4Mul(a, b math32.Matrix4) math32.Matrix4 {
	a11 := a[0]
	a12 := a[4]
	a13 := a[8]
	a14 := a[12]
	a21 := a[1]
	a22 := a[5]
	a23 := a[9]
	a24 := a[13]
	a31 := a[2]
	a32 := a[6]
	a33 := a[10]
	a34 := a[14]
	a41 := a[3]
	a42 := a[7]
	a43 := a[11]
	a44 := a[15]

	b11 := b[0]
	b12 := b[4]
	b13 := b[8]
	b14 := b[12]
	b21 := b[1]
	b22 := b[5]
	b23 := b[9]
	b24 := b[13]
	b31 := b[2]
	b32 := b[6]
	b33 := b[10]
	b34 := b[14]
	b41 := b[3]
	b42 := b[7]
	b43 := b[11]
	b44 := b[15]

	return Mat4Set(
		a11*b11+a12*b21+a13*b31+a14*b41,
		a11*b12+a12*b22+a13*b32+a14*b42,
		a11*b13+a12*b23+a13*b33+a14*b43,
		a11*b14+a12*b24+a13*b34+a14*b44,

		a21*b11+a22*b21+a23*b31+a24*b41,
		a21*b12+a22*b22+a23*b32+a24*b42,
		a21*b13+a22*b23+a23*b33+a24*b43,
		a21*b14+a22*b24+a23*b34+a24*b44,

		a31*b11+a32*b21+a33*b31+a34*b41,
		a31*b12+a32*b22+a33*b32+a34*b42,
		a31*b13+a32*b23+a33*b33+a34*b43,
		a31*b14+a32*b24+a33*b34+a34*b44,

		a41*b11+a42*b21+a43*b31+a44*b41,
		a41*b12+a42*b22+a43*b32+a44*b42,
		a41*b13+a42*b23+a43*b33+a44*b43,
		a41*b14+a42*b24+a43*b34+a44*b44)
}

// Mat4MulScalar returns the matrix with each element
// multiplied by the given scalar.
func Mat4MulScalar(m math32.Matrix4, s float32) math32.Matrix4 {
	return Mat4Set(
		m[0]*s, m[4]*s, m[8]*s, m[12]*s,
		m[1]*s, m[5]*s, m[9]*s, m[13]*s,
		m[2]*s, m[6]*s, m[10]*s, m[14]*s,
		m[3]*s, m[7]*s, m[11]*s, m[15]*s)
}

// Mat4MulVector3 returns the [math32.Vector3] multiplied by the matrix,
// including the translation component (i.e., as a point).
func Mat4MulVector3(m math32.Matrix4, v math32.Vector3) math32.Vector3 {
	return math32.Vec3(
		m[0]*v.X+m[4]*v.Y+m[8]*v.Z+m[12],
		m[1]*v.X+m[5]*v.Y+m[9]*v.Z+m[13],
		m[2]*v.X+m[6]*v.Y+m[10]*v.Z+m[14])
}

// Mat4MulVector4 returns the [math32.Vector4] multiplied by the matrix.
func Mat4MulVector4(m math32.Matrix4, v math32.Vector4) math32.Vector4 {
	return math32.Vec4(
		m[0]*v.X+m[4]*v.Y+m[8]*v.Z+m[12]*v.W,
		m[1]*v.X+m[5]*v.Y+m[9]*v.Z+m[13]*v.W,
		m[2]*v.X+m[6]*v.Y+m[10]*v.Z+m[14]*v.W,
		m[3]*v.X+m[7]*v.Y+m[11]*v.Z+m[15]*v.W)
}

// Mat4MulVector3AsVector4 returns the 3D vector multiplied by the matrix
// using a 4D vector with the given 4th dimensional value, then reduced back
// to a 3D vector. This is different from [Mat4MulVector3]. Use 0 for normals
// and 1 for positions as the 4th dim to set.
func Mat4MulVector3AsVector4(m math32.Matrix4, v math32.Vector3, w float32) math32.Vector3 {
	return Vec3FromVec4(Mat4MulVector4(m, Vec4FromVec3(v, w)))
}

// Mat4MulProjection3 returns the [math32.Vector3] multiplied by the
// projection matrix, with the perspective divide applied.
func Mat4MulProjection3(m math32.Matrix4, v math32.Vector3) math32.Vector3 {
	d := 1 / (m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]) // perspective divide
	return math32.Vec3(
		(m[0]*v.X+m[4]*v.Y+m[8]*v.Z+m[12])*d,
		(m[1]*v.X+m[5]*v.Y+m[9]*v.Z+m[13])*d,
		(m[2]*v.X+m[6]*v.Y+m[10]*v.Z+m[14])*d)
}

// Mat4Determinant returns the determinant of the matrix.
func Mat4Determinant(m math32.Matrix4) float32 {
	n11 := m[0]
	n12 := m[4]
	n13 := m[8]
	n14 := m[12]
	n21 := m[1]
	n22 := m[5]
	n23 := m[9]
	n24 := m[13]
	n31 := m[2]
	n32 := m[6]
	n33 := m[10]
	n34 := m[14]
	n41 := m[3]
	n42 := m[7]
	n43 := m[11]
	n44 := m[15]

	// note: no leading unary + on the parenthesized terms, as in math32:
	// WGSL has no unary + operator.
	return n41*(n14*n23*n32-n13*n24*n32-n14*n22*n33+n12*n24*n33+n13*n22*n34-n12*n23*n34) +
		n42*(n11*n23*n34-n11*n24*n33+n14*n21*n33-n13*n21*n34+n13*n24*n31-n14*n23*n31) +
		n43*(n11*n24*n32-n11*n22*n34-n14*n21*n32+n12*n21*n34+n14*n22*n31-n12*n24*n31) +
		n44*(0-n13*n22*n31-n11*n23*n32+n11*n22*n33+n13*n21*n32-n12*n21*n33+n12*n23*n31)
}

// Mat4Inverse returns the inverse of the matrix, or the identity matrix
// if it cannot be inverted (i.e., the determinant is 0).
func Mat4Inverse(s math32.Matrix4) math32.Matrix4 {
	n11 := s[0]
	n12 := s[4]
	n13 := s[8]
	n14 := s[12]
	n21 := s[1]
	n22 := s[5]
	n23 := s[9]
	n24 := s[13]
	n31 := s[2]
	n32 := s[6]
	n33 := s[10]
	n34 := s[14]
	n41 := s[3]
	n42 := s[7]
	n43 := s[11]
	n44 := s[15]

	t11 := n23*n34*n42 - n24*n33*n42 + n24*n32*n43 - n22*n34*n43 - n23*n32*n44 + n22*n33*n44
	t12 := n14*n33*n42 - n13*n34*n42 - n14*n32*n43 + n12*n34*n43 + n13*n32*n44 - n12*n33*n44
	t13 := n13*n24*n42 - n14*n23*n42 + n14*n22*n43 - n12*n24*n43 - n13*n22*n44 + n12*n23*n44
	t14 := n14*n23*n32 - n13*n24*n32 - n14*n22*n33 + n12*n24*n33 + n13*n22*n34 - n12*n23*n34

	det := n11*t11 + n21*t12 + n31*t13 + n41*t14
	if det == 0 {
		return Mat4Identity()
	}
	di := 1 / det

	var m math32.Matrix4
	m[0] = t11 * di
	m[1] = (n24*n33*n41 - n23*n34*n41 - n24*n31*n43 + n21*n34*n43 + n23*n31*n44 - n21*n33*n44) * di
	m[2] = (n22*n34*n41 - n24*n32*n41 + n24*n31*n42 - n21*n34*n42 - n22*n31*n44 + n21*n32*n44) * di
	m[3] = (n23*n32*n41 - n22*n33*n41 - n23*n31*n42 + n21*n33*n42 + n22*n31*n43 - n21*n32*n43) * di

	m[4] = t12 * di
	m[5] = (n13*n34*n41 - n14*n33*n41 + n14*n31*n43 - n11*n34*n43 - n13*n31*n44 + n11*n33*n44) * di
	m[6] = (n14*n32*n41 - n12*n34*n41 - n14*n31*n42 + n11*n34*n42 + n12*n31*n44 - n11*n32*n44) * di
	m[7] = (n12*n33*n41 - n13*n32*n41 + n13*n31*n42 - n11*n33*n42 - n12*n31*n43 + n11*n32*n43) * di

	m[8] = t13 * di
	m[9] = (n14*n23*n41 - n13*n24*n41 - n14*n21*n43 + n11*n24*n43 + n13*n21*n44 - n11*n23*n44) * di
	m[10] = (n12*n24*n41 - n14*n22*n41 + n14*n21*n42 - n11*n24*n42 - n12*n21*n44 + n11*n22*n44) * di
	m[11] = (n13*n22*n41 - n12*n23*n41 - n13*n21*n42 + n11*n23*n42 + n12*n21*n43 - n11*n22*n43) * di

	m[12] = t14 * di
	m[13] = (n13*n24*n31 - n14*n23*n31 + n14*n21*n33 - n11*n24*n33 - n13*n21*n34 + n11*n23*n34) * di
	m[14] = (n14*n22*n31 - n12*n24*n31 - n14*n21*n32 + n11*n24*n32 + n12*n21*n34 - n11*n22*n34) * di
	m[15] = (n12*n23*n31 - n13*n22*n31 + n13*n21*n32 - n11*n23*n32 - n12*n21*n33 + n11*n22*n33) * di
	return m
}

// Mat4Transpose returns the transpose of the matrix.
func Mat4Transpose(m math32.Matrix4) math32.Matrix4 {
	return Mat4Set(
		m[0], m[1], m[2], m[3],
		m[4], m[5], m[6], m[7],
		m[8], m[9], m[10], m[11],
		m[12], m[13], m[14], m[15])
}

// Mat4ScaleCols returns the matrix with its first column multiplied by the
// vector X component, the second column by the vector Y component and the
// third column by the vector Z component. The fourth column is unchanged.
func Mat4ScaleCols(m math32.Matrix4, v math32.Vector3) math32.Matrix4 {
	return Mat4Set(
		m[0]*v.X, m[4]*v.Y, m[8]*v.Z, m[12],
		m[1]*v.X, m[5]*v.Y, m[9]*v.Z, m[13],
		m[2]*v.X, m[6]*v.Y, m[10]*v.Z, m[14],
		m[3]*v.X, m[7]*v.Y, m[11]*v.Z, m[15])
}

// Mat4MaxScaleOnAxis returns the maximum scale value of the 3 axes.
func Mat4MaxScaleOnAxis(m math32.Matrix4) float32 {
	sx := m[0]*m[0] + m[1]*m[1] + m[2]*m[2]
	sy := m[4]*m[4] + m[5]*m[5] + m[6]*m[6]
	sz := m[8]*m[8] + m[9]*m[9] + m[10]*m[10]
	return math32.Sqrt(max(sx, max(sy, sz)))
}

// Mat4Translate3D returns a translation matrix
// for the given x, y and z values.
func Mat4Translate3D(x, y, z float32) math32.Matrix4 {
	return Mat4Set(
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1)
}

// Mat4Scale3D returns a scale transformation matrix
// for the given x, y and z values.
func Mat4Scale3D(x, y, z float32) math32.Matrix4 {
	return Mat4Set(
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1)
}

// Mat4RotateX returns a rotation matrix of angle theta around the X axis.
func Mat4RotateX(theta float32) math32.Matrix4 {
	c := math32.Cos(theta)
	s := math32.Sin(theta)
	return Mat4Set(
		1, 0, 0, 0,
		0, c, -s, 0,
		0, s, c, 0,
		0, 0, 0, 1)
}

// Mat4RotateY returns a rotation matrix of angle theta around the Y axis.
func Mat4RotateY(theta float32) math32.Matrix4 {
	c := math32.Cos(theta)
	s := math32.Sin(theta)
	return Mat4Set(
		c, 0, s, 0,
		0, 1, 0, 0,
		-s, 0, c, 0,
		0, 0, 0, 1)
}

// Mat4RotateZ returns a rotation matrix of angle theta around the Z axis.
func Mat4RotateZ(theta float32) math32.Matrix4 {
	c := math32.Cos(theta)
	s := math32.Sin(theta)
	return Mat4Set(
		c, -s, 0, 0,
		s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1)
}

// Mat4RotateAxis returns a rotation matrix of the given angle
// around the given axis.
func Mat4RotateAxis(axis math32.Vector3, angle float32) math32.Matrix4 {
	c := math32.Cos(angle)
	s := math32.Sin(angle)
	t := 1 - c
	x := axis.X
	y := axis.Y
	z := axis.Z
	tx := t * x
	ty := t * y
	return Mat4Set(
		tx*x+c, tx*y-s*z, tx*z+s*y, 0,
		tx*y+s*z, ty*y+c, ty*z-s*x, 0,
		tx*z-s*y, ty*z+s*x, t*z*z+c, 0,
		0, 0, 0, 1)
}

// Mat4SetBasis returns a matrix with the given basis vectors.
func Mat4SetBasis(xAxis, yAxis, zAxis math32.Vector3) math32.Matrix4 {
	return Mat4Set(
		xAxis.X, yAxis.X, zAxis.X, 0,
		xAxis.Y, yAxis.Y, zAxis.Y, 0,
		xAxis.Z, yAxis.Z, zAxis.Z, 0,
		0, 0, 0, 1)
}

// Mat4Pos returns the position (translation) component of the matrix.
func Mat4Pos(m math32.Matrix4) math32.Vector3 {
	return math32.Vec3(m[12], m[13], m[14])
}

// Mat4SetPos returns the matrix with its position (translation)
// component set to the given vector.
func Mat4SetPos(m math32.Matrix4, v math32.Vector3) math32.Matrix4 {
	nm := m
	nm[12] = v.X
	nm[13] = v.Y
	nm[14] = v.Z
	return nm
}

// Mat4CopyPos returns the matrix with its position (translation)
// component copied from the given source matrix.
func Mat4CopyPos(m, src math32.Matrix4) math32.Matrix4 {
	nm := m
	nm[12] = src[12]
	nm[13] = src[13]
	nm[14] = src[14]
	return nm
}

// Mat4FromQuat returns a rotation matrix from the given [math32.Quat].
func Mat4FromQuat(q math32.Quat) math32.Matrix4 {
	x2 := q.X + q.X
	y2 := q.Y + q.Y
	z2 := q.Z + q.Z
	xx := q.X * x2
	xy := q.X * y2
	xz := q.X * z2
	yy := q.Y * y2
	yz := q.Y * z2
	zz := q.Z * z2
	wx := q.W * x2
	wy := q.W * y2
	wz := q.W * z2

	return Mat4Set(
		1-(yy+zz), xy-wz, xz+wy, 0,
		xy+wz, 1-(xx+zz), yz-wx, 0,
		xz-wy, yz+wx, 1-(xx+yy), 0,
		0, 0, 0, 1)
}

// Mat4FromEuler returns a rotation matrix from the given euler angles.
func Mat4FromEuler(euler math32.Vector3) math32.Matrix4 {
	a := math32.Cos(euler.X)
	b := math32.Sin(euler.X)
	c := math32.Cos(euler.Y)
	d := math32.Sin(euler.Y)
	e := math32.Cos(euler.Z)
	f := math32.Sin(euler.Z)

	ae := a * e
	af := a * f
	be := b * e
	bf := b * f

	return Mat4Set(
		c*e, -c*f, d, 0,
		af+be*d, ae-bf*d, -b*c, 0,
		bf-ae*d, be+af*d, a*c, 0,
		0, 0, 0, 1)
}

// Mat4ToEuler returns the euler angles from the given
// pure rotation matrix.
func Mat4ToEuler(m math32.Matrix4) math32.Vector3 {
	m11 := m[0]
	m12 := m[4]
	m13 := m[8]
	m22 := m[5]
	m23 := m[9]
	m32 := m[6]
	m33 := m[10]

	var v math32.Vector3
	v.Y = math32.Asin(math32.Clamp(m13, float32(-1), float32(1)))
	if math32.Abs(m13) < 0.99999 {
		v.X = math32.Atan2(-m23, m33)
		v.Z = math32.Atan2(-m12, m11)
	} else {
		v.X = math32.Atan2(m32, m22)
		v.Z = 0
	}
	return v
}

// Mat4Transform returns a transformation matrix for the given position,
// rotation specified by the quaternion, and scale.
func Mat4Transform(pos math32.Vector3, quat math32.Quat, scale math32.Vector3) math32.Matrix4 {
	return Mat4SetPos(Mat4ScaleCols(Mat4FromQuat(quat), scale), pos)
}

// Mat4ExtractRotation returns a pure rotation matrix
// from the given transformation matrix.
func Mat4ExtractRotation(s math32.Matrix4) math32.Matrix4 {
	sx := 1 / Length3(math32.Vec3(s[0], s[1], s[2]))
	sy := 1 / Length3(math32.Vec3(s[4], s[5], s[6]))
	sz := 1 / Length3(math32.Vec3(s[8], s[9], s[10]))

	return Mat4Set(
		s[0]*sx, s[4]*sy, s[8]*sz, 0,
		s[1]*sx, s[5]*sy, s[9]*sz, 0,
		s[2]*sx, s[6]*sy, s[10]*sz, 0,
		0, 0, 0, 1)
}

//gosl:end
