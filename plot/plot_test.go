// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"os"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	_ "cogentcore.org/core/paint/renderers"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestPlot(t *testing.T) {
	pt := New()
	pt.Title.Text = "Test Plot"
	pt.X.Range.Max = 100
	pt.X.Label.Text = "X Axis"
	pt.Y.Range.Max = 100
	pt.Y.Label.Text = "Y Axis"

	imagex.Assert(t, pt.RenderImage(), "plot.png")
}
