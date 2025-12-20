// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"cogentcore.org/core/math32"
)

// see: newton/geometry for lots of helpful methods.

//gosl:start

// newton: geometry/types.py

// Shapes are elemental shapes for rigid bodies.
// In general, size dimensions are half values
// (e.g., radius, half-height, etc), which is natural for
// center-based body coordinates.
type Shapes int32 //enums:enum

const (
	// Plane cannot be a dynamic shape, but is most efficient for
	// collision computations. Use size = 0 for an infinite plane.
	// Natively extends in the X-Z plane: SizeX x SizeZ.
	Plane Shapes = iota

	// todo: HeightField here (terrain)

	// Sphere. SizeX is the radius.
	Sphere

	// Capsule is a cylinder with half-spheres on the ends.
	// Natively oriented vertically along the Y axis.
	// SizeX = radius, SizeY = half-height.
	Capsule

	// todo: Ellipsoid goes here

	// Cylinder, natively oriented vertically along the Y axis.
	// SizeX = radius, SizeY = half-height in Y axis.
	// Cylinder can not collide with a Box.
	Cylinder

	// Box is a 3D rectalinear shape.
	// The sizes are _half_ sizes along each dimension,
	// relative to the center.
	Box

	// todo: Mesh, SDF here

	// Cone is like a cylinder with the top radius = 0,
	// oriented up. SizeX = bottom radius, SizeY = half-height in Y.
	// Cone does not support any collisions and is not recommended for
	// interacting bodies.
	Cone
)

// newton: geometry/kernels.py: count_contact_points_for_pair

// ShapePairContacts returns the number of contact points possible
// for given pair of shapes. a <= b ordering. returns from a to b,
// ba is from b to a (mostly 0).
// infPlane means that a is a Plane and it is infinite (size = 0).
func ShapePairContacts(a, b Shapes, infPlane bool, ba *int32) int32 {
	*ba = 0
	switch a {
	case Plane:
		switch b {
		case Plane:
			return 0
		case Sphere:
			return 1
		case Capsule:
			if infPlane {
				return 2
			} else {
				return 2 + 4
			}
		case Cylinder:
			return 4
		case Box:
			if infPlane {
				return 8
			} else {
				return 8 + 4
			}
		default:
			return 0
		}
	case Sphere:
		return 1
	case Capsule:
		switch b {
		case Capsule:
			return 2
		case Box:
			return 8
		default:
			return 0
		}
	case Cylinder:
		return 0 // no box collisions!
	case Box:
		*ba = 12
		return 12
	default: // note: Cone has no collision points!
		return 0
	}
}

//gosl:end

// Radius returns the shape radius for given size.
// this is used for broad-phase collision.
func (sh Shapes) Radius(sz math32.Vector3) float32 {
	switch sh {
	case Plane:
		if sz.X > 0 {
			return sz.Length()
		}
		return 1.0e6 // infinite
	case Sphere:
		return sz.X
	case Capsule, Cylinder:
		return max(sz.X, sz.Z) + sz.Y // over-estimate for cylinder
	case Box:
		return sz.Length()
	}
	return 0
}

// BBox returns the bounding box for shape of given size.
func (sh Shapes) BBox(sz math32.Vector3) math32.Box3 {
	var bb math32.Box3
	switch sh {
	case Sphere:
		bb.SetMinMax(math32.Vec3(-sz.X, -sz.X, -sz.X), math32.Vec3(sz.X, sz.X, sz.X))
	case Capsule:
		bb.SetMinMax(math32.Vec3(-sz.X, -sz.Y-sz.X, -sz.X), math32.Vec3(sz.Z, sz.Y+sz.Z, sz.Z))
	case Cylinder:
		bb.SetMinMax(math32.Vec3(-sz.X, -sz.Y, -sz.X), math32.Vec3(sz.Z, sz.Y, sz.Z))
	case Box:
		bb.SetMinMax(sz.Negate(), sz)
	}
	return bb
}

// Inertia returns the inertia tensor for solid shape of given size,
// with uniform density and given mass.
func (sh Shapes) Inertia(sz math32.Vector3, mass float32) math32.Matrix3 {
	var inertia math32.Matrix3
	switch sh {
	// todo: other shapes!!  see below.
	case Sphere:
		r := sz.X
		// v := 4.0 / 3.0 * math32.Pi * r * r * r
		ia := 2.0 / 5.0 * mass * r * r
		inertia = math32.Mat3(ia, 0.0, 0.0, 0.0, ia, 0.0, 0.0, 0.0, ia)
	case Box:
		w := 2 * sz.X
		h := 2 * sz.Y
		d := 2 * sz.Z
		ia := 1.0 / 12.0 * mass * (h*h + d*d)
		ib := 1.0 / 12.0 * mass * (w*w + d*d)
		ic := 1.0 / 12.0 * mass * (w*w + h*h)
		inertia = math32.Mat3(ia, 0.0, 0.0, 0.0, ib, 0.0, 0.0, 0.0, ic)
		// todo: others:
	}
	return inertia
}

