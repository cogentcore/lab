// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import "cogentcore.org/lab/tensor"

//go:generate gosl -keep -exclude=Update,Defaults,ShouldDisplay -max-buffer-size=2147483616

//gosl:start

// vars are all the global vars for axon GPU / CPU computation.
//
//gosl:vars
var (
	// Params are global parameters.
	//gosl:group Params
	//gosl:read-only
	Params []PhysParams

	// Bodies are the rigid body elements (dynamic and static),
	// specifying the constant, non-dynamic properties,
	// which is initial state for dynamics.
	// [body][BodyVarsN]
	//gosl:group Bodies
	//gosl:dims 2
	Bodies *tensor.Float32

	// Joints is a list of permanent joints connecting bodies,
	// which do not change (no dynamic variables, except temps).
	// [joint][JointVars]
	//gosl:dims 2
	Joints *tensor.Float32

	// BodyJoints is a list of joint indexes for each dynamic body, for aggregating.
	// [dyn body][parent, child][maxjointsperbody]
	//gosl:dims 3
	BodyJoints *tensor.Int32

	// Dynamics are the dynamic rigid body elements: these actually move.
	// Two alternating states are used: Params.Cur and Params.Next.
	// [dyn body][cur/next][DynamicVarsN]
	//gosl:group Bodies
	//gosl:dims 3
	Dynamics *tensor.Float32

	// Contacts are points of contact between bodies.
	// [contact][ContactVarsN]
	//gosl:dims 2
	Contacts *tensor.Float32

	// JointControls are dynamic joint control inputs.
	// [joint][JointControlVarsN]
	//gosl:group Controls
	//gosl:dims 2
	JointControls *tensor.Float32
)
