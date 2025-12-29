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
	// ObjectIndex is the index of body within parent [Object],
	// which is used for id in [Builder] context.
	ObjectIndex int

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

	// Pose has the position and rotation.
	Pose Pose

	// Com is the center-of-mass offset from the Pose.Pos.
	Com math32.Vector3

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

	// Optional [phyxyz.Skin] for visualizing the body.
	Skin *phyxyz.Skin

	// BodyIndex is the index of this body in the [physics.Model] Bodies list,
	// once built.
	BodyIndex int32

	// DynamicIndex is the index of this dynamic body in the
	// [physics.Model] Dynamics list, once built.
	DynamicIndex int32
}

// NewBody adds a new body with given parameters.
// Returns the [Body] which can then be further customized.
// Use this for Static elements; NewDynamic for dynamic elements.
func (ob *Object) NewBody(shape physics.Shapes, hsize, pos math32.Vector3, rot math32.Quat) *Body {
	idx := len(ob.Bodies)
	bd := Body{ObjectIndex: idx, Shape: shape, HSize: hsize}
	bd.Pose.Pos = pos
	bd.Pose.Quat = rot
	ob.Bodies = append(ob.Bodies, bd)
	return &(ob.Bodies[idx])
}

// NewDynamic adds a new dynamic body with given parameters.
// Returns the [Body] which can then be further customized.
func (ob *Object) NewDynamic(shape physics.Shapes, mass float32, hsize, pos math32.Vector3, rot math32.Quat) *Body {
	bd := ob.NewBody(shape, hsize, pos, rot)
	bd.Dynamic = true
	bd.Mass = mass
	return bd
}

// NewBodySkin adds a new body with given parameters, including name and
// color parameters used for intializing a [phyxyz.Skin] in given [phyxyz.Scene].
// Returns the [Body] which can then be further customized.
// Use this for Static elements; NewDynamicSkin for dynamic elements.
func (ob *Object) NewBodySkin(sc *phyxyz.Scene, name string, shape physics.Shapes, clr string, hsize, pos math32.Vector3, rot math32.Quat) *Body {
	bd := ob.NewBody(shape, hsize, pos, rot)
	bd.NewSkin(sc, name, clr)
	return bd
}

// NewSkin adds a new skin for body with given name and color parameters.
func (bd *Body) NewSkin(sc *phyxyz.Scene, name string, clr string) *phyxyz.Skin {
	sk := sc.NewSkin(bd.Shape, name, clr, bd.HSize, bd.Pose.Pos, bd.Pose.Quat)
	bd.Skin = sk
	return sk
}

// NewDynamicSkin adds a new dynamic body with given parameters,
// including name and color parameters used for intializing a [phyxyz.Skin]
// in given [phyxyz.Scene].
// Returns the [Body] which can then be further customized.
func (ob *Object) NewDynamicSkin(sc *phyxyz.Scene, name string, shape physics.Shapes, clr string, mass float32, hsize, pos math32.Vector3, rot math32.Quat) *Body {
	bd := ob.NewBodySkin(sc, name, shape, clr, hsize, pos, rot)
	bd.Dynamic = true
	bd.Mass = mass
	return bd
}

/////// Physics functions

func (bd *Body) NewPhysicsBody(ml *physics.Model, world int) {
	var bi, di int32
	if bd.Dynamic {
		bi, di = ml.NewDynamic(bd.Shape, bd.Mass, bd.HSize, bd.Pose.Pos, bd.Pose.Quat)
	} else {
		bi = ml.NewBody(bd.Shape, bd.HSize, bd.Pose.Pos, bd.Pose.Quat)
		di = -1
	}
	bd.BodyIndex = bi
	bd.DynamicIndex = di
	physics.SetBodyWorld(bi, int32(world))
	physics.SetBodyGroup(bi, int32(bd.Group))
	// fmt.Println("\t\t", bi, di, bd.Pose.Pos, bd.Pose.Quat)
	if bd.Skin != nil {
		bd.Skin.BodyIndex = bi
		bd.Skin.DynamicIndex = di
	}
	physics.SetBodyThick(bi, bd.Thick)
	physics.SetBodyCom(bi, bd.Com)
	physics.SetBodyBounce(bi, bd.Bounce)
	physics.SetBodyFriction(bi, bd.Friction)
	physics.SetBodyFrictionTortion(bi, bd.FrictionTortion)
	physics.SetBodyFrictionRolling(bi, bd.FrictionRolling)
}

// PoseToPhysics sets the current body poses to the physics current state.
// For Dynamic bodies, sets dynamic state. Also updates world-anchored joints.
func (bd *Body) PoseToPhysics() {
	if bd.DynamicIndex >= 0 {
		params := physics.GetParams(0)
		physics.SetDynamicPos(bd.DynamicIndex, params.Next, bd.Pose.Pos)
		physics.SetDynamicQuat(bd.DynamicIndex, params.Next, bd.Pose.Quat)
	} else {
		physics.SetBodyPos(bd.DynamicIndex, bd.Pose.Pos)
		physics.SetBodyQuat(bd.DynamicIndex, bd.Pose.Quat)
	}
}
