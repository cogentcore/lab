// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/physics/phyxyz"
)

// World is one world within the Builder.
type World struct {
	// World is the world index. -1 = globals, else positive.. are distinct
	// non-interacting worlds.
	World int

	// Objects are the objects within the [World].
	// Each object is a coherent collection of bodies, typically
	// connected by joints. This is an organizational convenience
	// for positioning elements; has no physical implications.
	Objects []Object
}

func (wl *World) Object(idx int) *Object {
	return &wl.Objects[idx]
}

func (wl *World) NewObject() *Object {
	idx := len(wl.Objects)
	wl.Objects = append(wl.Objects, Object{})
	return &wl.Objects[idx]
}

// Copy copies all objects from given source world into this one.
// (The worlds will be identical after, regardless of current starting
// condition).
func (wl *World) Copy(ow *World) {
	wl.Objects = make([]Object, len(ow.Objects))
	for i := range wl.Objects {
		wl.Object(i).Copy(ow.Object(i))
	}
}

// CopySkins makes new skins for bodies in world,
// based on those in source world, which must be a Copy.
func (wl *World) CopySkins(sc *phyxyz.Scene, ow *World) {
	for i := range wl.Objects {
		wl.Object(i).CopySkins(sc, ow.Object(i))
	}
}

// Transform applies positional and rotational transforms to all objects.
func (wl *World) Transform(pos math32.Vector3, rot math32.Quat) {
	for i := range wl.Objects {
		ob := wl.Object(i)
		ob.Transform(pos, rot)
	}
}
