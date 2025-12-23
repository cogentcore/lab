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

// Pendula has sim params
type Pendula struct {

	// Number of balls: if collide, then run out of memory above 1000 or so
	NPendula int

	ForceOn  int
	ForceOff int
	Force    float32

	HSize math32.Vector3

	// Mass of each ball (kg)
	Mass float32

	// do the elements collide with each other?
	Collide bool

	// Stiff is the strength of positional constraints
	// when imposing them.
	Stiff float32

	// Damp is the strength of velocity constraints
	// when imposing them.
	Damp float32
}

func (b *Pendula) Defaults() {
	b.NPendula = 2
	b.HSize.Set(0.05, .2, 0.05)
	b.Mass = 0.1

	b.ForceOn = 100
	b.ForceOff = 102
	b.Force = 0

	b.Damp = 0.5
	b.Stiff = 1e4
}

func main() {
	// gpu.Debug = true
	b := core.NewBody("test1").SetTitle("Physics Pendula")
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

	ps := &Pendula{}
	ps.Defaults()

	wl := physics.NewWorld()
	wl.GPU = true

	params := physics.GetParams(0)
	params.Dt = 0.0001    // leaks balls > 0.0005
	params.SubSteps = 100 // major speedup by inner-stepping
	// params.Gravity.Y = 0
	params.Restitution.SetBool(false) // not working!
	params.ContactMargin = 0.1        // 0.1 better than .01 -- leaks a few

	bpf.SetStruct(ps)
	wpf.SetStruct(params)

	var topJoint int32

	config := func() {
		wr.Reset()
		wl.Reset()
		rot := math32.NewQuat(0, 0, 0, 1)
		_ = rot
		rleft := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), -math32.Pi/2)
		_ = rleft
		stY := 2*ps.HSize.Y*float32(ps.NPendula+1) + 1
		clr := colors.Names[0]
		pb := wr.NewDynamic(wl, "top", physics.Box, clr, ps.Mass, ps.HSize,
			math32.Vec3(-ps.HSize.Y, stY, 0), rleft)
		if !ps.Collide {
			physics.SetBodyGroup(pb.Index, int32(1))
		}

		ji := wl.NewJointRevolute(-1, pb.DynamicIndex, math32.Vec3(0, stY, 0), math32.Vec3(0, ps.HSize.Y, 0), math32.Vec3(0, 0, 1))
		// physics.SetJointTargetPos(ji, 0, math32.Pi/2, 1) // let it swing!
		physics.SetJointTargetPos(ji, 0, 0, 0) // let it swing!
		physics.SetJointTargetVel(ji, 0, 0, 0) // let it swing!

		topJoint = ji

		for i := 1; i < ps.NPendula; i++ {
			clr := colors.Names[i%len(colors.Names)]
			x := -float32(i)*ps.HSize.Y*2 - ps.HSize.Y
			cb := wr.NewDynamic(wl, "child", physics.Box, clr, ps.Mass, ps.HSize,
				math32.Vec3(x, stY, 0), rleft)
			if !ps.Collide {
				physics.SetBodyGroup(cb.Index, int32(1+i))
			}
			ji = wl.NewJointRevolute(pb.DynamicIndex, cb.DynamicIndex, math32.Vec3(0, -ps.HSize.Y, 0), math32.Vec3(0, ps.HSize.Y, 0), math32.Vec3(0, 0, 1))
			physics.SetJointTargetPos(ji, 0, 0, 0) // let it swing!
			physics.SetJointTargetVel(ji, 0, 0, 0) // let it swing!
			pb = cb
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

	sc.Camera.Pose.Pos = math32.Vec3(0, 6, 4.5)
	sc.Camera.LookAt(math32.Vec3(0, 3, 0), math32.Vec3(0, 1, 0))
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
						fmt.Println("still running...")
						return
					}
					go func() {
						isRunning = true
						for range n {
							wl.Step()
							cycle++
							if cycle >= ps.ForceOn && cycle < ps.ForceOff {
								fmt.Println(cycle, "\tforce on:", ps.Force)
								physics.SetJointControlForce(topJoint, 0, ps.Force)
							} else {
								physics.SetJointControlForce(topJoint, 0, 0)
							}
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
							fmt.Println("still running...")
							return
						}
						stop = false
						cycle = 0
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
							fmt.Println("still running...")
							return
						}
						stop = false
						cycle = 0
						config()
						updateView()
					})
			})
		})
	})
	b.RunMainWindow()
}
