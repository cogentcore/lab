// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"testing"

	"cogentcore.org/core/math32"
	"cogentcore.org/lab/physics"
	"github.com/stretchr/testify/assert"
)

func TestReplicate(t *testing.T) {
	rot := math32.NewQuatIdentity()
	bl := NewBuilder()
	wls := bl.NewGlobalWorld()
	ob := wls.NewObject()
	ob.NewBody(physics.Box, math32.Vec3(1, 2, 3), math32.Vec3(0, 2, 0), rot)
	ob.NewBody(physics.Capsule, math32.Vec3(1, 2, 3), math32.Vec3(2, 2, 0), rot)

	wld := bl.NewWorld()
	obd := wld.NewObject()
	bdd := obd.NewDynamic(physics.Box, 0.5, math32.Vec3(1, 2, 3), math32.Vec3(0, 2, 0), rot)
	bj := obd.NewJointPlaneXZ(nil, bdd, math32.Vec3(0, 0, 0), math32.Vec3(0, -2, 0))

	bl.ReplicateWorld(nil, 1, 4, 2)

	assert.Equal(t, 8, bl.ReplicasN)
	assert.Equal(t, 1, bl.ReplicasStart)

	for wi, wl := range bl.Worlds {
		assert.Equal(t, wi, wl.WorldIndex)
		for oi, ob := range wl.Objects {
			assert.Equal(t, wi, ob.WorldIndex)
			assert.Equal(t, oi, ob.Object)
			for bi, bd := range ob.Bodies {
				assert.Equal(t, wi, bd.WorldIndex)
				assert.Equal(t, oi, bd.Object)
				assert.Equal(t, bi, bd.ObjectBody)
			}
		}
	}
	assert.Equal(t, wls, bl.World(0))
	assert.Equal(t, wld, bl.World(1))
	assert.Equal(t, wld, bl.ReplicaWorld(0))
	assert.Equal(t, obd, bl.ReplicaObject(obd, 0))
	assert.Equal(t, bdd, bl.ReplicaBody(bdd, 0))
	assert.Equal(t, bj, bl.ReplicaJoint(bj, 0))

	assert.Equal(t, 3, bl.ReplicaBody(bdd, 2).WorldIndex)
	assert.Equal(t, 0, bl.ReplicaBody(bdd, 2).Object)
	assert.Equal(t, 0, bl.ReplicaBody(bdd, 2).ObjectBody)
}
