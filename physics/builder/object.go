// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"slices"

	"cogentcore.org/core/math32"
	"cogentcore.org/lab/physics/phyxyz"
)

// Object is an object within the [World].
// Each object is a coherent collection of bodies, typically
// connected by joints. This is an organizational convenience
// for positioning elements; has no physical implications.
type Object struct {
	// World is the world number for physics: -1 = globals, else positive
	// are distinct non-interacting worlds.
	World int

	// WorldIndex is the index of world within builder Worlds list.
	WorldIndex int

	// Object is the index within World's Objects list.
	Object int

	// Bodies are the bodies in the object.
	Bodies []*Body

	// Joints are joints connecting object bodies.
	// Joint indexes here refer strictly within bodies.
	Joints []*Joint

	// Sensors are functions that can be configured to report arbitrary values
	// on given body element. The output must be stored directly somewhere via
	// the closure function: the utility of the sensor function is being able
	// to capture all the configuration-time parameters needed to make it work,
	// and to have it automatically called on replicated objects.
	Sensors []func(obj *Object)
}

func (ob *Object) Body(idx int) *Body {
	return ob.Bodies[idx]
}

func (ob *Object) Joint(idx int) *Joint {
	return ob.Joints[idx]
}

// Copy copies all bodies and joints from given source world into this one.
// (The objects will be identical after, regardless of current starting
// condition).
func (ob *Object) Copy(so *Object) {
	ob.World = so.World
	ob.Object = so.Object
	ob.Bodies = make([]*Body, len(so.Bodies))
	ob.Joints = make([]*Joint, len(so.Joints))
	ob.Sensors = make([]func(obj *Object), len(so.Sensors))
	for i := range ob.Bodies {
		ob.Bodies[i] = &Body{}
		ob.Body(i).Copy(so.Body(i))
	}
	for i := range ob.Joints {
		ob.Joints[i] = &Joint{}
		ob.Joint(i).Copy(so.Joint(i))
	}
	copy(ob.Sensors, so.Sensors)
}

// CopySkins makes new skins for bodies based on those in source object.
// Which must have same number of bodies.
func (ob *Object) CopySkins(sc *phyxyz.Scene, so *Object) {
	for i := range ob.Bodies {
		bd := ob.Body(i)
		sb := so.Body(i)
		bd.NewSkin(sc, sb.Skin.Name, sb.Skin.Color)
	}
}

// InitState initializes current state variables in the object.
func (ob *Object) InitState() {
	for _, jd := range ob.Joints {
		jd.InitState()
	}
}

// HasBodyIndex returns true if a body in the object has any of
// given body index(es).
func (ob *Object) HasBodyIndex(bodyIndex ...int32) bool {
	for _, bd := range ob.Bodies {
		if slices.Contains(bodyIndex, bd.BodyIndex) {
			return true
		}
	}
	return false
}

//////// Transforms

// PoseToPhysics sets the current body poses to the physics current state.
// For Dynamic bodies, sets dynamic state. Also updates world-anchored joints.
func (ob *Object) PoseToPhysics() {
	for _, bd := range ob.Bodies {
		bd.PoseToPhysics()
	}
	for _, jd := range ob.Joints {
		jd.PoseToPhysics()
	}
}

// PoseFromPhysics gets the current body poses from the physics current state.
// Also updates world-anchored joints.
func (ob *Object) PoseFromPhysics() {
	for _, bd := range ob.Bodies {
		bd.PoseFromPhysics()
	}
	for _, jd := range ob.Joints {
		jd.PoseFromPhysics()
	}
}

// Move applies positional and rotational transforms to all bodies,
// and world-anchored joints.
func (ob *Object) Move(pos math32.Vector3) {
	for _, bd := range ob.Bodies {
		bd.Pose.Move(pos)
	}
	for _, jd := range ob.Joints {
		if jd.IsGlobal() {
			jd.PPose.Move(pos)
		}
	}
}

// RotateAround rotates around a given point
func (ob *Object) RotateAround(rot math32.Quat, around math32.Vector3) {
	for _, bd := range ob.Bodies {
		bd.Pose.RotateAround(rot, around)
	}
	for _, jd := range ob.Joints {
		if jd.IsGlobal() {
			jd.PPose.RotateAround(rot, around)
		}
	}
}

// RotateAroundBody rotates around a given body in object.
func (ob *Object) RotateAroundBody(body int, rot math32.Quat) {
	bd := ob.Body(body)
	ob.RotateAround(rot, bd.Pose.Pos)
}

// MoveOnAxis moves (translates) the specified distance on the
// specified local axis, relative to the given body in object.
// The axis is normalized prior to aplying the distance factor.
func (ob *Object) MoveOnAxisBody(body int, x, y, z, dist float32) {
	bd := ob.Body(body)
	delta := bd.Pose.Quat.MulVector(math32.Vec3(x, y, z).Normal()).MulScalar(dist)
	ob.Move(delta)
}

// RotateOnAxisBody rotates around the specified local axis the
// specified angle in degrees, relative to the given body in the object.
func (ob *Object) RotateOnAxisBody(body int, x, y, z, angle float32) {
	rot := math32.NewQuatAxisAngle(math32.Vec3(x, y, z), math32.DegToRad(angle))
	ob.RotateAroundBody(body, rot)
}

//////// Sensors

// NewSensor adds a new sensor function for this object.
// The closure function can capture local variables at the time
// of configuration, and write results wherever and however it is useful.
func (ob *Object) NewSensor(fun func(obj *Object)) {
	ob.Sensors = append(ob.Sensors, fun)
}

// RunSensors runs the sensor functions for this object.
func (ob *Object) RunSensors() {
	for _, sf := range ob.Sensors {
		sf(ob)
	}
}
