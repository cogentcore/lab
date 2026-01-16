// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slmath defines special math functions that operate on vector
// and quaternion types. These must be called as functions, not methods,
// and be outside of math32 itself so that the math32.Vector3 -> vec3<f32>
// replacement operates correctly. Must explicitly import this package into
// gosl using: //gosl:import "cogentcore.org/lab/gosl/slmath"
package slmath
