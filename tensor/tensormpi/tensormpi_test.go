// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensormpi

import (
	"testing"

	"cogentcore.org/lab/base/mpi"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The tests run without MPI (the mpi build tag), as one process, for
// which the gathers and reductions are copies.

// TestAllocN tests allocating items to the one process.
func TestAllocN(t *testing.T) {
	st, end, err := AllocN(10)
	require.NoError(t, err)
	assert.Equal(t, 0, st)
	assert.Equal(t, 10, end)
}

// TestGatherReduceTensor tests gathering the rows of, and reducing,
// tensors of each type.
func TestGatherReduceTensor(t *testing.T) {
	comm, err := mpi.NewComm(nil)
	require.NoError(t, err)
	for _, src := range []tensor.Values{
		tensor.NewFloat64FromValues(1, 2, 3),
		tensor.NewFloat32FromValues(1, 2, 3),
		tensor.NewIntFromValues(1, 2, 3),
		tensor.NewStringFromValues("a", "b", "c"),
	} {
		dest := tensor.Clone(src)
		dest.SetZeros()
		require.NoError(t, GatherTensorRows(dest, src, comm))
		assert.Equal(t, src.String1D(1), dest.String1D(1), "%T", src)

		dest = tensor.Clone(src)
		dest.SetZeros()
		require.NoError(t, ReduceTensor(dest, src, comm, mpi.OpSum))
		if src.DataType().String() != "string" {
			assert.Equal(t, src.Float1D(2), dest.Float1D(2), "%T", src)
		}
	}
	s := tensor.NewStringFromValues("x", "y")
	d := tensor.NewString(2)
	require.NoError(t, GatherTensorRowsString(d, s, comm))
	assert.Equal(t, []string{"x", "y"}, d.Values)
}

// TestGatherReduceTable tests gathering the rows of, and reducing, tables.
func TestGatherReduceTable(t *testing.T) {
	comm, err := mpi.NewComm(nil)
	require.NoError(t, err)
	src := table.New()
	src.AddFloat64Column("x")
	src.AddStringColumn("name")
	src.SetNumRows(2)
	src.Column("x").SetFloat1D(5, 1)
	src.Column("name").SetString1D("b", 1)

	dest := table.New()
	GatherTableRows(dest, src, comm)
	assert.Equal(t, 2, dest.NumRows())
	assert.Equal(t, 5.0, dest.Column("x").Float1D(1))
	assert.Equal(t, "b", dest.Column("name").String1D(1))

	red := table.New()
	ReduceTable(red, src, comm, mpi.OpSum)
	assert.Equal(t, 5.0, red.Column("x").Float1D(1))
}

// TestRandCheck tests that the random numbers of the one process agree.
func TestRandCheck(t *testing.T) {
	comm, err := mpi.NewComm(nil)
	require.NoError(t, err)
	assert.NoError(t, RandCheck(comm))
}
