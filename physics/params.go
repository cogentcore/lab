// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import "cogentcore.org/core/math32"

//gosl:start

// PhysParams are the physics parameters
type PhysParams struct {
	// DynamicsN is number of dynamics bodies.
	DynamicsN int32

	// JointsN is number of joints.
	JointsN int32

	// Step is the global stepsize for numerical integration.
	Step float32 `default:"0.01"`

	// Gravity is the force of gravity.
	Gravity float32

	// GravityDir is the direction of Gravity.
	GravityDir math32.Vector4
}

func (pr *PhysParams) Defaults() {
	pr.Step = 0.01
	pr.GravityDir.Set(0, -1, 0, 0)
}

//gosl:end
