// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"testing"

	"cogentcore.org/core/gpu"
	"cogentcore.org/core/math32"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dropOn makes a model with a static ground plane and a dynamic body of
// the given shape above it, which it steps for 2 seconds, and returns the
// dynamic index of the body.
func dropOn(t *testing.T, ml *Model, shape Shapes) int32 {
	t.Helper()
	params := GetParams(0)
	params.Gravity.Y = -9.8
	rot := math32.NewQuatIdentity()
	gnd := ml.NewBody(Plane, math32.Vec3(10, 0, 10), math32.Vec3(0, 0, 0), rot)
	SetBodyFriction(gnd, 0.5)
	SetBodyBounce(gnd, 0)
	SetBodyFrictionTortion(gnd, 0.1)
	SetBodyFrictionRolling(gnd, 0.1)
	_, di := ml.NewDynamic(shape, 1, math32.Vec3(0.5, 0.5, 0.5), math32.Vec3(0, 2, 0), rot)
	ml.Config()
	for range 2000 {
		ml.Step()
	}
	return di
}

// TestContactShapes tests that bodies of each shape fall onto a ground
// plane and rest on it, through the contacts of their collisions.
func TestContactShapes(t *testing.T) {
	for _, shape := range []Shapes{Box, Sphere, Capsule, Cylinder} {
		t.Run(shape.String(), func(t *testing.T) {
			ml := testModel()
			di := dropOn(t, ml, shape)
			params := GetParams(0)
			pos := DynamicPos(di, params.Next)
			assert.Less(t, pos.Y, float32(2), "fell")
			assert.Greater(t, pos.Y, float32(0), "rests on the ground, not below it")
			if shape != Sphere { // (a sphere keeps bouncing high: see TestSphereBounce)
				assert.Less(t, ml.TotalKineticEnergy(), float32(0.5), "at rest")
			}
			assert.Equal(t, shape, GetBodyShape(DynamicBody(di)))
		})
	}
}

// stack makes a model with a static ground plane, a dynamic body of the
// first shape resting on it, and a dynamic body of the second shape above
// that, which it steps for 2 seconds, returning the dynamic index of the
// top body, and the height that it should rest at. (The half-height of a
// capsule is at least 1.01 times its radius, so the bottom body is placed
// at its actual half-height, not overlapping the ground, which would
// launch it.)
func stack(t *testing.T, bottom, top Shapes) (int32, float32) {
	t.Helper()
	ml := testModel()
	params := GetParams(0)
	params.Gravity.Y = -9.8
	rot := math32.NewQuatIdentity()
	hsz := math32.Vec3(0.5, 0.5, 0.5)
	gnd := ml.NewBody(Plane, math32.Vec3(10, 0, 10), math32.Vec3(0, 0, 0), rot)
	ba, _ := ml.NewDynamic(bottom, 1, hsz, math32.Vec3(0, 0.5, 0), rot)
	hb := BodyHSize(ba).Y
	SetBodyPos(ba, math32.Vec3(0, hb, 0))
	bb, db := ml.NewDynamic(top, 1, hsz, math32.Vec3(0, 2.5, 0), rot)
	SetBodyGroup(gnd, -1) // static: collides with the dynamic groups
	SetBodyGroup(ba, 1)   // collide with each other
	SetBodyGroup(bb, 1)
	SetBodyWorld(ba, 0)
	SetBodyWorld(bb, 0)
	ml.Config()
	for range 2000 {
		ml.Step()
	}
	return db, 2*hb + BodyHSize(bb).Y
}

// TestCollideBodies tests that a dynamic body that falls onto another
// dynamic body on the ground rests on top of it, for the pairs of shapes
// other than cylinders (which only collide with planes).
func TestCollideBodies(t *testing.T) {
	shapes := []Shapes{Box, Sphere, Capsule}
	for _, bottom := range shapes {
		for _, top := range shapes {
			db, want := stack(t, bottom, top)
			assert.InDelta(t, want, DynamicPos(db, GetParams(0).Next).Y, 0.02, "%v on a %v", top, bottom)
		}
	}
}

// TestStaticInverseInertia tests that a static body has no inverse mass
// or inverse inertia, so that it does not move in the contact constraints.
func TestStaticInverseInertia(t *testing.T) {
	ml := testModel()
	bi := ml.NewBody(Box, math32.Vec3(1, 1, 1), math32.Vec3(0, 0, 0), math32.NewQuatIdentity())
	assert.Equal(t, float32(0), Bodies.Value(int(bi), int(BodyInvMass)))
	assert.Equal(t, math32.Matrix3{}, BodyInvInertia(bi))
}

