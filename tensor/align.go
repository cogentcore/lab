// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import (
	"fmt"
	"slices"

	"cogentcore.org/core/base/errors"
)

// AlignShapes aligns the shapes of two tensors, a and b for a binary
// computation producing an output, returning the effective aligned shapes
// for a, b, and the output, all with the same number of dimensions.
// Alignment proceeds from the innermost dimension out, with 1s provided
// beyond the number of dimensions for a or b.
// The output has the max of the dimension sizes for each dimension.
// An error is returned if the rules of alignment are violated:
// each dimension size must be either the same, or one of them
// is equal to 1. This corresponds to the "broadcasting" logic of NumPy.
func AlignShapes(a, b Tensor) (as, bs, os *Shape, err error) {
	asz := a.ShapeSizes()
	bsz := b.ShapeSizes()
	an := len(asz)
	bn := len(bsz)
	n := max(an, bn)
	osizes := make([]int, n)
	asizes := make([]int, n)
	bsizes := make([]int, n)
	for d := range n {
		ai := an - 1 - d
		bi := bn - 1 - d
		oi := n - 1 - d
		ad := 1
		bd := 1
		if ai >= 0 {
			ad = asz[ai]
		}
		if bi >= 0 {
			bd = bsz[bi]
		}
		if ad != bd && ad != 1 && bd != 1 {
			err = fmt.Errorf("tensor.AlignShapes: output dimension %d does not align for a=%d b=%d: must be either the same or one of them is a 1", oi, ad, bd)
			return
		}
		od := max(ad, bd)
		osizes[oi] = od
		asizes[oi] = ad
		bsizes[oi] = bd
	}
	as = NewShape(asizes...)
	bs = NewShape(bsizes...)
	os = NewShape(osizes...)
	return
}

// WrapIndex1D returns the 1d flat index for given n-dimensional index
// based on given shape, where any singleton dimension sizes cause the
// resulting index value to remain at 0, effectively causing that dimension
// to wrap around, while the other tensor is presumably using the full range
// of the values along this dimension. See [AlignShapes] for more info.
func WrapIndex1D(sh *Shape, i ...int) int {
	nd := sh.NumDims()
	ai := slices.Clone(i)
	for d := range nd {
		if sh.DimSize(d) == 1 {
			ai[d] = 0
		}
	}
	return sh.IndexTo1D(ai...)
}

// broadcaster maps a 1D index into an output shape to the 1D indexes
// into the shapes of two inputs a and b that are aligned with it (see
// [AlignShapes]), without allocating, as [Shape.IndexFrom1D] and
// [WrapIndex1D] do. The common cases of the same shape and of a single
// value (scalar) are direct.
type broadcaster struct {
	osz        []int // output sizes
	astr, bstr []int // strides, 0 for broadcast dimensions
	a, b       bcast
	anyGeneral bool
}

// bcast is the mapping for one input of a [broadcaster].
type bcast int

const (
	bcastSame    bcast = iota // the same index as the output
	bcastScalar               // a single value: index 0
	bcastGeneral              // via the strides
)

// newBroadcaster returns a [broadcaster] for the given aligned
// input shapes and output shape.
func newBroadcaster(as, bs, os *Shape) *broadcaster {
	br := &broadcaster{osz: os.Sizes}
	olen := os.Len()
	mode := func(sh *Shape) bcast {
		switch sh.Len() {
		case olen:
			return bcastSame
		case 1:
			return bcastScalar
		}
		return bcastGeneral
	}
	br.a, br.b = mode(as), mode(bs)
	if br.a == bcastGeneral || br.b == bcastGeneral {
		br.anyGeneral = true
		br.astr = broadcastStrides(as)
		br.bstr = broadcastStrides(bs)
	}
	return br
}

// broadcastStrides returns the row-major strides of the given shape,
// with a stride of 0 for dimensions of size 1, which are broadcast.
func broadcastStrides(sh *Shape) []int {
	nd := len(sh.Sizes)
	str := make([]int, nd)
	s := 1
	for d := nd - 1; d >= 0; d-- {
		if sh.Sizes[d] != 1 {
			str[d] = s
		}
		s *= sh.Sizes[d]
	}
	return str
}

// index returns the 1D indexes into a and b for the given 1D output index.
func (br *broadcaster) index(idx int) (ai, bi int) {
	if br.anyGeneral {
		rem := idx
		for d := len(br.osz) - 1; d >= 0; d-- {
			c := rem % br.osz[d]
			rem /= br.osz[d]
			ai += c * br.astr[d]
			bi += c * br.bstr[d]
		}
	}
	switch br.a {
	case bcastSame:
		ai = idx
	case bcastScalar:
		ai = 0
	}
	switch br.b {
	case bcastSame:
		bi = idx
	case bcastScalar:
		bi = 0
	}
	return
}

