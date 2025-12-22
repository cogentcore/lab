// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate

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
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/world"
)

func main() {
	// gpu.Debug = true
	b := core.NewBody("test1").SetTitle("Physics Test")
	split := core.NewSplits(b)
	// tv := core.NewTree(core.NewFrame(split))
	fv := core.NewForm(split)
	tbvw := core.NewTabs(split)
	scfr, _ := tbvw.NewTab("3D View")

	se := xyzcore.NewSceneEditor(scfr)
	se.UpdateWidget()
	sc := se.SceneXYZ()

	sc.Background = colors.Scheme.Select.Container
	xyz.NewAmbient(sc, "ambient", 0.3, xyz.DirectSun)

	dir := xyz.NewDirectional(sc, "dir", 1, xyz.DirectSun)
	dir.Pos.Set(0, 2, 1)

	wr := world.NewWorld(sc)

	wl := physics.NewWorld()
	fv.SetStruct(wl)

	split.SetSplits(0.2, 0.8)

	rot := math32.NewQuat(0, 0, 0, 1)
	// thick := float32(0.1)
	wr.NewBody(wl, "floor", physics.Plane, "grey", math32.Vec3(10, 0, 10),
		math32.Vec3(0, 0, 0), rot)

	height := float32(.5)
	width := height * .4
	depth := height * .15
	_ = width
	b1 := wr.NewDynamic(wl, "body", physics.Sphere, "purple", 1.0, math32.Vec3(height, height, depth),
		math32.Vec3(0, height+2, 0), rot)
	_ = b1

	// bj := wl.NewJointRevolute(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0), math32.Vec3(0, 1, 0))
	// bj := wl.NewJointPrismatic(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0), math32.Vec3(1, 0, 0))
	// physics.SetJointControlForce(bj, 0, 5)
	// physics.SetJointTargetPos(bj, 0, 1)
	// physics.SetJointDoF(bj, 0, physics.JointDamp, 0.01)
	// physics.SetJointDoF(bj, 0, physics.JointStiff, 1) // this makes a big difference
	// bj := wl.NewJointFree(-1, b1.DynamicIndex, math32.Vec3(0, 0, 0), math32.Vec3(0, -height, 0))
	// _ = bj

	wr.Init(wl)

	params := physics.GetParams(0)
	params.Dt = 0.01
	// params.Gravity.Y = 0
	fmt.Println(params.ContactRelax)

	wr.Update()

	sc.Camera.Pose.Pos = math32.Vec3(0, 40, 3.5)
	sc.Camera.LookAt(math32.Vec3(0, 5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("3")

	sc.Camera.Pose.Pos = math32.Vec3(0, 20, 30)
	sc.Camera.LookAt(math32.Vec3(0, 5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("2")

	sc.Camera.Pose.Pos = math32.Vec3(-1.33, 2.24, 3.55)
	sc.Camera.LookAt(math32.Vec3(0, .5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("1")
	sc.SaveCamera("default")

	stepNButton := func(p *tree.Plan, n int) {
		nm := fmt.Sprintf("Step %d", n)
		tree.AddAt(p, nm, func(w *core.Button) {
			w.SetText(nm).SetIcon(icons.PlayArrow).
				SetTooltip(fmt.Sprintf("Step state %d times", n)).
				OnClick(func(e events.Event) {
					go func() {
						for range n {
							wl.Step()
							wr.Update()
							if se.IsVisible() {
								se.AsyncLock()
								se.NeedsRender()
								se.AsyncUnlock()
								time.Sleep(10 * time.Millisecond)
							}
						}
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
					SetTooltip("Reset state").
					OnClick(func(e events.Event) {
						wl.InitState()
						wr.Update()
						if se.IsVisible() {
							se.NeedsRender()
						}
					})
			})
			stepNButton(p, 1)
			stepNButton(p, 10)
			stepNButton(p, 100)
		})
	})
	b.RunMainWindow()
}
