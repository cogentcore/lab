// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
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
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, 1.0, math32.Vector3{X: 3.5, Y: 0.0, Z: 0.0}, 1.0, 1.5},  // Separated by 1.5
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, 1.0, math32.Vector3{X: 3.0, Y: 0.0, Z: 0.0}, 1.0, 1.0},  // Separated by 1.0
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, 1.0, math32.Vector3{X: 2.5, Y: 0.0, Z: 0.0}, 1.0, 0.5},  // Separated by 0.5
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, 1.0, math32.Vector3{X: 2.0, Y: 0.0, Z: 0.0}, 1.0, 0.0},  // Exactly touching
		{math32.Vector3{X: 0.0, Y: 1.0, Z: 0.0}, 1.0, math32.Vector3{X: 1.8, Y: 1.0, Z: 0.0}, 1.0, -0.2}, // Penetration = 0.2
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, 1.0, math32.Vector3{X: 1.5, Y: 0.0, Z: 0.0}, 1.0, -0.5}, // Penetration = 0.5
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 1.0}, 1.0, math32.Vector3{X: 1.2, Y: 0.0, Z: 1.0}, 1.0, -0.8}, // Penetration = 0.8
		// Different radii
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, 0.5, math32.Vector3{X: 2.0, Y: 0.0, Z: 0.0}, 1.0, 0.5},  // Separated
		{math32.Vector3{X: 0.0, Y: 1.0, Z: 0.0}, 0.5, math32.Vector3{X: 1.5, Y: 1.0, Z: 0.0}, 1.0, 0.0},  // Touching
		{math32.Vector3{X: 0.0, Y: 0.0, Z: 0.0}, 0.5, math32.Vector3{X: 1.2, Y: 0.0, Z: 0.0}, 1.0, -0.3}, // Penetration = 0.3
	}

	tol := 1e-5

	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		gdA := GeomData{Shape: Sphere, Radius: tc.radiusA, Size: math32.Vector3{X: tc.radiusA, Y: 0, Z: 0}, WbR: tc.posA, WbQ: rot}
		gdB := GeomData{Shape: Sphere, Radius: tc.radiusB, Size: math32.Vector3{X: tc.radiusB, Y: 0, Z: 0}, WbR: tc.posB, WbQ: rot}
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
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 2, Y: 2, Z: 0}, 1, 1},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 2, Y: 1.5, Z: 0}, 1, 0.5},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 2, Y: 1, Z: 0}, 1, 0},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 2, Y: 0.8, Z: 0}, 1, -0.2},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 2, Y: 0.5, Z: 0}, 1, -0.5},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 2, Y: 0.2, Z: 0}, 1, -0.8},
		// {math32.Vector3{1, 0, 0}, math32.Vector3{1, 0, 0}, math32.Vector3{2.0, 0, 0}, 0.5, 0.5},  // X-axis, separation = 0.5
		// {math32.Vector3{1, 0, 0}, math32.Vector3{1, 0, 0}, math32.Vector3{1.5, 0, 0}, 0.5, 0},    // X-axis, touching
		//	{math32.Vector3{1, 0, 0}, math32.Vector3{1, 0, 0}, math32.Vector3{1.3, 0, 0}, 0.5, -0.2}, // X-axis, penetration = 0.2
	}
	tol := 1e-5

	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		// note: A = sphere but pos = B..
		gdA := GeomData{Shape: Sphere, Radius: tc.radius, Size: math32.Vector3{X: tc.radius, Y: 0, Z: 0}, WbR: tc.posB, WbQ: rot}
		gdB := GeomData{Shape: Plane, Size: math32.Vector3{X: 0, Y: 0, Z: 0}, WbR: tc.posA, WbQ: rot}
		InitGeomData(0, &gdA)
		InitGeomData(0, &gdB)

		var ptA, ptB, norm math32.Vector3
		dist := ColSpherePlane(0, 10, &gdA, &gdB, &ptA, &ptB, &norm)
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

func TestCapsulePlane(t *testing.T) {
	// note: this data is already configured as A = plane, B = capsule, but function is CapsulePlane
	// so we're switching below..
	tests := []struct {
		normal, posA, posB math32.Vector3
		radius, dist       float32
	}{
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 3, Z: 0}, 0.5, 1.5},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.5, Z: 0}, 0.5, 1.0},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.0, Z: 0}, 0.5, 0.5},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 1.5, Z: 0}, 0.5, 0},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 1.4, Z: 0}, 0.5, -0.1},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 1.3, Z: 0}, 0.5, -0.2},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 1.2, Z: 0}, 0.5, -0.3},
	}
	tol := 1e-5

	// rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)
	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		// note: A = capsule but pos = B..
		// 1.5 hh = 1 raw hh
		gdA := GeomData{Shape: Capsule, Radius: tc.radius, Size: math32.Vector3{X: tc.radius, Y: 1.5, Z: tc.radius}, WbR: tc.posB, WbQ: rot}
		gdB := GeomData{Shape: Plane, Size: math32.Vector3{X: 0, Y: 0, Z: 0}, WbR: tc.posA, WbQ: rot}
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

