// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mparser

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// printNode returns the Go source of the given node.
func printNode(t *testing.T, n any) (s string) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			s = "unprintable"
		}
	}()
	var b bytes.Buffer
	require.NoError(t, printer.Fprint(&b, token.NewFileSet(), n))
	return b.String()
}

// normalize collapses the white space of the given code, as the printer
// lays out some code differently than it is written here.
func normalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// printStmts returns the Go source of the given statements, one per line.
func printStmts(t *testing.T, stmts []ast.Stmt) string {
	t.Helper()
	var s []string
	for _, st := range stmts {
		s = append(s, printNode(t, st))
	}
	return strings.Join(s, "\n")
}

// TestParseLineGo tests parsing lines of standard Go statements, which
// print back as the same code.
func TestParseLineGo(t *testing.T) {
	lines := []string{
		"x := 1",
		"x, y = y, x",
		"x += 2",
		"x++",
		"var x int",
		"var x, y = 1, \"a\"",
		"const c = 3",
		"type T struct {\n\tA int `json:\"a\"`\n\tB, C string\n\t*D\n}",
		"type I interface {\n\tM(a int) (b string, err error)\n\tE\n}",
		"type G[T any, U comparable] struct{ v T }",
		"type A = []map[string]chan<- *int",
		"f := func(a ...int) int { return len(a) }",
		"go f(1, 2)",
		"defer g()",
		"return",
		"if x := f(); x > 1 {\n\ty()\n} else if z {\n} else {\n}",
		"for i := 0; i < 10; i++ {\n\tcontinue\n}",
		"for k, v := range m {\n\tbreak\n}",
		"for range 3 {\n}",
		"for {\n}",
		"switch x := y.(type) {\ncase int, string:\n\tfallthrough\ndefault:\n}",
		"switch {\ncase a > b:\n}",
		"select {\ncase v := <-ch:\n\t_ = v\ncase ch <- 1:\ndefault:\n}",
		"L:\n\tfor {\n\t\tbreak L\n\t}",
		"goto L",
		"ch <- x",
		"v := <-ch",
		"p := &T{A: 1, B: []int{1, 2}}",
		"m := map[string]int{\"a\": 1}",
		"switch x {\ncase 1, 2:\n}",
		"m := map[int]string{1: \"a\"}",
		"x := y.(*T)",
		"var c chan<- int",
		"var d <-chan int",
		"var e [...]int",
		"x := struct{ a int }{1}",
		"x := interface{ M() }(nil)",
		"x := -a * (b + c) / d % e &^ f",
		"x := !a && b || c",
		"a, b := <-ch, ^x",
		"x := a.b.c",
		"a.b = 1",
		"type C interface {\n\t~int | ~float64\n\tString() string\n}",
		"var f func(int, string) error",
		"var g func(a, b int, c ...string) (d, e int)",
		"type S struct {\n\t*pkg.T\n\tpkg.U\n}",
		"type M interface{ pkg.I }",
		"x := (*T)(nil)",
		"x := []func(){}",
		"var (\n\ta = 1\n\tb string\n)",
		"type (\n\tA int\n\tB = A\n)",
	}
	for _, line := range lines {
		stmts, err := ParseLine(line, AllErrors|ParseComments)
		if assert.NoError(t, err, line) {
			assert.Equal(t, normalize(line), normalize(printStmts(t, stmts)), line)
		}
	}
}

// TestParseLineMultiple tests parsing several statements on a line.
func TestParseLineMultiple(t *testing.T) {
	stmts, err := ParseLine("a := 1; b := 2; c()", AllErrors)
	require.NoError(t, err)
	assert.Len(t, stmts, 3)
	stmts, err = ParseLine("x := 1\n", AllErrors)
	require.NoError(t, err)
	assert.Len(t, stmts, 1)
}

// TestParseLineMath tests the math extensions of the parser: full array
// list expressions, and ellipses and several indexes in index expressions.
func TestParseLineMath(t *testing.T) {
	stmts, err := ParseLine("x := [1, 2, 3]", AllErrors)
	require.NoError(t, err)
	as := stmts[0].(*ast.AssignStmt)
	il, ok := as.Rhs[0].(*ast.IndexListExpr)
	require.True(t, ok, "an array list")
	assert.Len(t, il.Indices, 3)

	stmts, err = ParseLine("x := [[1, 2], [3, 4]]", AllErrors)
	require.NoError(t, err)
	il = stmts[0].(*ast.AssignStmt).Rhs[0].(*ast.IndexListExpr)
	assert.Len(t, il.Indices, 2)

	for _, ex := range []string{"a[1, 2]", "a[:, 1]", "a[..., 1]", "a[1:3, ...]", "a[::2]"} {
		_, err := ParseExpr(ex, AllErrors)
		assert.NoError(t, err, ex)
	}
	// a slice in an index has no X (the X is the indexed expression)
	ix, err := ParseExpr("a[1:2:3]", AllErrors)
	require.NoError(t, err)
	sl := ix.(*ast.IndexExpr).Index.(*ast.IndexListExpr).Indices[0].(*ast.SliceExpr)
	assert.Nil(t, sl.X)
	assert.Equal(t, "1", sl.Low.(*ast.BasicLit).Value)
	assert.True(t, sl.Slice3)

	ix, err = ParseExpr("a[1, 2, 3]", AllErrors)
	require.NoError(t, err)
	assert.Len(t, ix.(*ast.IndexExpr).Index.(*ast.IndexListExpr).Indices, 3)
}

// TestParseExpr tests parsing expressions, which print back as the
// same code.
func TestParseExpr(t *testing.T) {
	exprs := []string{
		"a + b*c",
		"f(x, y...)",
		"[]int{1, 2}",
		"func(x int) bool { return x > 0 }",
		"x.(T)",
		"*p",
		"&x",
		"<-ch",
		"map[K]V{}",
		"'c'",
		"1.5e3",
		"0x1f",
		"2i",
		"`raw`",
	}
	for _, ex := range exprs {
		e, err := ParseExpr(ex, AllErrors)
		if assert.NoError(t, err, ex) {
			assert.Equal(t, normalize(ex), normalize(printNode(t, e)), ex)
		}
	}
}

// TestParseErrors tests that invalid code returns errors, and that the
// start of an incomplete statement (as typed so far in math mode) is
// allowed in some places.
func TestParseErrors(t *testing.T) {
	bad := []string{
		"x := ",
		"if {",
		"for i := 0; i < ; i++ {}",
		"a := )",
		"func(",
		"x := [1, 2",
		"type",
		"switch x { case }",
		"var 1 = 2",
		"a := := b",
	}
	for _, line := range bad {
		_, err := ParseLine(line, AllErrors)
		assert.Error(t, err, line)
	}
	_, err := ParseExpr("a +", AllErrors)
	assert.Error(t, err)
	_, err = ParseExpr("a b", AllErrors)
	assert.Error(t, err)
	_, err = ParseExpr("a +", 0)
	assert.Error(t, err)
}

// TestParseComments tests parsing with comments, and the trace mode
// (which prints the productions).
func TestParseComments(t *testing.T) {
	stmts, err := ParseLine("x := 1 // the x", ParseComments)
	require.NoError(t, err)
	assert.Len(t, stmts, 1)
	_, err = ParseLine("/* a */ y := 2", ParseComments|DeclarationErrors)
	assert.NoError(t, err)
	_, err = ParseLine("z := f(a.b, [1, 2])", Trace)
	assert.NoError(t, err)
	_, err = ParseExpr("a +", Trace)
	assert.Error(t, err)
}
