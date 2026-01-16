// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phyxyz

import (
	"fmt"
	"strconv"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/lab/physics"
)

// Skin has visualization functions for physics elements.
type Skin struct { //types:add -setters
	// Name is a name for element (index always appended, so it is unique).
	Name string

	// Shape is the physical shape of the element.
	Shape physics.Shapes

	// Color is the color of the element.
	Color string

	// HSize is the half-size (e.g., radius) of the body.
	// Values depend on shape type: X is generally radius,
	// Y is half-height.
	HSize math32.Vector3

	// Pos is the position.
	Pos math32.Vector3

	// Quat is the rotation as a quaternion.
	Quat math32.Quat

	// NewSkin is a function that returns a new [xyz.Node]
	// to represent this element. If nil, uses appropriate defaults.
	NewSkin func() tree.Node

	// InitSkin is a function that initializes a new [xyz.Node]
	// that represents this element. If nil, uses appropriate defaults.
	InitSkin func(sld *xyz.Solid)

	// BodyIndex is the index of the body in [physics.Bodies]
	BodyIndex int32

	// DynamicIndex is the index in [physics.Dynamics] (-1 if not dynamic).
	DynamicIndex int32
}

// NewBody adds a new body with given parameters.
// Returns the Skin which can then be further customized.
// Use this for Static elements; NewDynamic for dynamic elements.
func (sc *Scene) NewBody(ml *physics.Model, name string, shape physics.Shapes, clr string, hsize, pos math32.Vector3, rot math32.Quat) *Skin {
	idx := ml.NewBody(shape, hsize, pos, rot)
	sk := sc.NewSkin(shape, name, clr, hsize, pos, rot)
	sk.SetBodyIndex(idx)
	return sk
}

// NewDynamic adds a new dynamic body with given parameters.
// Returns the Skin which can then be further customized.
func (sc *Scene) NewDynamic(ml *physics.Model, name string, shape physics.Shapes, clr string, mass float32, hsize, pos math32.Vector3, rot math32.Quat) *Skin {
	idx, dyIdx := ml.NewDynamic(shape, mass, hsize, pos, rot)
	sk := sc.NewSkin(shape, name, clr, hsize, pos, rot)
	sk.SetBodyIndex(idx).SetDynamicIndex(dyIdx)
	return sk
}

// UpdateFromPhysics updates the Skin from physics state.
func (sk *Skin) UpdateFromPhysics(sc *Scene) {
	params := physics.GetParams(0)
	di := int32(sk.DynamicIndex)
	bi := int32(sk.BodyIndex)
	if sc.ReplicasView {
		bi, di = physics.CurModel.ReplicasBodyIndexes(bi, int32(sc.ReplicasIndex))
	}
	if di >= 0 {
		sk.Pos = physics.DynamicPos(di, params.Cur)
		sk.Quat = physics.DynamicQuat(di, params.Cur)
	} else {
		sk.Pos = physics.BodyPos(bi)
		sk.Quat = physics.BodyQuat(bi)
	}
}

// UpdatePose updates the xyz node pose from skin.
func (sk *Skin) UpdatePose(sld *xyz.Solid) {
	sld.Pose.Pos = sk.Pos
	sld.Pose.Quat = sk.Quat
}

// UpdateColor updates the xyz node color from skin.
func (sk *Skin) UpdateColor(clr string, sld *xyz.Solid) {
	if clr == "" {
		return
	}
	sld.Material.Color = errors.Log1(colors.FromString(clr))
}

// Add adds given physics node to the [tree.Plan], using NewSkin
// function on the node, or default.
func (sk *Skin) Add(p *tree.Plan) {
	nm := sk.Name + strconv.Itoa(int(sk.BodyIndex))
	newFunc := sk.NewSkin
	if newFunc == nil {
		newFunc = func() tree.Node {
			return any(tree.New[xyz.Solid]()).(tree.Node)
		}
	}
	p.Add(nm, newFunc, func(n tree.Node) { sk.Init(n.(*xyz.Solid)) })
}

// Init initializes xyz node using InitSkin function or default.
func (sk *Skin) Init(sld *xyz.Solid) {
	initFunc := sk.InitSkin
	if initFunc != nil {
		initFunc(sld)
		return
	}
	switch sk.Shape {
	case physics.Plane:
		sk.PlaneInit(sld)
	case physics.Sphere:
		sk.SphereInit(sld)
	case physics.Capsule:
		sk.CapsuleInit(sld)
	case physics.Cylinder:
		sk.CylinderInit(sld)
	case physics.Box:
		sk.BoxInit(sld)
	}
}

// BoxInit is the default InitSkin function for [physics.Box].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (sk *Skin) BoxInit(sld *xyz.Solid) {
	mnm := "physics.Box"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		xyz.NewBox(sld.Scene, mnm, 1, 1, 1)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale = sk.HSize.MulScalar(2)
	sk.UpdateColor(sk.Color, sld)
	sld.Updater(func() {
		sk.UpdatePose(sld)
	})
}

// PlaneInit is the default InitSkin function for [physics.Plane].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (sk *Skin) PlaneInit(sld *xyz.Solid) {
	mnm := "physics.Plane"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		pl := xyz.NewPlane(sld.Scene, mnm, 1, 1)
		pl.Segs.Set(4, 4)
	}
	sld.SetMeshName(mnm)
	if sk.HSize.X == 0 {
		inf := float32(1e3)
		sld.Pose.Scale = math32.Vec3(inf, 1, inf)
	} else {
		sld.Pose.Scale = sk.HSize.MulScalar(2)
	}
	sk.UpdateColor(sk.Color, sld)
	sld.Updater(func() {
		sk.UpdatePose(sld)
	})
}

