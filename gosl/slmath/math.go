// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

//gosl:start

const SLPi = 3.141592653589793

// MinAngleDiff returns the minimum difference between two angles
// (in radians): a-b, dealing with the wrap-around issues with angles.
func MinAngleDiff(a, b float32) float32 {
	d := a - b
	if d > SLPi {
		d -= 2 * SLPi
	}
	if d < -SLPi {
		d += 2 * SLPi
	}
	return d
}

//gosl:end
