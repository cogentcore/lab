// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
)

// TestVector4VsMath32 checks every Vector4 function against the
// equivalent [math32.Vector4] method, over all the test vectors.
func TestVector4VsMath32(t *testing.T) {
	for _, v := range testVec4s {
		eq4(t, v.Negate(), Negate4(v), "Negate4")
		eqf(t, v.Length(), Length4(v), "Length4")
		eqf(t, v.LengthSquared(), LengthSquared4(v), "LengthSquared4")
		eq4(t, v.Floor(), Floor4(v), "Floor4")
		eq4(t, v.Ceil(), Ceil4(v), "Ceil4")
		eq4(t, v.Round(), Round4(v), "Round4")
		eq4(t, v.Normal(), Normal4(v), "Normal4")
		if v.W != 0 {
			eq3(t, v.PerspDiv(), PerspDiv4(v), "PerspDiv4")
		}

		for d := int32(0); d < 4; d++ {
			dim := math32.Dims(d)
			eqf(t, v.Dim(dim), Dim4(v, d), "Dim4")
			for _, s := range testScalars {
				nv := v
				nv.SetDim(dim, s)
				eq4(t, nv, SetDim4(v, d, s), "SetDim4")
			}
		}

		for _, o := range testVec4s {
			eqf(t, v.Dot(o), Dot4(v, o), "Dot4")
			eq4(t, v.Max(o), Max4(v, o), "Max4")
			eq4(t, v.Min(o), Min4(v, o), "Min4")
			for _, a := range testScalars {
				eq4(t, v.Lerp(o, a), Lerp4(v, o, a), "Lerp4")
			}
			mn := v.Min(o)
			mx := v.Max(o)
			for _, c := range testVec4s {
				nc := c
				nc.Clamp(mn, mx)
				eq4(t, nc, Clamp4(c, mn, mx), "Clamp4")
			}
		}
	}

	for _, v := range testVec3s {
		for _, w := range testScalars {
			eq4(t, math32.Vector4FromVector3(v, w), Vec4FromVec3(v, w), "Vec4FromVec3")
		}
	}
}

func TestVector4Values(t *testing.T) {
	v := math32.Vec4(3, -4, 12, -84)
	eqf(t, 85, Length4(v), "Length4")
	eqf(t, 7225, LengthSquared4(v), "LengthSquared4")
	eq4(t, math32.Vec4(0, 0, 0, 0), Normal4(math32.Vec4(0, 0, 0, 0)), "Normal4 of zero")
	eqf(t, 1, Length4(Normal4(v)), "Normal4 is unit length")
	eq4(t, math32.Vec4(-3, 4, -12, 84), Negate4(v), "Negate4")
	eq4(t, math32.Vec4(3, 4, 12, 84), Abs4(v), "Abs4")
	eq3(t, math32.Vec3(2, 4, 6), PerspDiv4(math32.Vec4(4, 8, 12, 2)), "PerspDiv4")
}

func TestVector4DivSafe(t *testing.T) {
	eq4(t, math32.Vec4(2, -4, 3, 7), DivSafe4(math32.Vec4(4, -8, 3, 7), math32.Vec4(2, 2, 0, 0)), "DivSafe4")
}

func TestVector4ClampMagnitude(t *testing.T) {
	eq4(t, math32.Vec4(2, -2, 1, 2), ClampMagnitude4(math32.Vec4(4, -8, 1, 2), 2), "ClampMagnitude4")
}
