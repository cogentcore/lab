// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/phyxyz"
)

// Physics provides a container and manager for the main physics elements:
// [Builder], [physics.Model], and [phyxyz.Scene]. This is helpful for
// models used within other apps (e.g., an AI simulation), whereas
// [phyxyz.Editor] provides a standalone GUI interface for testing models.
type Physics struct {
	// Model has the physics Model.
	Model *physics.Model

	// Builder for configuring the Model.
	Builder *Builder

	// Scene for visualizing the Model
	Scene *phyxyz.Scene
}

// Build calls Builder.Build with Model and Scene args,
// and then Init on the Scene.
func (ph *Physics) Build() {
	ph.Builder.Build(ph.Model, ph.Scene)
	if ph.Scene != nil {
		ph.Scene.Init(ph.Model)
	}
}

// InitState calls Scene.InitState or Model.InitState and Builder InitState.
func (ph *Physics) InitState() {
	if ph.Scene != nil {
		ph.Scene.InitState(ph.Model)
	} else {
		ph.Model.InitState()
	}
	if ph.Builder != nil {
		ph.Builder.InitState()
	}
}

// Step advances the physics world n steps, updating the scene every time.
func (ph *Physics) Step(n int) {
	for range n {
		ph.Model.Step()
		ph.Scene.Update()
	}
}

// StepQuiet advances the physics world n steps.
func (ph *Physics) StepQuiet(n int) {
	for range n {
		ph.Model.Step()
	}
}
