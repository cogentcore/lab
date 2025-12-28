// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

//go:generate core generate -add-types -setters

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/phyxyz"
)

// Builder is the global container of [physics.Model] elements,
// organized into worlds that are independently updated.
type Builder struct {
	// Worlds are the independent world elements.
	Worlds []World

	// ReplicasStart is the starting Worlds index for replicated world bodies.
	// Set by ReplicateWorld, and used to set corresponding value in Model.
	ReplicasStart int

	// ReplicasN is the total number of replicated Worlds (including source).
	// Set by ReplicateWorld, and used to set corresponding value in Model.
	ReplicasN int
}

func NewBuilder() *Builder {
	return &Builder{}
}

// Reset starts over making a new model.
func (bl *Builder) Reset() {
	bl.Worlds = nil
}

func (bl *Builder) World(idx int) *World {
	return &bl.Worlds[idx]
}

// NewGlobalWorld creates a new world with World index = -1,
// which are globals that collide with all worlds.
func (bl *Builder) NewGlobalWorld() *World {
	idx := len(bl.Worlds)
	bl.Worlds = append(bl.Worlds, World{World: -1})
	return &bl.Worlds[idx]
}

// NewWorld creates a new standard (non-global) world, with
// world index = index of last one + 1.
func (bl *Builder) NewWorld() *World {
	wn := 0
	idx := len(bl.Worlds)
	if idx > 0 {
		wn = bl.Worlds[idx-1].World + 1
	}
	bl.Worlds = append(bl.Worlds, World{World: wn})
	return &bl.Worlds[idx]
}

// Build builds a physics model, with optional [phyxyz.Scene] for
// visualization (using Skin elements created for bodies).
func (bl *Builder) Build(ml *physics.Model, sc *phyxyz.Scene) {
	repSt := int32(0)
	repN := int32(0)
	for wi := range bl.Worlds {
		wl := bl.World(wi)
		// fmt.Println("\n######## World:", wl.World)
		for oi := range wl.Objects {
			ob := wl.Object(oi)
			// fmt.Println("\n\t#### Object")
			for bbi := range ob.Bodies {
				bd := ob.Body(bbi)
				bd.NewPhysicsBody(ml, wl.World)
				if bl.ReplicasN > 0 && wi == bl.ReplicasStart {
					repN++
					if bbi == 0 {
						repSt = bd.BodyIndex
					}
				}
			}
			for bji := range ob.Joints {
				jd := ob.Joint(bji)
				jd.NewPhysicsJoint(ml, ob)
			}
		}
	}
	if repN > 0 {
		ml.ReplicasStart = repSt
		ml.ReplicasN = repN
	}
}

// ReplicateWorld makes copies of given world to form an X,Y grid of
// worlds with given offsets added between world objects. Note that
// worldIdx is the index in Worlds, not the world number.
// Because different worlds do not interact, offsets are not necessary
// and can potentially affect numerical accuracy. Offsets can also be
// established purely in [phyxyz.Scene] viewing.
// If the given [phyxyz.Scene] is non-nil, then new skins will be made
// for the replicated bodies (else not).
func (bl *Builder) ReplicateWorld(sc *phyxyz.Scene, worldIdx, nY, nX int, Yoff, Xoff math32.Vector3) {
	rot := math32.NewQuat(0, 0, 0, 1)
	src := bl.World(worldIdx)
	for y := range nY {
		for x := range nX {
			if x == 0 && y == 0 {
				continue
			}
			nw := bl.NewWorld()
			wi := nw.World
			nw.Copy(src)
			nw.World = wi
			off := Yoff.MulScalar(float32(y)).Add(Xoff.MulScalar(float32(x)))
			nw.Transform(off, rot)
			if sc != nil {
				nw.CopySkins(sc, src)
			}
		}
	}
	if sc == nil {
		bl.ReplicasStart = worldIdx
		bl.ReplicasN = nY * nX
	}
}
