// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !mpi && !mpich

package mpi

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDummy tests the functions without MPI (the mpi build tag), which
// run as one process.
func TestDummy(t *testing.T) {
	assert.False(t, IsOn())
	Init()
	assert.NoError(t, InitThreadSafe())
	assert.Equal(t, 0, WorldRank())
	assert.Equal(t, 1, WorldSize())
	cm, err := NewComm(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, cm.Rank())
	assert.Equal(t, 1, cm.Size())
	assert.NoError(t, cm.Barrier())
	assert.NoError(t, cm.Abort())
	Finalize()
}

// TestDummyComm calls each method of the communicator without MPI,
// which do nothing and return no error.
func TestDummyComm(t *testing.T) {
	cm, err := NewComm([]int{0})
	require.NoError(t, err)
	v := reflect.ValueOf(cm)
	n := 0
	for i := range v.NumMethod() {
		m := v.Method(i)
		mt := m.Type()
		args := make([]reflect.Value, mt.NumIn())
		for a := range args {
			args[a] = reflect.Zero(mt.In(a))
		}
		out := m.Call(args)
		if len(out) == 1 && out[0].Type().Name() == "error" {
			assert.Nil(t, out[0].Interface(), v.Type().Method(i).Name)
			n++
		}
	}
	assert.Greater(t, n, 50)
}

// captureStdout returns what the given function prints to the standard
// output.
func captureStdout(t *testing.T, fun func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	old := os.Stdout
	os.Stdout = w
	fun()
	os.Stdout = old
	w.Close()
	b, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(b)
}

// TestPrintf tests printing from the root process, and from all of them,
// with the rank first.
func TestPrintf(t *testing.T) {
	assert.Equal(t, "printed 1\n", captureStdout(t, func() { Printf("%s %d\n", "printed", 1) }))
	assert.Equal(t, "P0: all 2\n", captureStdout(t, func() { AllPrintf("%s %d\n", "all", 2) }))
	assert.Equal(t, "a b\n", captureStdout(t, func() { Println("a", "b") }))
	assert.Equal(t, "P0:  a b\n", captureStdout(t, func() { AllPrintln("a", "b") }), "all of the values")
}
