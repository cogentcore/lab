// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnique(t *testing.T) {
	s := NewStringFromValues("b", "a", "c", "a", "b", "a")
	assert.Equal(t, []string{"a", "b", "c"}, Unique(s).(*String).Values)
	assert.Equal(t, []int{3, 2, 1}, UniqueCounts(s).Values)
	_, groups := UniqueGroups(s)
	assert.Equal(t, [][]int{{1, 3, 5}, {0, 4}, {2}}, groups, "rows in order")

	// numbers keep their type, and NaN is missing data
	f := NewFloat64FromValues(3, math.NaN(), 1, 3, math.NaN(), 2)
	assert.Equal(t, []float64{1, 2, 3}, Unique(f).(*Float64).Values)
	assert.Equal(t, []int{1, 1, 2}, UniqueCounts(f).Values)
	i := NewIntFromValues(5, 5, 4)
	assert.Equal(t, []int{4, 5}, Unique(i).(*Int).Values)

	// all of the values of a 2D tensor, and the rows of a view
	assert.Equal(t, []float64{1, 2, 3}, Unique(Reshape(NewFloat64FromValues(3, 1, 2, 1), 2, 2)).(*Float64).Values)
	rw := NewRows(s, 2, 4)
	vals, groups := UniqueGroups(rw)
	assert.Equal(t, []string{"b", "c"}, vals.(*String).Values)
	assert.Equal(t, [][]int{{4}, {2}}, groups, "rows of the underlying data")
	assert.Equal(t, 0, Unique(NewFloat64(0)).Len())
}

func TestColumnByName(t *testing.T) {
	assert.Panics(t, func() { ColumnByName(NewFloat64(2), "x") })
	assert.Panics(t, func() { SetColumnByName(3, "x", 1) })
}

// TestRowsString tests that a Rows view prints only its rows.
func TestRowsString(t *testing.T) {
	s := NewStringFromValues("a", "b", "c", "d")
	str := NewRows(s, 1, 3).String()
	assert.Contains(t, str, "[2]")
	assert.NotContains(t, str, "c")
}
