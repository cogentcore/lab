// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate

import (
	"image"
	"os"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/core/xyz/xyzcore"
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/builder"
	"cogentcore.org/lab/physics/phyxyz"
)

var NoGUI bool

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-nogui" {
		NoGUI = true
	}
	ev := &Env{}
	ev.Defaults()
	if NoGUI {
		ev.NoGUIRun()
		return
	}
	// core.RenderTrace = true
	b := ev.ConfigGUI()
	b.RunMainWindow()
}

// Env encapsulates the virtual environment
type Env struct { //types:add

	// if true, emer is angry: changes face color
	EmerAngry bool

	// height of emer
	EmerHt float32

	// how far to move every step
	MoveStep float32

	// how far to rotate every step
	RotStep float32

	// number of model steps to take
	ModelSteps int

	// width of room
	Width float32

	// depth of room
	Depth float32

	// height of room
	Height float32

	// thickness of walls of room
	Thick float32

	// current depth map
	DepthVals []float32

	// offscreen render camera settings
	Camera phyxyz.Camera

	// color map to use for rendering depth map
	DepthMap core.ColorMapName

	// The core physics elements: Model, Builder, Scene
	Physics builder.Physics

	// 3D visualization of the Scene
	SceneEditor *xyzcore.SceneEditor

	// emer object
	Emer *builder.Object `display:"-"`

	// emer PlaneXZ joint for controlling motion
	EmerJoint *builder.Joint

	// Right eye of emer
	EyeR *builder.Body `display:"-"`

	// snapshot image
	EyeRImg *core.Image `display:"-"`

	// ball joint for the neck.
	NeckJoint *builder.Joint

	// depth map image
	DepthImage *core.Image `display:"-"`
}

func (ev *Env) Defaults() {
	ev.Width = 10
	ev.Depth = 15
	ev.Height = 2
	ev.Thick = 0.2
	ev.EmerHt = 1
	ev.MoveStep = ev.EmerHt * .2
	ev.RotStep = 15
	ev.ModelSteps = 100
	ev.DepthMap = core.ColorMapName("ColdHot")
	ev.Camera.Defaults()
	ev.Camera.FOV = 90
}

func (ev *Env) MakeWorld(sc *xyz.Scene) {
	ev.Physics.Model = physics.NewModel()
	ev.Physics.Builder = builder.NewBuilder()
	ev.Physics.Model.GPU = false
	sc.Background = colors.Scheme.Select.Container
	xyz.NewAmbient(sc, "ambient", 0.3, xyz.DirectSun)

	dir := xyz.NewDirectional(sc, "dir", 1, xyz.DirectSun)
	dir.Pos.Set(0, 2, 1) // default: 0,1,1 = above and behind us (we are at 0,0,X)

	ev.Physics.Scene = phyxyz.NewScene(sc)
	wl := ev.Physics.Builder.NewGlobalWorld()
	ev.MakeRoom(wl, "room1", ev.Width, ev.Depth, ev.Height, ev.Thick)
	ew := ev.Physics.Builder.NewWorld()
	ev.MakeEmer(ew, "emer", ev.EmerHt)
	// ev.Physics.Builder.ReplicateWorld(1, 8, 2)
	ev.Physics.Build()
	params := physics.GetParams(0)
	params.Gravity.Y = 0
	params.MaxForce = 1.0e3
	params.AngularDamping = 0.5
	// params.SubSteps = 1
}

// Initstate reinitializes the physics model state.
func (ev *Env) InitState() { //types:add
	ev.Physics.InitState()
	ev.UpdateView()
}

// ConfigView3D makes the 3D view
func (ev *Env) ConfigView3D(sc *xyz.Scene) {
	// sc.MultiSample = 1 // we are using depth grab so we need this = 1
}

// RenderEyeImg returns a snapshot from the perspective of Emer's right eye
func (ev *Env) RenderEyeImg() image.Image {
	return ev.Physics.Scene.RenderFrom(ev.EyeR.Skin, &ev.Camera, 0)
}

