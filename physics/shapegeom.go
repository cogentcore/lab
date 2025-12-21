// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/gosl/slmath"
)

//gosl:start

func SphereSDF(center math32.Vector3, radius float32, p math32.Vector3) float32 {
	return slmath.Length3(p.Sub(center)) - radius
}

func BoxSDF(upper, p math32.Vector3) float32 {
	// adapted from https://www.iquilezles.org/www/articles/distfunctions/distfunctions.htm
	qx := math32.Abs(p.X) - upper.X
	qy := math32.Abs(p.Y) - upper.Y
	qz := math32.Abs(p.Z) - upper.Z
	e := math32.Vec3(max(qx, 0.0), max(qy, 0.0), max(qz, 0.0))
	return slmath.Length3(e) + min(max(qx, max(qy, qz)), 0.0)
}

func BoxSDFGrad(upper, p math32.Vector3) math32.Vector3 {
	qx := math32.Abs(p.X) - upper.X
	qy := math32.Abs(p.Y) - upper.Y
	qz := math32.Abs(p.Z) - upper.Z

	// exterior case
	if qx > 0.0 || qy > 0.0 || qz > 0.0 {
		x := math32.Clamp(p.X, -upper.X, upper.X)
		y := math32.Clamp(p.Y, -upper.Y, upper.Y)
		z := math32.Clamp(p.Z, -upper.Z, upper.Z)

		return slmath.Normal3(p.Sub(math32.Vec3(x, y, z)))
	}

	sx := math32.Sign(p.X)
	sy := math32.Sign(p.Y)
	sz := math32.Sign(p.Z)

	// x projection
	if (qx > qy && qx > qz) || (qy == 0.0 && qz == 0.0) {
		return math32.Vec3(sx, 0.0, 0.0)
	}

	// y projection
	if (qy > qx && qy > qz) || (qx == 0.0 && qz == 0.0) {
		return math32.Vec3(0.0, sy, 0.0)
	}

	// z projection
	return math32.Vec3(0.0, 0.0, sz)
}

func CapsuleSDF(radius, hh float32, p math32.Vector3) float32 {
	if p.Y > hh {
		return slmath.Length3(math32.Vec3(p.X, p.Y-hh, p.Z)) - radius
	}
	if p.Y < -hh {
		return slmath.Length3(math32.Vec3(p.X, p.Y+hh, p.Z)) - radius
	}
	return slmath.Length3(math32.Vec3(p.X, 0.0, p.Z)) - radius
}

func CylinderSDF(radius, hh float32, p math32.Vector3) float32 {
	dx := slmath.Length3(math32.Vec3(p.X, 0.0, p.Z)) - radius
	dy := math32.Abs(p.Y) - hh
	return min(max(dx, dy), 0.0) + slmath.Length2(math32.Vec2(max(dx, 0.0), max(dy, 0.0)))
}

// Cone with apex at +hh and base at -hh
func ConeSDF(radius, hh float32, p math32.Vector3) float32 {
	dx := slmath.Length3(math32.Vec3(p.X, 0.0, p.Z)) - radius*(hh-p.Y)/(2.0*hh)
	dy := math32.Abs(p.Y) - hh
	return min(max(dx, dy), 0.0) + slmath.Length2(math32.Vec2(max(dx, 0.0), max(dy, 0.0)))
}

// SDF for a quad in the xz plane
func PlaneSDF(width, length float32, p math32.Vector3) float32 {
	if width > 0.0 && length > 0.0 {
		d := max(math32.Abs(p.X)-width, math32.Abs(p.Z)-length)
		return max(d, math32.Abs(p.Y))
	}
	return p.Y
}

// ClosestPointPlane projects the point onto the quad in
// the xz plane (if size > 0.0, otherwise infinite.
func ClosestPointPlane(width, length float32, pt math32.Vector3) math32.Vector3 {
	cp := pt
	if width == 0.0 {
		cp.Y = 0
		return cp
	}
	cp.X = math32.Clamp(pt.X, -width, width)
	cp.Z = math32.Clamp(pt.Z, -length, length)
	return cp
}

