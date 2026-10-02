// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slmath_test

import (
	"os"
	"testing"

	"cogentcore.org/core/cli"
	"cogentcore.org/lab/gosl/gotosl"
	"github.com/stretchr/testify/assert"
)

// TestGoslTranslate runs gosl over testdata/slmathtest.go, which calls
// every slmath function, and compares the generated WGSL against the
// golden file. gosl also runs the WGSL compiler over the result, so this
// checks that every slmath function produces valid WGSL, not just that
// the translation is stable.
//
// To update the golden file after an intended change:
//
//	cd testdata && gosl && cp shaders/SLMathTest.wgsl SLMathTest.golden
func TestGoslTranslate(t *testing.T) {
	t.Chdir("testdata")

	opts := cli.DefaultOptions("gosl", "Go as a shader language converts Go code to WGSL WebGPU shader code, which can be run on the GPU through WebGPU.")
	opts.Fatal = false // otherwise an error calls os.Exit and we can't test it
	cfg := &gotosl.Config{}
	assert.NoError(t, cli.Run(opts, cfg, gotosl.Run))

	want, err := os.ReadFile("SLMathTest.golden")
	if err != nil {
		t.Error(err)
		return
	}
	got, err := os.ReadFile("shaders/SLMathTest.wgsl")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, string(want), string(got))
}
