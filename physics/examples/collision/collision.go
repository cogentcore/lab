// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate -add-types

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	_ "cogentcore.org/lab/gosl/slbool/slboolcore" // include to get gui views
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/phyxyz"
)

// Collide has sim params
type Collide struct {
	// Shape of left body
	ShapeA physics.Shapes

	// Shape of right body
	ShapeB physics.Shapes

	// Size of left body (radius, capsule, cylinder, box are 2x taller)
	SizeA float32

	// Size of right body (radius, capsule, cylinder, box are 2x taller)
	SizeB float32

	// Mass of left object: if lighter than B, it will bounce back more.
	MassA float32

	// Mass of right object: if lighter than B, it will move faster.
	MassB float32

	// Z (depth) position: offset to get different collision angles.
	ZposA float32

	// Z (depth) position: offset to get different collision angles.
	ZposB float32

	// Mass of the pusher panel: if lighter, it transfers less energy.
	PushMass float32

	// Friction is for sliding: around 0.01 seems pretty realistic
	Friction float32

	// FrictionTortion is for rotating. Not generally relevant here.
	FrictionTortion float32

	// FrictionRolling is for rolling: around 0.01 seems pretty realistic
	FrictionRolling float32
}

func (cl *Collide) Defaults() {
	cl.ShapeA = physics.Sphere
	cl.ShapeB = physics.Sphere
	cl.SizeA = 0.5
	cl.SizeB = 0.5
	cl.MassA = 1
	cl.MassB = 1
	cl.PushMass = 1
	cl.Friction = 0.01
	cl.FrictionTortion = 0.01
	cl.FrictionRolling = 0.01
}

func main() {
	b := core.NewBody("collide").SetTitle("Physics Collide")
	ed := phyxyz.NewEditor(b)
	ed.CameraPos = math32.Vec3(0, 20, 20)

	cl := &Collide{}
	cl.Defaults()

	ed.SetUserParams(cl)

	core.NewText(b).SetText("Pusher target position:")
	pos := float32(3)
	sld := core.NewSlider(b).SetMin(0).SetMax(5).SetStep(.1).SetEnforceStep(true)
	core.Bind(&pos, sld)

	ed.SetConfigFunc(func() {
		ml := ed.Model
		ml.GPU = false
		// ml.ReportTotalKE = true
		sc := ed.Scene
		rot := math32.NewQuatIdentity()
		fl := sc.NewBody(ml, "floor", physics.Plane, "#D0D0D080", math32.Vec3(0, 0, 0), math32.Vec3(0, 0, 0), rot)
		physics.SetBodyFriction(fl.BodyIndex, cl.Friction)
		physics.SetBodyFrictionRolling(fl.BodyIndex, cl.FrictionRolling)
		physics.SetBodyFrictionTortion(fl.BodyIndex, cl.FrictionTortion)

		hhA := 2 * cl.SizeA
		hhB := 2 * cl.SizeB
		if cl.ShapeA == physics.Sphere {
			hhA = cl.SizeA
		}
		if cl.ShapeB == physics.Sphere {
			hhB = cl.SizeB
		}

		ba := sc.NewDynamic(ml, "A", cl.ShapeA, "blue", cl.MassA, math32.Vec3(cl.SizeA, 2*cl.SizeA, cl.SizeA), math32.Vec3(-5, hhA, cl.ZposA), rot)
		physics.SetBodyFriction(ba.BodyIndex, cl.Friction)
		physics.SetBodyFrictionRolling(ba.BodyIndex, cl.FrictionRolling)
		physics.SetBodyFrictionTortion(ba.BodyIndex, cl.FrictionTortion)

		bb := sc.NewDynamic(ml, "B", cl.ShapeB, "red", cl.MassB, math32.Vec3(cl.SizeB, 2*cl.SizeB, cl.SizeB), math32.Vec3(0, hhB, cl.ZposB), rot)
		physics.SetBodyFriction(bb.BodyIndex, cl.Friction)
		physics.SetBodyFrictionRolling(bb.BodyIndex, cl.FrictionRolling)
		physics.SetBodyFrictionTortion(bb.BodyIndex, cl.FrictionTortion)

		push := sc.NewDynamic(ml, "push", physics.Box, "grey", cl.PushMass, math32.Vec3(.1, 2, 2), math32.Vec3(-8, 2, 0), rot)
		ml.NewObject()
		sc.NewJointPrismatic(ml, nil, push, math32.Vec3(-8, 0, 0), math32.Vec3(0, -2, 0), math32.Vec3(1, 0, 0))
	})

	ed.SetControlFunc(func(timeStep int) {
		physics.SetJointTargetPos(0, 0, pos, 100)
		physics.SetJointTargetVel(0, 0, 0, 20)
	})

	b.RunMainWindow()
}