func ClosestPointLineSegment(a, b, pt math32.Vector3) math32.Vector3 {
	ab := b.Sub(a)
	ap := pt.Sub(a)
	t := slmath.Dot3(ap, ab) / slmath.Dot3(ab, ab)
	t = math32.Clamp(t, 0.0, 1.0)
	return a.Add(ab.MulScalar(t))
}

// closest point to box surface
func ClosestPointBox(upper, pt math32.Vector3) math32.Vector3 {
	x := math32.Clamp(pt.X, -upper.X, upper.X)
	y := math32.Clamp(pt.Y, -upper.Y, upper.Y)
	z := math32.Clamp(pt.Z, -upper.Z, upper.Z)
	if math32.Abs(pt.X) <= upper.X && math32.Abs(pt.Y) <= upper.Y && math32.Abs(pt.Z) <= upper.Z {
		// the point is inside, find closest face
		sx := math32.Abs(math32.Abs(pt.X) - upper.X)
		sy := math32.Abs(math32.Abs(pt.Y) - upper.Y)
		sz := math32.Abs(math32.Abs(pt.Z) - upper.Z)
		//  return closest point on closest side, handle corner cases
		if (sx < sy && sx < sz) || (sy == 0.0 && sz == 0.0) {
			x = math32.Sign(pt.X) * upper.X
		} else if (sy < sx && sy < sz) || (sx == 0.0 && sz == 0.0) {
			y = math32.Sign(pt.Y) * upper.Y
		} else {
			z = math32.Sign(pt.Z) * upper.Z
		}
	}
	return math32.Vec3(x, y, z)
}

// box vertex numbering:
//
//	6---7
//	|\  |\       y
//	| 2-+-3      |
//	4-+-5 |   z \|
//	 \|  \|      o---x
//	  0---1
//
// get the vertex of the box given its ID (0-7)
func BoxVertex(ptId int32, upper math32.Vector3) math32.Vector3 {
	sign_x := float32(ptId%2)*2.0 - 1.0
	sign_y := float32((ptId/2)%2)*2.0 - 1.0
	sign_z := float32((ptId/4)%2)*2.0 - 1.0
	return math32.Vec3(sign_x*upper.X, sign_y*upper.Y, sign_z*upper.Z)
}

// get the edge of the box given its ID (0-11)
func BoxEdge(edgeId int32, upper math32.Vector3, edge0, edge1 *math32.Vector3) {
	eid := edgeId
	if eid < 4 {
		// edges along x: 0-1, 2-3, 4-5, 6-7
		i := eid * 2
		j := i + 1
		*edge0 = BoxVertex(i, upper)
		*edge1 = BoxVertex(j, upper)
	} else if eid < 8 {
		// edges along y: 0-2, 1-3, 4-6, 5-7
		eid -= 4
		i := eid%2 + eid // 2 * 4
		j := i + 2
		*edge0 = BoxVertex(i, upper)
		*edge1 = BoxVertex(j, upper)
	}
	// edges along z: 0-4, 1-5, 2-6, 3-7
	eid -= 8
	i := eid
	j := i + 4
	*edge0 = BoxVertex(i, upper)
	*edge1 = BoxVertex(j, upper)
}

// get the edge of the plane given its ID (0-3)
func PlaneEdge(edgeId int32, width, length float32, edge0, edge1 *math32.Vector3) {
	p0x := (2*float32(edgeId%2) - 1) * width
	p0z := (2*float32(edgeId/2) - 1) * length
	var p1x, p1z float32
	if edgeId == 0 || edgeId == 3 {
		p1x = p0x
		p1z = -p0z
	} else {
		p1x = -p0x
		p1z = p0z
	}
	*edge0 = math32.Vec3(p0x, 0, p0z)
	*edge1 = math32.Vec3(p1x, 0, p1z)
}

