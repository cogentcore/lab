// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package table

import (
	"math"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetColumn(t *testing.T) {
	dt := New()
	dt.AddFloat64Column("X")
	dt.SetNumRows(4)
	for i := range 4 {
		dt.Column("X").SetFloat1D(float64(i), i)
	}
	// a new column, as math mode does for dt["Y"] = x
	require.NoError(t, dt.SetColumn("Y", tensor.NewFloat64FromValues(10, 11, 12, 13)))
	assert.Equal(t, 12.0, dt.Column("Y").Float1D(2))
	// an existing column keeps its type, and a scalar sets all of the rows
	require.NoError(t, dt.SetColumn("X", tensor.NewIntScalar(7)))
	assert.Equal(t, 7.0, dt.Column("X").Float1D(3))
	require.NoError(t, dt.SetColumn("S", tensor.NewStringFromValues("a", "b", "c", "d")))
	assert.True(t, dt.Column("S").IsString())
	assert.Error(t, dt.SetColumn("Z", tensor.NewFloat64FromValues(1, 2)), "wrong number of rows")

	// a view sets the rows of the view, in the columns shared with its source
	vw := NewView(dt)
	_, err := vw.FilterMask(tensor.NewFloat64FromValues(0, 1, 0, 1))
	require.NoError(t, err)
	assert.Equal(t, 2, vw.NumRows())
	require.NoError(t, vw.SetColumn("V", tensor.NewFloat64FromValues(1, 3)))
	assert.True(t, math.IsNaN(dt.Column("V").Float1D(0)), "not in the view")
	assert.Equal(t, 3.0, dt.Column("V").Float1D(3))
	assert.Equal(t, []float64{1, 3}, []float64{vw.Column("V").Float1D(0), vw.Column("V").Float1D(1)})

	_, err = vw.FilterMask(tensor.NewFloat64FromValues(1))
	assert.Error(t, err)

	// the column by name, as math mode does for dt["Y"]
	assert.Equal(t, 11.0, tensor.ColumnByName(dt, "Y").Float1D(1))
	tensor.SetColumnByName(dt, "W", 2.5)
	assert.Equal(t, 2.5, dt.Column("W").Float1D(3))
	assert.Panics(t, func() { tensor.ColumnByName(dt, "Nope") })
}
