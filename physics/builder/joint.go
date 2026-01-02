// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/physics"
)

// Joint describes a joint between two bodies.
type Joint struct {
	// Parent is index within an Object for parent body.
	// -1 for world-anchored parent.
	Parent int

	// Parent is index within an Object for parent body.
	Child int

	// Type is the type of the joint.
	Type physics.JointTypes

	// PPose is the parent position and orientation of the joint
	// in the parent's body-centered coordinates.
	PPose Pose

	// CPose is the child position and orientation of the joint
	// in the parent's body-centered coordinates.
	CPose Pose

	// LinearDoFN is the number of linear degrees of freedom (3 max).
	LinearDoFN int

	// AngularDoFN is the number of linear degrees of freedom (3 max).
	AngularDoFN int

	// DoFs are the degrees-of-freedom for this joint.
	DoFs []DoF

	// JointIndex is the index of this joint in [physics.Joints] when built.
	JointIndex int32
}

// Controls are the per degrees-of-freedom (DoF) joint control inputs.
type Controls struct {
	// Force is the force input driving the joint.
	Force float32

	// Pos is the position target value, where 0 is the initial
	// position. For angular joints, this is in radians.
	Pos float32

	// Stiff determines how strongly the target position
	// is enforced: 0 = not at all; larger = stronger (e.g., 1000 or higher).
	// Set to 0 to allow the joint to be fully flexible.
	Stiff float32

	// Vel is the velocity target value. For example, 0
	// effectively damps joint movement in proportion to Damp parameter.
	Vel float32

	// Damp determines how strongly the target velocity is enforced:
	// 0 = not at all; larger = stronger (e.g., 1 is reasonable).
	// Set to 0 to allow the joint to be fully flexible.
	Damp float32
}

func (ct *Controls) Defaults() {
	ct.Stiff = 1000
	ct.Damp = 20
}

// DoF is a degree-of-freedom for a [Joint].
type DoF struct {
	// Axis is the axis of articulation.
	Axis math32.Vector3

	// Limit has the limits for motion of this DoF.
	Limit minmax.F32

	// Init are the initial control values.
	Init Controls

	// Current are the current control values (based on method calls).
	Current Controls
}

func (df *DoF) Defaults() {
	df.Limit.Min = -physics.JointLimitUnlimited
	df.Limit.Max = physics.JointLimitUnlimited
	df.Init.Defaults()
	df.Current.Defaults()
}

func (df *DoF) InitState() {
	df.Current = df.Init

}

func (jd *Joint) DoF(idx int) *DoF {
	return &jd.DoFs[idx]
}

// newJoint adds a new joint of given type.
func (ob *Object) newJoint(typ physics.JointTypes, parent, child *Body, ppos, cpos math32.Vector3, linDoF, angDoF int) *Joint {
	pidx := -1
	if parent != nil {
		pidx = parent.ObjectIndex
	}
	idx := len(ob.Joints)
	ob.Joints = append(ob.Joints, Joint{Parent: pidx, Child: child.ObjectIndex, Type: typ, LinearDoFN: linDoF, AngularDoFN: angDoF})
	jd := ob.Joint(idx)
	jd.PPose.Pos = ppos
	jd.CPose.Pos = cpos
	ndof := linDoF + angDoF
	if ndof > 0 {
		jd.DoFs = make([]DoF, linDoF+angDoF)
		for i := range ndof {
			dof := jd.DoF(i)
			dof.Defaults()
		}
	}
	return jd
}

// NewJointFixed adds a new Fixed joint as a child of given parent.
// Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
func (ob *Object) NewJointFixed(parent, child *Body, ppos, cpos math32.Vector3) *Joint {
	return ob.newJoint(physics.Fixed, parent, child, ppos, cpos, 0, 0)
}

// NewJointPrismatic adds a new Prismatic (slider) joint as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// axis is the axis of articulation for the joint.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (ob *Object) NewJointPrismatic(parent, child *Body, ppos, cpos, axis math32.Vector3) *Joint {
	jt := ob.newJoint(physics.Prismatic, parent, child, ppos, cpos, 1, 0)
	jt.DoFs[0].Axis = axis
	return jt
}

// NewJointRevolute adds a new Revolute (hinge, axel) joint as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// axis is the axis of articulation for the joint.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (ob *Object) NewJointRevolute(parent, child *Body, ppos, cpos, axis math32.Vector3) *Joint {
	jt := ob.newJoint(physics.Revolute, parent, child, ppos, cpos, 0, 1)
	jt.DoFs[0].Axis = axis
	return jt
}

