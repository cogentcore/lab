// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate -add-types

import (
	"fmt"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	_ "cogentcore.org/lab/gosl/slbool/slboolcore" // include to get gui views
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/world"
)

// Pendula has sim params
type Pendula struct {

	// Number of bar elements to add to the pendulum. More interesting the more you add!
	NPendula int

	// StartVert starts the pendulum in the vertical orientation
	// (else horizontal, so it has somewhere to go). Need to add force if vertical.
	StartVert bool

	// TargetDegFromVert is the target number of degrees off of vertical
	// for each joint. Critical for this to not be 0 for StartVert.
	TargetDegFromVert int

	// Timestep in msec to add a force
	ForceOn int

	// Timestep in msec to stop adding force
	ForceOff int

	// Force to add
	Force float32

	// half-size of the pendulum elements.
	HSize math32.Vector3

	// Mass of each bar (kg)
	Mass float32

	// do the elements collide with each other?  this is currently bad!
	Collide bool

	// Stiff is the strength of the positional constraint to keep
	// each bar in a vertical position.
	Stiff float32

	// Damp is the strength of the velocity constraint to keep each
	// bar not moving.
	Damp float32
}

func (b *Pendula) Defaults() {
	b.NPendula = 2
	b.HSize.Set(0.05, .2, 0.05)
	b.Mass = 0.1

	b.ForceOn = 100
	b.ForceOff = 102
	b.Force = 0

	b.Damp = 0
	b.Stiff = 0
}

func main() {
	b := core.NewBody("test1").SetTitle("Physics Pendula")
	ed := world.NewEditor(b)

	ps := &Pendula{}
	ps.Defaults()

	ed.SetUserParams(ps)

	var botJoint int32

	ed.SetConfigFunc(func() {
		ph := ed.Physics
		wr := ed.World
		rot := math32.NewQuat(0, 0, 0, 1)
		rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)

		if ps.StartVert {
			rleft = rot
		}

		stY := 2*ps.HSize.Y*float32(ps.NPendula+1) + 1
		clr := colors.Names[0]
		x := -ps.HSize.Y
		y := stY
		if ps.StartVert {
			x = 0
			y -= ps.HSize.Y
		}
		pb := wr.NewDynamic(ph, "top", physics.Box, clr, ps.Mass, ps.HSize, math32.Vec3(x, y, 0), rleft)
		if !ps.Collide {
			physics.SetBodyGroup(pb.Index, int32(1))
		}

		targ := math32.DegToRad(float32(ps.TargetDegFromVert))

		ji := ph.NewJointRevolute(-1, pb.DynamicIndex, math32.Vec3(0, stY, 0), math32.Vec3(0, ps.HSize.Y, 0), math32.Vec3(0, 0, 1))
		physics.SetJointTargetPos(ji, 0, targ, ps.Stiff)
		physics.SetJointTargetVel(ji, 0, 0, ps.Damp)

		for i := 1; i < ps.NPendula; i++ {
			clr := colors.Names[i%len(colors.Names)]
			x := -float32(i)*ps.HSize.Y*2 - ps.HSize.Y
			y := stY
			if ps.StartVert {
				y = stY + x
				x = 0
			}
			cb := wr.NewDynamic(ph, "child", physics.Box, clr, ps.Mass, ps.HSize,
				math32.Vec3(x, y, 0), rleft)
			if !ps.Collide {
				physics.SetBodyGroup(cb.Index, int32(1+i))
			}
			ji = ph.NewJointRevolute(pb.DynamicIndex, cb.DynamicIndex, math32.Vec3(0, -ps.HSize.Y, 0), math32.Vec3(0, ps.HSize.Y, 0), math32.Vec3(0, 0, 1))
			physics.SetJointTargetPos(ji, 0, targ, ps.Stiff)
			physics.SetJointTargetVel(ji, 0, 0, ps.Damp)
			pb = cb
			botJoint = ji
		}
	})

	ed.SetControlFunc(func(timeStep int) {
		if timeStep >= ps.ForceOn && timeStep < ps.ForceOff {
			fmt.Println(timeStep, "\tforce on:", ps.Force)
			physics.SetJointControlForce(botJoint, 0, ps.Force)
		} else {
			physics.SetJointControlForce(botJoint, 0, 0)
		}
	})

	b.RunMainWindow()
}
