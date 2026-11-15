// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"testing"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/xyz"
	"cogentcore.org/lab/physics"
	"cogentcore.org/lab/physics/phyxyz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestPhysics returns a Physics without a scene, with a model that runs
// on the CPU.
func newTestPhysics() *Physics {
	ml := physics.NewModel()
	ml.GPU = false
	return &Physics{Model: ml, Builder: NewBuilder()}
}

// TestBuildFall tests building a world with a static ground and a
// dynamic box, which falls under gravity onto the ground, and the sensors
// and poses of the bodies.
func TestBuildFall(t *testing.T) {
	ph := newTestPhysics()
	rot := math32.NewQuatIdentity()
	wl := ph.Builder.NewWorld()
	gnd := wl.NewObject()
	ground := gnd.NewBody(physics.Plane, math32.Vec3(10, 0, 10), math32.Vec3(0, 0, 0), rot)
	ground.Friction = 0.5
	ob := wl.NewObject()
	box := ob.NewDynamic(physics.Box, 1, math32.Vec3(0.5, 0.5, 0.5), math32.Vec3(0, 3, 0), rot)
	box.Bounce = 0.1
	var sensed int
	ob.NewSensor(func(o *Object) { sensed++ })

	ph.Build()
	ph.InitState()
	require.GreaterOrEqual(t, box.DynamicIndex, int32(0), "a dynamic body")
	assert.Equal(t, int32(-1), ground.DynamicIndex)
	assert.True(t, ob.HasBodyIndex(box.BodyIndex))
	assert.False(t, ob.HasBodyIndex(ground.BodyIndex))

	start := box.Pose.Pos.Y
	ph.StepQuiet(2000) // 2 seconds
	box.PoseFromPhysics()
	assert.Less(t, box.Pose.Pos.Y, start, "fell")
	assert.InDelta(t, 0.5, box.Pose.Pos.Y, 0.2, "on the ground")

	ph.Builder.RunSensors()
	assert.Equal(t, 1, sensed)

	// moving the pose and setting it in the physics
	box.Pose.Pos = math32.Vec3(1, 2, 3)
	box.PoseToPhysics()
	box.PoseFromPhysics()
	assert.Equal(t, math32.Vec3(1, 2, 3), box.Pose.Pos)
	ground.PoseFromPhysics()
	assert.Equal(t, math32.Vec3(0, 0, 0), ground.Pose.Pos)
	ground.PoseToPhysics()

	ph.Step(1) // without a scene
	ph.Builder.Reset()
	assert.Empty(t, ph.Builder.Worlds)
}

// TestBuildJoints tests building the joints of each type, with their
// target positions, velocities, and angles.
func TestBuildJoints(t *testing.T) {
	ph := newTestPhysics()
	params := physics.GetParams(0)
	params.Gravity.Y = 0
	rot := math32.NewQuatIdentity()
	hsz := math32.Vec3(0.2, 0.2, 0.2)
	wl := ph.Builder.NewWorld()
	ob := wl.NewObject()
	b0 := ob.NewDynamic(physics.Box, 1, hsz, math32.Vec3(0, 0, 0), rot)
	b1 := ob.NewDynamic(physics.Box, 1, hsz, math32.Vec3(1, 0, 0), rot)
	b2 := ob.NewDynamic(physics.Box, 1, hsz, math32.Vec3(2, 0, 0), rot)
	b3 := ob.NewDynamic(physics.Box, 1, hsz, math32.Vec3(3, 0, 0), rot)
	b4 := ob.NewDynamic(physics.Box, 1, hsz, math32.Vec3(4, 0, 0), rot)
	b5 := ob.NewDynamic(physics.Box, 1, hsz, math32.Vec3(5, 0, 0), rot)
	z := math32.Vec3(0, 0, 0)
	axis := math32.Vec3(0, 0, 1)
	rev := ob.NewJointRevolute(nil, b0, z, z, axis)
	pri := ob.NewJointPrismatic(b0, b1, math32.Vec3(1, 0, 0), z, math32.Vec3(1, 0, 0))
	fix := ob.NewJointFixed(b1, b2, math32.Vec3(1, 0, 0), z)
	ball := ob.NewJointBall(b2, b3, math32.Vec3(1, 0, 0), z)
	dist := ob.NewJointDistance(b3, b4, math32.Vec3(1, 0, 0), z, 0.5, 1.5)
	free := ob.NewJointFree(nil, b5, z, z)
	for _, j := range []*Joint{rev, pri, fix, ball, dist, free} {
		require.NotNil(t, j)
	}
	assert.True(t, rev.IsGlobal())
	assert.False(t, pri.IsGlobal())
	assert.Same(t, rev, ob.Joint(0))
	assert.Same(t, b1, ob.Body(1))

	ph.Build()
	ph.InitState()
	assert.GreaterOrEqual(t, rev.JointIndex, int32(0))

	rev.SetTargetVel(0, 0, 10)
	rev.SetTargetPos(0, 0.5, 100)
	rev.AddTargetPos(0, 0.25, 100)
	assert.InDelta(t, 0.75, rev.DoF(0).Current.Pos, 1e-6)
	rev.SetTargetAngle(0, 90, 100)
	assert.InDelta(t, math32.Pi/2, rev.DoF(0).Current.Pos, 1e-5)
	rev.AddTargetAngle(0, 90, 100)
	assert.InDelta(t, math32.Pi, math32.Abs(rev.DoF(0).Current.Pos), 1e-5)
	ph.StepQuiet(10)

	rev.PPose.Pos = math32.Vec3(0, 1, 0)
	rev.PoseToPhysics()
	rev.PoseFromPhysics()
	assert.Equal(t, math32.Vec3(0, 1, 0), rev.PPose.Pos)
	pri.PoseToPhysics() // not global: nothing
	pri.PoseFromPhysics()
}

