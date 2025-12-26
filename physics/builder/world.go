// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

// World is one world within the Builder.
type World struct {
	// World is the world index.
	World int

	// Objects are the objects within the world.
	// Each object is a coherent collection of bodies, typically
	// connected by joints. This is an organizational convenience
	// for positioning elements as well.
	Objects []Object
}
