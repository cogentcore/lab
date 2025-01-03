// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vector

import (
	"math"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

func TestVector(t *testing.T) {
	v := tensor.NewFloat64FromValues(1, 2, 3)
	ip := Mul(v, v).(*tensor.Float64)
	assert.Equal(t, []float64{1, 4, 9}, ip.Values)

	smv := Sum(ip).(*tensor.Float64)
	assert.Equal(t, 14.0, smv.Values[0])

	dpv := Dot(v, v).(*tensor.Float64)
	assert.Equal(t, 14.0, dpv.Values[0])

	nl2v := L2Norm(v).(*tensor.Float64)
	assert.Equal(t, math.Sqrt(14.0), nl2v.Values[0])

	nl1v := L1Norm(v).(*tensor.Float64)
	assert.Equal(t, 6.0, nl1v.Values[0])
}
