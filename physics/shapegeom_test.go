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

	//
	//       // Check that normal points from geom 0 (sphere 1) into geom 1 (sphere 2)
	//       for i in range(len(test_cases)):
	//           pos1 = np.array(test_cases[i][0])
	//           pos2 = np.array(test_cases[i][2])
	//           normal = normals_np[i]
	//           self.assertTrue(
	//               check_normal_direction_sphere_sphere(pos1, pos2, normal),
	//               msg=f"Test case {i}: Normal does not point from sphere 1 toward sphere 2",
	//           )
	//
	//       // Check that contact position is at midpoint between surfaces
	//       for i in range(len(test_cases)):
	//           pos1 = np.array(test_cases[i][0])
	//           radius1 = test_cases[i][1]
	//           pos2 = np.array(test_cases[i][2])
	//           radius2 = test_cases[i][3]
	//           contact_pos = positions_np[i]
	//           normal = normals_np[i]
	//           penetration_depth = distances_np[i]
	//
	//           self.assertTrue(
	//               check_contact_position_midpoint(contact_pos, normal, penetration_depth, pos1, radius1, pos2, radius2),
	//               msg=f"Test case {i}: Contact position is not at midpoint between surfaces",
	//           )
	//
}

func TestPlaneSphere(t *testing.T) {
	//       Analytical calculation:
	//       - Distance = (sphere_center - plane_point) · plane_normal - sphere_radius
	//       - Negative distance indicates penetration
	tests := []struct {
		normal, posA, posB math32.Vector3
		radius, dist       float32
	}{
		{math32.Vector3{0.0, 0.0, 1.0}, math32.Vector3{0.0, 0.0, 0.0}, math32.Vector3{0.0, 0.0, 2.0}, 1.0, 1.0},  // Above plane, separation = 1.0
		{math32.Vector3{0.0, 0.0, 1.0}, math32.Vector3{0.0, 0.0, 0.0}, math32.Vector3{0.0, 0.0, 1.5}, 1.0, 0.5},  // Above plane, separation = 0.5
		{math32.Vector3{0.0, 0.0, 1.0}, math32.Vector3{0.0, 0.0, 0.0}, math32.Vector3{0.0, 0.0, 1.0}, 1.0, 0.0},  // Just touching
		{math32.Vector3{0.0, 0.0, 1.0}, math32.Vector3{0.0, 0.0, 0.0}, math32.Vector3{0.0, 0.0, 0.8}, 1.0, -0.2}, // Penetration = 0.2
		{math32.Vector3{0.0, 0.0, 1.0}, math32.Vector3{0.0, 0.0, 0.0}, math32.Vector3{0.0, 0.0, 0.5}, 1.0, -0.5}, // Penetration = 0.5
		{math32.Vector3{0.0, 0.0, 1.0}, math32.Vector3{0.0, 0.0, 0.0}, math32.Vector3{0.0, 0.0, 0.2}, 1.0, -0.8}, // Penetration = 0.8
		{math32.Vector3{1.0, 0.0, 0.0}, math32.Vector3{1.0, 0.0, 0.0}, math32.Vector3{2.0, 0.0, 0.0}, 0.5, 0.5},  // X-axis, separation = 0.5
		{math32.Vector3{1.0, 0.0, 0.0}, math32.Vector3{1.0, 0.0, 0.0}, math32.Vector3{1.5, 0.0, 0.0}, 0.5, 0.0},  // X-axis, touching
		{math32.Vector3{1.0, 0.0, 0.0}, math32.Vector3{1.0, 0.0, 0.0}, math32.Vector3{1.3, 0.0, 0.0}, 0.5, -0.2}, // X-axis, penetration = 0.2
	}
	_ = tests

	// for _, tc := range tests {
	// 	ColSphereSphere(0, )
	// }
	//
	//       wp.launch(
	//           test_plane_sphere_kernel,
	//           dim=len(test_cases),
	//           inputs=[plane_normals, plane_positions, sphere_positions, sphere_radii, distances, contact_positions],
	//       )
	//       wp.synchronize()
	//
	//       distances_np = distances.numpy()
	//       positions_np = contact_positions.numpy()
	//
	//       // Verify expected distances with analytical validation
	//       for i, expected_dist in enumerate([tc[4] for tc in test_cases]):
	//           self.assertAlmostEqual(
	//               distances_np[i],
	//               expected_dist,
	//               places=5,
	//               msg=f"Test case {i}: Expected distance {expected_dist:.4f}, got {distances_np[i]:.4f}",
	//           )
	//
	//       // Check that contact position lies between sphere and plane
	//       for i in range(len(test_cases)):
	//           if distances_np[i] >= 0:
	//               // Skip separated cases
	//               continue
	//
	//           plane_normal = np.array(test_cases[i][0])
	//           plane_pos = np.array(test_cases[i][1])
	//           sphere_pos = np.array(test_cases[i][2])
	//           sphere_radius = test_cases[i][3]
	//           contact_pos = positions_np[i]
	//
	//           // Contact position should be between sphere surface and plane
	//           // Distance from contact to sphere center should be less than sphere radius
	//           dist_to_sphere_center = np.linalg.norm(contact_pos - sphere_pos)
	//           self.assertLess(
	//               dist_to_sphere_center,
	//               sphere_radius + 0.01,
	//               msg=f"Test case {i}: Contact position too far from sphere (dist: {dist_to_sphere_center:.4f})",
	//           )
	//
	//           // Contact position should be on the plane side of the sphere center
	//           // (or at most slightly past the plane)
	//           dist_contact_to_plane = np.dot(contact_pos - plane_pos, plane_normal)
	//           dist_sphere_to_plane = np.dot(sphere_pos - plane_pos, plane_normal)
	//           self.assertLessEqual(
	//               dist_contact_to_plane,
	//               dist_sphere_to_plane + 0.01,
	//               msg=f"Test case {i}: Contact position on wrong side of sphere center",
	//           )
	//
}
