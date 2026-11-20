// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

//go:generate core generate

import (
	"fmt"
	"math"
	"strings"

	"cogentcore.org/core/base/indent"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/tensor"
)

// todo: all of this data goes into the tensorfs
// Cluster makes a new dir, stuffs results in there!
// need a global "cwd" that it uses, so basically you cd
// to a dir, then call it.

// Node is one node in the cluster
type Node struct {
	// index into original distance matrix; only valid for for terminal leaves.
	Index int

	// Distance value for this node, i.e., how far apart were all the kids from
	// each other when this node was created. is 0 for leaf nodes
	Dist float64

	// ParDist is total aggregate distance from parents; The X axis offset at which our cluster starts.
	ParDist float64

	// Y is y-axis value for this node; if a parent, it is the average of its kids Y's,
	// otherwise it counts down.
	Y float64

	// Kids are child nodes under this one.
	Kids []*Node
}

// IsLeaf returns true if node is a leaf of the tree with no kids
func (nn *Node) IsLeaf() bool {
	return len(nn.Kids) == 0
}

// Sprint prints to string
func (nn *Node) Sprint(labels tensor.Tensor, depth int) string {
	if nn.IsLeaf() && labels != nil {
		return labels.String1D(nn.Index) + " "
	}
	var sv strings.Builder
	fmt.Fprintf(&sv, "\n%v%v: ", indent.Tabs(depth), nn.Dist)
	for _, kn := range nn.Kids {
		sv.WriteString(kn.Sprint(labels, depth+1))
	}
	return sv.String()
}

// Indexes collects all the indexes in this node
func (nn *Node) Indexes(ix []int, ctr *int) {
	if nn.IsLeaf() {
		ix[*ctr] = nn.Index
		(*ctr)++
	} else {
		for _, kn := range nn.Kids {
			kn.Indexes(ix, ctr)
		}
	}
}

// NewNode merges two nodes into a new node
func NewNode(na, nb *Node, dst float64) *Node {
	nn := &Node{Dist: dst}
	nn.Kids = []*Node{na, nb}
	return nn
}

// TODO: this call signature does not fit with standard
// not sure how one might pack Node into a tensor

// Cluster implements agglomerative clustering, based on a
// distance matrix dmat, e.g., as computed by [metric.Matrix] method,
// using a metric that increases in value with greater dissimilarity.
// labels provides an optional String tensor list of labels for the elements
// of the distance matrix.
// This calls InitAllLeaves to initialize the root node with all of the leaves,
// and then Glom to do the iterative agglomerative clustering process.
// If you want to start with pre-defined initial clusters,
// then call Glom with a root node so-initialized.
func Cluster(metric Metrics, dmat, labels tensor.Tensor) *Node {
	ntot := dmat.DimSize(0) // number of leaves
	root := InitAllLeaves(ntot)
	return Glom(root, metric, dmat)
}

// InitAllLeaves returns a standard root node initialized with all of the leaves.
func InitAllLeaves(ntot int) *Node {
	root := &Node{}
	root.Kids = make([]*Node, ntot)
	for i := range ntot {
		root.Kids[i] = &Node{Index: i}
	}
	return root
}

