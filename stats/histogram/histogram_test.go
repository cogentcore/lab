// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package histogram

import (
	"testing"

	"cogentcore.org/lab/table"
	"github.com/stretchr/testify/assert"
)

// TestHistogram64 tests the counts of evenly spaced values in 3 bins.
func TestHistogram64(t *testing.T) {
	vals := []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1}
	ex := []float64{4, 3, 4}
	res := []float64{}

	F64(&res, vals, 3, 0, 1)

	assert.Equal(t, ex, res)
}

// TestHistogram64Range tests that values outside the range are ignored,
// that the max value goes in the last bin, and that the destination
// is resized and cleared.
func TestHistogram64Range(t *testing.T) {
	vals := []float64{-1, 0, 0.49, 0.5, 0.99, 1, 2}
	res := []float64{9, 9, 9, 9, 9}
	F64(&res, vals, 2, 0, 1)
	assert.Equal(t, []float64{2, 3}, res)

	F64(&res, nil, 4, 0, 1)
	assert.Equal(t, []float64{0, 0, 0, 0}, res)
}

// TestHistogram64Table tests the Value and Count columns of the table.
func TestHistogram64Table(t *testing.T) {
	vals := []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1}
	dt := table.New()
	dt.AddStringColumn("Old") // replaced
	F64Table(dt, vals, 3, 0, 1)
	assert.Equal(t, 2, dt.NumColumns())
	assert.Equal(t, 3, dt.NumRows())
	exvals := []float64{0, 0.3333, 0.6667}
	excounts := []float64{4, 3, 4}
	for ri := range 3 {
		assert.InDelta(t, exvals[ri], dt.Column("Value").Float1D(ri), 1e-12)
		assert.Equal(t, excounts[ri], dt.Column("Count").Float1D(ri))
	}
}
