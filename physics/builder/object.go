// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

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
