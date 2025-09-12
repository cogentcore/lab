// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cluster

import (
	"math"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/lab/tensor"
)

func init() {
	tensor.AddFunc(Min.FuncName(), MinFunc)
	tensor.AddFunc(Max.FuncName(), MaxFunc)
	tensor.AddFunc(Avg.FuncName(), AvgFunc)
	tensor.AddFunc(Contrast.FuncName(), ContrastFunc)
}

// Metrics are standard clustering distance metric functions,
// specifying how a node computes its distance based on its leaves.
type Metrics int32 //enums:enum

const (
	// Min is the minimum-distance or single-linkage weighting function.
	Min Metrics = iota

	// Max is the maximum-distance or complete-linkage weighting function.
	Max

	// Avg is the average-distance or average-linkage weighting function.
	Avg

	// Contrast computes maxd + (average within distance - average between distance).
	Contrast
)

// MetricFunc is a clustering distance metric function that evaluates aggregate distance
// between nodes, given the indexes of leaves in a and b clusters
// which are indexs into an ntot x ntot distance matrix dmat.
// maxd is the maximum distance value in the dmat, which is needed by the
// ContrastDist function and perhaps others.
type MetricFunc func(aix, bix []int, ntot int, maxd float64, dmat tensor.Tensor) float64

// FuncName returns the package-qualified function name to use
// in tensor.Call to call this function.
func (m Metrics) FuncName() string {
	return "cluster." + m.String()
}

// Func returns function for given metric.
func (m Metrics) Func() MetricFunc {
	fn := errors.Log1(tensor.FuncByName(m.FuncName()))
	return fn.Fun.(func(aix, bix []int, ntot int, maxd float64, dmat tensor.Tensor) float64)
}

// Call calls a standard Metrics enum function on given tensors.
// Output results are in the out tensor.
func (m Metrics) Call(aix, bix []int, ntot int, maxd float64, dmat tensor.Tensor) float64 {
	return m.Func()(aix, bix, ntot, maxd, dmat)
}

// MinFunc is the minimum-distance or single-linkage weighting function for comparing
// two clusters a and b, given by their list of indexes.
// ntot is total number of nodes, and dmat is the square similarity matrix [ntot x ntot].
func MinFunc(aix, bix []int, ntot int, maxd float64, dmat tensor.Tensor) float64 {
	md := math.MaxFloat64
	for _, ai := range aix {
		for _, bi := range bix {
			d := dmat.Float(ai, bi)
			if d < md {
				md = d
			}
		}
	}
	return md
}

// MaxFunc is the maximum-distance or complete-linkage weighting function for comparing
// two clusters a and b, given by their list of indexes.
// ntot is total number of nodes, and dmat is the square similarity matrix [ntot x ntot].
func MaxFunc(aix, bix []int, ntot int, maxd float64, dmat tensor.Tensor) float64 {
	md := -math.MaxFloat64
	for _, ai := range aix {
		for _, bi := range bix {
			d := dmat.Float(ai, bi)
			if d > md {
				md = d
			}
		}
	}
	return md
}

// AvgFunc is the average-distance or average-linkage weighting function for comparing
// two clusters a and b, given by their list of indexes.
// ntot is total number of nodes, and dmat is the square similarity matrix [ntot x ntot].
func AvgFunc(aix, bix []int, ntot int, maxd float64, dmat tensor.Tensor) float64 {
	md := 0.0
	n := 0
	for _, ai := range aix {
		for _, bi := range bix {
			d := dmat.Float(ai, bi)
			md += d
			n++
		}
	}
	if n > 0 {
		md /= float64(n)
	}
	return md
}

// ContrastFunc computes maxd + (average within distance - average between distance)
// for two clusters a and b, given by their list of indexes.
// avg between is average distance between all items in a & b versus all outside that.
// ntot is total number of nodes, and dmat is the square similarity matrix [ntot x ntot].
// maxd is the maximum distance and is needed to ensure distances are positive.
func ContrastFunc(aix, bix []int, ntot int, maxd float64, dmat tensor.Tensor) float64 {
	wd := AvgFunc(aix, bix, ntot, maxd, dmat)
	nab := len(aix) + len(bix)
	abix := append(aix, bix...)
	abmap := make(map[int]struct{}, ntot-nab)
	for _, ix := range abix {
		abmap[ix] = struct{}{}
	}
	oix := make([]int, ntot-nab)
	octr := 0
	for ix := 0; ix < ntot; ix++ {
		if _, has := abmap[ix]; !has {
			oix[octr] = ix
			octr++
		}
	}
	bd := AvgFunc(abix, oix, ntot, maxd, dmat)
	return maxd + (wd - bd)
}
