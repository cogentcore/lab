// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/lab/physics/examples/collide"
)

func main() {
	b := core.NewBody("collide").SetTitle("Physics Collide")
	collide.Config(b)
	b.RunMainWindow()
}
