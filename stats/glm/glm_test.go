// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glm

import (
	"math/rand"
	"strings"
	"testing"

	"cogentcore.org/lab/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// linearTable returns a table with n rows of two independent variables
// in column "X" and two dependent variables in column "Y":
// y0 = 2*x0 - x1 + 1, y1 = -x0 + 0.5*x1 - 2, plus prediction and error columns.
func linearTable(n int, noise float64) *table.Table {
	rnd := rand.New(rand.NewSource(1))
	dt := table.New()
	dt.AddFloat64Column("X", 2)
	dt.AddFloat64Column("Y", 2)
	dt.AddFloat64Column("Pred", 2)
	dt.AddFloat64Column("Err", 2)
	dt.AddFloat64Column("Bad", 3)
	dt.SetNumRows(n)
	x := dt.Column("X")
	y := dt.Column("Y")
	for i := range n {
		x0 := rnd.Float64()*4 - 2
		x1 := rnd.Float64()*4 - 2
		x.SetFloatRow(x0, i, 0)
		x.SetFloatRow(x1, i, 1)
		y.SetFloatRow(2*x0-x1+1+noise*rnd.NormFloat64(), i, 0)
		y.SetFloatRow(-x0+0.5*x1-2+noise*rnd.NormFloat64(), i, 1)
	}
	return dt
}

// TestGLMFit tests fitting two dependent variables from two independent
// variables with known coefficients, including the prediction and
// error outputs and the summary strings.
func TestGLMFit(t *testing.T) {
	dt := linearTable(100, 0)
	glm := NewGLM()
	glm.MaxIters = 200
	glm.LRate = 0.5 // the rate halves every 10 iterations
	glm.StopTolerance = 1.0e-12
	require.NoError(t, glm.SetTable(dt, "X", "Y", "Pred", "Err"))
	assert.Equal(t, 2, glm.NIndepVars)
	assert.Equal(t, 2, glm.NDepVars)
	glm.Run()

	exp := [][]float64{{2, -1, 1}, {-1, 0.5, -2}}
	for di := range 2 {
		for ii := range 3 {
			assert.InDelta(t, exp[di][ii], glm.Coeff.Float(di, ii), 0.02, "coeff %d %d", di, ii)
		}
		assert.Greater(t, glm.R2[di], 0.999)
	}
	assert.Less(t, glm.MSE, 1.0e-3)
	for i := range dt.NumRows() {
		for di := range 2 {
			y := dt.Column("Y").FloatRow(i, di)
			p := dt.Column("Pred").FloatRow(i, di)
			e := dt.Column("Err").FloatRow(i, di)
			assert.InDelta(t, y, p+e, 1.0e-9)
			assert.InDelta(t, y, p, 0.05)
		}
	}

	glm.DepNames[0] = "Y0"
	glm.IndepNames[0] = "X0"
	glm.IndepNames[1] = "X1"
	v := glm.Variance()
	assert.True(t, strings.HasPrefix(v, "Y0\tR^2: "), v)
	assert.Contains(t, v, "DV 1\tR^2: ")
	c := glm.Coeffs()
	lines := strings.Split(strings.TrimSpace(c), "\n")
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "Y0 = ")
	assert.Contains(t, lines[0], " * X0 + ")
	assert.Contains(t, lines[0], " * X1 + ")
	assert.Contains(t, lines[1], "DV 1 = ")
}

// TestGLMZeroOffsetNoise tests fitting with noise, a zero offset and
// no prediction or error columns, and a view of the table with indexes.
func TestGLMZeroOffsetNoise(t *testing.T) {
	dt := linearTable(200, 0.1)
	// remove the offsets so a zero offset fit is correct
	y := dt.Column("Y")
	for i := range dt.NumRows() {
		y.SetFloatRow(y.FloatRow(i, 0)-1, i, 0)
		y.SetFloatRow(y.FloatRow(i, 1)+2, i, 1)
	}
	dv := table.NewView(dt)
	dv.Indexes = nil
	for i := dt.NumRows() - 1; i >= 0; i -= 2 { // every other row, reversed
		dv.Indexes = append(dv.Indexes, i)
	}
	glm := NewGLM()
	glm.ZeroOffset = true
	glm.MaxIters = 200
	glm.LRate = 0.5 // the rate halves every 10 iterations
	glm.StopTolerance = 1.0e-12
	require.NoError(t, glm.SetTable(dv, "X", "Y", "", ""))
	glm.Run()
	assert.InDelta(t, 2, glm.Coeff.Float(0, 0), 0.05)
	assert.InDelta(t, -1, glm.Coeff.Float(0, 1), 0.05)
	assert.Equal(t, 0.0, glm.Coeff.Float(0, 2))
	assert.InDelta(t, -1, glm.Coeff.Float(1, 0), 0.05)
	assert.InDelta(t, 0.5, glm.Coeff.Float(1, 1), 0.05)
	assert.Equal(t, 0.0, glm.Coeff.Float(1, 2))
	assert.Greater(t, glm.R2[0], 0.99)
	assert.Greater(t, glm.R2[1], 0.95)
	assert.InDelta(t, 0.02, glm.MSE, 0.005) // noise variance, summed over the 2 DVs
}

// TestGLMCost tests that L1 and L2 costs shrink the coefficients.
func TestGLMCost(t *testing.T) {
	dt := linearTable(100, 0)
	fit := func(l1, l2 float64) float64 {
		glm := NewGLM()
		glm.MaxIters = 100
		glm.LRate = 0.5 // the rate halves every 10 iterations
		glm.StopTolerance = 1.0e-12
		glm.L1Cost = l1
		glm.L2Cost = l2
		require.NoError(t, glm.SetTable(dt, "X", "Y", "", ""))
		glm.Run()
		return glm.Coeff.Float(0, 0)
	}
	base := fit(0, 0)
	assert.Less(t, fit(20, 0), base-0.01)
	assert.Less(t, fit(0, 20), base-0.01)
}

// TestGLMSetTableErrors tests that SetTable checks the shapes of the
// prediction and error columns.
func TestGLMSetTableErrors(t *testing.T) {
	dt := linearTable(10, 0)
	glm := NewGLM()
	assert.Error(t, glm.SetTable(dt, "X", "Y", "Bad", ""))
	assert.Error(t, glm.SetTable(dt, "X", "Y", "", "Bad"))
}
