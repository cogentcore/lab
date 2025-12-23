// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate -add-types

import (
	"fmt"

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
	"cogentcore.org/lab/physics/world"
)

// Test has sim params
type Test struct {
	// Height to start from
	Height float32

	// Size of test item
	Size float32

	// Mass of test item
	Mass float32

	Bounce          float32
	Friction        float32
	FrictionTortion float32
	FrictionRolling float32
}

func (b *Test) Defaults() {
	b.Height = 50
	b.Size = 0.5
	b.Mass = 0.1

	b.Bounce = 0.5
	b.Friction = 0
	b.FrictionTortion = 0
	b.FrictionRolling = 0
}

func main() {
	// gpu.Debug = true
	b := core.NewBody("test1").SetTitle("Physics Test")
	split := core.NewSplits(b)
	fpanel := core.NewFrame(split)
	fpanel.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
	})

	bpf := core.NewForm(fpanel)
	wpf := core.NewForm(fpanel)

	tbvw := core.NewTabs(split)
	scfr, _ := tbvw.NewTab("3D View")
	split.SetSplits(0.2, 0.8)

	se := xyzcore.NewSceneEditor(scfr)
	se.UpdateWidget()
	sc := se.SceneXYZ()

	sc.Background = colors.Scheme.Select.Container
	xyz.NewAmbient(sc, "ambient", 0.3, xyz.DirectSun)

	dir := xyz.NewDirectional(sc, "dir", 1, xyz.DirectSun)
	dir.Pos.Set(0, 2, 1)

	wr := world.NewWorld(sc)

	bs := &Test{}
	bs.Defaults()

	wl := physics.NewWorld()
	wl.GPU = false

	params := physics.GetParams(0)
	params.Dt = 0.0001    // leaks balls > 0.0005
	params.SubSteps = 100 // major speedup by inner-stepping
	// params.Gravity.Y = 0
	params.ContactRelax = 0.2         // 0.1 seems most physical -- 0.2 getting a bit more ke?
	params.Restitution.SetBool(false) // not working!
	params.ContactMargin = 0          // 0 for restitution

	bpf.SetStruct(bs)
	wpf.SetStruct(params)

	config := func() {
		wr.Reset()
		wl.Reset()
		rot := math32.NewQuat(0, 0, 0, 1)
		// thick := float32(0.1)
		fl := wr.NewBody(wl, "floor", physics.Plane, "#D0D0D080", math32.Vec3(10, 0, 10),
			math32.Vec3(0, 0, 0), rot)
		fl.SetBodyBounce(bs.Bounce)

		height := float32(bs.Size)
		width := height * .4
		depth := height * .15
		_, _ = width, depth
		// b1 := wr.NewDynamic(wl, "body", physics.Box, "purple", 1.0, math32.Vec3(height, height, depth),
		// 	math32.Vec3(0, height*2, 0), rot)
		b1 := wr.NewDynamic(wl, "body", physics.Sphere, "purple", 1.0, math32.Vec3(height, height, height),
			math32.Vec3(0, height*bs.Height, 0), rot)
		// b1.SetBodyBounce(bs.Bounce)
		_ = b1

		// bj := wl.NewJointRevolute(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0), math32.Vec3(0, 1, 0))
		// bj := wl.NewJointPrismatic(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0), math32.Vec3(1, 0, 0))
		// // // physics.SetJointControlForce(bj, 0, .1)
		// physics.SetJointTargetPos(bj, 0, 1, 1)
		// physics.SetJointTargetVel(bj, 0, 0, 1)
		wr.Init(wl)
		wr.Update()
	}

	config()

	cycle := 0

	updateView := func() {
		bpf.Update()
		wpf.Update()
		if se.IsVisible() {
			se.NeedsRender()
		}
	}

	sc.Camera.Pose.Pos = math32.Vec3(0, 40, 3.5)
	sc.Camera.LookAt(math32.Vec3(0, 5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("3")

	sc.Camera.Pose.Pos = math32.Vec3(-1.33, 2.24, 3.55)
	sc.Camera.LookAt(math32.Vec3(0, .5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("2")

	sc.Camera.Pose.Pos = math32.Vec3(0, 20, 30)
	sc.Camera.LookAt(math32.Vec3(0, 5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("1")
	sc.SaveCamera("default")

	isRunning := false
	stop := false

	stepNButton := func(p *tree.Plan, n int) {
		nm := fmt.Sprintf("Step %d", n)
		tree.AddAt(p, nm, func(w *core.Button) {
			w.SetText(nm).SetIcon(icons.PlayArrow).
				SetTooltip(fmt.Sprintf("Step state %d times", n)).
				OnClick(func(e events.Event) {
					if isRunning {
						return
					}
					go func() {
						isRunning = true
						for range n {
							wl.Step()
							cycle++
							wr.Update()
							if se.IsVisible() {
								se.AsyncLock()
								se.NeedsRender()
								se.AsyncUnlock()
								// time.Sleep(1 * time.Nanosecond)
							}
							if stop {
								stop = false
								break
							}
						}
						isRunning = false
					}()
				})
			w.Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
		})
	}

	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(func(p *tree.Plan) {
			tree.Add(p, func(w *core.Button) {
				w.SetText("Init").SetIcon(icons.Reset).
					SetTooltip("Reset physics state back to starting.").
					OnClick(func(e events.Event) {
						if isRunning {
							return
						}
						wl.InitState()
						wr.Update()
						updateView()
					})
			})
			tree.Add(p, func(w *core.Button) {
				w.SetText("Stop").SetIcon(icons.Stop).
					SetTooltip("Stop running").
					OnClick(func(e events.Event) {
						stop = true
					})
			})
			tree.Add(p, func(w *core.Separator) {})

			stepNButton(p, 1)
			stepNButton(p, 10)
			stepNButton(p, 100)
			stepNButton(p, 1000)
			stepNButton(p, 10000)
			tree.Add(p, func(w *core.Separator) {})

			tree.Add(p, func(w *core.Button) {
				w.SetText("Rebuild").SetIcon(icons.Reset).
					SetTooltip("Rebuild the environment, when you change parameters").
					OnClick(func(e events.Event) {
						if isRunning {
							return
						}
						config()
						updateView()
					})
			})
		})
	})
	b.RunMainWindow()
}
