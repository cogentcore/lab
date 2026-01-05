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
	ShapeA physics.Shapes
	ShapeB physics.Shapes

	SizeA float32
	SizeB float32

	MassA float32
	MassB float32

	ZposA float32
	ZposB float32
}

func (cl *Collide) Defaults() {
	cl.ShapeA = physics.Sphere
	cl.ShapeB = physics.Sphere
	cl.SizeA = 0.5
	cl.SizeB = 0.5
	cl.MassA = 1
	cl.MassB = 1
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
		sc := ed.Scene
		rot := math32.NewQuatIdentity()
		sc.NewBody(ml, "floor", physics.Plane, "#D0D0D080", math32.Vec3(0, 0, 0), math32.Vec3(0, 0, 0), rot)

		sc.NewDynamic(ml, "A", cl.ShapeA, "blue", cl.MassA, math32.Vec3(cl.SizeA, cl.SizeA, cl.SizeA), math32.Vec3(-5, cl.SizeA, cl.ZposA), rot)
		sc.NewDynamic(ml, "B", cl.ShapeB, "red", cl.MassB, math32.Vec3(cl.SizeB, cl.SizeB, cl.SizeB), math32.Vec3(5, cl.SizeB, cl.ZposB), rot)

		push := sc.NewDynamic(ml, "push", physics.Box, "grey", 1.0, math32.Vec3(.1, 2, 2), math32.Vec3(-8, 2, 0), rot)
		ml.NewObject()
		sc.NewJointPrismatic(ml, nil, push, math32.Vec3(-8, 0, 0), math32.Vec3(0, -2, 0), math32.Vec3(1, 0, 0))
	})

	ed.SetControlFunc(func(timeStep int) {
		physics.SetJointTargetPos(0, 0, pos, 100)
		physics.SetJointTargetVel(0, 0, 0, 20)
	})

	b.RunMainWindow()
}
