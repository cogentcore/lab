// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"bytes"
	"strings"
	"testing"
)

// TestEval runs Goal code in the interpreter, line by line as in the
// interactive shell, checking what it prints.
func TestEval(t *testing.T) {
	var out bytes.Buffer
	in := NewInterpreter(Options{Stdout: &out, Stderr: &out})
	lines := []string{
		`x := 3`,
		`func add(a, b int) int {`,
		`	return a + b`,
		`}`,
		`fmt.Println(add(x, 4))`,
		`x := "redeclared"`,
		`fmt.Println(x)`,
		`s := $echo hello$`,
		`fmt.Println(strings.ToUpper(s))`,
		`# a := zeros(2, 3)`,
		`# b := a + 2`,
		`fmt.Println(b.Len(), b.Float1D(0))`,
		`fmt.Println(errors.Log1(5, nil))`,
		`t := tensor.NewFloat64(2, 2)`,
		`t.Set(1.5, 1, 1)`,
		`fmt.Println(t.Value(1, 1))`,
		`var w = []string{`,
		`	"a",`,
		`}`,
		`fmt.Println(len(w))`,
	}
	for _, line := range lines {
		if _, _, err := in.Eval(line); err != nil {
			t.Fatalf("%s: %v", line, err)
		}
	}
	want := "7\nredeclared\nHELLO\n6 2\n5\n1.5\n1\n"
	if got := out.String(); got != want {
		t.Errorf("output:\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(out.String(), "rror") {
		t.Error(out.String())
	}
}
