// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
)

// Pose represents the 3D position and rotation.
type Pose struct {

	// Pos is the position of center of mass of object.
	Pos math32.Vector3

	// Quat is the rotation specified as a quaternion.
	Quat math32.Quat
}

// Defaults sets defaults only if current values are nil
func (ps *Pose) Defaults() {
	if ps.Quat.IsNil() {
		ps.Quat.SetIdentity()
	}
}

// Transform applies positional and rotational transform to pose.
func (ps *Pose) Transform(pos math32.Vector3, rot math32.Quat) {
	ps.Pos = rot.MulVector(ps.Pos).Add(pos)
	ps.Quat = rot.Mul(ps.Quat)
}

//////// Moving

// Move moves (translates) Pos by given amount, and sets the LinVel to the given
// delta -- this can be useful for Scripted motion to track movement.
func (ps *Pose) Move(delta math32.Vector3) {
	ps.Pos.SetAdd(delta)
}

// MoveOnAxis moves (translates) the specified distance on the specified local axis,
// relative to the current rotation orientation.
// The axis is normalized prior to aplying the distance factor.
// Sets the LinVel to motion vector.
func (ps *Pose) MoveOnAxis(x, y, z, dist float32) { //types:add
	delta := ps.Quat.MulVector(math32.Vec3(x, y, z).Normal()).MulScalar(dist)
	ps.Pos.SetAdd(delta)
}

// MoveOnAxisAbs moves (translates) the specified distance on the specified local axis,
// in absolute X,Y,Z coordinates (does not apply the Quat rotation factor.
// The axis is normalized prior to aplying the distance factor.
// Sets the LinVel to motion vector.
func (ps *Pose) MoveOnAxisAbs(x, y, z, dist float32) { //types:add
	delta := math32.Vec3(x, y, z).Normal().MulScalar(dist)
	ps.Pos.SetAdd(delta)
}

//////// Rotating

func (ps *Pose) RotateAround(rot math32.Quat, around math32.Vector3) {
	ps.Pos = rot.MulVector(ps.Pos.Sub(around)).Add(around)
	ps.Quat = rot.Mul(ps.Quat)
}

// SetEulerRotation sets the rotation in Euler angles (degrees).
func (ps *Pose) SetEulerRotation(x, y, z float32) { //types:add
	ps.Quat.SetFromEuler(math32.Vec3(x, y, z).MulScalar(math32.DegToRadFactor))
}

// SetEulerRotationRad sets the rotation in Euler angles (radians).
func (ps *Pose) SetEulerRotationRad(x, y, z float32) {
	ps.Quat.SetFromEuler(math32.Vec3(x, y, z))
}

// EulerRotation returns the current rotation in Euler angles (degrees).
func (ps *Pose) EulerRotation() math32.Vector3 { //types:add
	return ps.Quat.ToEuler().MulScalar(math32.RadToDegFactor)
}

// EulerRotationRad returns the current rotation in Euler angles (radians).
func (ps *Pose) EulerRotationRad() math32.Vector3 {
	return ps.Quat.ToEuler()
}

// SetAxisRotation sets rotation from local axis and angle in degrees.
func (ps *Pose) SetAxisRotation(x, y, z, angle float32) { //types:add
	ps.Quat.SetFromAxisAngle(math32.Vec3(x, y, z), math32.DegToRad(angle))
}

// SetAxisRotationRad sets rotation from local axis and angle in radians.
func (ps *Pose) SetAxisRotationRad(x, y, z, angle float32) {
	ps.Quat.SetFromAxisAngle(math32.Vec3(x, y, z), angle)
}

// RotateOnAxis rotates around the specified local axis the specified angle in degrees.
func (ps *Pose) RotateOnAxis(x, y, z, angle float32) { //types:add
	ps.Quat.SetMul(math32.NewQuatAxisAngle(math32.Vec3(x, y, z), math32.DegToRad(angle)))
}

// RotateOnAxisRad rotates around the specified local axis the specified angle in radians.
func (ps *Pose) RotateOnAxisRad(x, y, z, angle float32) {
	ps.Quat.SetMul(math32.NewQuatAxisAngle(math32.Vec3(x, y, z), angle))
}

// RotateEuler rotates by given Euler angles (in degrees) relative to existing rotation.
func (ps *Pose) RotateEuler(x, y, z float32) { //types:add
	ps.Quat.SetMul(math32.NewQuatEuler(math32.Vec3(x, y, z).MulScalar(math32.DegToRadFactor)))
}

// RotateEulerRad rotates by given Euler angles (in radians) relative to existing rotation.
func (ps *Pose) RotateEulerRad(x, y, z, angle float32) {
	ps.Quat.SetMul(math32.NewQuatEuler(math32.Vec3(x, y, z)))
}

// MakePoseToolbar returns a toolbar function for physics state updates,
// calling the given updt function after making the change.
func MakePoseToolbar(ps *Pose, updt func()) func(p *tree.Plan) {
	return func(p *tree.Plan) {
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.SetEulerRotation).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.SetAxisRotation).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.RotateEuler).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.RotateOnAxis).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.EulerRotation).SetAfterFunc(updt).SetShowReturn(true).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.Separator) {})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.MoveOnAxis).SetAfterFunc(updt).SetIcon(icons.MoveItem)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.MoveOnAxisAbs).SetAfterFunc(updt).SetIcon(icons.MoveItem)
		})

	}
}
