// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGoalFiles transpiles each testdata/*.goal file to Go, builds them,
// and runs each, checking that its output is that of testdata/*.want,
// if it exists.
func TestGoalFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("builds Go programs")
	}
	files, err := filepath.Glob("testdata/*.goal")
	assert.NoError(t, err)

	// The programs are built in the module, in testdata, for its dependencies.
	dir, err := os.MkdirTemp("testdata", "build")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	var names []string
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".goal")
		pdir := filepath.Join(dir, name)
		assert.NoError(t, os.Mkdir(pdir, 0755))
		st := NewState()
		assert.NoError(t, st.TranspileFile(f, filepath.Join(pdir, "main.go")), name)
		names = append(names, name)
	}

	bin := filepath.Join(dir, "bin") + string(filepath.Separator)
	built := buildGoalFiles(t, dir, bin, names)

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			if !built[name] {
				t.Skip("only builds with unused variables")
			}
			out, err := exec.Command(filepath.Join(bin, name)).CombinedOutput()
			assert.NoError(t, err, string(out))
			want, err := os.ReadFile(filepath.Join("testdata", name+".want"))
			if err != nil {
				return
			}
			assert.Equal(t, string(want), string(out))
		})
	}
}

// buildGoalFiles builds the transpiled programs of the given names in dir
// into bin, returning which ones were built. A program whose only compile
// errors are unused variables is not built, but is not an error: those are
// errors of Go, but not of the Goal interpreter, which can run it.
func buildGoalFiles(t *testing.T, dir, bin string, names []string) map[string]bool {
	built := map[string]bool{}
	pkg := func(name string) string { return "./" + filepath.ToSlash(filepath.Join(dir, name)) }
	if _, err := exec.Command("go", "build", "-o", bin, pkg("...")).CombinedOutput(); err == nil {
		for _, name := range names {
			built[name] = true
		}
		return built
	}
	for _, name := range names {
		out, err := exec.Command("go", "build", "-o", filepath.Join(bin, name), pkg(name)).CombinedOutput()
		if err == nil {
			built[name] = true
			continue
		}
		for _, ln := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if !strings.HasPrefix(ln, "#") && !strings.Contains(ln, "declared and not used") {
				t.Errorf("building %s: %v\n%s", name, err, out)
				break
			}
		}
	}
	return built
}
