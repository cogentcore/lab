// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"fmt"
	"math"
	"testing"

	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

// sameTree reports whether the two trees have the same structure,
// leaves, and distances (within a tolerance).
func sameTree(t *testing.T, a, b *Node, path string) {
	if !assert.Equal(t, len(a.Kids), len(b.Kids), path) {
		return
	}
	assert.Equal(t, a.Index, b.Index, path)
	assert.InDelta(t, a.Dist, b.Dist, 1e-9*max(1, math.Abs(a.Dist)), path)
	for i := range a.Kids {
		sameTree(t, a.Kids[i], b.Kids[i], fmt.Sprintf("%s/%d", path, i))
	}
}

// TestGlomGeneric tests that Glom gives the same trees as the direct
// implementation, for all of the metrics, including with ties.
func TestGlomGeneric(t *testing.T) {
	for _, ties := range []bool{false, true} {
		dmat := randDistMatrix(40)
		if ties { // integer distances, with many ties
			for i := range dmat.Len() {
				dmat.SetFloat1D(math.Round(dmat.Float1D(i)), i)
			}
		}
		for m := range MetricsN {
			fast := Glom(InitAllLeaves(40), m, dmat)
			slow := glomGeneric(InitAllLeaves(40), m, dmat)
			sameTree(t, fast, slow, fmt.Sprintf("%s ties=%v", m, ties))
		}
	}
	// predefined initial clusters
	dmat := randDistMatrix(30)
	mk := func() *Node {
		root := InitAllLeaves(30)
		root.Kids = append(root.Kids[3:], NewNode(root.Kids[0], NewNode(root.Kids[1], root.Kids[2], 0), 0))
		return root
	}
	for m := range MetricsN {
		sameTree(t, Glom(mk(), m, dmat), glomGeneric(mk(), m, dmat), "init "+m.String())
	}
	// a single cluster
	root := InitAllLeaves(1)
	assert.Equal(t, root, Glom(root, Min, tensor.NewFloat64(1, 1)))
}
