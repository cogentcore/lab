// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"

	"cogentcore.org/core/core"
	"cogentcore.org/core/pages"
)

//go:embed content
var content embed.FS

func main() {
	b := core.NewBody("Cogent Lab")
	pg := pages.NewPage(b).SetContent(content)
	b.AddTopBar(func(bar *core.Frame) {
		core.NewToolbar(bar).Maker(pg.MakeToolbar)
	})
	b.RunMainWindow()
}