// TestSphereBounce tests that a sphere dropped on the ground stays above
// it. (It keeps bouncing almost as high as it was dropped from, unlike a
// box, which comes to rest, even with no bounce of the ground.)
func TestSphereBounce(t *testing.T) {
	ml := testModel()
	di := dropOn(t, ml, Sphere)
	params := GetParams(0)
	for range 1000 {
		ml.Step()
		assert.Greater(t, DynamicPos(di, params.Next).Y, float32(0.3))
	}
}

// TestBodyValues tests setting and getting the values of bodies.
func TestBodyValues(t *testing.T) {
	ml := testModel()
	rot := math32.NewQuatIdentity()
	bi := ml.NewBody(Box, math32.Vec3(1, 2, 3), math32.Vec3(0, 0, 0), rot)
	SetBodyThick(bi, 0.1)
	SetBodyCom(bi, math32.Vec3(0.1, 0.2, 0.3))
	assert.Equal(t, math32.Vec3(0.1, 0.2, 0.3), BodyCom(bi))
	SetBodyPos(bi, math32.Vec3(1, 2, 3))
	assert.Equal(t, math32.Vec3(1, 2, 3), BodyPos(bi))
	q := math32.NewQuatAxisAngle(math32.Vec3(0, 1, 0), 1)
	SetBodyQuat(bi, q)
	assert.InDelta(t, q.W, BodyQuat(bi).W, 1e-6)
	assert.Equal(t, Box, GetBodyShape(bi))
}

// TestContactValues tests setting and getting the values of contacts.
func TestContactValues(t *testing.T) {
	ml := testModel()
	dropOn(t, ml, Box) // so that the contacts are allocated
	SetContactA(0, 1)
	SetContactB(0, 2)
	SetContactPointIdx(0, 3)
	assert.Equal(t, int32(1), GetContactA(0))
	assert.Equal(t, int32(2), GetContactB(0))
	assert.Equal(t, int32(3), GetContactPointIdx(0))
	SetBroadContactA(0, 4)
	assert.Equal(t, int32(4), GetBroadContactA(0))
	v := math32.Vec3(1, 2, 3)
	for _, f := range []struct {
		set func(int32, math32.Vector3)
		get func(int32) math32.Vector3
	}{
		{SetContactAPoint, ContactAPoint}, {SetContactBPoint, ContactBPoint},
		{SetContactAOff, ContactAOff}, {SetContactBOff, ContactBOff},
		{SetContactNorm, ContactNorm},
		{SetContactADelta, ContactADelta}, {SetContactAAngDelta, ContactAAngDelta},
		{SetContactBDelta, ContactBDelta}, {SetContactBAngDelta, ContactBAngDelta},
	} {
		f.set(0, v)
		assert.Equal(t, v, f.get(0))
	}
}

// TestDynamicValues tests setting and getting the values of dynamic
// bodies, and the angular velocities and accelerations at points.
func TestDynamicValues(t *testing.T) {
	ml := testModel()
	rot := math32.NewQuatIdentity()
	_, di := ml.NewDynamic(Box, 1, math32.Vec3(1, 1, 1), math32.Vec3(0, 0, 0), rot)
	ml.Config()
	params := GetParams(0)
	SetDynamicDelta(di, params.Next, math32.Vec3(1, 0, 0))
	assert.Equal(t, math32.Vec3(1, 0, 0), DynamicDelta(di, params.Next))
	SetDynamicAngVel(di, params.Next, math32.Vec3(0, 2, 0))
	wp := AngularVelocityAt(di, math32.Vec3(1, 0, 0), math32.Vec3(0, 1, 0))
	assert.InDelta(t, -2, wp.Z, 1e-5, "rotation around Y at X shows in Z")
	SetDynamicAngAcc(di, params.Next, math32.Vec3(0, 3, 0))
	wa := AngularAccelAt(di, math32.Vec3(1, 0, 0), math32.Vec3(0, 1, 0))
	assert.InDelta(t, -3, wa.Z, 1e-5)
	assert.Greater(t, ml.TotalKineticEnergy(), float32(0), "rotating")
}

