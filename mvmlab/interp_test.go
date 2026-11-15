// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mvmlab

import (
	"bytes"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/mvmcore"
	"cogentcore.org/core/text/textcore"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoalInterpreter tests the Goal interpreter registered in
// mvmcore.Interpreters: Go code, Goal math, and widgets added to b.
func TestGoalInterpreter(t *testing.T) {
	f := mvmcore.Interpreters["Goal"]
	require.NotNil(t, f)
	var out bytes.Buffer
	in, err := f(&out)
	require.NoError(t, err)

	_, err = in.Eval(`fmt.Println("hi", 1+2)`)
	require.NoError(t, err)
	assert.Equal(t, "hi 3\n", out.String())
	assert.Nil(t, tensorfs.ListOutput) // restored after Eval

	out.Reset()
	_, err = in.Eval("##\nx := arange(4) * 2\n##\nfmt.Println(x.Float1D(3))")
	require.NoError(t, err)
	assert.Equal(t, "6\n", out.String())

	fr := core.NewFrame(core.NewBody())
	mvmcore.SetParent(fr)
	defer func() { core.ExternalParent = nil }()
	_, err = in.Eval(`core.NewButton(b).SetText("Hello")`)
	require.NoError(t, err)
	require.Equal(t, 1, fr.NumChildren())
	bt, ok := fr.Child(0).(*core.Button)
	require.True(t, ok)
	assert.Equal(t, "Hello", bt.Text)

	_, err = in.Eval(`undefinedName + 1`)
	assert.Error(t, err)
}

// TestBindTextEditor tests that the Goal code of a text editor runs when
// it is bound, with its output shown in the parent.
func TestBindTextEditor(t *testing.T) {
	mvmcore.ResetGoalInterpreter()
	defer mvmcore.ResetGoalInterpreter()
	b := core.NewBody()
	ed := textcore.NewEditor(b)
	ed.Lines.SetString(`fmt.Println("bound", 6*7)`)
	parent := core.NewFrame(b)
	BindTextEditor(ed, parent, "Goal")
	require.Equal(t, 1, parent.NumChildren())
	oed, ok := parent.Child(0).(*textcore.Editor)
	require.True(t, ok)
	assert.Equal(t, "bound 42\n", oed.Lines.String())
}

// TestFormatGoalErrors tests that FormatGoal keeps a missing final newline,
// and returns the source unchanged with an error for invalid code.
func TestFormatGoalErrors(t *testing.T) {
	res, err := FormatGoal("x:=1")
	assert.NoError(t, err)
	assert.Equal(t, "x := 1", res)

	res, err = FormatGoal("x:=1\n")
	assert.NoError(t, err)
	assert.Equal(t, "x := 1\n", res)

	src := "func {\n"
	res, err = FormatGoal(src)
	assert.Error(t, err)
	assert.Equal(t, src, res)
}
