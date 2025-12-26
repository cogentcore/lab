// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phyxyz

import (
	"strconv"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/lab/physics"
)

// View has visualization functions for physics elements.
type View struct {
	// Name is a name for element (index always appended).
	Name string

	// Shape is the physical shape of the element.
	Shape physics.Shapes

	// Color is the color of the element.
	Color string

	// Size is the size (per shape).
	Size math32.Vector3

	// Pos is the position.
	Pos math32.Vector3

	// Quat is the rotation as a quaternion.
	Quat math32.Quat

	// NewView is a function that returns a new [xyz.Node]
	// to represent this element. If nil, uses appropriate defaults.
	NewView func() tree.Node

	// InitView is a function that initializes a new [xyz.Node]
	// that represents this element. If nil, uses appropriate defaults.
	InitView func(sld *xyz.Solid)

	// Index is the index of the element in a list.
	Index int32

	// DynamicIndex is the index of a dynamic element (-1 if not dynamic).
	DynamicIndex int32
}

// NewBody adds a new body with given parameters.
// Returns the View which can then be further customized.
// Use this for Static elements; NewDynamic for dynamic elements.
func (wr *Scene) NewBody(ph *physics.Model, name string, shape physics.Shapes, clr string, size, pos math32.Vector3, rot math32.Quat) *View {
	idx := ph.NewBody(shape, size, pos, rot)
	vw := &View{Name: name, Index: idx, DynamicIndex: -1, Shape: shape, Color: clr, Size: size, Pos: pos, Quat: rot}
	wr.Views = append(wr.Views, vw)
	return vw
}

// NewDynamic adds a new dynamic body with given parameters.
// Returns the View which can then be further customized.
func (wr *Scene) NewDynamic(ph *physics.Model, name string, shape physics.Shapes, clr string, mass float32, size, pos math32.Vector3, rot math32.Quat) *View {
	idx, dyIdx := ph.NewDynamic(shape, mass, size, pos, rot)
	vw := &View{Name: name, Index: idx, DynamicIndex: dyIdx, Shape: shape, Color: clr, Size: size, Pos: pos, Quat: rot}
	wr.Views = append(wr.Views, vw)
	return vw
}

// UpdateFromPhysics updates the View from physics state.
func (vw *View) UpdateFromPhysics() {
	params := physics.GetParams(0)
	if vw.DynamicIndex >= 0 {
		ix := int32(vw.DynamicIndex)
		vw.Pos = physics.DynamicPos(ix, params.Cur)
		vw.Quat = physics.DynamicQuat(ix, params.Cur)
	} else {
		ix := int32(vw.Index)
		vw.Pos = physics.BodyPos(ix)
		vw.Quat = physics.BodyQuat(ix)
	}
}

// UpdatePose updates the xyz node pose from view.
func (vw *View) UpdatePose(sld *xyz.Solid) {
	sld.Pose.Pos = vw.Pos
	sld.Pose.Quat = vw.Quat
}

// UpdateColor updates the xyz node color from view.
func (vw *View) UpdateColor(clr string, sld *xyz.Solid) {
	if clr == "" {
		return
	}
	sld.Material.Color = errors.Log1(colors.FromString(clr))
}

// Add adds given physics node to the [tree.Plan], using NewView
// function on the node, or default.
func (vw *View) Add(p *tree.Plan) {
	nm := vw.Name + strconv.Itoa(int(vw.Index))
	newFunc := vw.NewView
	if newFunc == nil {
		newFunc = func() tree.Node {
			return any(tree.New[xyz.Solid]()).(tree.Node)
		}
	}
	p.Add(nm, newFunc, func(n tree.Node) { vw.Init(n.(*xyz.Solid)) })
}

// Init initializes xyz node using InitView function or default.
func (vw *View) Init(sld *xyz.Solid) {
	initFunc := vw.InitView
	if initFunc != nil {
		initFunc(sld)
		return
	}
	switch vw.Shape {
	case physics.Plane:
		vw.PlaneInit(sld)
	case physics.Sphere:
		vw.SphereInit(sld)
	case physics.Capsule:
		vw.CapsuleInit(sld)
	case physics.Cylinder:
		vw.CylinderInit(sld)
	case physics.Box:
		vw.BoxInit(sld)
	}
}

