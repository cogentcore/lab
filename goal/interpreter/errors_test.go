// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// goalErrorTests are Goal code with errors, of the kinds that users make,
// for checking that the errors are reported at the right place in the
// Goal code, with a message that makes sense.
var goalErrorTests = []struct {
	name, src string
}{
	// Go code: syntax
	{"go-missing-paren", "x := 1\nfmt.Println(x"},
	{"go-extra-paren", "x := 1\nfmt.Println(x))"},
	{"go-missing-brace", "x := 1\nif x > 0 {\n\tx = 2\n"},
	{"go-unterminated-string", "s := \"abc\nfmt.Println(s)"},
	{"go-bad-operator", "x := 1\ny := x +* 2"},
	{"go-missing-operand", "x := 1\ny := x +"},

	// Go code: types and names
	{"go-undefined", "x := 1\ny := x + z\nfmt.Println(y)"},
	{"go-undefined-func", "x := 1\ny := double(x)\nfmt.Println(y)"},
	{"go-mismatched-types", "a := 10\nb := 12.5\nc := a + b\nfmt.Println(c)"},
	{"go-assign-string", "x := 1\nx = \"one\""},
	{"go-too-many-args", "f := func(a int) int { return a }\nfmt.Println(f(1, 2))"},
	{"go-too-few-args", "f := func(a, b int) int { return a + b }\nfmt.Println(f(1))"},
	{"go-not-a-function", "x := 1\nx(2)"},
	{"go-no-field", "type P struct{ X int }\np := P{X: 1}\nfmt.Println(p.Y)"},
	{"go-no-new-vars", "x := 1\nx := 2"},
	{"go-return-type", "f := func() int {\n\treturn \"a\"\n}\nfmt.Println(f())"},

	// Go code: run time
	{"go-index-range", "s := []int{1, 2}\ni := 5\nfmt.Println(s[i])"},
	{"go-nil-map", "var m map[string]int\nm[\"a\"] = 1"},
	{"go-nil-pointer", "type P struct{ X int }\nvar p *P\nfmt.Println(p.X)"},
	{"go-divide-zero", "a := 1\nb := 0\nfmt.Println(a / b)"},
	{"go-type-assert", "var x any = \"a\"\nn := x.(int)\nfmt.Println(n)"},
	{"go-panic", "panic(\"boom\")"},
	{"go-error-in-loop", "s := []int{1, 2}\nfor i := range 3 {\n\tfmt.Println(s[i])\n}"},

	// math mode: syntax
	{"math-unclosed-array", "##\na := [1, 2\n##"},
	{"math-unclosed-index", "##\na := zeros(3)\nb := a[1,\n##"},
	{"math-missing-operand", "##\na := zeros(3)\nb := a +\n##"},
	{"math-bad-operator", "##\na := zeros(3)\nb := a +* 2\n##"},
	{"math-unclosed-call", "##\na := zeros(3, 4\n##"},

	// math mode: names, types and shapes
	{"math-undefined", "##\na := zeros(3)\nb := a + c\n##"},
	{"math-unknown-func", "##\na := zeros(3)\nb := foo(a)\n##"},
	{"math-wrong-args", "##\na := full(3)\n##"},
	{"math-reshape-size", "##\na := zeros(2, 3)\nb := a.reshape(4, 4)\nfmt.Println(b)\n##"},
	{"math-shape-mismatch", "##\na := [1., 2., 3.]\nb := [1., 2.]\nc := a + b\nfmt.Println(c)\n##"},
	{"math-matmul-shapes", "##\na := zeros(2, 3)\nb := zeros(2, 3)\nc := a @ b\nfmt.Println(c)\n##"},
	{"math-index-range", "##\na := zeros(3)\nb := a[5]\nfmt.Println(b)\n##"},
	{"math-switch-tensor", "##\na := [1, 2]\nswitch a[0] {\ncase 1:\n\tfmt.Println(\"one\")\n}\n##"},
	{"math-range-map", "m := map[string]int{\"a\": 1}\n##\nfor k, v := range m {\n\tfmt.Println(k, v)\n}\n##"},
	{"math-go-int-float", "##\nvar n int = 3\nvar x float64 = 0.5\ny := n * x\nfmt.Println(y)\n##"},
	{"math-range-float", "##\nvar x float64 = 2.5\nfor i := range x {\n\tfmt.Println(i)\n}\n##"},
	{"math-string-tensor", "##\na := [\"a\", \"b\"]\nb := a * 2\nfmt.Println(b)\n##"},

	// inline math in Go code
	{"inline-math-syntax", "x := 1\nfmt.Println(# x + #)"},
	{"inline-math-undefined", "fmt.Println(# zeros(3) + q #)"},

	// shell commands
	{"exec-not-found", "notacommand -x"},
	{"exec-unterminated-quote", "echo \"abc"},
	{"exec-cd-missing", "cd /no/such/dir"},
	{"exec-bad-redirect", "echo abc >"},
}

// TestGoalErrors runs Goal code with errors, writing the errors to
// testdata/errors.want, which the test checks, to review them: run it
// with GOAL_ERRORS_UPDATE=1 to update the file.
func TestGoalErrors(t *testing.T) {
	var b strings.Builder
	for _, test := range goalErrorTests {
		var out syncBuffer
		in := NewInterpreter(Options{Stdout: &out, Stderr: &out})
		_, _, err := in.EvalCode(test.src)
		fmt.Fprintf(&b, "######## %s\n", test.name)
		for i, ln := range strings.Split(test.src, "\n") {
			fmt.Fprintf(&b, "%3d | %s\n", i+1, ln)
		}
		if err == nil {
			b.WriteString("-------- no error\n")
		} else {
			fmt.Fprintf(&b, "-------- error:\n%s\n", err.Error())
		}
		b.WriteString("\n")
	}
	if os.Getenv("GOAL_ERRORS_UPDATE") != "" {
		assert.NoError(t, os.WriteFile("testdata/errors.want", []byte(b.String()), 0666))
		return
	}
	want, err := os.ReadFile("testdata/errors.want")
	assert.NoError(t, err)
	assert.Equal(t, string(want), b.String())
}
