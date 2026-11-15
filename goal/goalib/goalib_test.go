// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goalib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSplitLines tests that lines are split and trimmed of whitespace.
func TestSplitLines(t *testing.T) {
	assert.Equal(t, []string{"a", "b c", "", "d"}, SplitLines("  a \n\tb c\n\n d\t"))
	assert.Equal(t, []string{"x", "y"}, SplitLines("x\r\ny"))
}

// TestFileDirExists tests FileExists and DirExists on a file, a directory
// and a missing path.
func TestFileDirExists(t *testing.T) {
	dir := t.TempDir()
	fn := filepath.Join(dir, "f.txt")
	require.NoError(t, os.WriteFile(fn, []byte("x"), 0666))
	assert.True(t, FileExists(fn))
	assert.False(t, FileExists(filepath.Join(dir, "missing")))
	assert.True(t, DirExists(dir))
	assert.False(t, DirExists(filepath.Join(dir, "missing")))
}

// TestReadWriteReplaceFile tests writing, reading and replacing in a file,
// and the errors for a missing directory.
func TestReadWriteReplaceFile(t *testing.T) {
	dir := t.TempDir()
	fn := filepath.Join(dir, "f.txt")
	require.NoError(t, WriteFile(fn, "one two one"))
	assert.Equal(t, "one two one", ReadFile(fn))
	assert.Equal(t, "1 two 1", ReplaceInFile(fn, "one", "1"))
	assert.Equal(t, "1 two 1", ReadFile(fn))

	bad := filepath.Join(dir, "missing", "f.txt")
	assert.Error(t, WriteFile(bad, "x"))
	assert.Equal(t, "", ReadFile(bad))
}

// TestStringsToAnys tests converting strings to a slice of any.
func TestStringsToAnys(t *testing.T) {
	assert.Equal(t, []any{"a", "b"}, StringsToAnys([]string{"a", "b"}))
	assert.Empty(t, StringsToAnys(nil))
}

// TestAllFiles tests listing all files, but not directories, under a path,
// and an empty result for a missing path.
func TestAllFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub", "deeper"), 0777))
	files := []string{
		filepath.Join(dir, "a.txt"),
		filepath.Join(dir, "sub", "b.txt"),
		filepath.Join(dir, "sub", "deeper", "c.txt"),
	}
	for _, f := range files {
		require.NoError(t, os.WriteFile(f, nil, 0666))
	}
	assert.Equal(t, files, AllFiles(dir))
	assert.Nil(t, AllFiles(filepath.Join(dir, "missing")))
}