// TestJointControls tests setting and getting the target positions,
// angles, and velocities of joints, and their control forces.
func TestJointControls(t *testing.T) {
	ml := testModel()
	rot := math32.NewQuatIdentity()
	_, di := ml.NewDynamic(Box, 1, math32.Vec3(0.2, 0.2, 0.2), math32.Vec3(0, 0, 0), rot)
	ml.NewObject()
	ji := ml.NewJointRevolute(-1, di, math32.Vec3(0, 0, 0), math32.Vec3(0, 0, 0), math32.Vec3(0, 0, 1))
	ml.Config()
	SetJointTargetPos(ji, 0, 0.5, 100)
	assert.InDelta(t, 0.5, GetJointTargetPos(ji, 0), 1e-6)
	SetJointTargetAngle(ji, 0, 90, 100)
	assert.InDelta(t, math32.Pi/2, GetJointTargetPos(ji, 0), 1e-5)
	SetJointTargetVel(ji, 0, 2, 10)
	assert.InDelta(t, 2, GetJointTargetVel(ji, 0), 1e-6)
	SetJointControlForce(ji, 0, 3)
	assert.InDelta(t, 3, JointControl(ji, 0, JointControlForce), 1e-6)
	SetJointControl(ji, 0, JointControlForce, 4)
	assert.InDelta(t, 4, JointControl(ji, 0, JointControlForce), 1e-6)
	for range 100 {
		ml.Step()
	}
}

// TestGPU tests stepping a model on the GPU, whose results are those of
// the CPU.
func TestGPU(t *testing.T) {
	gpu.SkipIfNoGPU(t)
	cpu := testModel()
	dic := dropOn(t, cpu, Box)
	params := GetParams(0)
	want := DynamicPos(dic, params.Next)

	ml := NewModel()
	ml.GPU = true
	di := dropOn(t, ml, Box)
	ml.SetAsCurrent()
	got := DynamicPos(di, GetParams(0).Next)
	require.InDelta(t, want.Y, got.Y, 0.05, "the GPU is like the CPU")
}

// TestSetStateFirstStep tests that the state of a dynamic body set before
// the first step, in Next (where the state is between steps), is used by
// it: a position, and a delta (the velocity state of the integration;
// DynamicVel is computed from the change of position after each step).
func TestSetStateFirstStep(t *testing.T) {
	ml := testModel()
	params := GetParams(0)
	params.Gravity.Y = 0
	rot := math32.NewQuatIdentity()
	_, di := ml.NewDynamic(Sphere, 1, math32.Vec3(0.5, 0.5, 0.5), math32.Vec3(0, 0, 0), rot)
	ml.Config()
	assert.NotEqual(t, params.Cur, params.Next)
	SetDynamicPos(di, params.Next, math32.Vec3(5, 0, 0))
	SetDynamicDelta(di, params.Next, math32.Vec3(1, 0, 0))
	ml.Step()
	pos := DynamicPos(di, params.Next)
	assert.Greater(t, pos.X, float32(5), "moved from the position set, with the delta")
	assert.Greater(t, DynamicVel(di, params.Next).X, float32(0))
}

// TestCollideCapsules tests that a capsule (longer than its radius) that
// falls onto another standing capsule on the ground rests on it, end to
// end.
func TestCollideCapsules(t *testing.T) {
	ml := testModel()
	params := GetParams(0)
	params.Gravity.Y = -9.8
	rot := math32.NewQuatIdentity()
	hsz := math32.Vec3(0.5, 1, 0.5)
	gnd := ml.NewBody(Plane, math32.Vec3(10, 0, 10), math32.Vec3(0, 0, 0), rot)
	ba, _ := ml.NewDynamic(Capsule, 1, hsz, math32.Vec3(0, 1, 0), rot)
	bb, db := ml.NewDynamic(Capsule, 1, hsz, math32.Vec3(0, 4, 0), rot)
	SetBodyGroup(gnd, -1)
	SetBodyGroup(ba, 1)
	SetBodyGroup(bb, 1)
	ml.Config()
	for range 2000 {
		ml.Step()
	}
	assert.InDelta(t, 3, DynamicPos(db, params.Next).Y, 0.02)
}

// TestCollideWall tests that a sphere moving toward a wall, a plane that is
// moved and rotated from the ground, stops at it: the broad phase uses the
// position of the sphere in the frame of the plane.
func TestCollideWall(t *testing.T) {
	ml := testModel()
	params := GetParams(0)
	params.Gravity.Y = 0
	// a plane at x=3, facing -x (its normal, y, rotated to -x)
	wrot := math32.NewQuatAxisAngle(math32.Vec3(0, 0, 1), math32.Pi/2)
	wall := ml.NewBody(Plane, math32.Vec3(5, 0, 5), math32.Vec3(3, 0, 0), wrot)
	SetBodyGroup(wall, -1)
	_, di := ml.NewDynamic(Sphere, 1, math32.Vec3(0.5, 0.5, 0.5), math32.Vec3(0, 0, 0), math32.NewQuatIdentity())
	ml.Config()
	SetDynamicDelta(di, params.Next, math32.Vec3(2, 0, 0))
	maxX := float32(0)
	for range 3000 {
		ml.Step()
		maxX = max(maxX, DynamicPos(di, params.Next).X)
	}
	assert.InDelta(t, 2.5, maxX, 0.02, "stopped at the wall")
}