// find point on edge closest to box, return its barycentric edge coordinate
func ClosestEdgeBox(upper, edgeA, edgeB math32.Vector3, maxIter int32) float32 {
	// Golden-section search
	a := float32(0.0)
	b := float32(1.0)
	h := b - a
	invphi := float32(0.61803398875)  // 1 / phi
	invphi2 := float32(0.38196601125) // 1 / phi^2
	c := a + invphi2*h
	d := a + invphi*h
	query := edgeA.MulScalar(1.0 - c).Add(edgeB.MulScalar(c))
	yc := BoxSDF(upper, query)
	query = edgeA.MulScalar(1.0 - d).Add(edgeB.MulScalar(d))
	yd := BoxSDF(upper, query)

	for range maxIter {
		if yc < yd { // yc > yd to find the maximum
			b = d
			d = c
			yd = yc
			h = invphi * h
			c = a + invphi2*h
			query = edgeA.MulScalar(1.0 - c).Add(edgeB.MulScalar(c))
			yc = BoxSDF(upper, query)
		} else {
			a = c
			c = d
			yc = yd
			h = invphi * h
			d = a + invphi*h
			query = edgeA.MulScalar(1.0 - d).Add(edgeB.MulScalar(d))
			yd = BoxSDF(upper, query)
		}
	}

	if yc < yd {
		return 0.5 * (a + d)
	}
	return 0.5 * (c + b)
}

// find point on edge closest to plane, return its barycentric edge coordinate
func ClosestEdgePlane(width, length float32, edgeA, edgeB math32.Vector3, maxIter int32) float32 {
	// Golden-section search
	a := float32(0.0)
	b := float32(1.0)
	h := b - a
	invphi := float32(0.61803398875)  // 1 / phi
	invphi2 := float32(0.38196601125) // 1 / phi^2
	c := a + invphi2*h
	d := a + invphi*h
	query := edgeA.MulScalar(1.0 - c).Add(edgeB.MulScalar(c))
	yc := PlaneSDF(width, length, query)
	query = edgeA.MulScalar(1.0 - d).Add(edgeB.MulScalar(d))
	yd := PlaneSDF(width, length, query)

	for range maxIter {
		if yc < yd { // yc > yd to find the maximum
			b = d
			d = c
			yd = yc
			h = invphi * h
			c = a + invphi2*h
			query = edgeA.MulScalar(1.0 - c).Add(edgeB.MulScalar(c))
			yc = PlaneSDF(width, length, query)
		} else {
			a = c
			c = d
			yc = yd
			h = invphi * h
			d = a + invphi*h
			query = edgeA.MulScalar(1.0 - d).Add(edgeB.MulScalar(d))
			yd = PlaneSDF(width, length, query)
		}
	}

	if yc < yd {
		return 0.5 * (a + d)
	}
	return 0.5 * (c + b)
}

// find point on edge closest to capsule, return its barycentric edge coordinate
func ClosestEdgeCapsule(radius, hh float32, edgeA, edgeB math32.Vector3, maxIter int32) float32 {
	// Golden-section search
	a := float32(0.0)
	b := float32(1.0)
	h := b - a
	invphi := float32(0.61803398875)  // 1 / phi
	invphi2 := float32(0.38196601125) // 1 / phi^2
	c := a + invphi2*h
	d := a + invphi*h
	query := edgeA.MulScalar(1.0 - c).Add(edgeB.MulScalar(c))
	yc := CylinderSDF(radius, hh, query)
	query = edgeA.MulScalar(1.0 - d).Add(edgeB.MulScalar(d))
	yd := CylinderSDF(radius, hh, query)

	for range maxIter {
		if yc < yd { // yc > yd to find the maximum
			b = d
			d = c
			yd = yc
			h = invphi * h
			c = a + invphi2*h
			query = edgeA.MulScalar(1.0 - c).Add(edgeB.MulScalar(c))
			yc = CylinderSDF(radius, hh, query)
		} else {
			a = c
			c = d
			yc = yd
			h = invphi * h
			d = a + invphi*h
			query = edgeA.MulScalar(1.0 - d).Add(edgeB.MulScalar(d))
			yd = CylinderSDF(radius, hh, query)
		}
	}

	if yc < yd {
		return 0.5 * (a + d)
	}
	return 0.5 * (c + b)
}

//gosl:end
