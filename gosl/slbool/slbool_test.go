// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slbool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBool tests the conversions to and from Go bools and the predicates.
func TestBool(t *testing.T) {
	assert.Equal(t, True, FromBool(true))
	assert.Equal(t, False, FromBool(false))
	assert.Equal(t, Bool(1), True)
	assert.Equal(t, Bool(0), False)

	var b Bool
	assert.False(t, b.Bool())
	assert.True(t, b.IsFalse())
	assert.False(t, b.IsTrue())
	assert.True(t, IsFalse(b))
	assert.False(t, IsTrue(b))

	b.SetBool(true)
	assert.Equal(t, True, b)
	assert.True(t, b.Bool())
	assert.True(t, b.IsTrue())
	assert.False(t, b.IsFalse())
	assert.True(t, IsTrue(b))
	assert.False(t, IsFalse(b))

	// other values are neither true nor false
	b = 2
	assert.False(t, b.IsTrue())
	assert.False(t, b.IsFalse())
	assert.Equal(t, "false", b.String())
}

// TestBoolText tests the string and text marshaling conversions.
func TestBoolText(t *testing.T) {
	assert.Equal(t, "true", True.String())
	assert.Equal(t, "false", False.String())

	var b Bool
	for _, s := range []string{"true", "True"} {
		b = False
		b.FromString(s)
		assert.Equal(t, True, b, s)
	}
	for _, s := range []string{"false", "TRUE", "1", ""} {
		b = True
		b.FromString(s)
		assert.Equal(t, False, b, s)
	}

	txt, err := True.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "true", string(txt))
	txt, err = False.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, "false", string(txt))

	require.NoError(t, b.UnmarshalText([]byte("true")))
	assert.Equal(t, True, b)
	require.NoError(t, b.UnmarshalText([]byte("false")))
	assert.Equal(t, False, b)
}
