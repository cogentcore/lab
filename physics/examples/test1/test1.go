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

// Test has sim params
type Test struct {
	// Height to start from
	Height float32

	// Size of test item
	Size float32

	// Mass of test item
	Mass float32

	Bounce          float32
	Friction        float32
	FrictionTortion float32
	FrictionRolling float32
}

func (b *Test) Defaults() {
	b.Height = 50
	b.Size = 0.5
	b.Mass = 0.1

	b.Bounce = 0.5
	b.Friction = 0
	b.FrictionTortion = 0
	b.FrictionRolling = 0
}

func main() {
	b := core.NewBody("test1").SetTitle("Physics Test1")
	ed := phyxyz.NewEditor(b)
	ed.CameraPos = math32.Vec3(0, 3, 3)

	bs := &Test{}
	bs.Defaults()

	ed.SetUserParams(bs)

	ed.SetConfigFunc(func() {
		ml := ed.Model
		ml.GPU = false
		sc := ed.Scene
		rot := math32.NewQuat(0, 0, 0, 1)
		// thick := float32(0.1)
		fl := sc.NewBody(ml, "floor", physics.Plane, "#D0D0D080", math32.Vec3(10, 0, 10),
			math32.Vec3(0, 0, 0), rot)
		fl.SetBodyBounce(bs.Bounce)

		height := float32(bs.Size)
		width := height * .4
		depth := height * .15
		_, _ = width, depth
		// b1 := wr.NewDynamic(ml, "body", physics.Box, "purple", 1.0, math32.Vec3(height, height, depth),
		// 	math32.Vec3(0, height*2, 0), rot)
		// b1 := sc.NewDynamic(ml, "body", physics.Sphere, "purple", 1.0, math32.Vec3(height, height, height), math32.Vec3(0, height*bs.Height, 0), rot)
		b1 := sc.NewDynamic(ml, "body", physics.Box, "purple", 1.0, math32.Vec3(width, height, depth), math32.Vec3(0, bs.Size, 0), rot)
		physics.SetBodyGroup(b1.BodyIndex, 0)

		// rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)
		// b2 := sc.NewDynamic(ml, "nose", physics.Capsule, "blue", .001, math32.Vec3(0.1*depth, 0.1*height, 0.1*depth), math32.Vec3(-depth, 2*bs.Size, 0), rleft)
		// b1.SetBodyBounce(bs.Bounce)
		_ = b1

		// bj := ml.NewJointRevolute(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0), math32.Vec3(0, 1, 0))
		// bj := ml.NewJointPrismatic(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0), math32.Vec3(1, 0, 0))
		// // // physics.SetJointControlForce(bj, 0, .1)
		// physics.SetJointTargetPos(bj, 0, 1, 1)
		// physics.SetJointTargetVel(bj, 0, 0, 1)
		// ml.NewJointFixed(b1.DynamicIndex, b2.DynamicIndex, math32.Vec3(-depth, bs.Size, 0), math32.Vec3(-0.1*depth, 0, 0))
		bj := ml.NewJointBall(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0))
		physics.SetJointTargetPos(bj, 0, 0, 1000)
		physics.SetJointTargetVel(bj, 0, 0, 20)
		physics.SetJointTargetPos(bj, 1, 0, 1000)
		physics.SetJointTargetVel(bj, 1, 0, 20)
		physics.SetJointTargetPos(bj, 2, 0, 1000)
		physics.SetJointTargetVel(bj, 2, 0, 20)
	})

	// ed.SetControlFunc(func(timeStep int) {
	// 	if timeStep >= ps.ForceOn && timeStep < ps.ForceOff {
	// 		fmt.Println(timeStep, "\tforce on:", ps.Force)
	// 		physics.SetJointControlForce(botJoint.JointIndex, 0, ps.Force)
	// 	} else {
	// 		physics.SetJointControlForce(botJoint.JointIndex, 0, 0)
	// 	}
	// })

	b.RunMainWindow()
}
