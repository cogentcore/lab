// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package physics

import (
	"cogentcore.org/core/math32"
	"cogentcore.org/lab/gosl/slmath"
)

//gosl:start

// newton: geometry/kernels.py class GeoData

// GeomData contains all geometric data for narrow-phase collision.
type GeomData struct {
	BodyIdx int32

	Shape Shapes

	// MinSize is the min of the Size dimensions.
	MinSize float32

	Thickness float32

	// Radius is the effective radius for sphere-like elements (Sphere, Capsule, Cone)
	Radius float32

	Size math32.Vector3

	// World-to-Body transform
	// Position (R) (i.e., BodyPos)
	WtoBR math32.Vector3
	// Quaternion (Q) (i.e., BodyQuat)
	WtoBQ math32.Quat

	// Body-to-World transform (inverse)
	// Position (R)
	BtoWR math32.Vector3
	// Quaternion (Q)
	BtoWQ math32.Quat
}

func NewGeomData(bi, cni int32, shp Shapes) GeomData {
	var gd GeomData
	gd.BodyIdx = bi
	gd.Shape = shp
	gd.Size = BodySize(bi)
	gd.MinSize = min(gd.Size.X, gd.Size.Y)
	gd.MinSize = min(gd.MinSize, gd.Size.Z)
	gd.WtoBR = BodyDynamicPos(bi, cni)
	gd.WtoBQ = BodyDynamicQuat(bi, cni)
	var bwR math32.Vector3
	var bwQ math32.Quat
	slmath.SpatialTransformInverse(gd.WtoBR, gd.WtoBQ, &bwR, &bwQ)
	gd.BtoWR = bwR
	gd.BtoWQ = bwQ
	gd.Radius = 0
	if shp == Sphere || shp == Capsule { // todo: cone is separate
		gd.Radius = gd.Size.X
	}
	return gd
}

/////// Collision methods:
// note: have to pass a non-pointer arg as first arg, due to gosl issue.

func ColSphereSphere(cni int32, gdA *GeomData, gdB *GeomData, ptA, ptB, norm *math32.Vector3) float32 {
	*ptA = gdA.WtoBR
	*ptB = gdB.WtoBR
	diff := (*ptA).Sub(*ptB)
	*norm = slmath.Normal3(diff)
	return slmath.Dot3(diff, *norm)
}

//gosl:end
