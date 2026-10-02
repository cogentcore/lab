// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slmath defines special math functions that operate on the
// [cogentcore.org/core/math32] vector, matrix and quaternion types.
// These must be called as functions, not methods, and be outside of math32
// itself so that the math32.Vector3 -> vec3<f32> replacement operates
// correctly. Must explicitly import this package into gosl using:
//
//	//gosl:import "cogentcore.org/lab/gosl/slmath"
//
// WGSL natively supports the basic vector operations (Add, Sub, Mul, Div,
// MulScalar, DivScalar), which gosl transpiles into the equivalent
// operators, so there are no slmath functions for those.
//
// Go has no overloading, so each type gets a distinct name: the Vector
// types take a numeric suffix (Length2, Length3, Length4), and the Quat
// and Matrix types take a prefix (QuatSlerp, Mat2Inverse, Mat3Determinant,
// Mat4MulVector3). Where a math32 method mutates through a pointer
// receiver, the slmath function takes and returns values instead.
//
// See the README for what is omitted, where slmath deliberately differs
// from math32 so that the CPU and GPU results agree, and how the matrix
// types map onto the WGSL matrix types.
package slmath
