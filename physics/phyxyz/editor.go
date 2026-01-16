// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phyxyz

import (
	"fmt"
	"time"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/core/xyz/xyzcore"
	_ "cogentcore.org/lab/gosl/slbool/slboolcore" // include to get gui views
	"cogentcore.org/lab/physics"
)

// Editor provides a basic viewer and parameter controller widget
// for exploring physics models. It creates and manages its own
// [physics.Model] and [phyxyz.Scene].
type Editor struct { //types:add
	core.Frame

	// Model has the physics simulation.
	Model *physics.Model

	// Scene has the 3D GUI visualization.
	Scene *Scene

	// UserParams is a struct with parameters for configuring the physics sim.
	// These are displayed in the editor.
	UserParams any

	// ConfigFunc is the function that configures the [physics.Model].
	ConfigFunc func()

	// ControlFunc is the function that sets control parameters,
	// based on the current timestep (in milliseconds, converted from physics time).
	ControlFunc func(timeStep int)

	// CameraPos provides the default initial camera position, looking at the origin.
	// Set this to larger numbers to zoom out, and smaller numbers to zoom in.
	// Defaults to math32.Vec3(0, 25, 20).
	CameraPos math32.Vector3

	// Replica is the replica world to view, if replicas are present in model.
	Replica int

	// IsRunning is true if currently running sim.
	isRunning bool

	// Stop triggers topping of running.
	stop bool

	// TimeStep is current time step in physics update cycles.
	TimeStep int

	// editor is the xyz GUI visualization widget.
	editor *xyzcore.SceneEditor

	// Toolbar is the top toolbar.
	toolbar *core.Toolbar

	// Splits is the container for elements.
	splits *core.Splits

	// UserParamsForm has the user's config parameters.
	userParamsForm *core.Form

	// ParamsForm has the Physics parameters.
	paramsForm *core.Form
}

func (pe *Editor) CopyFieldsFrom(frm tree.Node) {
	fr := frm.(*Editor)
	pe.Frame.CopyFieldsFrom(&fr.Frame)
}

func (pe *Editor) Init() {
	pe.Frame.Init()
	pe.CameraPos = math32.Vec3(0, 25, 20)

	pe.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
		s.Direction = styles.Column
	})

	tree.AddChildAt(pe, "tb", func(w *core.Toolbar) {
		pe.toolbar = w
		w.Maker(pe.MakeToolbar)
	})

	tree.AddChildAt(pe, "splits", func(w *core.Splits) {
		pe.splits = w
		pe.splits.SetSplits(0.2, 0.8)
		tree.AddChildAt(w, "forms", func(w *core.Frame) {
			w.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				s.Grow.Set(1, 1)
			})
			tree.AddChildAt(w, "users", func(w *core.Form) {
				pe.userParamsForm = w
			})
			tree.AddChildAt(w, "params", func(w *core.Form) {
				pe.paramsForm = w
				if pe.UserParams != nil {
					pe.userParamsForm.SetStruct(pe.UserParams)
				}
				params := &pe.Model.Params[0]
				pe.paramsForm.SetStruct(params)
			})
		})

		tree.AddChildAt(w, "scene", func(w *xyzcore.SceneEditor) {
			pe.editor = w
			w.UpdateWidget()
			sc := pe.editor.SceneXYZ()

			sc.Background = colors.Scheme.Select.Container
			xyz.NewAmbient(sc, "ambient", 0.3, xyz.DirectSun)

			dir := xyz.NewDirectional(sc, "dir", 1, xyz.DirectSun)
			dir.Pos.Set(0, 2, 1)

			pe.Scene = NewScene(sc)
			pe.Model = physics.NewModel()

			sc.Camera.Pose.Pos = math32.Vec3(0, 40, 3.5)
			sc.Camera.LookAt(math32.Vec3(0, 5, 0), math32.Vec3(0, 1, 0))
			sc.SaveCamera("3")

			sc.Camera.Pose.Pos = math32.Vec3(-1.33, 2.24, 3.55)
			sc.Camera.LookAt(math32.Vec3(0, .5, 0), math32.Vec3(0, 1, 0))
			sc.SaveCamera("2")

			sc.Camera.Pose.Pos = pe.CameraPos
			sc.Camera.LookAt(math32.Vec3(0, 0, 0), math32.Vec3(0, 1, 0))
			sc.SaveCamera("1")
			sc.SaveCamera("default")

			pe.ConfigModel()
		})
	})
}

