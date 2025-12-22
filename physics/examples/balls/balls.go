// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate

import (
	"fmt"
	"math/rand/v2"

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
	b := core.NewBody("test1").SetTitle("Physics Balls")
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
	wl.GPU = true
	fv.SetStruct(wl)

	split.SetSplits(0.2, 0.8)

	rot := math32.NewQuat(0, 0, 0, 1)
	wr.NewBody(wl, "floor", physics.Plane, "#D0D0D040", math32.Vec3(0, 0, 0),
		math32.Vec3(0, 0, 0), rot)

	nballs := 10000
	size := float32(0.1)
	bounce := float32(0.5)
	box := 2 * size * math32.Sqrt(float32(nballs))
	height := float32(20)
	for i := range nballs {
		_ = i
		ht := rand.Float32() * height
		x := rand.Float32()*box - 0.5*box
		z := rand.Float32()*box - 0.5*box
		clr := colors.Names[i%len(colors.Names)]
		b1 := wr.NewDynamic(wl, "body", physics.Sphere, clr, 1.0, math32.Vec3(size, size, size),
			math32.Vec3(x, size+ht, z), rot)
		// todo: helper methods on view to set this stuff:
		physics.Bodies.Set(bounce, int(b1.Index), int(physics.BodyBounce))
		physics.SetBodyGroup(b1.Index, int32(i)) // no self collisions
	}
	wr.Init(wl)

	params := physics.GetParams(0)
	params.Dt = 0.001
	// params.Gravity.Y = 0
	params.ContactRelax = 0.1
	params.Restitution.SetBool(false)
	fmt.Println(params.ContactRelax)

	wl.Config()
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
					SetTooltip("Reset state").
					OnClick(func(e events.Event) {
						wl.InitState()
						wr.Update()
						if se.IsVisible() {
							se.NeedsRender()
						}
					})
			})
			tree.Add(p, func(w *core.Button) {
				w.SetText("Stop").SetIcon(icons.Stop).
					SetTooltip("Stop running").
					OnClick(func(e events.Event) {
						stop = true
					})
			})
			stepNButton(p, 1)
			stepNButton(p, 10)
			stepNButton(p, 100)
			stepNButton(p, 1000)
			stepNButton(p, 10000)
		})
	})
	b.RunMainWindow()
}
