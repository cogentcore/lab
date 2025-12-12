// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import "cogentcore.org/core/math32"

//gosl:start

// Shapes are elemental shapes for rigid bodies.
type Shapes int32 //enums:enum

const (
	// Box is a 3D rectalinear shape.
	Box Shapes = iota

	// Sphere. SizeX is the radius.
	Sphere

	// Cylinder, natively oriented vertically along the Y axis.
	// If one radius is 0, then it is a cone.
	// SizeX = bottom radius, SizeY = height in Y axis, SizeZ = top radius.
	Cylinder

	// Capsule, which is a cylinder with half-spheres on the ends.
	// Natively oriented vertically along the Y axis.
	// SizeX = bottom radius, SizeY = height, SizeZ = top radius.
	Capsule
)

//gosl:end

func (sh Shapes) ShapeBBox(sz math32.Vector3) math32.Box3 {
	var bb math32.Box3

	switch sh {
	case Box:
		bb.SetMinMax(sz.MulScalar(-.5), sz.MulScalar(.5))
	case Sphere:
		bb.SetMinMax(math32.Vec3(-sz.X, -sz.X, -sz.X), math32.Vec3(sz.X, sz.X, sz.X))
	case Cylinder:
		h2 := sz.Y / 2
		bb.SetMinMax(math32.Vec3(-sz.X, -h2, -sz.X), math32.Vec3(sz.Z, h2, sz.Z))
	case Capsule:
		th := sz.X + sz.Y + sz.Z
		h2 := th / 2
		bb.SetMinMax(math32.Vec3(-sz.X, -h2, -sz.X), math32.Vec3(sz.Z, h2, sz.Z))
	}
	// bb.Area = 2*sz.X + 2*sz.Y + 2*sz.Z
	// bb.Volume = sz.X * sz.Y * sz.Z
	return bb
}