// GrabEyeImg takes a snapshot from the perspective of Emer's right eye
func (ev *Env) GrabEyeImg() { //types:add
	img := ev.RenderEyeImg()
	if img != nil {
		ev.EyeRImg.SetImage(img)
		ev.EyeRImg.NeedsRender()
	}
	// depth, err := ev.View3D.DepthImage()
	// if err == nil && depth != nil {
	// 	ev.DepthVals = depth
	// 	ev.ViewDepth(depth)
	// }
}

// ViewDepth updates depth bitmap with depth data
func (ev *Env) ViewDepth(depth []float32) {
	cmap := colormap.AvailableMaps[string(ev.DepthMap)]
	img := image.NewRGBA(image.Rectangle{Max: ev.Camera.Size})
	ev.DepthImage.SetImage(img)
	phyxyz.DepthImage(img, depth, cmap, &ev.Camera)
	ev.DepthImage.NeedsRender()
}

// UpdateView tells 3D view it needs to update.
func (ev *Env) UpdateView() {
	if ev.SceneEditor.IsVisible() {
		ev.SceneEditor.NeedsRender()
	}
}

// ModelStep does one step of the physics model.
func (ev *Env) ModelStep() { //types:add
	// physics.ToGPU(physics.DynamicsVar)
	ev.Physics.Step(ev.ModelSteps)
	// cts := pw.WorldCollide(physics.DynsTopGps)
	// ev.Contacts = nil
	// for _, cl := range cts {
	// 	if len(cl) > 1 {
	// 		for _, c := range cl {
	// 			if c.A.AsTree().Name == "body" {
	// 				ev.Contacts = cl
	// 			}
	// 			fmt.Printf("A: %v  B: %v\n", c.A.AsTree().Name, c.B.AsTree().Name)
	// 		}
	// 	}
	// }
	ev.EmerAngry = false
	// if len(ev.Contacts) > 1 { // turn around
	// 	ev.EmerAngry = true
	// 	fmt.Printf("hit wall: turn around!\n")
	// 	rot := 100.0 + 90.0*rand.Float32()
	// 	ev.Emer.Rel.RotateOnAxis(0, 1, 0, rot)
	// }
	ev.GrabEyeImg()
	ev.UpdateView()
}

// StepForward moves Emer forward in current facing direction one step,
// and takes GrabEyeImg
func (ev *Env) StepForward() { //types:add
	ev.Emer.PoseFromPhysics()
	// doesn't integrate well with joints..
	// ev.Emer.MoveOnAxisBody(0, 0, 0, 1, -ev.MoveStep)
	// ev.Emer.PoseToPhysics()
	ev.EmerJoint.AddPlaneXZPos(math32.Pi*.5, -ev.MoveStep, 1000)
	ev.ModelStep()
}

// StepBackward moves Emer backward in current facing direction one step, and takes GrabEyeImg
func (ev *Env) StepBackward() { //types:add
	ev.Emer.PoseFromPhysics()
	// ev.Emer.MoveOnAxisBody(0, 0, 0, 1, ev.MoveStep)
	// ev.Emer.PoseToPhysics()
	ev.EmerJoint.AddPlaneXZPos(math32.Pi*.5, ev.MoveStep, 1000)
	ev.ModelStep()
}

// RotBodyLeft rotates emer left and takes GrabEyeImg
func (ev *Env) RotBodyLeft() { //types:add
	ev.Emer.PoseFromPhysics()
	// ev.Emer.RotateOnAxisBody(0, 0, 1, 0, ev.RotStep)
	// ev.Emer.PoseToPhysics()
	ev.EmerJoint.AddTargetPos(2, math32.DegToRad(ev.RotStep), 1000)
	ev.ModelStep()
}

// RotBodyRight rotates emer right and takes GrabEyeImg
func (ev *Env) RotBodyRight() { //types:add
	ev.Emer.PoseFromPhysics()
	// ev.Emer.RotateOnAxisBody(0, 0, 1, 0, -ev.RotStep)
	// ev.Emer.PoseToPhysics()
	ev.EmerJoint.AddTargetPos(2, math32.DegToRad(-ev.RotStep), 1000)
	ev.ModelStep()
}

// RotHeadLeft rotates emer left and takes GrabEyeImg
func (ev *Env) RotHeadLeft() { //types:add
	ev.Emer.PoseFromPhysics()
	ev.NeckJoint.AddTargetAngle(1, ev.RotStep, 1000)
	ev.ModelStep()
}