// NewJointBall adds a new Ball joint (3 angular DoF) as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (ob *Object) NewJointBall(parent, child *Body, ppos, cpos math32.Vector3) *Joint {
	jt := ob.newJoint(physics.Ball, parent, child, ppos, cpos, 0, 3)
	return jt
}

// NewJointDistance adds a new Distance joint (6 DoF),
// with distance constrained only on the first linear X axis,
// as a child of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (ob *Object) NewJointDistance(parent, child *Body, ppos, cpos math32.Vector3, minDist, maxDist float32) *Joint {
	jt := ob.newJoint(physics.Ball, parent, child, ppos, cpos, 3, 3)
	jt.DoFs[0].Limit.Min = minDist
	jt.DoFs[0].Limit.Max = maxDist
	return jt
}

// NewJointFree adds a new Free joint as a child
// of given parent. Use nil for parent to add a world-anchored joint.
// ppos, cpos are the relative positions from the parent, child.
// These are for the non-rotated body (i.e., body rotation is applied
// to these positions as well).
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (ob *Object) NewJointFree(parent, child *Body, ppos, cpos math32.Vector3) *Joint {
	jt := ob.newJoint(physics.Free, parent, child, ppos, cpos, 0, 0)
	return jt
}

// NewJointPlaneXZ adds a new 3 DoF Planar motion joint suitable for
// controlling the motion of a body on the standard X-Z play (Y = up).
// The first two linear DoF control position in X, Z, and the third
// angular DoF controls rotation in the plane (along the Y axis).
// Use -1 for parent to add a world-anchored joint (typical).
// ppos, cpos are the relative positions from the parent, child.
// Sets relative rotation matricies to identity by default.
// Use [SetJointDoF] to set the remaining DoF parameters.
func (ob *Object) NewJointPlaneXZ(parent, child *Body, ppos, cpos math32.Vector3) *Joint {
	jt := ob.newJoint(physics.PlaneXZ, parent, child, ppos, cpos, 2, 1)
	return jt
}

// NewPhysicsJoint makes the physics joint for joint
func (jd *Joint) NewPhysicsJoint(ml *physics.Model, ob *Object) int32 {
	pi := jd.Parent
	pdi := int32(-1)
	if pi >= 0 {
		pb := ob.Body(pi)
		pdi = pb.DynamicIndex // todo: validate
	}
	cb := ob.Body(jd.Child)
	cdi := cb.DynamicIndex
	ji := int32(0)
	switch jd.Type {
	case physics.Prismatic:
		ji = ml.NewJointPrismatic(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos, jd.DoFs[0].Axis)
	case physics.Revolute:
		ji = ml.NewJointRevolute(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos, jd.DoFs[0].Axis)
	case physics.Ball:
		ji = ml.NewJointBall(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos)
	case physics.Fixed:
		ji = ml.NewJointFixed(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos)
	case physics.Distance:
		ji = ml.NewJointBall(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos)
	case physics.Free:
		ji = ml.NewJointFree(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos)
	case physics.PlaneXZ:
		ji = ml.NewJointPlaneXZ(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos)
	}
	for i := range jd.LinearDoFN {
		d := jd.DoF(i)
		di := int32(i)
		physics.SetJointDoF(ji, di, physics.JointLimitLower, d.Limit.Min)
		physics.SetJointDoF(ji, di, physics.JointLimitUpper, d.Limit.Max)
		physics.SetJointTargetPos(ji, di, d.Init.Pos, d.Init.Stiff)
		physics.SetJointTargetVel(ji, di, d.Init.Vel, d.Init.Damp)
		d.Axis = physics.JointAxis(ji, di)
	}
	for i := range jd.AngularDoFN {
		di := int32(i + jd.LinearDoFN)
		d := jd.DoF(int(di))
		physics.SetJointDoF(ji, di, physics.JointLimitLower, d.Limit.Min)
		physics.SetJointDoF(ji, di, physics.JointLimitUpper, d.Limit.Max)
		physics.SetJointTargetPos(ji, di, d.Init.Pos, d.Init.Stiff)
		physics.SetJointTargetVel(ji, di, d.Init.Vel, d.Init.Damp)
		d.Axis = physics.JointAxis(ji, di)
	}
	jd.JointIndex = ji
	// fmt.Println("\t\tjoint:", pdi, cdi, jd.Type)
	// if pdi < 0 {
	// 	fmt.Println("\t\t\t", jd.PPose.Pos)
	// }
	return ji
}

