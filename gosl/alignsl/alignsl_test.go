// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package alignsl

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

// testSrc has struct types that are and are not aligned for the GPU.
const testSrc = `package p

// Good is 16 bytes of 32 bit fields.
type Good struct {
	A float32
	B int32
	C uint32
	D float32
}

// Short is only 8 bytes.
type Short struct {
	A float32
	B float32
}

// BadField has an int field that is not 32 bits.
type BadField struct {
	A int
	B float32
	C float32
}

// Nested has a struct field at an offset of 4.
type Nested struct {
	A   float32
	G   Good
	pad float32
	pd1 float32
	pd2 float32
}

// Inner is used in Outer, but not listed, so it is checked from the stack.
type Inner struct {
	A float32
}

// Outer has the unlisted Inner struct field.
type Outer struct {
	I   Inner
	A   float32
	B   float32
	C   float32
}

// Unsupported has a slice field.
type Unsupported struct {
	S []float32
	A float32
	B float32
}

// MyVec is a vector type, which can have any total size.
type MyVec struct {
	X float32
	Y float32
}

// Empty has no fields.
type Empty struct{}

// U64 has a uint64, which is allowed.
type U64 struct {
	A uint64
	B uint64
}
`

// testPackage type checks testSrc into a package.
func testPackage(t *testing.T, src string) *packages.Package {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	require.NoError(t, err)
	sizes := types.SizesFor("gc", "amd64")
	conf := types.Config{Importer: importer.Default(), Sizes: sizes}
	pkg, err := conf.Check("p", fset, []*ast.File{f}, nil)
	require.NoError(t, err)
	return &packages.Package{Types: pkg, TypesSizes: sizes}
}

// structOf returns the struct type and name of the named type in pkg.
func structOf(t *testing.T, pkg *packages.Package, name string) *types.Struct {
	ob := pkg.Types.Scope().Lookup(name)
	require.NotNil(t, ob, name)
	return ob.Type().Underlying().(*types.Struct)
}

// TestCheckStruct tests the errors found in individual struct types.
func TestCheckStruct(t *testing.T) {
	pkg := testPackage(t, testSrc)
	all := map[string]bool{"Good": true, "Short": true, "BadField": true, "Nested": true, "Unsupported": true, "MyVec": true, "Empty": true, "U64": true}
	check := func(name string) (bool, []string) {
		cx := NewContext(pkg.TypesSizes, all)
		return CheckStruct(cx, structOf(t, pkg, name), name), cx.Errs
	}

	for _, nm := range []string{"Good", "MyVec", "Empty", "U64"} {
		hasErr, errs := check(nm)
		assert.False(t, hasErr, nm)
		assert.Empty(t, errs, nm)
	}

	hasErr, errs := check("Short")
	assert.True(t, hasErr)
	assert.Equal(t, []string{"Short", "    total size: 8 not even multiple of 16 -- needs 2 extra 32bit padding fields"}, errs)

	hasErr, errs = check("BadField")
	assert.True(t, hasErr)
	assert.Equal(t, "BadField", errs[0])
	assert.Contains(t, errs[1], "A:  basic type != [U]Int32 or Float32: int")
	assert.Len(t, errs, 2) // int is 8 bytes, so the total size is 16

	hasErr, errs = check("Nested")
	assert.True(t, hasErr)
	assert.Equal(t, []string{"Nested", "    G:  struct type: Good is not at mod-16 byte offset: 4"}, errs)

	hasErr, errs = check("Unsupported")
	assert.True(t, hasErr)
	assert.Equal(t, "Unsupported", errs[0])
	assert.Contains(t, errs[1], "S:  unsupported type: []float32")

	// not in the list of struct types to check
	cx := NewContext(pkg.TypesSizes, map[string]bool{})
	assert.False(t, CheckStruct(cx, structOf(t, pkg, "Short"), "Short"))
	assert.Empty(t, cx.Errs)

	// each struct is only checked once
	cx = NewContext(pkg.TypesSizes, all)
	st := structOf(t, pkg, "Short")
	assert.True(t, CheckStruct(cx, st, "Short"))
	assert.False(t, CheckStruct(cx, st, "Short"))
}

// TestCheckStack tests that struct fields of unlisted types are checked
// in a second pass.
func TestCheckStack(t *testing.T) {
	pkg := testPackage(t, testSrc)
	cx := NewContext(pkg.TypesSizes, map[string]bool{"Outer": true})
	assert.False(t, CheckStruct(cx, structOf(t, pkg, "Outer"), "Outer"))
	assert.Len(t, cx.Stack, 1)
	assert.True(t, CheckStack(cx))
	assert.Empty(t, cx.Stack)
	assert.Equal(t, []string{"Inner", "    total size: 4 not even multiple of 16 -- needs 3 extra 32bit padding fields"}, cx.Errs)
}

// TestCheckPackage tests checking a whole package, with and without errors.
func TestCheckPackage(t *testing.T) {
	pkg := testPackage(t, testSrc)
	assert.NoError(t, CheckPackage(pkg, map[string]bool{"Good": true, "MyVec": true}))

	err := CheckPackage(pkg, map[string]bool{"Good": true, "Short": true, "Outer": true})
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "WARNING: in struct type alignment checking")
	assert.Contains(t, msg, "Short\n    total size: 8")
	assert.Contains(t, msg, "Inner\n    total size: 4")
	assert.NotContains(t, msg, "Good")
}

// TestCheckScopeChildren tests that the child scopes are checked when
// a scope has no struct types, as for types declared in functions.
func TestCheckScopeChildren(t *testing.T) {
	src := `package p

func f() {
	type Local struct {
		A float32
	}
	var l Local
	_ = l
}
`
	pkg := testPackage(t, src)
	cx := NewContext(pkg.TypesSizes, map[string]bool{"Local": true})
	assert.True(t, CheckScope(cx, pkg.Types.Scope(), 0))
	assert.Equal(t, "Local", cx.Errs[0])
}

// TestTypeName tests the names of named and unnamed types.
func TestTypeName(t *testing.T) {
	pkg := testPackage(t, testSrc)
	assert.Equal(t, "Good", TypeName(pkg.Types.Scope().Lookup("Good").Type()))
	assert.Equal(t, "[]float32", TypeName(types.NewSlice(types.Typ[types.Float32])))
}