// RotHeadRight rotates emer right and takes GrabEyeImg
func (ev *Env) RotHeadRight() { //types:add
	ev.Emer.PoseFromPhysics()
	ev.NeckJoint.AddTargetAngle(1, -ev.RotStep, 1000)
	ev.ModelStep()
}

// MakeRoom constructs a new room with given params
func (ev *Env) MakeRoom(wl *builder.World, name string, width, depth, height, thick float32) {
	rot := math32.NewQuatIdentity()
	hw := width / 2
	hd := depth / 2
	hh := height / 2
	ht := thick / 2
	obj := wl.NewObject()
	sc := ev.Physics.Scene
	obj.NewBodySkin(sc, name+"_floor", physics.Box, "grey", math32.Vec3(hw, ht, hd),
		math32.Vec3(0, -ht, 0), rot)
	obj.NewBodySkin(sc, name+"_back-wall", physics.Box, "blue", math32.Vec3(hw, hh, ht),
		math32.Vec3(0, hh, -hd), rot)
	obj.NewBodySkin(sc, name+"_left-wall", physics.Box, "red", math32.Vec3(ht, hh, hd),
		math32.Vec3(-hw, hh, 0), rot)
	obj.NewBodySkin(sc, name+"_right-wall", physics.Box, "green", math32.Vec3(ht, hh, hd),
		math32.Vec3(hw, hh, 0), rot)
	obj.NewBodySkin(sc, name+"_front-wall", physics.Box, "yellow", math32.Vec3(hw, hh, ht),
		math32.Vec3(0, hh, hd), rot)
}

// MakeEmer constructs a new Emer virtual robot of given height (e.g., 1).
func (ev *Env) MakeEmer(wl *builder.World, name string, height float32) {
	hh := height / 2
	hw := hh * .4
	hd := hh * .15
	headsz := hd * 1.5
	eyesz := headsz * .2
	mass := float32(1) // kg
	rot := math32.NewQuatIdentity()
	obj := wl.NewObject()
	ev.Emer = obj
	sc := ev.Physics.Scene
	emr := obj.NewDynamicSkin(sc, name+"_body", physics.Box, "purple", mass, math32.Vec3(hw, hh, hd), math32.Vec3(0, hh, 0), rot)
	// body := physics.NewCapsule(emr, "body", math32.Vec3(0, hh, 0), hh, hw)
	// body := physics.NewCylinder(emr, "body", math32.Vec3(0, hh, 0), hh, hw)
	ev.EmerJoint = obj.NewJointPlaneXZ(nil, emr, math32.Vec3(0, 0, 0), math32.Vec3(0, -hh, 0))
	emr.Group = 0

	headPos := math32.Vec3(0, 2*hh+headsz, 0)
	head := obj.NewDynamicSkin(sc, name+"_head", physics.Box, "tan", mass*.1, math32.Vec3(headsz, headsz, headsz), headPos, rot)
	head.Group = 0
	hdsk := head.Skin
	hdsk.InitSkin = func(sld *xyz.Solid) {
		hdsk.BoxInit(sld)
		sld.Updater(func() {
			clr := hdsk.Color
			if ev.EmerAngry {
				clr = "pink"
			}
			hdsk.UpdateColor(clr, sld)
		})
	}
	ev.NeckJoint = obj.NewJointBall(emr, head, math32.Vec3(0, hh, 0), math32.Vec3(0, -headsz, 0))

	eyeoff := math32.Vec3(-headsz*.6, headsz*.1, -(headsz + eyesz*.3))
	bd := obj.NewDynamicSkin(sc, name+"_eye-l", physics.Box, "green", mass*.01, math32.Vec3(eyesz, eyesz*.5, eyesz*.2), headPos.Add(eyeoff), rot)
	bd.Group = 0
	obj.NewJointFixed(head, bd, eyeoff, math32.Vec3(0, 0, -eyesz*.3))

	eyeoff = math32.Vec3(headsz*.6, headsz*.1, -(headsz + eyesz*.3))
	ev.EyeR = obj.NewDynamicSkin(sc, name+"_eye-r", physics.Box, "green", mass*.01, math32.Vec3(eyesz, eyesz*.5, eyesz*.2), headPos.Add(eyeoff), rot)
	ev.EyeR.Group = 0
	obj.NewJointFixed(head, ev.EyeR, eyeoff, math32.Vec3(0, 0, -eyesz*.3))
}

