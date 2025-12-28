// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/physics/phyxyz"
)

// Object is an object within the [World].
// Each object is a coherent collection of bodies, typically
// connected by joints. This is an organizational convenience
// for positioning elements; has no physical implications.
type Object struct {
	// Bodies are the bodies in the object.
	Bodies []Body

	// Joints are joints connecting object bodies.
	// Joint indexes here refer strictly within bodies.
	Joints []Joint
}

func (ob *Object) Body(idx int) *Body {
	return &ob.Bodies[idx]
}

func (ob *Object) Joint(idx int) *Joint {
	return &ob.Joints[idx]
}

// Copy copies all bodies and joints from given source world into this one.
// (The objects will be identical after, regardless of current starting
// condition).
func (ob *Object) Copy(so *Object) {
	ob.Bodies = make([]Body, len(so.Bodies))
	ob.Joints = make([]Joint, len(so.Joints))
	copy(ob.Bodies, so.Bodies)
	copy(ob.Joints, so.Joints)
	for i := range ob.Bodies {
		bd := ob.Body(i)
		bd.Skin = nil
	}
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

// Transform applies positional and rotational transforms to all bodies,
// and world-anchored joints.
func (ob *Object) Transform(pos math32.Vector3, rot math32.Quat) {
	for i := range ob.Bodies {
		ob.Body(i).Pose.Transform(pos, rot)
	}
	for i := range ob.Joints {
		ob.Joint(i).Transform(pos, rot) // only for world-anchored joints
	}
}

// PoseToPhysics sets the current body poses to the physics current state.
// For Dynamic bodies, sets dynamic state. Also updates world-anchored joints.
func (ob *Object) PoseToPhysics() {
	for i := range ob.Bodies {
		ob.Body(i).PoseToPhysics()
	}
	for i := range ob.Joints {
		ob.Joint(i).PoseToPhysics()
	}
}