// TestPlaneXZ tests a PlaneXZ joint, whose target positions move in the
// direction of an angle.
func TestPlaneXZ(t *testing.T) {
	ph := newTestPhysics()
	rot := math32.NewQuatIdentity()
	ob := ph.Builder.NewWorld().NewObject()
	bd := ob.NewDynamic(physics.Box, 1, math32.Vec3(0.5, 0.5, 0.5), math32.Vec3(0, 0.5, 0), rot)
	jd := ob.NewJointPlaneXZ(nil, bd, math32.Vec3(0, 0, 0), math32.Vec3(0, -0.5, 0))
	ph.Build()
	ph.InitState()
	jd.AddPlaneXZPos(0, 1, 100)
	assert.InDelta(t, 1, jd.DoF(0).Current.Pos, 1e-6)
	assert.InDelta(t, 0, jd.DoF(1).Current.Pos, 1e-6)
	jd.AddPlaneXZPos(math32.Pi/2, 2, 100)
	assert.InDelta(t, 1, jd.DoF(0).Current.Pos, 1e-5)
	assert.InDelta(t, 2, jd.DoF(1).Current.Pos, 1e-5)
}

// TestPose tests moving and rotating poses, and objects and their bodies.
func TestPose(t *testing.T) {
	var ps Pose
	ps.Defaults()
	assert.Equal(t, math32.NewQuatIdentity(), ps.Quat)
	ps.MoveOnAxis(1, 0, 0, 2)
	assert.Equal(t, math32.Vec3(2, 0, 0), ps.Pos)
	ps.MoveOnAxisAbs(0, 2, 0, 1)
	assert.Equal(t, math32.Vec3(2, 1, 0), ps.Pos)
	ps.Move(math32.Vec3(-2, -1, 0))
	assert.Equal(t, math32.Vec3(0, 0, 0), ps.Pos)

	ps.SetEulerRotation(0, 90, 0)
	e := ps.EulerRotation()
	assert.InDelta(t, 90, e.Y, 1e-3)
	ps.SetEulerRotationRad(0, 0, 0)
	assert.InDelta(t, 0, ps.EulerRotationRad().Y, 1e-5)
	ps.SetAxisRotation(0, 1, 0, 90)
	assert.InDelta(t, 90, ps.EulerRotation().Y, 1e-3)
	ps.SetAxisRotationRad(0, 1, 0, 0)
	ps.RotateOnAxis(0, 1, 0, 90)
	assert.InDelta(t, 90, ps.EulerRotation().Y, 1e-3)
	ps.RotateOnAxisRad(0, 1, 0, -math32.Pi/2)
	assert.InDelta(t, 0, ps.EulerRotation().Y, 1e-3)
	ps.RotateEuler(0, 90, 0)
	ps.RotateEulerRad(0, -math32.Pi/2, 0)
	assert.InDelta(t, 0, ps.EulerRotation().Y, 1e-3)

	// a moved pose, rotated around the origin
	ps = Pose{Pos: math32.Vec3(1, 0, 0)}
	ps.Defaults()
	ps.RotateAround(math32.NewQuatAxisAngle(math32.Vec3(0, 1, 0), math32.Pi/2), math32.Vec3(0, 0, 0))
	assert.InDelta(t, 0, ps.Pos.X, 1e-5)
	assert.InDelta(t, -1, ps.Pos.Z, 1e-5)
	ps.Transform(math32.Vec3(0, 1, 0), math32.NewQuatIdentity())
	assert.InDelta(t, 1, ps.Pos.Y, 1e-5)

	rot := math32.NewQuatIdentity()
	ob := NewBuilder().NewWorld().NewObject()
	ob.NewBody(physics.Box, math32.Vec3(1, 1, 1), math32.Vec3(1, 0, 0), rot)
	ob.NewBody(physics.Box, math32.Vec3(1, 1, 1), math32.Vec3(3, 0, 0), rot)
	ob.MoveOnAxisBody(0, 0, 1, 0, 2)
	assert.InDelta(t, 2, ob.Body(0).Pose.Pos.Y, 1e-5)
	assert.InDelta(t, 2, ob.Body(1).Pose.Pos.Y, 1e-5, "the whole object")
	ob.RotateOnAxisBody(0, 0, 1, 0, 180)
	assert.InDelta(t, -1, ob.Body(1).Pose.Pos.X, 1e-4, "around the first body")
	ob.RotateAroundBody(0, math32.NewQuatAxisAngle(math32.Vec3(0, 1, 0), math32.Pi))
	assert.InDelta(t, 3, ob.Body(1).Pose.Pos.X, 1e-4, "back")
}

