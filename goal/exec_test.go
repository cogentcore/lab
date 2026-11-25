// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	assert.Equal(t, "hi", NewGoal().Output("echo", "hi"))
}

// TestExpand tests that only the [ExpandWord] args of a command have their
// environment variables expanded, with those of the goal (set) first, and
// that the others, as the values of Go expressions, are passed as is.
func TestExpand(t *testing.T) {
	gl := NewGoal()
	gl.Config.Env["GOALFOO"] = "foo"
	t.Setenv("GOALBAR", "bar")
	assert.Equal(t, "foo bar", gl.Output("echo", gl.Expand("$GOALFOO"), gl.Expand("${GOALBAR}")))
	assert.Equal(t, "$GOALFOO", gl.Output("echo", "$GOALFOO"))
	// a script with $ in it, as in a Go value, gets its args
	script := `set -- a b; echo "$#: $@"`
	assert.Equal(t, "2: a b", gl.Output("sh", "-c", script))
	// an unset variable is an empty word, which is not passed
	assert.Equal(t, "0", gl.Output("sh", "-c", `echo $#`, gl.Expand("$GOALUNSET")))
	// and the command itself
	t.Setenv("GOALECHO", "echo")
	assert.Equal(t, "hi", gl.Output(gl.Expand("$GOALECHO"), "hi"))
}
