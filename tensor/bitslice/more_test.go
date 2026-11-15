// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitslice

import (
	"strings"
	"testing"
)

// expectPanic fails the test if fun does not panic.
func expectPanic(t *testing.T, name string, fun func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s did not panic", name)
		}
	}()
	fun()
}

// fromBools returns a new Slice with the given values.
func fromBools(vals ...bool) Slice {
	bs := Make(len(vals), 0)
	for i, v := range vals {
		bs.Set(v, i)
	}
	return bs
}

// equalBools reports whether the bits of bs are the given values.
func equalBools(bs Slice, vals ...bool) bool {
	bb := bs.ToBools()
	if len(bb) != len(vals) {
		return false
	}
	for i := range bb {
		if bb[i] != vals[i] {
			return false
		}
	}
	return true
}

// TestMakeCap tests the length and capacity of new slices.
func TestMakeCap(t *testing.T) {
	bs := Make(10, 0)
	if bs.Len() != 10 || bs.Cap() != 16 {
		t.Errorf("Make(10, 0): len %d cap %d, want 10, 16", bs.Len(), bs.Cap())
	}
	bs = Make(10, 100)
	if bs.Len() != 10 || bs.Cap() != 104 {
		t.Errorf("Make(10, 100): len %d cap %d, want 10, 104", bs.Len(), bs.Cap())
	}
	bs = Make(0, 0)
	if bs.Len() != 0 || len(bs) != 1 {
		t.Errorf("Make(0, 0): len %d bytes %d, want 0, 1", bs.Len(), len(bs))
	}
	var nb Slice
	if nb.Len() != 0 {
		t.Errorf("nil Len: %d", nb.Len())
	}
}

// TestAppendNil tests appending to a nil slice, and across byte boundaries.
func TestAppendNil(t *testing.T) {
	var bs Slice
	bs.Append(true)
	if !equalBools(bs, true) {
		t.Errorf("Append to nil: %v", bs.String())
	}
	var want []bool
	want = append(want, true)
	for i := range 20 {
		v := i%3 == 0
		bs.Append(v)
		want = append(want, v)
		if !equalBools(bs, want...) {
			t.Fatalf("after %d appends: %v", i+2, bs.String())
		}
	}
}

// TestSetLenShrinkGrow tests shrinking and growing within capacity.
func TestSetLenShrinkGrow(t *testing.T) {
	bs := Make(20, 0)
	bs.Set(true, 3)
	bs.SetLen(5)
	if bs.Len() != 5 || !bs.Index(3) {
		t.Errorf("SetLen(5): %v", bs.String())
	}
	bs.SetLen(16)
	if bs.Len() != 16 || !bs.Index(3) {
		t.Errorf("SetLen(16): %v", bs.String())
	}
	bs.SetLen(0)
	if bs.Len() != 0 {
		t.Errorf("SetLen(0): %d", bs.Len())
	}
}

// TestSetAllClone tests setting all bits, and that a clone is independent.
func TestSetAllClone(t *testing.T) {
	bs := Make(12, 0)
	bs.SetAll(true)
	for i := range 12 {
		if !bs.Index(i) {
			t.Errorf("SetAll(true): bit %d is false", i)
		}
	}
	cl := bs.Clone()
	bs.SetAll(false)
	for i := range 12 {
		if bs.Index(i) {
			t.Errorf("SetAll(false): bit %d is true", i)
		}
		if !cl.Index(i) {
			t.Errorf("Clone: bit %d is false", i)
		}
	}
	if cl.Len() != 12 {
		t.Errorf("Clone len: %d", cl.Len())
	}
}

// TestSubSlice tests the default end, empty results and the panics.
func TestSubSlice(t *testing.T) {
	bs := fromBools(true, false, true, true, false)
	if ss := bs.SubSlice(2, 0); !equalBools(ss, true, true, false) {
		t.Errorf("SubSlice(2, 0): %v", ss.String())
	}
	if ss := bs.SubSlice(3, 3); ss.Len() != 0 {
		t.Errorf("SubSlice(3, 3): %v", ss.String())
	}
	expectPanic(t, "SubSlice end > len", func() { bs.SubSlice(0, 6) })
	expectPanic(t, "SubSlice start > end", func() { bs.SubSlice(4, 2) })
}

// TestDeleteInsert tests deleting and inserting at the ends,
// deleting everything, and the panics.
func TestDeleteInsert(t *testing.T) {
	bs := fromBools(true, false, true, true, false)
	if ds := bs.Delete(0, 1); !equalBools(ds, false, true, true, false) {
		t.Errorf("Delete(0, 1): %v", ds.String())
	}
	if ds := bs.Delete(3, 2); !equalBools(ds, true, false, true) {
		t.Errorf("Delete(3, 2): %v", ds.String())
	}
	if ds := bs.Delete(0, 5); ds.Len() != 0 {
		t.Errorf("Delete(0, 5): %v", ds.String())
	}
	expectPanic(t, "Delete n <= 0", func() { bs.Delete(0, 0) })
	expectPanic(t, "Delete start >= len", func() { bs.Delete(5, 1) })
	expectPanic(t, "Delete end > len", func() { bs.Delete(3, 3) })

	if is := bs.Insert(5, 2); !equalBools(is, true, false, true, true, false, false, false) {
		t.Errorf("Insert(5, 2): %v", is.String())
	}
	if is := bs.Insert(0, 1); !equalBools(is, false, true, false, true, true, false) {
		t.Errorf("Insert(0, 1): %v", is.String())
	}
	expectPanic(t, "Insert n <= 0", func() { bs.Insert(0, 0) })
	expectPanic(t, "Insert start > len", func() { bs.Insert(6, 1) })
}

// TestString tests the strings of nil, empty, line-wrapped and long slices.
func TestString(t *testing.T) {
	var nb Slice
	if s := nb.String(); s != "nil" {
		t.Errorf("nil: %q", s)
	}
	eb := Slice{}
	if s := eb.String(); s != "[]" {
		t.Errorf("empty: %q", s)
	}
	bs := Make(81, 0)
	bs.Set(true, 80)
	want := "[" + strings.Repeat("0 ", 80) + "\n1 ]"
	if s := bs.String(); s != want {
		t.Errorf("81 bits: %q", s)
	}
	bs = Make(1001, 0)
	s := bs.String()
	if !strings.HasSuffix(s, "...(len=1001)]") || strings.Count(s, "0 ") != 1000 {
		t.Errorf("1001 bits: suffix %q", s[len(s)-20:])
	}
}
