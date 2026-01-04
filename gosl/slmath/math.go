// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

//gosl:start

const Pi = 3.141592653589793

// MinAngleDiff returns the minimum difference between two angles
// (in radians): a-b
func MinAngleDiff(a, b float32) float32 {
	d := a - b
	if d > Pi {
		d -= 2 * Pi
	}
	if d < -Pi {
		d += 2 * Pi
	}
	return d
}

//gosl:end