func TestCylinderPlane(t *testing.T) {
	// note: this data is already configured as A = plane, B = capsule, but function is CapsulePlane
	// so we're switching below..
	tests := []struct {
		normal, posA, posB math32.Vector3
		radius, dist       float32
	}{
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.5, Z: 0}, 0.5, 1.5},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.0, Z: 0}, 0.5, 1.0},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 1.5, Z: 0}, 0.5, 0.5},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 1.0, Z: 0}, 0.5, 0},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 0.9, Z: 0}, 0.5, -0.1},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 0.8, Z: 0}, 0.5, -0.2},
		{math32.Vector3{X: 0, Y: 1, Z: 0}, math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 0.7, Z: 0}, 0.5, -0.3},
	}
	tol := 1e-5

	// rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)
	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		// note: A = capsule but pos = B..
		// 1 hh
		gdA := GeomData{Shape: Cylinder, Radius: tc.radius, Size: math32.Vector3{X: tc.radius, Y: 1.0, Z: tc.radius}, WbR: tc.posB, WbQ: rot}
		gdB := GeomData{Shape: Plane, Size: math32.Vector3{X: 0, Y: 0, Z: 0}, WbR: tc.posA, WbQ: rot}
		InitGeomData(0, &gdA)
		InitGeomData(0, &gdB)

		var ptA, ptB, norm math32.Vector3
		// important: we know that the lower axis, cpi = 0, is closest here
		// for cpi := range int32(4) {
		dist := ColCylinderPlane(0, 10, &gdA, &gdB, &ptA, &ptB, &norm)
		// fmt.Println(cpi, dist, ptA, ptB, norm)
		// }
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

func TestSphereCapsule(t *testing.T) {
	tests := []struct {
		posA, posB             math32.Vector3
		radiusA, radiusB, dist float32
	}{
		{math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 4, Z: 0}, 1.0, 0.5, 1.5},
		{math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 3.5, Z: 0}, 1.0, 0.5, 1.0},
		{math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 3.0, Z: 0}, 1.0, 0.5, 0.5},
		{math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.5, Z: 0}, 1.0, 0.5, 0},
		{math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.4, Z: 0}, 1.0, 0.5, -0.1},
		{math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.3, Z: 0}, 1.0, 0.5, -0.2},
		{math32.Vector3{X: 0, Y: 0, Z: 0}, math32.Vector3{X: 0, Y: 2.2, Z: 0}, 1.0, 0.5, -0.3},
	}
	tol := 1e-5

	rot := math32.NewQuatIdentity()
	for _, tc := range tests {
		// note: A = capsule but pos = B..
		// 1.5 hh = 1 raw hh
		gdA := GeomData{Shape: Sphere, Radius: tc.radiusA, Size: math32.Vector3{X: tc.radiusA, Y: 0, Z: 0}, WbR: tc.posA, WbQ: rot}
		gdB := GeomData{Shape: Capsule, Radius: tc.radiusB, Size: math32.Vector3{X: tc.radiusB, Y: 1.5, Z: tc.radiusB}, WbR: tc.posB, WbQ: rot}
		InitGeomData(0, &gdA)
		InitGeomData(0, &gdB)

		var ptA, ptB, norm math32.Vector3
		// important: we know that the lower axis, cpi = 0, is closest here
		dist := ColSphereCapsule(0, 10, &gdA, &gdB, &ptA, &ptB, &norm)
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

// TestBoxEdge tests that the 12 edges of a box join the vertexes that
// differ in one coordinate: 4 edges along each of x, y, and z.
func TestBoxEdge(t *testing.T) {
	upper := math32.Vec3(1, 2, 3)
	seen := map[[2]math32.Vector3]bool{}
	var n [3]int
	for eid := range int32(12) {
		var e0, e1 math32.Vector3
		BoxEdge(eid, upper, &e0, &e1)
		d := e1.Sub(e0)
		dims := 0
		for dim := range 3 {
			if d.Dim(math32.Dims(dim)) != 0 {
				dims++
				n[dim]++
				assert.Equal(t, 2*upper.Dim(math32.Dims(dim)), d.Dim(math32.Dims(dim)), "edge %d", eid)
			}
		}
		assert.Equal(t, 1, dims, "edge %d is along one axis: %v %v", eid, e0, e1)
		assert.False(t, seen[[2]math32.Vector3{e0, e1}], "edge %d is new", eid)
		seen[[2]math32.Vector3{e0, e1}] = true
	}
	assert.Equal(t, [3]int{4, 4, 4}, n)
}
