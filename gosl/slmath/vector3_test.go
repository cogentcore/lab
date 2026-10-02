// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
)

// TestVector3VsMath32 checks every Vector3 function against the
// equivalent [math32.Vector3] method, over all the test vectors.
func TestVector3VsMath32(t *testing.T) {
	for _, v := range testVec3s {
		eq3(t, v.Negate(), Negate3(v), "Negate3")
		eqf(t, v.Length(), Length3(v), "Length3")
		eqf(t, v.LengthSquared(), LengthSquared3(v), "LengthSquared3")
		eq3(t, v.Abs(), Abs3(v), "Abs3")
		eq3(t, v.Floor(), Floor3(v), "Floor3")
		eq3(t, v.Ceil(), Ceil3(v), "Ceil3")
		eq3(t, v.Round(), Round3(v), "Round3")
		eq3(t, v.Normal(), Normal3(v), "Normal3")

		for d := int32(0); d < 3; d++ {
			dim := math32.Dims(d)
			eqf(t, v.Dim(dim), Dim3(v, d), "Dim3")
			for _, s := range testScalars {
				nv := v
				nv.SetDim(dim, s)
				eq3(t, nv, SetDim3(v, d, s), "SetDim3")
			}
		}

		for _, o := range testVec3s {
			eqf(t, v.Dot(o), Dot3(v, o), "Dot3")
			eq3(t, v.Max(o), Max3(v, o), "Max3")
			eq3(t, v.Min(o), Min3(v, o), "Min3")
			eq3(t, v.Cross(o), Cross3(v, o), "Cross3")
			eqf(t, v.DistanceTo(o), DistanceTo3(v, o), "DistanceTo3")
			eqf(t, v.DistanceToSquared(o), DistanceToSquared3(v, o), "DistanceToSquared3")
			for _, a := range testScalars {
				eq3(t, v.Lerp(o, a), Lerp3(v, o, a), "Lerp3")
			}
			if v.Length() != 0 && o.Length() != 0 {
				eqf(t, v.CosTo(o), CosTo3(v, o), "CosTo3")
				eqf(t, v.AngleTo(o), AngleTo3(v, o), "AngleTo3")
			}
			nv := v
			eq3(t, nv.ProjectOnVector(o), ProjectOnVector3(v, o), "ProjectOnVector3")
			eq3(t, nv.ProjectOnPlane(o), ProjectOnPlane3(v, o), "ProjectOnPlane3")
			eq3(t, nv.Reflect(o), Reflect3(v, o), "Reflect3")

			mn := v.Min(o)
			mx := v.Max(o)
			for _, c := range testVec3s {
				nc := c
				nc.Clamp(mn, mx)
				eq3(t, nc, Clamp3(c, mn, mx), "Clamp3")
			}
		}
	}
}

func TestVector3NDCToWindow(t *testing.T) {
	size := math32.Vec2(800, 600)
	off := math32.Vec2(10, 20)
	for _, v := range testVec3s {
		eq3(t, v.NDCToWindow(size, off, 0, 1, false), NDCToWindow3(v, size, off, 0, 1, false), "NDCToWindow3")
		eq3(t, v.NDCToWindow(size, off, 0, 1, true), NDCToWindow3(v, size, off, 0, 1, true), "NDCToWindow3 flipY")
		eq3(t, v.NDCToWindow(size, off, -1, 2, false), NDCToWindow3(v, size, off, -1, 2, false), "NDCToWindow3 near far")
	}
}

func TestVector3FromVector4(t *testing.T) {
	for _, v := range testVec4s {
		eq3(t, math32.Vector3FromVector4(v), Vec3FromVec4(v), "Vec3FromVec4")
	}
}

func TestVector3Values(t *testing.T) {
	v := math32.Vec3(3, -4, 12)
	eqf(t, 13, Length3(v), "Length3")
	eqf(t, 169, LengthSquared3(v), "LengthSquared3")
	eq3(t, math32.Vec3(3.0/13, -4.0/13, 12.0/13), Normal3(v), "Normal3")
	eq3(t, math32.Vec3(0, 0, 0), Normal3(math32.Vec3(0, 0, 0)), "Normal3 of zero")
	eq3(t, math32.Vec3(0, 0, 1), Cross3(math32.Vec3(1, 0, 0), math32.Vec3(0, 1, 0)), "Cross3")
	eqf(t, math32.Pi/2, AngleTo3(math32.Vec3(1, 0, 0), math32.Vec3(0, -1, 0)), "AngleTo3")
	eqf(t, -math32.Pi/2, AngleTo3(math32.Vec3(1, 0, 0), math32.Vec3(0, 1, 0)), "AngleTo3")
	// reflecting off the X=0 plane flips X
	eq3(t, math32.Vec3(-1, 2, 3), Reflect3(math32.Vec3(1, 2, 3), math32.Vec3(1, 0, 0)), "Reflect3")
	eq3(t, math32.Vec3(0, 2, 3), ProjectOnPlane3(math32.Vec3(1, 2, 3), math32.Vec3(1, 0, 0)), "ProjectOnPlane3")
	eq3(t, math32.Vec3(1, 0, 0), ProjectOnVector3(math32.Vec3(1, 2, 3), math32.Vec3(5, 0, 0)), "ProjectOnVector3")
}

func TestVector3DivSafe(t *testing.T) {
	eq3(t, math32.Vec3(2, -4, 3), DivSafe3(math32.Vec3(4, -8, 3), math32.Vec3(2, 2, 0)), "DivSafe3")
}

func TestVector3ClampMagnitude(t *testing.T) {
	eq3(t, math32.Vec3(2, -2, 1), ClampMagnitude3(math32.Vec3(4, -8, 1), 2), "ClampMagnitude3")
}