// TestBuildScene tests building with a scene of skins for the bodies,
// replicated worlds, whose skins are cloned, and stepping with the scene.
func TestBuildScene(t *testing.T) {
	ph := newTestPhysics()
	ph.Scene = phyxyz.NewScene(xyz.NewScene())
	rot := math32.NewQuatIdentity()
	gw := ph.Builder.NewGlobalWorld()
	gw.NewObject().NewBodySkin(ph.Scene, "ground", physics.Plane, "grey", math32.Vec3(10, 0, 10), math32.Vec3(0, 0, 0), rot)
	wl := ph.Builder.NewWorld()
	ob := wl.NewObject()
	bd := ob.NewDynamicSkin(ph.Scene, "box", physics.Box, "red", 1, math32.Vec3(0.5, 0.5, 0.5), math32.Vec3(0, 3, 0), rot)
	require.NotNil(t, bd.Skin)
	ph.Builder.ReplicateWorld(ph.Scene, 1, 1, 2)
	assert.Equal(t, 2, ph.Builder.ReplicasN)
	assert.Len(t, ph.Scene.Skins, 3, "the box is cloned for the second world")

	ph.Build()
	ph.InitState()
	assert.True(t, ph.Scene.ReplicasView)
	ph.Step(5)
	assert.Equal(t, bd.BodyIndex, bd.Skin.BodyIndex)
	wl.PoseToPhysics()
	wl.Move(math32.Vec3(1, 0, 0))
	assert.InDelta(t, 1, bd.Pose.Pos.X, 1e-6)
}

// TestPoseBeforeFirstStep tests that a pose set in the physics before the
// first step is kept by it.
func TestPoseBeforeFirstStep(t *testing.T) {
	ph := newTestPhysics()
	physics.GetParams(0).Gravity.Y = 0
	ob := ph.Builder.NewWorld().NewObject()
	bd := ob.NewDynamic(physics.Box, 1, math32.Vec3(0.5, 0.5, 0.5), math32.Vec3(0, 0, 0), math32.NewQuatIdentity())
	ph.Build()
	ph.InitState()
	bd.Pose.Pos = math32.Vec3(5, 5, 5)
	bd.PoseToPhysics()
	ph.StepQuiet(1)
	bd.PoseFromPhysics()
	assert.Equal(t, math32.Vec3(5, 5, 5), bd.Pose.Pos)
}
