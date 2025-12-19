// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package world

import (
	"strconv"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/lab/physics"
)

// View has visualization functions for physics elements.
type View struct {
	// Name is a name for element (index always appended).
	Name string

	// Shape is the physical shape of the element.
	Shape physics.Shapes

	// Color is the color of the element.
	Color string

	// Size is the size (per shape).
	Size math32.Vector3

	// Pos is the position.
	Pos math32.Vector3

	// Quat is the rotation as a quaternion.
	Quat math32.Quat

	// NewView is a function that returns a new [xyz.Node]
	// to represent this element. If nil, uses appropriate defaults.
	NewView func() tree.Node

	// InitView is a function that initializes a new [xyz.Node]
	// that represents this element. If nil, uses appropriate defaults.
	InitView func(sld *xyz.Solid)

	// Index is the index of the element in a list.
	Index int32

	// DynamicIndex is the index of a dynamic element (-1 if not dynamic).
	DynamicIndex int32
}

// NewBody adds a new body with given parameters.
// Returns the View which can then be further customized.
// Use this for Static elements; NewDynamic for dynamic elements.
func (wr *World) NewBody(wl *physics.World, name string, shape physics.Shapes, clr string, size, pos math32.Vector3, rot math32.Quat) *View {
	idx := wl.NewBody(shape, size, pos, rot)
	vw := &View{Name: name, Index: idx, DynamicIndex: -1, Shape: shape, Color: clr, Size: size, Pos: pos, Quat: rot}
	wr.Views = append(wr.Views, vw)
	return vw
}

// NewDynamic adds a new dynamic body with given parameters.
// Returns the View which can then be further customized.
func (wr *World) NewDynamic(wl *physics.World, name string, shape physics.Shapes, clr string, mass float32, size, pos math32.Vector3, rot math32.Quat) *View {
	idx, dyIdx := wl.NewDynamic(shape, mass, size, pos, rot)
	vw := &View{Name: name, Index: idx, DynamicIndex: dyIdx, Shape: shape, Color: clr, Size: size, Pos: pos, Quat: rot}
	wr.Views = append(wr.Views, vw)
	return vw
}

// UpdateFromPhysics updates the View from physics state.
func (vw *View) UpdateFromPhysics() {
	params := physics.GetParams(0)
	if vw.DynamicIndex >= 0 {
		ix := int32(vw.DynamicIndex)
		vw.Pos = physics.DynamicPos(ix, params.Cur)
		vw.Quat = physics.DynamicQuat(ix, params.Cur)
	} else {
		ix := int32(vw.Index)
		vw.Pos = physics.BodyPos(ix)
		vw.Quat = physics.BodyQuat(ix)
	}
}

// UpdatePose updates the xyz node pose from view.
func (vw *View) UpdatePose(sld *xyz.Solid) {
	sld.Pose.Pos = vw.Pos
	sld.Pose.Quat = vw.Quat
}

// UpdateColor updates the xyz node color from view.
func (vw *View) UpdateColor(clr string, sld *xyz.Solid) {
	if clr == "" {
		return
	}
	sld.Material.Color = errors.Log1(colors.FromString(clr))
}

// Add adds given physics node to the [tree.Plan], using NewView
// function on the node, or default.
func (vw *View) Add(p *tree.Plan) {
	nm := vw.Name + strconv.Itoa(int(vw.Index))
	newFunc := vw.NewView
	if newFunc == nil {
		newFunc = func() tree.Node {
			return any(tree.New[xyz.Solid]()).(tree.Node)
		}
	}
	p.Add(nm, newFunc, func(n tree.Node) { vw.Init(n.(*xyz.Solid)) })
}

// Init initializes xyz node using InitView function or default.
func (vw *View) Init(sld *xyz.Solid) {
	initFunc := vw.InitView
	if initFunc != nil {
		initFunc(sld)
		return
	}
	switch vw.Shape {
	case physics.Plane:
		vw.PlaneInit(sld)
	case physics.Sphere:
		vw.SphereInit(sld)
	case physics.Capsule:
		vw.CapsuleInit(sld)
	case physics.Cylinder:
		vw.CylinderInit(sld)
	case physics.Box:
		vw.BoxInit(sld)
	}
}

// BoxInit is the default InitView function for [physics.Box].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) BoxInit(sld *xyz.Solid) {
	mnm := "physics.Box"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		xyz.NewBox(sld.Scene, mnm, 1, 1, 1)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale = vw.Size.MulScalar(2)
	vw.UpdateColor(vw.Color, s6ld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// PlaneInit is the default InitView function for [physics.Plane].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) PlaneInit(sld *xyz.Solid) {
	mnm := "physics.Plane"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		xyz.NewPlane(sld.Scene, mnm, 1, 1)
	}
	sld.SetMeshName(mnm)
	if vw.Size.X == 0 {
		inf := 1e6
		sld.Pose.Scale = math32.Vec3(inf, inf, 1)
	} else {
		sld.Pose.Scale = vw.Size.MulScalar(2)
	}
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// CylinderInit is the default InitView function for [physics.Cylinder].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) CylinderInit(sld *xyz.Solid) {
	mnm := "physics.Cylinder"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		xyz.NewCylinder(sld.Scene, mnm, 1, 1, 32, 1, true, true)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale = vw.Size
	sld.Pose.Scale.Y *= 2
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// CapsuleInit is the default InitView function for [physics.Capsule].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) CapsuleInit(sld *xyz.Solid) {
	mnm := "physics.Capsule"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		ms = xyz.NewCapsule(sld.Scene, mnm, 1, .2, 32, 1)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale.Set(vw.Size.X/.2, 2*(vw.Size.Y/1.4), vw.Size.Z/.2)
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}

// SphereInit is the default InitView function for [physics.Sphere].
// Only updates Pose in Updater: if node will change size or color,
// add updaters for that.
func (vw *View) SphereInit(sld *xyz.Solid) {
	mnm := "physics.Sphere"
	if ms, _ := sld.Scene.MeshByName(mnm); ms == nil {
		ms = xyz.NewSphere(sld.Scene, mnm, 1, 32)
	}
	sld.SetMeshName(mnm)
	sld.Pose.Scale.SetScalar(vw.Size.X)
	vw.UpdateColor(vw.Color, sld)
	sld.Updater(func() {
		vw.UpdatePose(sld)
	})
}