/*
def compute_capsule_inertia(density: float, r: float, h: float) -> tuple[float, wp.vec3, wp.mat33]:
    """Helper to compute mass and inertia of a solid capsule extending along the z-axis

    Args:
        density: The capsule density
        r: The capsule radius
        h: The capsule height (full height of the interior cylinder)

    Returns:

        A tuple of (mass, inertia) with inertia specified around the origin
    """

    ms = density * (4.0 / 3.0) * wp.pi * r * r * r
    mc = density * wp.pi * r * r * h

    # total mass
    m = ms + mc

    # adapted from ODE
    Ia = mc * (0.25 * r * r + (1.0 / 12.0) * h * h) + ms * (0.4 * r * r + 0.375 * r * h + 0.25 * h * h)
    Ib = (mc * 0.5 + ms * 0.4) * r * r

    # For Z-axis orientation: I_xx = I_yy = Ia, I_zz = Ib
    I = wp.mat33([[Ia, 0.0, 0.0], [0.0, Ia, 0.0], [0.0, 0.0, Ib]])

    return (m, wp.vec3(), I)


def compute_cylinder_inertia(density: float, r: float, h: float) -> tuple[float, wp.vec3, wp.mat33]:
    """Helper to compute mass and inertia of a solid cylinder extending along the z-axis

    Args:
        density: The cylinder density
        r: The cylinder radius
        h: The cylinder height (extent along the z-axis)

    Returns:

        A tuple of (mass, inertia) with inertia specified around the origin
    """

    m = density * wp.pi * r * r * h

    Ia = 1 / 12 * m * (3 * r * r + h * h)
    Ib = 1 / 2 * m * r * r

    # For Z-axis orientation: I_xx = I_yy = Ia, I_zz = Ib
    I = wp.mat33([[Ia, 0.0, 0.0], [0.0, Ia, 0.0], [0.0, 0.0, Ib]])

    return (m, wp.vec3(), I)


def compute_cone_inertia(density: float, r: float, h: float) -> tuple[float, wp.vec3, wp.mat33]:
    """Helper to compute mass and inertia of a solid cone extending along the z-axis

    Args:
        density: The cone density
        r: The cone radius
        h: The cone height (extent along the z-axis)

    Returns:

        A tuple of (mass, center of mass, inertia) with inertia specified around the center of mass
    """

    m = density * wp.pi * r * r * h / 3.0

    # Center of mass is at -h/4 from the geometric center
    # Since the cone has base at -h/2 and apex at +h/2, the COM is 1/4 of the height from base toward apex
    com = wp.vec3(0.0, 0.0, -h / 4.0)

    # Inertia about the center of mass
    Ia = 3 / 20 * m * r * r + 3 / 80 * m * h * h
    Ib = 3 / 10 * m * r * r

    # For Z-axis orientation: I_xx = I_yy = Ia, I_zz = Ib
    I = wp.mat33([[Ia, 0.0, 0.0], [0.0, Ia, 0.0], [0.0, 0.0, Ib]])

    return (m, com, I)


def compute_ellipsoid_inertia(density: float, a: float, b: float, c: float) -> tuple[float, wp.vec3, wp.mat33]:
    """Helper to compute mass and inertia of a solid ellipsoid

    The ellipsoid is centered at the origin with semi-axes a, b, c along the x, y, z axes respectively.

    Args:
        density: The ellipsoid density
        a: The semi-axis along the x-axis
        b: The semi-axis along the y-axis
        c: The semi-axis along the z-axis

    Returns:

        A tuple of (mass, center of mass, inertia) with inertia specified around the center of mass
    """
    # Volume of ellipsoid: V = (4/3) * pi * a * b * c
    v = (4.0 / 3.0) * wp.pi * a * b * c
    m = density * v

    # Inertia tensor for a solid ellipsoid about its center of mass:
    # Ixx = (1/5) * m * (b² + c²)
    # Iyy = (1/5) * m * (a² + c²)
    # Izz = (1/5) * m * (a² + b²)
    Ixx = (1.0 / 5.0) * m * (b * b + c * c)
    Iyy = (1.0 / 5.0) * m * (a * a + c * c)
    Izz = (1.0 / 5.0) * m * (a * a + b * b)

    I = wp.mat33([[Ixx, 0.0, 0.0], [0.0, Iyy, 0.0], [0.0, 0.0, Izz]])

    return (m, wp.vec3(), I)

*/

/*
def compute_box_inertia(density: float, w: float, h: float, d: float) -> tuple[float, wp.vec3, wp.mat33]:
    """Helper to compute mass and inertia of a solid box

    Args:
        density: The box density
        w: The box width along the x-axis
        h: The box height along the y-axis
        d: The box depth along the z-axis

    Returns:

        A tuple of (mass, inertia) with inertia specified around the origin
    """

    v = w * h * d
    m = density * v
    I = compute_box_inertia_from_mass(m, w, h, d)

    return (m, wp.vec3(), I)

}
*/
