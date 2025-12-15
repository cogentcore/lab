// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"cogentcore.org/lab/gosl/slbool"
	"cogentcore.org/lab/gosl/slvec"
)

//gosl:start

// PhysParams are the physics parameters
type PhysParams struct {
	// DynamicsN is number of dynamics bodies.
	DynamicsN int32

	// JointsN is number of joints.
	JointsN int32

	// Iters is the number of iterations to perform.
	Iters int32 `default:"2"`

	// Dt is the integration stepsize.
	Dt float32 `default:"0.01"`

	// SoftRelax is soft-body relaxation constant.
	SoftRelax float32 `default:"0.9"`

	// JointLinearRelax is joint linear relaxation constant.
	JointLinearRelax float32 `default:"0.7"`

	// JointAngularRelax is joint angular relaxation constant.
	JointAngularRelax float32 `default:"0.4"`

	// JointLinearComply is joint linear compliance constant.
	JointLinearComply float32 `default:"0"`

	// JointAngularComply is joint angular compliance constant.
	JointAngularComply float32 `default:"0"`

	// ContactRelax is rigid contact relaxation constant.
	ContactRelax float32 `default:"0.8"`

	// AngularDamping is damping of angular motion.
	AngularDamping float32 `default:"0"`

	// Contact weighting
	ContactWeighting slbool.Bool `default:"true"`

	// Restitution
	Restitution slbool.Bool `default:"false"`

	pad, pad1, pad2 float32

	// Gravity is the gravity acceleration function
	Gravity slvec.Vector3
}

func (pr *PhysParams) Defaults() {
	pr.Iters = 2
	pr.Dt = 0.01
	pr.Gravity.Set(0, -9.81, 0)
	pr.SoftRelax = 0.9
	pr.JointLinearRelax = 0.7
	pr.JointAngularRelax = 0.4
	pr.JointLinearComply = 0
	pr.JointAngularComply = 0
	pr.ContactRelax = 0.8
	pr.AngularDamping = 0
	pr.ContactWeighting.SetBool(true)
	pr.Restitution.SetBool(false)
}

//gosl:end
