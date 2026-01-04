// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import "cogentcore.org/lab/tensor"

// note: add -keep to inspect intermediate .go code
//go:generate gosl -exclude=Update,Defaults,ShouldDisplay -max-buffer-size=2147483616

// CurModel is the currently active [Model].
var CurModel *Model

//gosl:start

// vars are all the global vars for axon GPU / CPU computation.
//
//gosl:vars
var (
	// Params are global parameters.
	//gosl:group Params
	Params []PhysicsParams

	// Bodies are the rigid body elements (dynamic and static),
	// specifying the constant, non-dynamic properties,
	// which is initial state for dynamics.
	// [body][BodyVarsN]
	//gosl:group Bodies
	//gosl:dims 2
	Bodies *tensor.Float32

	// Objects is a list of joint indexes for each object, where each object
	// contains all the joints interconnecting an overlapping set of bodies.
	// This is known as an articulation in other physics software.
	// Joints must be added in parent -> child order within objects, as joints
	// are updated in sequential order within object. First element is n joints.
	// [object][MaxObjectJoints+1]
	//gosl:dims 2
	Objects *tensor.Int32

	// BodyJoints is a list of joint indexes for each dynamic body, for aggregating.
	// [dyn body][parent, child][maxjointsperbody]
	//gosl:dims 3
	BodyJoints *tensor.Int32

	// Joints is a list of permanent joints connecting bodies,
	// which do not change (no dynamic variables, except temps).
	// [joint][JointVars]
	//gosl:dims 2
	Joints *tensor.Float32

	// JointDoFs is a list of joint DoF parameters, allocated per joint.
	// [dof][JointDoFVars]
	//gosl:dims 2
	JointDoFs *tensor.Float32

	// BodyCollidePairs are pairs of Body indexes that could potentially collide
	// based on precomputed collision logic, using World, Group, and Joint indexes.
	// [BodyCollidePairsN][2]
	//gosl:dims 2
	BodyCollidePairs *tensor.Int32

	// Dynamics are the dynamic rigid body elements: these actually move.
	// Two alternating states are used: Params.Cur and Params.Next.
	// [dyn body][cur/next][DynamicVarsN]
	//gosl:group Dynamics
	//gosl:dims 3
	Dynamics *tensor.Float32

	// BroadContactsN has number of points of broad contact
	// between bodies. [1]
	//gosl:dims 1
	BroadContactsN *tensor.Int32

	// BroadContacts are the results of broad-phase contact processing,
	// establishing possible points of contact between bodies.
	// [ContactsMax][BroadContactVarsN]
	//gosl:dims 2
	BroadContacts *tensor.Float32

	// ContactsN has number of points of narrow (final) contact
	// between bodies. [1]
	//gosl:dims 1
	ContactsN *tensor.Int32

	// Contacts are the results of narrow-phase contact processing,
	// where only actual contacts with fully-specified values are present.
	// [ContactsMax][ContactVarsN]
	//gosl:dims 2
	Contacts *tensor.Float32

	// JointControls are dynamic joint control inputs, per joint DoF
	// (in correspondence with [JointDoFs]). This can be uploaded to the
	// GPU at every step.
	// [dof][JointControlVarsN]
	//gosl:group Controls
	//gosl:dims 2
	JointControls *tensor.Float32
)

//gosl:end
