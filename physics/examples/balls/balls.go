// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package balls

//go:generate core generate -add-types

import (
	"math/rand/v2"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
	_ "cogentcore.org/lab/gosl/slbool/slboolcore" // include to get gui views
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/phyxyz"
)

// Balls has sim params
type Balls struct {

	// Number of balls: if collide, then run out of memory above 1000 or so
	NBalls int

	// Collide is whether the balls collide with each other
	Collide bool

	// Size of each ball (m)
	Size float32

	// Mass of each ball (kg)
	Mass float32

	// size of the box (m)
	Width  float32
	Depth  float32
	Height float32
	Thick  float32

	Bounce          float32
	Friction        float32
	FrictionTortion float32
	FrictionRolling float32
}

func (b *Balls) Defaults() {
	b.NBalls = 1000
	b.Collide = true
	b.Size = 0.2
	b.Mass = 0.1

	b.Width = 50
	b.Depth = 50
	b.Height = 20
	b.Thick = .1

	b.Bounce = 0.5
	b.Friction = 0
	b.FrictionTortion = 0
	b.FrictionRolling = 0
}

func Config(b tree.Node) {
	ed := phyxyz.NewEditor(b)

	bs := &Balls{}
	bs.Defaults()
	ed.CameraPos = math32.Vec3(0, bs.Width, bs.Width)

	ed.SetUserParams(bs)

	ed.SetConfigFunc(func() {
		ml := ed.Model
		ml.Params[0].SubSteps = 100
		ml.Params[0].Dt = 0.001
		// ml.GPU = false
		// ml.ReportTotalKE = true
		sc := ed.Scene
		rot := math32.NewQuatIdentity()
		sc.NewBody(ml, "floor", physics.Plane, "#D0D0D080", math32.Vec3(0, 0, 0),
			math32.Vec3(0, 0, 0), rot)

		hw := bs.Width / 2
		hd := bs.Depth / 2
		hh := bs.Height / 2
		ht := bs.Thick / 2
		sc.NewBody(ml, "back-wall", physics.Box, "#0000FFA0", math32.Vec3(hw, hh, ht),
			math32.Vec3(0, hh, -hd), rot)
		sc.NewBody(ml, "left-wall", physics.Box, "#FF0000A0", math32.Vec3(ht, hh, hd),
			math32.Vec3(-hw, hh, 0), rot)
		sc.NewBody(ml, "right-wall", physics.Box, "#00FF00A0", math32.Vec3(ht, hh, hd),
			math32.Vec3(hw, hh, 0), rot)
		sc.NewBody(ml, "front-wall", physics.Box, "#FFFF00A0", math32.Vec3(hw, hh, ht),
			math32.Vec3(0, hh, hd), rot)

		box := bs.Width * .9
		size := bs.Size
		for i := range bs.NBalls {
			ht := rand.Float32() * bs.Height
			x := rand.Float32()*box - 0.5*box
			z := rand.Float32()*box - 0.5*box
			clr := colors.Names[i%len(colors.Names)]
			bl := sc.NewDynamic(ml, "ball", physics.Sphere, clr, bs.Mass, math32.Vec3(size, size, size),
				math32.Vec3(x, size+ht, z), rot)
			if !bs.Collide {
				physics.SetBodyGroup(bl.BodyIndex, int32(i+1)) // only collide within same group
			}
			bl.SetBodyBounce(bs.Bounce)
			bl.SetBodyFriction(bs.Friction)
			bl.SetBodyFrictionTortion(bs.FrictionTortion)
			bl.SetBodyFrictionRolling(bs.FrictionRolling)
		}
	})
}
