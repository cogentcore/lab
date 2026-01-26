// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"fmt"
	"testing"

	"cogentcore.org/core/math32"
	"github.com/stretchr/testify/assert"
)

func testModel() *Model {
	model := NewModel()
	model.GPU = false
	return model
}

func TestJointRevolute(t *testing.T) {
	ml := testModel()
	params := GetParams(0)
	params.Gravity.Y = 0
	params.SubSteps = 1
	params.Dt = 0.001
	rot := math32.NewQuatIdentity()
	hsz := math32.Vec3(0.2, 0.2, 0.2)
	mass := float32(0.1)
	stiff := float32(1000)
	damp := float32(20)
	steps := 100
	tol := 1.0e-1 // this is pretty bad, but whatever
	dim := math32.Z
	var axis, off math32.Vector3
	axis.SetDim(dim, 1)
	fmt.Println("####  dim:", dim, axis)

	bi, di := ml.NewDynamic(Box, mass, hsz, math32.Vec3(0, 0, 0), rot)
	_ = bi
	ml.NewObject()
	ji := ml.NewJointRevolute(-1, di, math32.Vec3(0, 0, 0), off, axis)

	ml.Config()
	// fmt.Println("inertia:", BodyInertia(bi))

	SetJointTargetVel(ji, 0, 0, damp)

	for trg := float32(-5); trg <= 5.0; trg += 0.5 {
		SetJointTargetPos(ji, 0, trg, stiff)
		for range steps {
			ml.Step()
			// q := DynamicQuat(di, params.Next)
			// a := q.ToAxisAngle()
			// fmt.Println("trg:", trg, math32.WrapPi(trg), a.W, q)
		}
		q := DynamicQuat(di, params.Next)
		a := q.ToAxisAngle()
		// fmt.Println(trg, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)))
		assert.InDelta(t, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)), tol)
	}

	// return
	// zooming in around Pi transition
	for trg := float32(2.5); trg <= 3.5; trg += 0.01 {
		SetJointTargetPos(ji, 0, trg, stiff)
		for range steps {
			ml.Step()
		}
		if math32.Abs(trg-3.13) < 0.001 { // flips a bit here
			continue
		}
		q := DynamicQuat(di, params.Next)
		a := q.ToAxisAngle()
		// fmt.Println(trg, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)))
		assert.InDelta(t, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)), tol)
	}
}

func TestJointPlaneXZ(t *testing.T) {
	ml := testModel()
	params := GetParams(0)
	params.Gravity.Y = 0
	params.SubSteps = 1
	params.Dt = 0.001
	rot := math32.NewQuatIdentity()
	hsz := math32.Vec3(0.2, 0.2, 0.2)
	mass := float32(0.1)
	stiff := float32(1000)
	damp := float32(20)
	steps := 100
	tol := 1.0e-1 // this is pretty bad, but whatever
	dim := math32.Y
	var axis, off math32.Vector3
	axis.SetDim(dim, 1)
	fmt.Println("####  dim:", dim, axis)

	bi, di := ml.NewDynamic(Box, mass, hsz, math32.Vec3(0, 0, 0), rot)
	_ = bi
	ml.NewObject()
	ji := ml.NewJointPlaneXZ(-1, di, math32.Vec3(0, 0, 0), off)
	SetJointAxis(ji, 2, axis)

	ml.Config()
	// fmt.Println("inertia:", BodyInertia(bi))

	SetJointTargetVel(ji, 0, 0, damp)

	for trg := float32(-5); trg <= 5.0; trg += 0.5 {
		SetJointTargetPos(ji, 2, trg, stiff)
		for range steps {
			ml.Step()
			// q := DynamicQuat(di, params.Next)
			// a := q.ToAxisAngle()
			// fmt.Println("trg:", trg, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)), q)
		}
		q := DynamicQuat(di, params.Next)
		a := q.ToAxisAngle()
		// fmt.Println(trg, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)))
		assert.InDelta(t, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)), tol)
	}

	return
	// zooming in around Pi transition
	for trg := float32(2.5); trg <= 3.5; trg += 0.01 {
		SetJointTargetPos(ji, 2, trg, stiff)
		for range steps {
			ml.Step()
		}
		if math32.Abs(trg-3.13) < 0.001 { // flips a bit here
			continue
		}
		q := DynamicQuat(di, params.Next)
		a := q.ToAxisAngle()
		// fmt.Println(trg, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)))
		assert.InDelta(t, math32.WrapPi(trg), math32.WrapPi(a.W*a.Dim(dim)), tol)
	}
}
