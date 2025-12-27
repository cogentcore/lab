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

// DoF is a degree-of-freedom for a [Joint].
type DoF struct {
	// Axis is the axis of articulation.
	Axis math32.Vector3

	// Limit has the limits for motion of this DoF.
	Limit minmax.F32

	// TargetPos is the position target value, where 0 is the initial
	// position. For angular joints, this is in radians.
	TargetPos float32

	// TargetStiff determines how strongly the target position
	// is enforced: 0 = not at all; larger = stronger (e.g., 1000 or higher).
	// Set to 0 to allow the joint to be fully flexible.
	TargetStiff float32

	// TargetVel is the velocity target value. For example, 0
	// effectively damps joint movement in proportion to Damp parameter.
	TargetVel float32

	// TargetDamp determines how strongly the target velocity is enforced:
	// 0 = not at all; larger = stronger (e.g., 1 is reasonable).
	// Set to 0 to allow the joint to be fully flexible.
	TargetDamp float32
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
	jt := Joint{Parent: pidx, Child: child.ObjectIndex, Type: typ, LinearDoFN: linDoF, AngularDoFN: angDoF}
	jt.PPose.Pos = ppos
	jt.CPose.Pos = cpos
	ndof := linDoF + angDoF
	if ndof > 0 {
		jt.DoFs = make([]DoF, linDoF+angDoF)
		for i := range ndof {
			jt.DoFs[i].Limit.Min = -physics.JointLimitUnlimited
			jt.DoFs[i].Limit.Max = physics.JointLimitUnlimited
		}
	}
	ob.Joints = append(ob.Joints, jt)
	return &ob.Joints[idx]
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
func (ob *Object) NewJointFree(ml *physics.Model, parent, child *Body, ppos, cpos math32.Vector3) *Joint {
	jt := ob.newJoint(physics.Free, parent, child, ppos, cpos, 0, 0)
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
	case physics.Distance:
		ji = ml.NewJointBall(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos)
	case physics.Free:
		ji = ml.NewJointFree(pdi, cdi, jd.PPose.Pos, jd.CPose.Pos)
	}
	for i := range jd.LinearDoFN {
		d := jd.DoF(i)
		di := int32(i)
		physics.SetJointDoF(ji, di, physics.JointLimitLower, d.Limit.Min)
		physics.SetJointDoF(ji, di, physics.JointLimitUpper, d.Limit.Max)
		physics.SetJointTargetPos(ji, di, d.TargetPos, d.TargetStiff)
		physics.SetJointTargetVel(ji, di, d.TargetVel, d.TargetDamp)
	}
	for i := range jd.AngularDoFN {
		d := jd.DoF(i)
		di := int32(i + jd.LinearDoFN)
		physics.SetJointDoF(ji, di, physics.JointLimitLower, d.Limit.Min)
		physics.SetJointDoF(ji, di, physics.JointLimitUpper, d.Limit.Max)
		physics.SetJointTargetPos(ji, di, d.TargetPos, d.TargetStiff)
		physics.SetJointTargetVel(ji, di, d.TargetVel, d.TargetDamp)
	}
	jd.JointIndex = ji
	return ji
}
