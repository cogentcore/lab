// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"fmt"
	"testing"

	"cogentcore.org/core/math32"
	"github.com/stretchr/testify/assert"
)

func TestSphereSphere(t *testing.T) {
	// Test sphere-sphere collision with analytical penetration depth validation.
	//       Analytical calculation:
	//       - Distance = ||center2 - center1|| - (radius1 + radius2)
	//       - Negative distance indicates penetration

	tests := []struct {
		posA          math32.Vector3
		radiusA       float32
		posB          math32.Vector3
		radiusB, dist float32
	}{
		{math32.Vector3{0.0, 0.0, 0.0}, 1.0, math32.Vector3{3.5, 0.0, 0.0}, 1.0, 1.5},  // Separated by 1.5
		{math32.Vector3{0.0, 0.0, 0.0}, 1.0, math32.Vector3{3.0, 0.0, 0.0}, 1.0, 1.0},  // Separated by 1.0
		{math32.Vector3{0.0, 0.0, 0.0}, 1.0, math32.Vector3{2.5, 0.0, 0.0}, 1.0, 0.5},  // Separated by 0.5
		{math32.Vector3{0.0, 0.0, 0.0}, 1.0, math32.Vector3{2.0, 0.0, 0.0}, 1.0, 0.0},  // Exactly touching
		{math32.Vector3{0.0, 1.0, 0.0}, 1.0, math32.Vector3{1.8, 1.0, 0.0}, 1.0, -0.2}, // Penetration = 0.2
		{math32.Vector3{0.0, 0.0, 0.0}, 1.0, math32.Vector3{1.5, 0.0, 0.0}, 1.0, -0.5}, // Penetration = 0.5
		{math32.Vector3{0.0, 0.0, 1.0}, 1.0, math32.Vector3{1.2, 0.0, 1.0}, 1.0, -0.8}, // Penetration = 0.8
		// Different radii
		{math32.Vector3{0.0, 0.0, 0.0}, 0.5, math32.Vector3{2.0, 0.0, 0.0}, 1.0, 0.5},  // Separated
		{math32.Vector3{0.0, 1.0, 0.0}, 0.5, math32.Vector3{1.5, 1.0, 0.0}, 1.0, 0.0},  // Touching
		{math32.Vector3{0.0, 0.0, 0.0}, 0.5, math32.Vector3{1.2, 0.0, 0.0}, 1.0, -0.3}, // Penetration = 0.3
	}

	tol := 1e-5

	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		gdA := GeomData{Shape: Sphere, Radius: tc.radiusA, Size: math32.Vector3{tc.radiusA, 0, 0}, WbR: tc.posA, WbQ: rot}
		gdB := GeomData{Shape: Sphere, Radius: tc.radiusB, Size: math32.Vector3{tc.radiusB, 0, 0}, WbR: tc.posB, WbQ: rot}
		InitGeomData(0, &gdA)
		InitGeomData(0, &gdB)

		var ptA, ptB, norm math32.Vector3
		dist := ColSphereSphere(0, 10, &gdA, &gdB, &ptA, &ptB, &norm)
		margin := float32(0.01)

		var ctA, ctB, offA, offB math32.Vector3
		var distActual, offMagA, offMagB float32
		actual := ContactPoints(dist, margin, &gdA, &gdB, ptA, ptB, norm, &ctA, &ctB, &offA, &offB, &distActual, &offMagA, &offMagB)
		_ = actual

		// fmt.Println(distActual, tc.dist, actual)
		// if actual {
		// 	fmt.Println(ptA, ptB, ctA, ctB, offA, offB, offMagA, offMagB)
		// }

		assert.InDelta(t, tc.dist, distActual, tol)
		assert.InDelta(t, 1.0, norm.Length(), tol)
		assert.Equal(t, distActual < margin, actual)

		if !actual {
			continue
		}
		// cpA := ctA.Add(offA)
		// cpB := ctB.Add(offB)
		// fmt.Println(cpA, cpB, tc.dist, actual)
	}
}

