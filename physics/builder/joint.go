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
}

// DoF is a degree-of-freedom for a [Joint].
type DoF struct {
	// Axis is the axis of articulation.
	Axis math32.Vector3

	// Limit has the limits for motion of this DoF.
	Limit minmax.F32
}
