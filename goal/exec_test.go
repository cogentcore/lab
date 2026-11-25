// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"os"
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

// TestEnvAssign tests NAME=value before a command: the variable has the
// value only for that command, with ~ and $ expanded, and the args of the
// command are expanded before, as in bash.
func TestEnvAssign(t *testing.T) {
	gl := NewGoal()
	home, _ := os.UserHomeDir()
	gl.Config.Env["GOALOLD"] = "old"
	assert.Equal(t, "new", gl.Output(gl.EnvVar("GOALOLD", "new"), "sh", "-c", `echo $GOALOLD`))
	assert.Equal(t, "old", gl.Config.Env["GOALOLD"])
	assert.Equal(t, home+"/x", gl.Output(gl.EnvVar("GOALNEW", "~/x"), "sh", "-c", `echo $GOALNEW`))
	_, has := gl.Config.Env["GOALNEW"]
	assert.False(t, has)
	assert.Equal(t, "old-a", gl.Output(gl.EnvVar("GOALX", gl.Expand("$GOALOLD-a")), "sh", "-c", `echo $GOALX`))
	assert.Equal(t, "old", gl.Output(gl.EnvVar("GOALOLD", "new"), "echo", gl.Expand("$GOALOLD")))
}
