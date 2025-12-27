// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

//go:generate core generate -add-types -setters

import (
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/phyxyz"
)

// Builder is the global container of [physics.Model] elements,
// organized into worlds that are independently updated.
type Builder struct {

	// Worlds are the independent world elements.
	Worlds []World
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
		wn = bl.Worlds[idx-1].World
	}
	bl.Worlds = append(bl.Worlds, World{World: wn})
	return &bl.Worlds[idx]
}

// Build builds a physics model, with optional [phyxyz.Scene] for
// visualization (using Skin elements created for bodies).
func (bl *Builder) Build(ml *physics.Model, sc *phyxyz.Scene) {
	for wi := range bl.Worlds {
		wl := bl.World(wi)
		for oi := range wl.Objects {
			ob := wl.Object(oi)
			for bbi := range ob.Bodies {
				bd := ob.Body(bbi)
				bd.NewPhysicsBody(ml, wl.World)
			}
			for bji := range ob.Joints {
				jd := ob.Joint(bji)
				jd.NewPhysicsJoint(ml, ob)
			}
		}
	}
}
