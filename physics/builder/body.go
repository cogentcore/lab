// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/phyxyz"
)

// Body is a rigid body.
type Body struct {
	// Shape of the body.
	Shape physics.Shapes

	// Dynamic makes this a dynamic body.
	Dynamic bool

	// Group partitions bodies within worlds into different groups
	// for collision detection. 0 does not collide with anything.
	// Negative numbers are global within a world, except they don't
	// collide amongst themselves (all non-dynamic bodies should go
	// in -1 because they don't collide amongst each-other, but do
	// potentially collide with dynamics).
	// Positive numbers only collide amongst themselves, and with
	// negative groups, but not other positive groups. To avoid
	// unwanted collisions, put bodies into separate groups.
	// There is an automatic constraint that the two objects
	// within a single joint do not collide with each other, so this
	// does not need to be handled here.
	Group int

	// HSize is the half-size (e.g., radius) of the body.
	// Values depend on shape type: X is generally radius,
	// Y is half-height.
	HSize math32.Vector3

	// Thick is the thickness of the body, as a hollow shape.
	// If 0, then it is a solid shape (default).
	Thick float32

	// Mass of the object. Only relevant for Dynamic bodies.
	Mass float32

	// Bounce specifies the COR or coefficient of restitution (0..1),
	// which determines how elastic the collision is,
	// i.e., final velocity / initial velocity.
	Bounce float32

	// Friction is the standard coefficient for linear friction (mu).
	Friction float32

	// FrictionTortion is resistance to spinning at the contact point.
	FrictionTortion float32

	// FrictionRolling is resistance to rolling motion at contact.
	FrictionRolling float32

	// Pose has the position and rotation.
	Pose Pose

	// Com is the center-of-mass offset from the Pose.Pos.
	Com math32.Vector3

	// View is the view element for this Body (optional).
	View *phyxyz.View

	// Index is the index of this body in the physics.World,
	// once built.
	Index int32

	// DynamicIndex is the index of this dynamic body in
	// the physics.World, once built.
	DynamicIndex int32
}
