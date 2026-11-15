// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"slices"
	"strings"
	"testing"

	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pointsDist returns the L1 distance matrix for points on a line
// at 0, 1, 5, 7, 20, with labels a..e.
func pointsDist() (dmat, labels tensor.Tensor) {
	pts := tensor.NewFloat64(5, 1)
	copy(pts.Values, []float64{0, 1, 5, 7, 20})
	dmat = metric.Matrix(metric.L1Norm, pts)
	labels = tensor.NewStringFromValues("a", "b", "c", "d", "e")
	return
}

// leaves returns the sorted leaf indexes of the node.
func leaves(nn *Node) []int {
	ix := make([]int, 5)
	ctr := 0
	nn.Indexes(ix, &ctr)
	ix = ix[:ctr]
	slices.Sort(ix)
	return ix
}

// TestClusterMetrics tests the clustering of points on a line with each
// of the cluster metrics, for the merge distances and structure.
func TestClusterMetrics(t *testing.T) {
	dmat, labels := pointsDist()
	// min: {a,b}@1, {c,d}@2, {ab,cd}@4, {all}@13
	// max: {a,b}@1, {c,d}@2, {ab,cd}@7, {all}@20
	// avg: {a,b}@1, {c,d}@2, {ab,cd}@5.5, {all}@(20+19+15+13)/4
	want := map[Metrics]float64{Min: 13, Max: 20, Avg: (20.0 + 19 + 15 + 13) / 4}
	sub := map[Metrics]float64{Min: 4, Max: 7, Avg: 5.5}
	for m, top := range want {
		cl := Cluster(m, dmat, labels)
		require.Len(t, cl.Kids, 1, m.String())
		root := cl.Kids[0]
		assert.InDelta(t, top, root.Dist, 1.0e-9, m.String())
		require.Len(t, root.Kids, 2)
		var abcd *Node
		for _, k := range root.Kids {
			if k.IsLeaf() {
				assert.Equal(t, 4, k.Index)
			} else {
				abcd = k
			}
		}
		require.NotNil(t, abcd, m.String())
		assert.InDelta(t, sub[m], abcd.Dist, 1.0e-9, m.String())
		assert.Equal(t, []int{0, 1, 2, 3}, leaves(abcd))
	}
}

// TestClusterContrast tests the Contrast metric function on known values
// and that clustering with it includes every leaf.
func TestClusterContrast(t *testing.T) {
	dmat, labels := pointsDist()
	// within {a},{b} = 1; between {a,b} and {c,d,e}: (5+7+20+4+6+19)/6
	bd := (5.0 + 7 + 20 + 4 + 6 + 19) / 6
	assert.InDelta(t, 20+1-bd, ContrastFunc([]int{0}, []int{1}, 5, 20, dmat), 1.0e-9)
	assert.InDelta(t, 20+1-bd, Contrast.Call([]int{0}, []int{1}, 5, 20, dmat), 1.0e-9)

	cl := Cluster(Contrast, dmat, labels)
	require.Len(t, cl.Kids, 1)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, leaves(cl))
}

// TestClusterFuncs tests the Min, Max and Avg metric functions directly.
func TestClusterFuncs(t *testing.T) {
	dmat, _ := pointsDist()
	a := []int{0, 1}
	b := []int{2, 3}
	assert.Equal(t, 4.0, MinFunc(a, b, 5, 20, dmat))
	assert.Equal(t, 7.0, MaxFunc(a, b, 5, 20, dmat))
	assert.Equal(t, 5.5, AvgFunc(a, b, 5, 20, dmat))
	assert.Equal(t, 0.0, AvgFunc(nil, b, 5, 20, dmat))
	assert.Equal(t, 7.0, Max.Call(a, b, 5, 20, dmat))
	assert.Equal(t, "cluster.Max", Max.FuncName())
}

// TestSprint tests the string representation of a cluster tree,
// with and without labels.
func TestSprint(t *testing.T) {
	dmat, labels := pointsDist()
	cl := Cluster(Min, dmat, labels)
	s := cl.Sprint(labels, 0)
	assert.Equal(t, "\n0: \n\t13: \n\t\t4: \n\t\t\t2: d c \n\t\t\t1: b a e ", s)

	nl := cl.Sprint(nil, 0)
	assert.NotContains(t, nl, "a")
	assert.Equal(t, strings.Count(s, "\n"), strings.Count(nl, "\n")-5) // leaves are lines without labels
}