// AlignForAssign ensures that the shapes of two tensors, a and b
// have the proper alignment for assigning b into a.
// Alignment proceeds from the innermost dimension out, with 1s provided
// beyond the number of dimensions for a or b.
// An error is returned if the rules of alignment are violated:
// each dimension size must be either the same, or b is equal to 1.
// This corresponds to the "broadcasting" logic of NumPy.
func AlignForAssign(a, b Tensor) (as, bs *Shape, err error) {
	asz := a.ShapeSizes()
	bsz := b.ShapeSizes()
	an := len(asz)
	bn := len(bsz)
	n := max(an, bn)
	asizes := make([]int, n)
	bsizes := make([]int, n)
	for d := range n {
		ai := an - 1 - d
		bi := bn - 1 - d
		oi := n - 1 - d
		ad := 1
		bd := 1
		if ai >= 0 {
			ad = asz[ai]
		}
		if bi >= 0 {
			bd = bsz[bi]
		}
		if ad != bd && bd != 1 {
			err = fmt.Errorf("tensor.AlignShapes: dimension %d does not align for a=%d b=%d: must be either the same or b is a 1", oi, ad, bd)
			return
		}
		asizes[oi] = ad
		bsizes[oi] = bd
	}
	as = NewShape(asizes...)
	bs = NewShape(bsizes...)
	return
}

// SplitAtInnerDims returns the sizes of the given tensor's shape
// with the given number of inner-most dimensions retained as is,
// and those above collapsed to a single dimension.
// If the total number of dimensions is < nInner the result is nil.
func SplitAtInnerDims(tsr Tensor, nInner int) []int {
	sizes := tsr.ShapeSizes()
	nd := len(sizes)
	if nd < nInner {
		return nil
	}
	rsz := make([]int, nInner+1)
	split := nd - nInner
	rows := sizes[:split]
	copy(rsz[1:], sizes[split:])
	nr := 1
	for _, r := range rows {
		nr *= r
	}
	rsz[0] = nr
	return rsz
}

// FloatAssignFunc sets a to a binary function of a and b float64 values.
func FloatAssignFunc(fun func(a, b float64) float64, a, b Tensor) error {
	as, bs, err := AlignForAssign(a, b)
	if err != nil {
		return err
	}
	alen := as.Len()
	br := newBroadcaster(as, bs, as)
	VectorizeThreaded(1, func(tsr ...Tensor) int { return alen },
		func(idx int, tsr ...Tensor) {
			_, bi := br.index(idx)
			tsr[0].SetFloat1D(fun(tsr[0].Float1D(idx), tsr[1].Float1D(bi)), idx)
		}, a, b)
	return nil
}

// StringAssignFunc sets a to a binary function of a and b string values.
func StringAssignFunc(fun func(a, b string) string, a, b Tensor) error {
	as, bs, err := AlignForAssign(a, b)
	if err != nil {
		return err
	}
	alen := as.Len()
	br := newBroadcaster(as, bs, as)
	VectorizeThreaded(1, func(tsr ...Tensor) int { return alen },
		func(idx int, tsr ...Tensor) {
			_, bi := br.index(idx)
			tsr[0].SetString1D(fun(tsr[0].String1D(idx), tsr[1].String1D(bi)), idx)
		}, a, b)
	return nil
}

// FloatBinaryFunc sets output to a binary function of a, b float64 values.
// The flops (floating point operations) estimate is used to control parallel
// threading using goroutines, and should reflect number of flops in the function.
// See [VectorizeThreaded] for more information.
func FloatBinaryFunc(flops int, fun func(a, b float64) float64, a, b Tensor) Tensor {
	return CallOut2Gen2(FloatBinaryFuncOut, flops, fun, a, b)
}

// FloatBinaryFuncOut sets output to a binary function of a, b float64 values.
func FloatBinaryFuncOut(flops int, fun func(a, b float64) float64, a, b Tensor, out Values) error {
	as, bs, os, err := AlignShapes(a, b)
	if err != nil {
		return err
	}
	out.SetShapeSizes(os.Sizes...)
	olen := os.Len()
	br := newBroadcaster(as, bs, os)
	VectorizeThreaded(flops, func(tsr ...Tensor) int { return olen },
		func(idx int, tsr ...Tensor) {
			ai, bi := br.index(idx)
			out.SetFloat1D(fun(tsr[0].Float1D(ai), tsr[1].Float1D(bi)), idx)
		}, a, b, out)
	return nil
}

// StringBinaryFunc sets output to a binary function of a, b string values.
func StringBinaryFunc(fun func(a, b string) string, a, b Tensor) Tensor {
	return CallOut2Gen1(StringBinaryFuncOut, fun, a, b)
}

// StringBinaryFuncOut sets output to a binary function of a, b string values.
func StringBinaryFuncOut(fun func(a, b string) string, a, b Tensor, out Values) error {
	as, bs, os, err := AlignShapes(a, b)
	if err != nil {
		return err
	}
	out.SetShapeSizes(os.Sizes...)
	olen := os.Len()
	br := newBroadcaster(as, bs, os)
	VectorizeThreaded(1, func(tsr ...Tensor) int { return olen },
		func(idx int, tsr ...Tensor) {
			ai, bi := br.index(idx)
			out.SetString1D(fun(tsr[0].String1D(ai), tsr[1].String1D(bi)), idx)
		}, a, b, out)
	return nil
}

