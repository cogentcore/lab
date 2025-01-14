// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"

	"cogentcore.org/core/content"
	"cogentcore.org/core/core"
	_ "cogentcore.org/lab/yaegilab"
)

//go:embed content
var econtent embed.FS

func main() {
	b := core.NewBody("Cogent Lab")
	ct := content.NewContent(b).SetContent(econtent)
	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(ct.MakeToolbar)
	})
	b.RunMainWindow()
}
