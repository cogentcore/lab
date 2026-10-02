// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath

import (
	"testing"

	"cogentcore.org/core/math32"
)

func TestMinAngleDiff(t *testing.T) {
	eqf(t, math32.Pi, SLPi, "SLPi")
	for _, a := range testScalars {
		for _, b := range testScalars {
			eqf(t, math32.MinAngleDiff(a, b), MinAngleDiff(a, b), "MinAngleDiff")
		}
	}
	eqf(t, 0, MinAngleDiff(0, 0), "same angle")
	eqf(t, 0.2, MinAngleDiff(0.1, -0.1), "small difference")
	// wrap-around: just past +Pi is just past -Pi
	eqf(t, -0.2, MinAngleDiff(math32.Pi-0.1, -math32.Pi+0.1), "wraps at +Pi")
	eqf(t, 0.2, MinAngleDiff(-math32.Pi+0.1, math32.Pi-0.1), "wraps at -Pi")
}