// FloatFunc sets output to a function of tensor float64 values.
// The flops (floating point operations) estimate is used to control parallel
// threading using goroutines, and should reflect number of flops in the function.
// See [VectorizeThreaded] for more information.
func FloatFunc(flops int, fun func(in float64) float64, in Tensor) Values {
	return CallOut1Gen2(FloatFuncOut, flops, fun, in)
}

// FloatFuncOut sets output to a function of tensor float64 values.
func FloatFuncOut(flops int, fun func(in float64) float64, in Tensor, out Values) error {
	SetShapeFrom(out, in)
	n := in.Len()
	VectorizeThreaded(flops, func(tsr ...Tensor) int { return n },
		func(idx int, tsr ...Tensor) {
			tsr[1].SetFloat1D(fun(tsr[0].Float1D(idx)), idx)
		}, in, out)
	return nil
}

// FloatSetFunc sets tensor float64 values from a function,
// which gets the index. Must be parallel threadsafe.
// The flops (floating point operations) estimate is used to control parallel
// threading using goroutines, and should reflect number of flops in the function.
// See [VectorizeThreaded] for more information.
func FloatSetFunc(flops int, fun func(idx int) float64, a Tensor) error {
	n := a.Len()
	VectorizeThreaded(flops, func(tsr ...Tensor) int { return n },
		func(idx int, tsr ...Tensor) {
			tsr[0].SetFloat1D(fun(idx), idx)
		}, a)
	return nil
}

//////// Bool

// BoolStringsFunc sets boolean output value based on a function involving
// string values from the two tensors.
func BoolStringsFunc(fun func(a, b string) bool, a, b Tensor) *Bool {
	out := NewBool()
	errors.Log(BoolStringsFuncOut(fun, a, b, out))
	return out
}

// BoolStringsFuncOut sets boolean output value based on a function involving
// string values from the two tensors.
func BoolStringsFuncOut(fun func(a, b string) bool, a, b Tensor, out *Bool) error {
	as, bs, os, err := AlignShapes(a, b)
	if err != nil {
		return err
	}
	out.SetShapeSizes(os.Sizes...)
	olen := os.Len()
	br := newBroadcaster(as, bs, os)
	VectorizeThreaded(5, func(tsr ...Tensor) int { return olen },
		func(idx int, tsr ...Tensor) {
			ai, bi := br.index(idx)
			out.SetBool1D(fun(tsr[0].String1D(ai), tsr[1].String1D(bi)), idx)
		}, a, b, out)
	return nil
}

// BoolFloatsFunc sets boolean output value based on a function involving
// float64 values from the two tensors.
func BoolFloatsFunc(fun func(a, b float64) bool, a, b Tensor) *Bool {
	out := NewBool()
	errors.Log(BoolFloatsFuncOut(fun, a, b, out))
	return out
}

// BoolFloatsFuncOut sets boolean output value based on a function involving
// float64 values from the two tensors.
func BoolFloatsFuncOut(fun func(a, b float64) bool, a, b Tensor, out *Bool) error {
	as, bs, os, err := AlignShapes(a, b)
	if err != nil {
		return err
	}
	out.SetShapeSizes(os.Sizes...)
	olen := os.Len()
	br := newBroadcaster(as, bs, os)
	VectorizeThreaded(5, func(tsr ...Tensor) int { return olen },
		func(idx int, tsr ...Tensor) {
			ai, bi := br.index(idx)
			out.SetBool1D(fun(tsr[0].Float1D(ai), tsr[1].Float1D(bi)), idx)
		}, a, b, out)
	return nil
}

// BoolIntsFunc sets boolean output value based on a function involving
// int values from the two tensors.
func BoolIntsFunc(fun func(a, b int) bool, a, b Tensor) *Bool {
	out := NewBool()
	errors.Log(BoolIntsFuncOut(fun, a, b, out))
	return out
}

// BoolIntsFuncOut sets boolean output value based on a function involving
// int values from the two tensors.
func BoolIntsFuncOut(fun func(a, b int) bool, a, b Tensor, out *Bool) error {
	as, bs, os, err := AlignShapes(a, b)
	if err != nil {
		return err
	}
	out.SetShapeSizes(os.Sizes...)
	olen := os.Len()
	br := newBroadcaster(as, bs, os)
	VectorizeThreaded(5, func(tsr ...Tensor) int { return olen },
		func(idx int, tsr ...Tensor) {
			ai, bi := br.index(idx)
			out.SetBool1D(fun(tsr[0].Int1D(ai), tsr[1].Int1D(bi)), idx)
		}, a, b, out)
	return nil
}
