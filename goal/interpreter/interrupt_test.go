// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"testing"
	"time"

	"cogentcore.org/lab/goal/transpile"
	"github.com/stretchr/testify/assert"
)

// TestInterrupt checks that interrupting the interpreter, as Ctrl-C does,
// stops Go code that runs, with an error at its position in the Goal code.
func TestInterrupt(t *testing.T) {
	var out syncBuffer
	in := NewInterpreter(Options{Stdout: &out, Stderr: &out})
	for _, src := range []string{
		"x := 1\nfor i := range 10000000000 { }",
		"##\na := zeros(3)\nfor i := 0; i < 10000000000; i++ {\n\ta[0] += 1\n}\n##",
	} {
		timer := time.AfterFunc(100*time.Millisecond, in.Interp.Interrupt)
		_, _, err := in.EvalCode(src)
		timer.Stop()
		var te *transpile.Error
		if assert.ErrorAs(t, err, &te, src) {
			assert.Equal(t, "interrupted", te.Msg)
			assert.Greater(t, te.Line, 1)
		}
	}
	_, _, err := in.EvalCode("fmt.Println(\"after\")")
	assert.NoError(t, err)
	assert.Contains(t, out.String(), "after")
}
