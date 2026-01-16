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
	Worlds []*World

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
	return bl.Worlds[idx]
}

// NewGlobalWorld creates a new world with World index = -1,
// which are globals that collide with all worlds.
func (bl *Builder) NewGlobalWorld() *World {
	idx := len(bl.Worlds)
	bl.Worlds = append(bl.Worlds, &World{World: -1})
	return bl.Worlds[idx]
}

// NewWorld creates a new standard (non-global) world, with
// world index = index of last one + 1.
func (bl *Builder) NewWorld() *World {
	wn := 0
	idx := len(bl.Worlds)
	if idx > 0 {
		wn = bl.Worlds[idx-1].World + 1
	}
	bl.Worlds = append(bl.Worlds, &World{World: wn, WorldIndex: idx})
	return bl.Worlds[idx]
}

// Build builds a physics model, with optional [phyxyz.Scene] for
// visualization (using Skin elements created for bodies).
func (bl *Builder) Build(ml *physics.Model, sc *phyxyz.Scene) {
	bSt := int32(-1)
	bN := int32(0)
	jSt := int32(-1)
	jN := int32(0)
	for wi, wl := range bl.Worlds {
		// fmt.Println("\n######## World:", wl.World)
		for _, ob := range wl.Objects {
			// fmt.Println("\n\t#### Object")
			for _, bd := range ob.Bodies {
				bd.NewPhysicsBody(ml, wl.World)
				if bl.ReplicasN > 0 && wi == bl.ReplicasStart {
					bN++
					if bSt < 0 {
						bSt = bd.BodyIndex
					}
				}
			}
			if len(ob.Joints) == 0 {
				continue
			}
			ml.NewObject()
			for _, jd := range ob.Joints {
				jd.NewPhysicsJoint(ml, ob)
				if bl.ReplicasN > 0 && wi == bl.ReplicasStart {
					jN++
					if jSt < 0 {
						jSt = jd.JointIndex
					}
				}
			}
		}
	}
	if bN > 0 {
		ml.ReplicasN = int32(bl.ReplicasN)
		ml.ReplicaBodiesStart = bSt
		ml.ReplicaBodiesN = bN
		ml.ReplicaJointsStart = jSt
		ml.ReplicaJointsN = jN
	}
}

// InitState initializes the current state variables in the builder.
// This does not call InitState in physics, because that depends on
// whether the Sccene is being used.
func (bl *Builder) InitState() {
	for _, wl := range bl.Worlds {
		for _, ob := range wl.Objects {
			ob.InitState()
		}
	}
}

// RunSensors runs the sensor functions for this Builder.
func (bl *Builder) RunSensors() {
	for _, wl := range bl.Worlds {
		wl.RunSensors()
	}
}

// ReplicateWorld makes copies of given world to form an X,Y grid of
// worlds with given optional offsets (Y, X) added between world objects.
// Note that worldIdx is the index in Worlds, not the world number.
// Because different worlds do not interact, offsets are not necessary
// and can potentially affect numerical accuracy.
// If the given [phyxyz.Scene] is non-nil, then new skins will be made
// for the replicated bodies. Otherwise, the [phyxyz.Scene] can view
// different replicas.
func (bl *Builder) ReplicateWorld(sc *phyxyz.Scene, worldIdx, nY, nX int, offs ...math32.Vector3) {
	src := bl.World(worldIdx)
	var Yoff, Xoff math32.Vector3
	if len(offs) > 0 {
		Yoff = offs[0]
	}
	if len(offs) > 1 {
		Xoff = offs[1]
	}

	for y := range nY {
		for x := range nX {
			if x == 0 && y == 0 {
				continue
			}
			nw := bl.NewWorld()
			wi := nw.WorldIndex
			nw.Copy(src)
			nw.SetWorldIndex(wi)
			off := Yoff.MulScalar(float32(y)).Add(Xoff.MulScalar(float32(x)))
			nw.Move(off)
			if sc != nil {
				nw.CopySkins(sc, src)
			}
		}
	}
	bl.ReplicasStart = worldIdx
	bl.ReplicasN = nY * nX
}

// CloneSkins copies existing Body skins into the given [phyxyz.Scene],
// thereby configuring the given scene to view the physics model for this builder.
func (bl *Builder) CloneSkins(sc *phyxyz.Scene) {
	for _, wl := range bl.Worlds {
		for _, ob := range wl.Objects {
			for _, bd := range ob.Bodies {
				if bd.Skin == nil {
					continue
				}
				sc.AddSkinClone(bd.Skin)
			}
		}
	}
}

// ReplicaWorld returns the replica World at given replica index,
// Where replica is index into replicated worlds (0 = original).
func (bl *Builder) ReplicaWorld(replica int) *World {
	return bl.Worlds[bl.ReplicasStart+replica]
}

// ReplicaObject returns the replica corresponding to given [Object],
// Where replica is index into replicated worlds (0 = original).
func (bl *Builder) ReplicaObject(ob *Object, replica int) *Object {
	wl := bl.ReplicaWorld(replica)
	return wl.Object(ob.Object)
}

// ReplicaBody returns the replica corresponding to given [Body],
// Where replica is index into replicated worlds (0 = original).
func (bl *Builder) ReplicaBody(bd *Body, replica int) *Body {
	wl := bl.ReplicaWorld(replica)
	return wl.Object(bd.Object).Body(bd.ObjectBody)
}

// ReplicaJoint returns the replica corresponding to given [Joint],
// Where replica is index into replicated worlds (0 = original).
func (bl *Builder) ReplicaJoint(bd *Joint, replica int) *Joint {
	wl := bl.ReplicaWorld(replica)
	return wl.Object(bd.Object).Joint(bd.ObjectJoint)
}
