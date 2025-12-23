// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate -add-types

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
	_ "cogentcore.org/lab/gosl/slbool/slboolcore" // include to get gui views
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/world"
)

// Balls has sim params
type Balls struct {

	// Number of balls: if collide, then run out of memory above 1000 or so
	NBalls int

	// Collide is whether the balls collide with each other
	Collide bool

	// Size of each ball (m)
	Size float32

	// Mass of each ball (kg)
	Mass float32

	// size of the box (m)
	Width  float32
	Depth  float32
	Height float32
	Thick  float32

	Bounce          float32
	Friction        float32
	FrictionTortion float32
	FrictionRolling float32
}

func (b *Balls) Defaults() {
	b.NBalls = 1000
	b.Collide = true
	b.Size = 0.2
	b.Mass = 0.1

	b.Width = 50
	b.Depth = 50
	b.Height = 20
	b.Thick = .1

	b.Bounce = 0.5
	b.Friction = 0
	b.FrictionTortion = 0
	b.FrictionRolling = 0
}

func main() {
	// gpu.Debug = true
	b := core.NewBody("test1").SetTitle("Physics Balls")
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

	bs := &Balls{}
	bs.Defaults()

	wl := physics.NewWorld()
	wl.GPU = true

	params := physics.GetParams(0)
	params.Dt = 0.0001    // leaks balls > 0.0005
	params.SubSteps = 100 // major speedup by inner-stepping
	// params.Gravity.Y = 0
	params.ContactRelax = 0.2         // 0.1 seems most physical -- 0.2 getting a bit more ke?
	params.Restitution.SetBool(false) // not working!
	params.ContactMargin = 0          // 0.1 better than .01 -- leaks a few

	bpf.SetStruct(bs)
	wpf.SetStruct(params)

	config := func() {
		wr.Reset()
		wl.Reset()
		rot := math32.NewQuat(0, 0, 0, 1)
		wr.NewBody(wl, "floor", physics.Plane, "#D0D0D080", math32.Vec3(0, 0, 0),
			math32.Vec3(0, 0, 0), rot)

		hw := bs.Width / 2
		hd := bs.Depth / 2
		hh := bs.Height / 2
		ht := bs.Thick / 2
		wr.NewBody(wl, "back-wall", physics.Box, "#0000FFA0", math32.Vec3(hw, hh, ht),
			math32.Vec3(0, hh, -hd), rot)
		wr.NewBody(wl, "left-wall", physics.Box, "#FF0000A0", math32.Vec3(ht, hh, hd),
			math32.Vec3(-hw, hh, 0), rot)
		wr.NewBody(wl, "right-wall", physics.Box, "#00FF00A0", math32.Vec3(ht, hh, hd),
			math32.Vec3(hw, hh, 0), rot)
		wr.NewBody(wl, "front-wall", physics.Box, "#FFFF00A0", math32.Vec3(hw, hh, ht),
			math32.Vec3(0, hh, hd), rot)

		box := bs.Width * .9
		size := bs.Size
		for i := range bs.NBalls {
			ht := rand.Float32() * bs.Height
			x := rand.Float32()*box - 0.5*box
			z := rand.Float32()*box - 0.5*box
			clr := colors.Names[i%len(colors.Names)]
			bl := wr.NewDynamic(wl, "ball", physics.Sphere, clr, bs.Mass, math32.Vec3(size, size, size),
				math32.Vec3(x, size+ht, z), rot)
			if !bs.Collide {
				physics.SetBodyGroup(bl.Index, int32(i+1)) // only collide within same group
			}
			bl.SetBodyBounce(bs.Bounce)
			bl.SetBodyFriction(bs.Friction)
			bl.SetBodyFrictionTortion(bs.FrictionTortion)
			bl.SetBodyFrictionRolling(bs.FrictionRolling)
		}
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

	sc.Camera.Pose.Pos = math32.Vec3(0, 80, 75)
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