func (ev *Env) ConfigGUI() *core.Body {
	// vgpu.Debug = true

	b := core.NewBody("virtroom").SetTitle("Physics Virtual Room")
	split := core.NewSplits(b)

	core.NewForm(split).SetStruct(ev)
	imfr := core.NewFrame(split)
	tbvw := core.NewTabs(split)
	scfr, _ := tbvw.NewTab("3D View")

	split.SetSplits(.2, .2, .6)

	////////    3D Scene

	etb := core.NewToolbar(scfr)
	_ = etb
	ev.SceneEditor = xyzcore.NewSceneEditor(scfr)
	ev.SceneEditor.UpdateWidget()
	sc := ev.SceneEditor.SceneXYZ()
	ev.MakeWorld(sc)

	// local toolbar for manipulating emer
	// etb.Maker(phyxyz.MakeStateToolbar(&ev.Emer.Rel, func() {
	// 	ev.World.Update()
	// 	ev.SceneEditor.NeedsRender()
	// }))

	sc.Camera.Pose.Pos = math32.Vec3(0, 40, 3.5)
	sc.Camera.LookAt(math32.Vec3(0, 5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("3")

	sc.Camera.Pose.Pos = math32.Vec3(0, 20, 30)
	sc.Camera.LookAt(math32.Vec3(0, 5, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("2")

	sc.Camera.Pose.Pos = math32.Vec3(-.86, .97, 2.7)
	sc.Camera.LookAt(math32.Vec3(0, .8, 0), math32.Vec3(0, 1, 0))
	sc.SaveCamera("1")
	sc.SaveCamera("default")

	////////    Image

	imfr.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
	})
	core.NewText(imfr).SetText("Right Eye Image:")
	ev.EyeRImg = core.NewImage(imfr)
	ev.EyeRImg.SetName("eye-r-img")
	ev.EyeRImg.Image = image.NewRGBA(image.Rectangle{Max: ev.Camera.Size})

	core.NewText(imfr).SetText("Right Eye Depth:")
	ev.DepthImage = core.NewImage(imfr)
	ev.DepthImage.SetName("depth-img")
	ev.DepthImage.Image = image.NewRGBA(image.Rectangle{Max: ev.Camera.Size})

	////////    Toolbar

	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(ev.MakeToolbar)
	})
	return b
}

func (ev *Env) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.InitState).SetText("Init").SetIcon(icons.Update)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.GrabEyeImg).SetText("Grab Image").SetIcon(icons.Image)
	})
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.ModelStep).SetText("Step").SetIcon(icons.SkipNext).
			Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.StepForward).SetText("Fwd").SetIcon(icons.SkipNext).
			Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.StepBackward).SetText("Bkw").SetIcon(icons.SkipPrevious).
			Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.RotBodyLeft).SetText("Body Left").SetIcon(icons.KeyboardArrowLeft).
			Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.RotBodyRight).SetText("Body Right").SetIcon(icons.KeyboardArrowRight).
			Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.RotHeadLeft).SetText("Head Left").SetIcon(icons.KeyboardArrowLeft).
			Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(ev.RotHeadRight).SetText("Head Right").SetIcon(icons.KeyboardArrowRight).
			Styler(func(s *styles.Style) {
				s.SetAbilities(true, abilities.RepeatClickable)
			})
	})
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Button) {
		w.SetText("README").SetIcon(icons.FileMarkdown).
			SetTooltip("Open browser on README.").
			OnClick(func(e events.Event) {
				core.TheApp.OpenURL("https://github.com/cogentcore/core/blob/master/xyz/examples/physics/README.md")
			})
	})
}

func (ev *Env) NoGUIRun() {
	gp, dev, err := gpu.NoDisplayGPU()
	if err != nil {
		panic(err)
	}
	sc := phyxyz.NoDisplayScene(gp, dev)
	ev.MakeWorld(sc)

	img := ev.RenderEyeImg()
	if img != nil {
		imagex.Save(img, "eyer_0.png")
	}
}
