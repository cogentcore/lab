// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
	"github.com/stretchr/testify/assert"
)

// TestVector2VsMath32 checks every Vector2 function against the
// equivalent [math32.Vector2] method, over all the test vectors.
func TestVector2VsMath32(t *testing.T) {
	for _, v := range testVec2s {
		eq2(t, v.Negate(), Negate2(v), "Negate2")
		eqf(t, v.Length(), Length2(v), "Length2")
		eqf(t, v.LengthSquared(), LengthSquared2(v), "LengthSquared2")
		eq2(t, v.Abs(), Abs2(v), "Abs2")
		eq2(t, v.Floor(), Floor2(v), "Floor2")
		eq2(t, v.Ceil(), Ceil2(v), "Ceil2")
		eq2(t, v.Round(), Round2(v), "Round2")
		eq2(t, v.Normal(), Normal2(v), "Normal2")
		eq2(t, v.Rot90CW(), Rot90CW2(v), "Rot90CW2")
		eq2(t, v.Rot90CCW(), Rot90CCW2(v), "Rot90CCW2")

		for d := int32(0); d < 2; d++ {
			dim := math32.Dims(d)
			eqf(t, v.Dim(dim), Dim2(v, d), "Dim2")
			for _, s := range testScalars {
				nv := v
				nv.SetDim(dim, s)
				eq2(t, nv, SetDim2(v, d, s), "SetDim2")
				eq2(t, v.AddDim(dim, s), AddDim2(v, d, s), "AddDim2")
				eq2(t, v.SubDim(dim, s), SubDim2(v, d, s), "SubDim2")
				eq2(t, v.MulDim(dim, s), MulDim2(v, d, s), "MulDim2")
				if s != 0 {
					eq2(t, v.DivDim(dim, s), DivDim2(v, d, s), "DivDim2")
				}
			}
		}

		for _, o := range testVec2s {
			eqf(t, v.Dot(o), Dot2(v, o), "Dot2")
			eq2(t, v.Max(o), Max2(v, o), "Max2")
			eq2(t, v.Min(o), Min2(v, o), "Min2")
			eqf(t, v.Cross(o), Cross2(v, o), "Cross2")
			eqf(t, v.DistanceTo(o), DistanceTo2(v, o), "DistanceTo2")
			eqf(t, v.DistanceToSquared(o), DistanceToSquared2(v, o), "DistanceToSquared2")
			for _, a := range testScalars {
				eq2(t, v.Lerp(o, a), Lerp2(v, o, a), "Lerp2")
			}
			if v.Length() != 0 && o.Length() != 0 {
				eqf(t, v.CosTo(o), CosTo2(v, o), "CosTo2")
				eqf(t, v.AngleTo(o), AngleTo2(v, o), "AngleTo2")
			}
			// Clamp needs min < max, so use the component-wise ordering.
			mn := v.Min(o)
			mx := v.Max(o)
			for _, c := range testVec2s {
				nc := c
				nc.Clamp(mn, mx)
				eq2(t, nc, Clamp2(c, mn, mx), "Clamp2")
			}
		}

		for _, phi := range testScalars {
			for _, p0 := range testVec2s {
				eq2(t, v.Rot(phi, p0), Rot2(v, phi, p0), "Rot2")
			}
		}
	}

	for _, a := range testScalars {
		for _, r := range testScalars {
			eq2(t, math32.Vector2Polar(a, r), Polar2(a, r), "Polar2")
		}
	}
}

func TestVector2InTriangle(t *testing.T) {
	p0 := math32.Vec2(0, 0)
	p1 := math32.Vec2(4, 0)
	p2 := math32.Vec2(0, 4)
	for _, v := range testVec2s {
		eqb(t, v.InTriangle(p0, p1, p2), InTriangle2(v, p0, p1, p2), "InTriangle2")
	}
	// explicit cases
	assert.True(t, InTriangle2(math32.Vec2(1, 1), p0, p1, p2))
	assert.False(t, InTriangle2(math32.Vec2(3, 3), p0, p1, p2))
	assert.False(t, InTriangle2(math32.Vec2(-1, 1), p0, p1, p2))
}

func TestVector2WindowToNDC(t *testing.T) {
	size := math32.Vec2(800, 600)
	off := math32.Vec2(10, 20)
	for _, v := range testVec2s {
		eq3(t, v.WindowToNDC(size, off, false), WindowToNDC2(v, size, off, false), "WindowToNDC2")
		eq3(t, v.WindowToNDC(size, off, true), WindowToNDC2(v, size, off, true), "WindowToNDC2 flipY")
	}
}

func TestVector2Values(t *testing.T) {
	v := math32.Vec2(3, -4)
	eqf(t, 5, Length2(v), "Length2")
	eqf(t, 25, LengthSquared2(v), "LengthSquared2")
	eq2(t, math32.Vec2(0.6, -0.8), Normal2(v), "Normal2")
	eq2(t, math32.Vec2(0, 0), Normal2(math32.Vec2(0, 0)), "Normal2 of zero")
	eq2(t, math32.Vec2(-3, 4), Negate2(v), "Negate2")
	eq2(t, math32.Vec2(3, 4), Abs2(v), "Abs2")
	eqf(t, -11, Dot2(v, math32.Vec2(1, 3.5)), "Dot2")
	eqf(t, 14.5, Cross2(v, math32.Vec2(1, 3.5)), "Cross2")
	eq2(t, math32.Vec2(-4, -3), Rot90CW2(v), "Rot90CW2")
	eq2(t, math32.Vec2(4, 3), Rot90CCW2(v), "Rot90CCW2")
	eqf(t, math32.Pi/2, AngleTo2(math32.Vec2(1, 0), math32.Vec2(0, -1)), "AngleTo2")
	eqf(t, -math32.Pi/2, AngleTo2(math32.Vec2(1, 0), math32.Vec2(0, 1)), "AngleTo2")
}

func TestVector2DivSafe(t *testing.T) {
	eq2(t, math32.Vec2(2, -4), DivSafe2(math32.Vec2(4, -8), math32.Vec2(2, 2)), "DivSafe2")
	eq2(t, math32.Vec2(4, -4), DivSafe2(math32.Vec2(4, -8), math32.Vec2(0, 2)), "DivSafe2 zero X")
	eq2(t, math32.Vec2(4, -8), DivSafe2(math32.Vec2(4, -8), math32.Vec2(0, 0)), "DivSafe2 both zero")
}

func TestVector2ClampMagnitude(t *testing.T) {
	eq2(t, math32.Vec2(2, -2), ClampMagnitude2(math32.Vec2(4, -8), 2), "ClampMagnitude2")
	eq2(t, math32.Vec2(1, -0.5), ClampMagnitude2(math32.Vec2(1, -0.5), 2), "ClampMagnitude2 under")
}