func TestSpherePlane(t *testing.T) {
	//       Analytical calculation:
	//       - Distance = (sphere_center - plane_point) · plane_normal - sphere_radius
	//       - Negative distance indicates penetration
	// note: this data is already configured as A = plane, B = sphere, but function is SpherePlane
	// so we're switching below..
	tests := []struct {
		normal, posA, posB math32.Vector3
		radius, dist       float32
	}{
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 2, 0}, 1, 1},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 1.5, 0}, 1, 0.5},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 1, 0}, 1, 0},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 0.8, 0}, 1, -0.2},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 0.5, 0}, 1, -0.5},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 0.2, 0}, 1, -0.8},
		// {math32.Vector3{1, 0, 0}, math32.Vector3{1, 0, 0}, math32.Vector3{2.0, 0, 0}, 0.5, 0.5},  // X-axis, separation = 0.5
		// {math32.Vector3{1, 0, 0}, math32.Vector3{1, 0, 0}, math32.Vector3{1.5, 0, 0}, 0.5, 0},    // X-axis, touching
		//	{math32.Vector3{1, 0, 0}, math32.Vector3{1, 0, 0}, math32.Vector3{1.3, 0, 0}, 0.5, -0.2}, // X-axis, penetration = 0.2
	}
	tol := 1e-5

	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		// note: A = sphere but pos = B..
		gdA := GeomData{Shape: Sphere, Radius: tc.radius, Size: math32.Vector3{tc.radius, 0, 0}, WbR: tc.posB, WbQ: rot}
		gdB := GeomData{Shape: Plane, Size: math32.Vector3{0, 0, 0}, WbR: tc.posA, WbQ: rot}
		InitGeomData(0, &gdA)
		InitGeomData(0, &gdB)

		var ptA, ptB, norm math32.Vector3
		dist := ColSpherePlane(0, 10, &gdA, &gdB, &ptA, &ptB, &norm)
		margin := float32(0.01)

		var ctA, ctB, offA, offB math32.Vector3
		var distActual, offMagA, offMagB float32
		actual := ContactPoints(dist, margin, &gdA, &gdB, ptA, ptB, norm, &ctA, &ctB, &offA, &offB, &distActual, &offMagA, &offMagB)
		_ = actual

		fmt.Println(dist, distActual, tc.dist, actual, norm, ptA, ptB)
		// if actual {
		// 	fmt.Println(ptA, ptB, ctA, ctB, offA, offB, offMagA, offMagB)
		// }

		assert.InDelta(t, tc.dist, distActual, tol)
		assert.InDelta(t, 1.0, norm.Length(), tol)
		assert.Equal(t, distActual < margin, actual)

		if !actual {
			continue
		}
		// cpA := ctA.Add(offA)
		// cpB := ctB.Add(offB)
		// fmt.Println(cpA, cpB, tc.dist, actual)
	}
}

func TestCapsulePlane(t *testing.T) {
	// note: this data is already configured as A = plane, B = capsule, but function is CapsulePlane
	// so we're switching below..
	tests := []struct {
		normal, posA, posB math32.Vector3
		radius, dist       float32
	}{
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 3, 0}, 0.5, 1.5},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 2.5, 0}, 0.5, 1.0},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 2.0, 0}, 0.5, 0.5},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 1.5, 0}, 0.5, 0},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 1.4, 0}, 0.5, -0.1},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 1.3, 0}, 0.5, -0.2},
		{math32.Vector3{0, 1, 0}, math32.Vector3{0, 0, 0}, math32.Vector3{0, 1.2, 0}, 0.5, -0.3},
	}
	tol := 1e-5

	// rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)
	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		// note: A = capsule but pos = B..
		// 1.5 hh = 1 raw hh
		gdA := GeomData{Shape: Capsule, Radius: tc.radius, Size: math32.Vector3{tc.radius, 1.5, tc.radius}, WbR: tc.posB, WbQ: rot}
		gdB := GeomData{Shape: Plane, Size: math32.Vector3{0, 0, 0}, WbR: tc.posA, WbQ: rot}
		InitGeomData(0, &gdA)
		InitGeomData(0, &gdB)

		var ptA, ptB, norm math32.Vector3
		// important: we know that the lower axis, cpi = 0, is closest here
		dist := ColCapsulePlane(0, 10, &gdA, &gdB, &ptA, &ptB, &norm)
		margin := float32(0.01)

		var ctA, ctB, offA, offB math32.Vector3
		var distActual, offMagA, offMagB float32
		actual := ContactPoints(dist, margin, &gdA, &gdB, ptA, ptB, norm, &ctA, &ctB, &offA, &offB, &distActual, &offMagA, &offMagB)
		_ = actual

		// fmt.Println(dist, distActual, tc.dist, actual, norm, ptA, ptB)
		// if actual {
		// 	fmt.Println(ptA, ptB, ctA, ctB, offA, offB, offMagA, offMagB)
		// }

		assert.InDelta(t, tc.dist, distActual, tol)
		assert.InDelta(t, 1.0, norm.Length(), tol)
		assert.Equal(t, distActual < margin, actual)

		if !actual {
			continue
		}
		// cpA := ctA.Add(offA)
		// cpB := ctB.Add(offB)
		// fmt.Println(cpA, cpB, tc.dist, actual)
	}
}
