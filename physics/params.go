// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/gosl/slbool"
	"cogentcore.org/lab/gosl/slvec"
)

//gosl:start

// PhysParams are the physics parameters
type PhysParams struct {
	// Iterations is the number of integration iterations to perform
	// within each solver step. Muller et al (2020) report that 1 is best.
	Iterations int32 `default:"1"`

	// Dt is the integration stepsize.
	// For highly kinetic situations (e.g., rapidly moving bouncing balls)
	// 0.0001 is needed to ensure contact registration. Use SubSteps to
	// accomplish a target effective read-out step size.
	Dt float32 `default:"0.0001"`

	// SubSteps is the number of integration steps to take per Step()
	// function call. These sub steps are taken without any sync to/from
	// the GPU and are therefore much faster.
	SubSteps int32 `default:"10" min:"1"`

	// Contact margin is the extra distance for broadphase collision
	// around rigid bodies. This can make some joints potentially unstable if > 0
	ContactMargin float32 `default:"0,0.1"`

	// ContactRelax is rigid contact relaxation constant.
	// Higher values cause errros
	ContactRelax float32 `default:"0.8"` // 0.8 def

	// Contact weighting: balances contact forces?
	ContactWeighting slbool.Bool `default:"true"` // true

	// Restitution takes into account bounciness of objects.
	Restitution slbool.Bool `default:"false"` // false

	// JointLinearRelax is joint linear relaxation constant.
	JointLinearRelax float32 `default:"0.7"` // 0.7 def

	// JointAngularRelax is joint angular relaxation constant.
	JointAngularRelax float32 `default:"0.4"` // 0.4 def

	// JointLinearComply is joint linear compliance constant.
	JointLinearComply float32 `default:"0"` // 0 def

	// JointAngularComply is joint angular compliance constant.
	JointAngularComply float32 `default:"0"` // 0 def

	// AngularDamping is damping of angular motion.
	AngularDamping float32 `default:"0"` // 0 def

	// SoftRelax is soft-body relaxation constant.
	SoftRelax float32 `default:"0.9"`

	// MaxGeomIter is number of iterations to perform in shape-based
	// geometry collision computations
	MaxGeomIter int32 `default:"10"`

	// Maximum number of contacts to process at any given point.
	ContactsMax int32 `edit:"-"`

	// Index for the current state (0 or 1, alternates with Next).
	Cur int32 `edit:"-"`

	// Index for the next state (1 or 0, alternates with Cur).
	Next int32 `edit:"-"`

	// BodiesN is number of rigid bodies.
	BodiesN int32 `edit:"-"`

	// DynamicsN is number of dynamics bodies.
	DynamicsN int32 `edit:"-"`

	// JointsN is number of joints.
	JointsN int32 `edit:"-"`

	// JointDoFsN is number of joint DoFs.
	JointDoFsN int32 `edit:"-"`

	// BodyJointsMax is max number of joints per body + 1 for actual n.
	BodyJointsMax int32 `edit:"-"`

	// BodyCollidePairsN is the total number of pre-compiled collision pairs
	// to examine.
	BodyCollidePairsN int32 `edit:"-"`

	pad int32

	// Gravity is the gravity acceleration function
	Gravity slvec.Vector3
}

func (pr *PhysParams) Defaults() {
	pr.Iterations = 1
	pr.Dt = 0.0001
	pr.SubSteps = 10
	pr.Gravity.Set(0, -9.81, 0)

	pr.ContactMargin = 0
	pr.ContactRelax = 0.8
	pr.ContactWeighting.SetBool(true)
	pr.Restitution.SetBool(false)

	pr.JointLinearRelax = 0.7
	pr.JointAngularRelax = 0.4
	pr.JointLinearComply = 0
	pr.JointAngularComply = 0

	pr.AngularDamping = 0
	pr.SoftRelax = 0.9
	pr.MaxGeomIter = 10
}

// StepsToMsec returns the given number of individual Step calls
// converted into milliseconds, suitable for driving controls.
func (pr *PhysParams) StepsToMsec(steps int) int {
	msper := 1000 * pr.Dt * float32(pr.SubSteps)
	return int(math32.Round(float32(steps) * msper))
}

// StepsToMsec returns the given number of individual Step calls
// converted into milliseconds, suitable for driving controls,
// Using the currently-set Params.
func StepsToMsec(steps int) int {
	return GetParams(0).StepsToMsec(steps)
}

//gosl:end
