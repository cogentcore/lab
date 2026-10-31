// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCompleteJunction checks that a junction to a directory is completed
// as a directory.
func TestCompleteJunction(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", filepath.Join(dir, "junction"), filepath.Join(dir, "sub")).CombinedOutput(); err != nil {
		t.Fatalf("mklink: %v: %s", err, out)
	}
	gl := NewGoal()
	for _, de := range gl.localDirEntries(dir) {
		if de.name == "junction" && !de.isDir {
			t.Error("the junction is not a directory")
		}
	}
}
