// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAddOnce tests that a plotter is added to a plot once, as the New
// functions of the plotters add them, and NewTablePlot then also does.
func TestAddOnce(t *testing.T) {
	plt := New()
	a := &testPlotter{}
	b := &testPlotter{}
	plt.Add(a)
	plt.Add(a, b)
	assert.Len(t, plt.Plotters, 2)
}