// CylinderInit is the default InitSkin function for [physics.Cylinder].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (sk *Skin) CylinderInit(sld *xyz.Solid) {
	mnm := "physics.Cylinder"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		xyz.NewCylinder(sld.Scene, mnm, 1, 1, 32, 1, true, true)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale = sk.HSize
	sld.Pose.Scale.Y *= 2
	sk.UpdateColor(sk.Color, sld)
	sld.Updater(func() {
		sk.UpdatePose(sld)
	})
}

// CapsuleInit is the default InitSkin function for [physics.Capsule].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (sk *Skin) CapsuleInit(sld *xyz.Solid) {
	rat := sk.HSize.Y / sk.HSize.X
	mnm := fmt.Sprintf("physics.Capsule_%g", math32.Truncate(rat, 3))
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		ms = xyz.NewCapsule(sld.Scene, mnm, 2*(sk.HSize.Y-sk.HSize.X)/sk.HSize.X, 1, 32, 1)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale.Set(sk.HSize.X, sk.HSize.X, sk.HSize.X)
	sk.UpdateColor(sk.Color, sld)
	sld.Updater(func() {
		sk.UpdatePose(sld)
	})
}

// SphereInit is the default InitSkin function for [physics.Sphere].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (sk *Skin) SphereInit(sld *xyz.Solid) {
	mnm := "physics.Sphere"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		ms = xyz.NewSphere(sld.Scene, mnm, 1, 32)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale.SetScalar(sk.HSize.X)
	sk.UpdateColor(sk.Color, sld)
	sld.Updater(func() {
		sk.UpdatePose(sld)
	})
}

// SetBodyWorld partitions bodies into different worlds for
// collision detection: Global bodies = -1 can collide with
// everything; otherwise only items within the same world collide.
func (sk *Skin) SetBodyWorld(world int) {
	physics.SetBodyWorld(sk.BodyIndex, int32(world))
}

// SetBodyGroup partitions bodies within worlds into different groups
// for collision detection. 0 does not collide with anything.
// Negative numbers are global within a world, except they don't
// collide amongst themselves (all non-dynamic bodies should go
// in -1 because they don't collide amongst each-other, but do
// potentially collide with dynamics).
// Positive numbers only collide amongst themselves, and with
// negative groups, but not other positive groups. This is for
// more special-purpose dynamics: in general use 1 for all dynamic
// bodies. There is an automatic constraint that the two objects
// within a single joint do not collide with each other, so this
// does not need to be handled here.
func (sk *Skin) SetBodyGroup(group int) {
	physics.SetBodyGroup(sk.BodyIndex, int32(group))
}

// SetBodyBounce specifies the COR or coefficient of restitution (0..1),
// which determines how elastic the collision is,
// i.e., final velocity / initial velocity.
func (sk *Skin) SetBodyBounce(val float32) {
	physics.Bodies.Set(val, int(sk.BodyIndex), int(physics.BodyBounce))
}

// SetBodyFriction is the standard coefficient for linear friction (mu).
func (sk *Skin) SetBodyFriction(val float32) {
	physics.Bodies.Set(val, int(sk.BodyIndex), int(physics.BodyFriction))
}

// SetBodyFrictionTortion is resistance to spinning at the contact point.
func (sk *Skin) SetBodyFrictionTortion(val float32) {
	physics.Bodies.Set(val, int(sk.BodyIndex), int(physics.BodyFrictionTortion))
}

// SetBodyFrictionRolling is resistance to rolling motion at contact.
func (sk *Skin) SetBodyFrictionRolling(val float32) {
	physics.Bodies.Set(val, int(sk.BodyIndex), int(physics.BodyFrictionRolling))
}

// NewJointFixed adds a new Fixed joint between given parent and child.
// Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
func (sc *Scene) NewJointFixed(ml *physics.Model, parent, child *Skin, ppos, cpos math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ml.NewJointFixed(pidx, child.DynamicIndex, ppos, cpos)
}

// NewJointPrismatic adds a new Prismatic (slider) joint between given
// parent and child. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// axis is the axis of articulation for the joint.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (sc *Scene) NewJointPrismatic(ml *physics.Model, parent, child *Skin, ppos, cpos, axis math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ml.NewJointPrismatic(pidx, child.DynamicIndex, ppos, cpos, axis)
}

// NewJointRevolute adds a new Revolute (hinge, axel) joint between given
// parent and child. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// axis is the axis of articulation for the joint.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (sc *Scene) NewJointRevolute(ml *physics.Model, parent, child *Skin, ppos, cpos, axis math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ml.NewJointRevolute(pidx, child.DynamicIndex, ppos, cpos, axis)
}

// NewJointBall adds a new Ball joint (3 angular DoF) between given parent
// and child. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (sc *Scene) NewJointBall(ml *physics.Model, parent, child *Skin, ppos, cpos math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ml.NewJointBall(pidx, child.DynamicIndex, ppos, cpos)
}

// NewJointDistance adds a new Distance joint (6 DoF),
// with distance constrained only on the first linear X axis,
// between given parent and child. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (sc *Scene) NewJointDistance(ml *physics.Model, parent, child *Skin, ppos, cpos math32.Vector3, minDist, maxDist float32) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ml.NewJointDistance(pidx, child.DynamicIndex, ppos, cpos, minDist, maxDist)
}

// NewJointFree adds a new Free joint between given parent and child.
// Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (sc *Scene) NewJointFree(ml *physics.Model, parent, child *Skin, ppos, cpos math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ml.NewJointFree(pidx, child.DynamicIndex, ppos, cpos)
}
