// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phyxyz

import (
	"image"
	"testing"

	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/gpu"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/xyz"
	"cogentcore.org/lab/physics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testModel returns a new model that runs on the CPU.
func testModel() *physics.Model {
	ml := physics.NewModel()
	ml.GPU = false
	return ml
}

// TestScene tests making the skins of bodies of each shape and joints,
// initializing the scene, and updating it from the physics, as a dynamic
// body falls.
func TestScene(t *testing.T) {
	ml := testModel()
	sc := NewScene(xyz.NewScene())
	rot := math32.NewQuatIdentity()
	hsz := math32.Vec3(0.5, 0.5, 0.5)
	gnd := sc.NewBody(ml, "ground", physics.Plane, "grey", math32.Vec3(10, 0, 10), math32.Vec3(0, 0, 0), rot)
	gnd.SetBodyFriction(0.5)
	gnd.SetBodyBounce(0.1)
	gnd.SetBodyFrictionTortion(0.1)
	gnd.SetBodyFrictionRolling(0.1)
	box := sc.NewDynamic(ml, "box", physics.Box, "red", 1, hsz, math32.Vec3(0, 3, 0), rot)
	box.SetBodyWorld(0)
	box.SetBodyGroup(1)
	shapes := []physics.Shapes{physics.Cylinder, physics.Capsule, physics.Sphere}
	var dyns []*Skin
	for i, sh := range shapes {
		dyns = append(dyns, sc.NewDynamic(ml, sh.String(), sh, "blue", 1, hsz, math32.Vec3(float32(2+2*i), 3, 0), rot))
	}
	ml.NewObject()
	z := math32.Vec3(0, 0, 0)
	axis := math32.Vec3(0, 0, 1)
	assert.GreaterOrEqual(t, sc.NewJointRevolute(ml, nil, dyns[0], z, z, axis), int32(0))
	sc.NewJointPrismatic(ml, dyns[0], dyns[1], math32.Vec3(2, 0, 0), z, math32.Vec3(1, 0, 0))
	sc.NewJointBall(ml, dyns[1], dyns[2], math32.Vec3(2, 0, 0), z)
	ml.NewObject()
	sc.NewJointFree(ml, nil, box, z, z)

	sc.Init(ml)
	assert.False(t, sc.ReplicasView)
	assert.Equal(t, len(sc.Skins), sc.Root.NumChildren(), "a solid for each skin")
	sc.Init(ml) // again: updates
	sc.InitState(ml)

	y0 := box.Pos.Y
	for range 500 {
		ml.Step()
	}
	sc.Update()
	assert.Less(t, box.Pos.Y, y0, "the skin follows the falling body")
	sld := sc.Root.ChildByName("box" + "1").(*xyz.Solid)
	assert.Equal(t, box.Pos, sld.Pose.Pos)

	sc.Reset()
	assert.Empty(t, sc.Skins)
}

// TestSceneJoints tests making the fixed and distance joints of skins.
func TestSceneJoints(t *testing.T) {
	ml := testModel()
	sc := NewScene(xyz.NewScene())
	rot := math32.NewQuatIdentity()
	hsz := math32.Vec3(0.2, 0.2, 0.2)
	a := sc.NewDynamic(ml, "a", physics.Box, "red", 1, hsz, math32.Vec3(0, 0, 0), rot)
	b := sc.NewDynamic(ml, "b", physics.Box, "red", 1, hsz, math32.Vec3(1, 0, 0), rot)
	c := sc.NewDynamic(ml, "c", physics.Box, "red", 1, hsz, math32.Vec3(2, 0, 0), rot)
	ml.NewObject()
	z := math32.Vec3(0, 0, 0)
	sc.NewJointFixed(ml, a, b, math32.Vec3(1, 0, 0), z)
	sc.NewJointDistance(ml, b, c, math32.Vec3(1, 0, 0), z, 0.5, 1.5)
	sc.Init(ml)
	ml.Step()
	sc.Update()
}

// TestAddSkinClone tests that a clone of a skin is a copy.
func TestAddSkinClone(t *testing.T) {
	sc := NewScene(xyz.NewScene())
	sk := sc.NewSkin(physics.Box, "a", "red", math32.Vec3(1, 1, 1), math32.Vec3(0, 0, 0), math32.NewQuatIdentity())
	sc.AddSkinClone(sk)
	require.Len(t, sc.Skins, 2)
	assert.NotSame(t, sk, sc.Skins[1])
	assert.Equal(t, *sk, *sc.Skins[1])
}

// TestDepth tests normalizing depth maps and rendering them as images.
func TestDepth(t *testing.T) {
	var cam Camera
	cam.Defaults()
	cam.Size = image.Pt(2, 2)
	depth := []float32{0, 0.5, 0.9, 1} // bottom row first
	var nd []float32
	DepthNorm(&nd, depth, &cam, false)
	require.Len(t, nd, 4)
	assert.Less(t, nd[0], nd[1], "farther is larger")
	assert.Equal(t, float32(1), nd[3], "beyond MaxD")
	var fl []float32
	DepthNorm(&fl, depth, &cam, true)
	assert.Equal(t, nd[0], fl[2], "flipped in Y")
	cam.LogD = false
	DepthNorm(&nd, depth, &cam, false)
	assert.Less(t, nd[0], nd[1])

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	DepthImage(img, depth, colormap.AvailableMaps["ColdHot"], &cam)
	assert.NotEqual(t, img.RGBAAt(0, 1), img.RGBAAt(1, 0), "different depths have different colors")
	DepthImage(nil, depth, nil, &cam)
}

// TestRenderFrom tests rendering the scene offscreen from the point of
// view of a skin, as for the eyes of a body.
func TestRenderFrom(t *testing.T) {
	gpu.SkipIfNoGPU(t)
	gp, dev, err := gpu.NoDisplayGPU()
	require.NoError(t, err)
	ml := testModel()
	sc := NewScene(NoDisplayScene(gp, dev))
	rot := math32.NewQuatIdentity()
	sc.NewBody(ml, "ground", physics.Plane, "grey", math32.Vec3(10, 0, 10), math32.Vec3(0, 0, 0), rot)
	eye := sc.NewDynamic(ml, "eye", physics.Sphere, "blue", 1, math32.Vec3(0.1, 0.1, 0.1), math32.Vec3(0, 2, 5), rot)
	sc.Init(ml)
	var cam Camera
	cam.Defaults()
	imgs := sc.RenderFrom(eye, &cam)
	require.Len(t, imgs, 1)
	assert.Equal(t, cam.Size, imgs[0].Bounds().Size())
}