// BoxInit is the default InitView function for [physics.Box].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) BoxInit(sld *xyz.Solid) {
	mnm := "physics.Box"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		xyz.NewBox(sld.Scene, mnm, 1, 1, 1)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale = vw.Size.MulScalar(2)
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// PlaneInit is the default InitView function for [physics.Plane].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) PlaneInit(sld *xyz.Solid) {
	mnm := "physics.Plane"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		pl := xyz.NewPlane(sld.Scene, mnm, 1, 1)
		pl.Segs.Set(4, 4)
	}
	sld.SetMeshName(mnm)
	if vw.Size.X == 0 {
		inf := float32(1e3)
		sld.Pose.Scale = math32.Vec3(inf, 1, inf)
	} else {
		sld.Pose.Scale = vw.Size.MulScalar(2)
	}
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// CylinderInit is the default InitView function for [physics.Cylinder].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) CylinderInit(sld *xyz.Solid) {
	mnm := "physics.Cylinder"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		xyz.NewCylinder(sld.Scene, mnm, 1, 1, 32, 1, true, true)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale = vw.Size
	sld.Pose.Scale.Y *= 2
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// CapsuleInit is the default InitView function for [physics.Capsule].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) CapsuleInit(sld *xyz.Solid) {
	mnm := "physics.Capsule"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		ms = xyz.NewCapsule(sld.Scene, mnm, 1, .2, 32, 1)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale.Set(vw.Size.X/.2, 2*(vw.Size.Y/1.4), vw.Size.Z/.2)
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// SphereInit is the default InitView function for [physics.Sphere].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) SphereInit(sld *xyz.Solid) {
	mnm := "physics.Sphere"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		ms = xyz.NewSphere(sld.Scene, mnm, 1, 32)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale.SetScalar(vw.Size.X)
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// SetBodyWorld partitions bodies into different worlds for
// collision detection: Global bodies = -1 can collide with
// everything; otherwise only items within the same world collide.
func (vw *View) SetBodyWorld(world int) {
	physics.SetBodyWorld(vw.Index, int32(world))
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
func (vw *View) SetBodyGroup(group int) {
	physics.SetBodyGroup(vw.Index, int32(group))
}

// SetBodyBounce specifies the COR or coefficient of restitution (0..1),
// which determines how elastic the collision is,
// i.e., final velocity / initial velocity.
func (vw *View) SetBodyBounce(val float32) {
	physics.Bodies.Set(val, int(vw.Index), int(physics.BodyBounce))
}

// SetBodyFriction is the standard coefficient for linear friction (mu).
func (vw *View) SetBodyFriction(val float32) {
	physics.Bodies.Set(val, int(vw.Index), int(physics.BodyFriction))
}

// SetBodyFrictionTortion is resistance to spinning at the contact point.
func (vw *View) SetBodyFrictionTortion(val float32) {
	physics.Bodies.Set(val, int(vw.Index), int(physics.BodyFrictionTortion))
}

// SetBodyFrictionRolling is resistance to rolling motion at contact.
func (vw *View) SetBodyFrictionRolling(val float32) {
	physics.Bodies.Set(val, int(vw.Index), int(physics.BodyFrictionRolling))
}

// NewJointFixed adds a new Fixed joint as a child of given parent.
// Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
func (vw *View) NewJointFixed(ph *physics.Model, parent *View, ppos, cpos math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ph.NewJointFixed(pidx, vw.DynamicIndex, ppos, cpos)
}

// NewJointPrismatic adds a new Prismatic (slider) joint as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// axis is the axis of articulation for the joint.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (vw *View) NewJointPrismatic(ph *physics.Model, parent *View, ppos, cpos, axis math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ph.NewJointPrismatic(pidx, vw.DynamicIndex, ppos, cpos, axis)
}

// NewJointRevolute adds a new Revolute (hinge, axel) joint as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// axis is the axis of articulation for the joint.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (vw *View) NewJointRevolute(ph *physics.Model, parent *View, ppos, cpos, axis math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ph.NewJointRevolute(pidx, vw.DynamicIndex, ppos, cpos, axis)
}

// NewJointBall adds a new Ball joint (3 angular DoF) as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (vw *View) NewJointBall(ph *physics.Model, parent *View, ppos, cpos math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ph.NewJointBall(pidx, vw.DynamicIndex, ppos, cpos)
}

// NewJointDistance adds a new Distance joint (6 DoF),
// with distance constrained only on the first linear X axis,
// as a child of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (vw *View) NewJointDistance(ph *physics.Model, parent *View, ppos, cpos math32.Vector3, minDist, maxDist float32) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ph.NewJointDistance(pidx, vw.DynamicIndex, ppos, cpos, minDist, maxDist)
}

// NewJointFree adds a new Free joint as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (vw *View) NewJointFree(ph *physics.Model, parent *View, ppos, cpos math32.Vector3) int32 {
	pidx := int32(-1)
	if parent != nil {
		pidx = parent.DynamicIndex
	}
	return ph.NewJointFree(pidx, vw.DynamicIndex, ppos, cpos)
}