// IsGlobal returns true if this joint has a global world anchor parent.
func (jd *Joint) IsGlobal() bool {
	return jd.Parent < 0
}

// InitState initializes current state variables in the Joint.
func (jd *Joint) InitState() {
	ji := jd.JointIndex
	for di := range jd.DoFs {
		d := jd.DoF(di)
		d.InitState()
		physics.SetJointTargetPos(ji, int32(di), d.Init.Pos, d.Init.Stiff)
		physics.SetJointTargetVel(ji, int32(di), d.Init.Vel, d.Init.Damp)
	}
}

// PoseToPhysics sets the current world-anchored joint pose
// to the physics current state.
func (jd *Joint) PoseToPhysics() {
	if !jd.IsGlobal() {
		return
	}
	physics.SetJointPPos(jd.JointIndex, jd.PPose.Pos)
	physics.SetJointPQuat(jd.JointIndex, jd.PPose.Quat)
}

// PoseFromPhysics gets the current world-anchored joint pose
// from the physics current state.
func (jd *Joint) PoseFromPhysics() {
	if !jd.IsGlobal() {
		return
	}
	jd.PPose.Pos = physics.JointPPos(jd.JointIndex)
	jd.PPose.Quat = physics.JointPQuat(jd.JointIndex)
}

// SetTargetVel sets the target position for given DoF for
// this joint in the physics model. Records into [DoF.Current].
func (jd *Joint) SetTargetVel(dof int32, vel, damp float32) {
	d := jd.DoF(int(dof))
	d.Current.Vel = vel
	d.Current.Damp = damp
	physics.SetJointTargetVel(jd.JointIndex, dof, vel, damp)
}

// SetTargetPos sets the target position for given DoF for
// this joint in the physics model. Records into [DoF.Current].
func (jd *Joint) SetTargetPos(dof int32, pos, stiff float32) {
	d := jd.DoF(int(dof))
	d.Current.Pos = pos
	d.Current.Stiff = stiff
	physics.SetJointTargetPos(jd.JointIndex, dof, pos, stiff)
}

// AddTargetPos adds to the Current target position for given DoF for
// this joint in the physics model, setting stiffness.
func (jd *Joint) AddTargetPos(dof int32, pos, stiff float32) {
	d := jd.DoF(int(dof))
	d.Current.Pos += pos
	d.Current.Stiff = stiff
	physics.SetJointTargetPos(jd.JointIndex, dof, d.Current.Pos, stiff)
}

// SetTargetAngle sets the target angular position
// and stiffness for given joint, DoF to given values.
// Stiffness determines how strongly the joint constraint is enforced
// (0 = not at all; 1000+ = strongly).
// Angle is in Degrees, not radians. Usable range is within -180..180
// which is enforced, and values near the edge can be unstable at higher
// stiffness levels.
func (jd *Joint) SetTargetAngle(dof int32, angDeg, stiff float32) {
	pos := math32.WrapPi(math32.DegToRad(angDeg))
	d := jd.DoF(int(dof))
	d.Current.Pos = pos
	d.Current.Stiff = stiff
	physics.SetJointTargetPos(jd.JointIndex, dof, pos, stiff)
}

// AddTargetAngle adds to the Current target angular position,
// and sets stiffness for given joint, DoF to given values.
// Stiffness determines how strongly the joint constraint is enforced
// (0 = not at all; 1000+ = strongly).
// Angle is in Degrees, not radians. Usable range is within -180..180
// which is enforced, and values near the edge can be unstable at higher
// stiffness levels.
func (jd *Joint) AddTargetAngle(dof int32, angDeg, stiff float32) {
	d := jd.DoF(int(dof))
	d.Current.Pos = math32.WrapPi(d.Current.Pos + math32.DegToRad(angDeg))
	d.Current.Stiff = stiff
	physics.SetJointTargetPos(jd.JointIndex, dof, d.Current.Pos, stiff)
}

// AddPlaneXZPos adds to the Current target X and Z axis positions for
// a PlaneXZ joint, using the current Y axis rotation angle to project
// along the current angle direction. angOff provides an angle offset to
// add to the Y axis angle.
func (jd *Joint) AddPlaneXZPos(angOff, delta, stiff float32) {
	ang := angOff - jd.DoF(2).Current.Pos
	dx := delta * math32.Cos(ang)
	dz := delta * math32.Sin(ang)
	jd.AddTargetPos(0, dx, stiff)
	jd.AddTargetPos(1, dz, stiff)
}