// ConfigModel configures the physics model.
func (pe *Editor) ConfigModel() {
	if pe.isRunning {
		core.MessageSnackbar(pe, "Simulation is still running...")
		return
	}
	pe.Scene.Reset()
	pe.Model.Reset()
	if pe.ConfigFunc != nil {
		pe.ConfigFunc()
	}
	pe.Scene.Init(pe.Model)
	pe.stop = false
	pe.TimeStep = 0
	pe.editor.NeedsRender()
}

// Restart restarts the simulation, returning true if successful (i.e., not running).
func (pe *Editor) Restart() bool {
	if pe.isRunning {
		core.MessageSnackbar(pe, "Simulation is still running...")
		return false
	}
	pe.stop = false
	pe.TimeStep = 0
	pe.Scene.InitState(pe.Model)
	pe.editor.NeedsRender()
	return true
}

// Step steps the world n times, with updates. Must be called as a goroutine.
func (pe *Editor) Step(n int) {
	if pe.isRunning {
		return
	}
	pe.isRunning = true
	pe.Model.SetAsCurrent()
	pe.toolbar.AsyncLock()
	pe.toolbar.UpdateRender()
	pe.toolbar.AsyncUnlock()
	for range n {
		if pe.ControlFunc != nil {
			pe.ControlFunc(physics.StepsToMsec(pe.TimeStep))
		}
		pe.Model.Step()
		pe.TimeStep++
		pe.Scene.Update()
		pe.editor.AsyncLock()
		pe.editor.NeedsRender()
		pe.editor.AsyncUnlock()
		if !pe.Model.GPU {
			time.Sleep(time.Nanosecond) // this is essential for web (wasm) running to actually update
			// if running in GPU mode, it works, but otherwise the thread never yields and it never updates.
		}
		if pe.stop {
			pe.stop = false
			break
		}
	}
	pe.isRunning = false
	pe.AsyncLock()
	pe.Update()
	pe.AsyncUnlock()
}

func (pe *Editor) MakeToolbar(p *tree.Plan) {
	stepNButton := func(n int) {
		nm := fmt.Sprintf("Step %d", n)
		tree.AddAt(p, nm, func(w *core.Button) {
			w.FirstStyler(func(s *styles.Style) { s.SetEnabled(!pe.isRunning) })
			w.SetText(nm).SetIcon(icons.PlayArrow).
				SetTooltip(fmt.Sprintf("Step state %d times", n)).
				OnClick(func(e events.Event) {
					if pe.isRunning {
						fmt.Println("still running...")
						return
					}
					go pe.Step(n)
				})
			w.Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
		})
	}

	tree.Add(p, func(w *core.Button) {
		w.SetText("Restart").SetIcon(icons.Reset).
			SetTooltip("Reset physics state back to starting.").
			OnClick(func(e events.Event) {
				pe.Restart()
			})
		w.FirstStyler(func(s *styles.Style) { s.SetEnabled(!pe.isRunning) })
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Stop").SetIcon(icons.Stop).
			SetTooltip("Stop running").
			OnClick(func(e events.Event) {
				pe.stop = true
			})
		w.FirstStyler(func(s *styles.Style) { s.SetEnabled(pe.isRunning) })
	})
	tree.Add(p, func(w *core.Separator) {})

	stepNButton(1)
	stepNButton(10)
	stepNButton(100)
	stepNButton(1000)
	stepNButton(10000)

	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Button) {
		w.SetText("Rebuild").SetIcon(icons.Reset).
			SetTooltip("Rebuild the environment, when you change parameters").
			OnClick(func(e events.Event) {
				pe.ConfigModel()
			})
		w.FirstStyler(func(s *styles.Style) { s.SetEnabled(!pe.isRunning) })
	})

	tree.Add(p, func(w *core.Separator) {})

	tt := "Replica world to view"
	tree.Add(p, func(w *core.Text) { w.SetText("Replica:").SetTooltip(tt) })

	tree.Add(p, func(w *core.Spinner) {
		core.Bind(&pe.Replica, w)
		w.SetMin(0).SetTooltip(tt)
		w.Styler(func(s *styles.Style) {
			replN := int32(0)
			if physics.CurModel != nil && pe.Scene != nil {
				replN = physics.CurModel.ReplicasN
				pe.Scene.ReplicasView = replN > 0
			}
			w.SetMax(float32(replN - 1))
			s.SetEnabled(replN > 1)
		})
		w.OnChange(func(e events.Event) {
			pe.Scene.ReplicasIndex = pe.Replica
			pe.Scene.Update()
			pe.NeedsRender()
		})
	})
}
