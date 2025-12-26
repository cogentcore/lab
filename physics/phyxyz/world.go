// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package phyxyz implements visualization of [physics] using [xyz]
// 3D graphics.
package phyxyz

//go:generate core generate -add-types

import (
	"image"

	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/lab/physics"
)

// Scene displays a [physics.Model] using a [xyz.Scene].
// One Scene can be used for multiple different [physics.Model]s which
// is more efficient when running multiple in parallel.
// Initial construction of the physics and visualization happens here.
type Scene struct {
	// Scene is the [xyz.Scene] object for visualizing.
	Scene *xyz.Scene

	// Root is the root Group node in the Scene under which the world is rendered.
	Root *xyz.Group

	// Views are the view elements for each body in [physics.Model].
	Views []*View
}

// NewScene returns a new Scene for visualizing a [physics.Model].
// with given [xyz.Scene], making a top-level Root group in the scene.
func NewScene(sc *xyz.Scene) *Scene {
	rgp := xyz.NewGroup(sc)
	rgp.SetName("world")
	xysc := &Scene{Scene: sc, Root: rgp}
	return xysc
}

// Init configures the visual world based on Views,
// and calls Config on [physics.Model].
// Call this _once_ after making all the new Views and Bodies.
// (will return if already called).
func (sc *Scene) Init(ph *physics.Model) {
	ph.Config()
	if len(sc.Root.Makers.Normal) > 0 {
		return
	}
	sc.Root.Maker(func(p *tree.Plan) {
		for _, vw := range sc.Views {
			vw.Add(p)
		}
	})
}

// Reset resets any existing views, starting fresh for a new configuration.
func (sc *Scene) Reset() {
	sc.Views = nil
	if sc.Scene != nil {
		sc.Scene.Update()
	}
}

// Update updates the xyz scene from current physics node state.
// (use physics.Model.SetAsCurrent()).
func (sc *Scene) Update() {
	sc.UpdateFromPhysics()
	if sc.Scene != nil {
		sc.Scene.Update()
	}
}

// UpdateFromPhysics updates the Scene from currently active
// physics state (use physics.Model.SetAsCurrent()).
func (sc *Scene) UpdateFromPhysics() {
	for _, vw := range sc.Views {
		vw.UpdateFromPhysics()
	}
}

// RenderFromView does an offscreen render using given [View]
// for the camera position and orientation, returning the render image.
// Current scene camera is saved and restored.
func (sc *Scene) RenderFromNode(vw *View, cam *Camera) image.Image {
	xysc := sc.Scene
	camnm := "physics-view-rendernode-save"
	xysc.SaveCamera(camnm)
	defer func() {
		xysc.SetCamera(camnm)
		xysc.UseMainFrame()
	}()

	xysc.Camera.FOV = cam.FOV
	xysc.Camera.Near = cam.Near
	xysc.Camera.Far = cam.Far
	xysc.Camera.Pose.Pos = vw.Pos
	xysc.Camera.Pose.Quat = vw.Quat
	xysc.Camera.Pose.Scale.Set(1, 1, 1)

	xysc.UseAltFrame(cam.Size)
	return xysc.RenderGrabImage()
}

// DepthImage returns the current rendered depth image
// func (vw *Scene) DepthImage() ([]float32, error) {
// 	return vw.Scene.DepthImage()
// }