// Glom does the iterative agglomerative clustering,
// based on a raw similarity matrix as given,
// using a root node that has already been initialized
// with the starting clusters, which is all of the
// leaves by default, but could be anything if you want
// to start with predefined clusters.
// The distance matrix must be symmetric.
//
// At each step, the closest pair of clusters (the first one, so that
// ties are deterministic) is merged. The distances between clusters are
// computed from the leaves once, and then updated for each merge from
// those of the merged clusters (the Lance-Williams method): the min or
// max distance, or the sums of distances for Avg and Contrast, along
// with the within-cluster sums and the sums of distances to all leaves
// for Contrast. This is O(n^3) in simple operations, instead of
// O(n^3) metric calls (O(n^4) for Contrast).
func Glom(root *Node, metric Metrics, dmat tensor.Tensor) *Node {
	ntot := dmat.DimSize(0) // number of leaves
	nk := len(root.Kids)
	if nk < 2 {
		return root
	}
	// the distance matrix and its max
	dm := make([]float64, ntot*ntot)
	maxd := math.Inf(-1)
	for i := range ntot {
		for j := range ntot {
			d := dmat.Float(i, j)
			dm[i*ntot+j] = d
			if !math.IsNaN(d) {
				maxd = max(maxd, d)
			}
		}
	}

	// per cluster, in slots that are reused for merged clusters:
	// size, the sum of the distances of its leaves to all leaves
	// (rowSum), and to its own leaves (within).
	size := make([]float64, nk)
	rowSum := make([]float64, nk)
	within := make([]float64, nk)
	// pair has the min, max, or sum of the distances between
	// the leaves of each pair of clusters (by slot).
	pair := make([]float64, nk*nk)
	leaves := make([][]int, nk)
	ix := make([]int, ntot)
	for s, kn := range root.Kids {
		ctr := 0
		kn.Indexes(ix, &ctr)
		leaves[s] = append([]int(nil), ix[:ctr]...)
		size[s] = float64(ctr)
		for _, i := range leaves[s] {
			for j := range ntot {
				rowSum[s] += dm[i*ntot+j]
			}
		}
	}
	// clusterPair returns the pair value for the leaves of a and b.
	clusterPair := func(a, b []int) float64 {
		v := 0.0
		switch metric {
		case Min:
			v = math.MaxFloat64
		case Max:
			v = -math.MaxFloat64
		}
		for _, i := range a {
			for _, j := range b {
				d := dm[i*ntot+j]
				switch metric {
				case Min:
					v = min(v, d)
				case Max:
					v = max(v, d)
				default:
					v += d
				}
			}
		}
		return v
	}
	for a := range nk {
		for b := range a {
			v := clusterPair(leaves[a], leaves[b])
			pair[a*nk+b], pair[b*nk+a] = v, v
		}
		within[a] = clusterPair(leaves[a], leaves[a])
	}
	// dist returns the distance between the clusters in slots a and b.
	dist := func(a, b int) float64 {
		switch metric {
		case Min, Max:
			return pair[a*nk+b]
		}
		sab := pair[a*nk+b]
		wd := sab / (size[a] * size[b])
		if metric == Avg {
			return wd
		}
		nab := size[a] + size[b]
		bd := 0.0
		if nout := float64(ntot) - nab; nout > 0 {
			bd = (rowSum[a] + rowSum[b] - within[a] - within[b] - 2*sab) / (nab * nout)
		}
		return maxd + (wd - bd)
	}

	// slots has the slot of each kid, in the order of root.Kids.
	slots := make([]int, nk)
	for i := range slots {
		slots[i] = i
	}
	for {
		// the first closest pair is merged, so ties are deterministic
		na, nb := 0, 0
		mval := math.MaxFloat64
		for ai, sa := range slots {
			for bi := range ai {
				dv := dist(sa, slots[bi])
				if dv < mval {
					mval = dv
					na, nb = ai, bi
				}
			}
		}
		sa, sb := slots[na], slots[nb]
		nn := NewNode(root.Kids[na], root.Kids[nb], mval)
		// merge b into the slot of a
		for _, sk := range slots {
			if sk == sa || sk == sb {
				continue
			}
			pa, pb := pair[sa*nk+sk], pair[sb*nk+sk]
			v := pa + pb
			switch metric {
			case Min:
				v = min(pa, pb)
			case Max:
				v = max(pa, pb)
			}
			pair[sa*nk+sk], pair[sk*nk+sa] = v, v
		}
		within[sa] += within[sb] + 2*pair[sa*nk+sb]
		rowSum[sa] += rowSum[sb]
		size[sa] += size[sb]
		for i := len(root.Kids) - 1; i >= 0; i-- {
			if i == na || i == nb {
				root.Kids = append(root.Kids[:i], root.Kids[i+1:]...)
				slots = append(slots[:i], slots[i+1:]...)
			}
		}
		root.Kids = append(root.Kids, nn)
		slots = append(slots, sa)
		if len(root.Kids) == 1 {
			break
		}
	}
	return root
}

// glomGeneric is the direct implementation of [Glom], which computes
// the metric function on the leaves of every pair of clusters at each step:
// O(n^3) metric calls, and O(n^4) for Contrast. It is used in tests of Glom.
// Glom does the iterative agglomerative clustering,
// based on a raw similarity matrix as given,
// using a root node that has already been initialized
// with the starting clusters, which is all of the
// leaves by default, but could be anything if you want
// to start with predefined clusters.
func glomGeneric(root *Node, metric Metrics, dmat tensor.Tensor) *Node {
	ntot := dmat.DimSize(0) // number of leaves
	mout := tensor.NewFloat64Scalar(0)
	stats.MaxOut(tensor.As1D(dmat), mout)
	maxd := mout.Float1D(0)
	// indexes in each group
	aidx := make([]int, ntot)
	bidx := make([]int, ntot)
	for {
		// the first closest pair is merged, so ties are deterministic
		na, nb := 0, 0
		mval := math.MaxFloat64
		for ai, ka := range root.Kids {
			actr := 0
			ka.Indexes(aidx, &actr)
			aix := aidx[0:actr]
			for bi := range ai {
				kb := root.Kids[bi]
				bctr := 0
				kb.Indexes(bidx, &bctr)
				bix := bidx[0:bctr]
				dv := metric.Call(aix, bix, ntot, maxd, dmat)
				if dv < mval {
					mval = dv
					na, nb = ai, bi
				}
			}
		}
		nn := NewNode(root.Kids[na], root.Kids[nb], mval)
		for i := len(root.Kids) - 1; i >= 0; i-- {
			if i == na || i == nb {
				root.Kids = append(root.Kids[:i], root.Kids[i+1:]...)
			}
		}
		root.Kids = append(root.Kids, nn)
		if len(root.Kids) == 1 {
			break
		}
	}
	return root
}
