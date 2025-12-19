// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package world implements visualization of [physics] using [xyz]
// 3D graphics.
package world

//go:generate core generate -add-types

import (
	"image"

	"cogentcore.org/core/core"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/lab/physics"
)

// World displays a [physics.World] using a [xyz.Scene].
// One World can be used for multiple different [physics.World]s which
// is more efficient when running multiple in parallel.
// Initial construction of the physics and visualization happens here.
type World struct {
	// Scene is the [xyz.Scene] object for visualizing.
	Scene *xyz.Scene

	// Root is the root Group node in the Scene under which the world is rendered.
	Root *xyz.Group

	// Views are the view elements for each body in [physics.World].
	Views []*View
}

// NewWorld returns a new World for visualizing a [physics.World].
// with given [xyz.Scene], making a top-level Root group in the scene.
func NewWorld(sc *xyz.Scene) *World {
	rgp := xyz.NewGroup(sc)
	rgp.SetName("world")
	wr := &World{Scene: sc, Root: rgp}
	return wr
}

// Init configures the visual world based on Views,
// and calls Config on [physics.World].
// Call this _once_ after making all the new Views and Bodies.
// (will return if already called).
func (wr *World) Init(wl *physics.World) {
	wl.Config()
	if len(wr.Root.Makers.Normal) > 0 {
		return
	}
	wr.Root.Maker(func(p *tree.Plan) {
		for _, vw := range wr.Views {
			vw.Add(p)
		}
	})
}

// Update updates the xyz scene from current physics node state.
// (use physics.World.SetAsCurrent()).
func (wr *World) Update() {
	wr.UpdateFromPhysics()
	if wr.Scene != nil {
		wr.Scene.Update()
	}
}

// UpdateFromPhysics updates the World from currently active
// physics state (use physics.World.SetAsCurrent()).
func (wr *World) UpdateFromPhysics() {
	for _, vw := range wr.Views {
		vw.UpdateFromPhysics()
	}
}

// RenderFromView does an offscreen render using given [View]
// for the camera position and orientation, returning the render image.
// Current scene camera is saved and restored.
func (wr *World) RenderFromNode(vw *View, cam *Camera) image.Image {
	sc := wr.Scene
	camnm := "physics-view-rendernode-save"
	sc.SaveCamera(camnm)
	defer func() {
		sc.SetCamera(camnm)
		sc.UseMainFrame()
	}()

	sc.Camera.FOV = cam.FOV
	sc.Camera.Near = cam.Near
	sc.Camera.Far = cam.Far
	sc.Camera.Pose.Pos = vw.Pos
	sc.Camera.Pose.Quat = vw.Quat
	sc.Camera.Pose.Scale.Set(1, 1, 1)

	sc.UseAltFrame(cam.Size)
	return sc.RenderGrabImage()
}

// DepthImage returns the current rendered depth image
// func (vw *World) DepthImage() ([]float32, error) {
// 	return vw.Scene.DepthImage()
// }

// MakeStateToolbar returns a toolbar function for physics state updates,
// calling the given updt function after making the change.
func MakeStateToolbar(ps *physics.State, updt func()) func(p *tree.Plan) {
	return func(p *tree.Plan) {
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.SetEulerRotation).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.SetAxisRotation).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.RotateEuler).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.RotateOnAxis).SetAfterFunc(updt).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.EulerRotation).SetAfterFunc(updt).SetShowReturn(true).SetIcon(icons.Rotate90DegreesCcw)
		})
		tree.Add(p, func(w *core.Separator) {})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.MoveOnAxis).SetAfterFunc(updt).SetIcon(icons.MoveItem)
		})
		tree.Add(p, func(w *core.FuncButton) {
			w.SetFunc(ps.MoveOnAxisAbs).SetAfterFunc(updt).SetIcon(icons.MoveItem)
		})

	}
}
