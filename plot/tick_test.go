// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefaultTicksZero tests that a tick at 0 is exactly 0, without
// the rounding errors of the steps that would show in its label.
func TestDefaultTicksZero(t *testing.T) {
	for _, n := range []int{5, 10, 30, 40} {
		for _, tk := range (DefaultTicks{}).Ticks(-0.88856439213005511, 0.88856232836900007, n) {
			if tk.IsMinor() {
				continue
			}
			assert.NotContains(t, tk.Label, "e-", "n = %d", n)
		}
	}
}
