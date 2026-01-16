// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"

	"cogentcore.org/core/core"
	"cogentcore.org/lab/physics/examples/virtroom"
)

var NoGUI bool

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-nogui" {
		NoGUI = true
	}
	ev := &virtroom.Env{}
	ev.Defaults()
	if NoGUI {
		ev.NoGUIRun()
		return
	}
	// core.RenderTrace = true
	b := core.NewBody("virtroom").SetTitle("Physics Virtual Room")
	ev.ConfigGUI(b)
	b.RunMainWindow()
}
