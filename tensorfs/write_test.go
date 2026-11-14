// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWrite tests the operations that change the filesystem by paths:
// MkdirAll, Rename, RemoveAll, and CopyPath, with paths that start with
// the name of the root, as those of a file tree of it do, or not.
func TestWrite(t *testing.T) {
	root, _ := NewDir("root")
	Value[float64](root.Dir("Sims"), "X", 3).SetFloat1D(1.5, 1)
	Value[float64](root.Dir("Sims/Run"), "Y", 2)

	require.NoError(t, root.MkdirAll("root/Sims/New/Deep", 0o755))
	d, err := root.DirAtPath("Sims/New/Deep")
	require.NoError(t, err)
	assert.Equal(t, "root/Sims/New/Deep", d.Path())
	require.NoError(t, root.MkdirAll("Sims/New", 0o755), "exists")
	assert.Error(t, root.MkdirAll("Sims/X/Z", 0o755), "not a directory")

	// rename a value, and move a directory, keeping its time
	require.NoError(t, root.Rename("root/Sims/X", "Sims/Z"))
	z, err := root.NodeAtPath("Sims/Z")
	require.NoError(t, err)
	assert.Equal(t, 1.5, z.Tensor.Float1D(1))
	_, err = root.NodeAtPath("Sims/X")
	assert.Error(t, err, "renamed")
	run, _ := root.NodeAtPath("Sims/Run")
	tm := run.ModTime()
	require.NoError(t, root.Rename("Sims/Run", "Sims/New/Run"))
	moved, err := root.NodeAtPath("Sims/New/Run/Y")
	require.NoError(t, err)
	assert.Equal(t, "root/Sims/New/Run/Y", moved.Path())
	assert.Equal(t, tm, run.ModTime())
	assert.ErrorIs(t, root.Rename("Sims/Z", "Sims/New"), fs.ErrExist)
	assert.Error(t, root.Rename("Sims/New", "Sims/New/Deep/New"), "into itself")
	assert.Error(t, root.Rename("root", "Other"), "the root")

	// copy a value and a directory, replacing
	require.NoError(t, root.CopyPath("Sims/Z2", "Sims/Z"))
	z2, _ := root.NodeAtPath("Sims/Z2")
	assert.Equal(t, 1.5, z2.Tensor.Float1D(1))
	z2.Tensor.SetFloat1D(3, 1)
	assert.Equal(t, 1.5, z.Tensor.Float1D(1), "a clone")
	require.NoError(t, root.CopyPath("Sims/Copy", "Sims/New"))
	cy, err := root.NodeAtPath("Sims/Copy/Run/Y")
	require.NoError(t, err)
	assert.Equal(t, "root/Sims/Copy/Run/Y", cy.Path(), "with the parents of the clones")
	require.NoError(t, root.CopyPath("Sims/Z2", "Sims/Z"))
	z2, _ = root.NodeAtPath("Sims/Z2")
	assert.Equal(t, 1.5, z2.Tensor.Float1D(1), "replaced")
	assert.Error(t, root.CopyPath("Sims/New/In", "Sims/New"), "into itself")

	// remove
	require.NoError(t, root.RemoveAll("Sims/Copy"))
	_, err = root.NodeAtPath("Sims/Copy/Run/Y")
	assert.Error(t, err)
	require.NoError(t, root.RemoveAll("Sims/Nothing"), "nothing to remove")
	assert.Error(t, root.RemoveAll("."), "the root")
	names, _ := root.Dir("Sims").Nodes()
	var ns []string
	for _, n := range names {
		ns = append(ns, n.Name())
	}
	assert.Equal(t, []string{"New", "Z", "Z2"}, ns)
}
