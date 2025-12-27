// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

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
