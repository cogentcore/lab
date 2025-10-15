// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"embed"

	"cogentcore.org/core/content"
	"cogentcore.org/core/core"
	"cogentcore.org/core/htmlcore"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/tree"
	_ "cogentcore.org/lab/yaegilab"
)

//go:embed content
var econtent embed.FS

func main() {
	b := core.NewBody("Cogent Lab")
	ct := content.NewContent(b).SetContent(econtent)
	ctx := ct.Context
	content.OfflineURL = "https://cogentcore.org/lab"
	ctx.AddWikilinkHandler(htmlcore.GoDocWikilink("doc", "cogentcore.org/lab"))
	b.AddTopBar(func(bar *core.Frame) {
		tb := core.NewToolbar(bar)
		tb.Maker(ct.MakeToolbar)
		tb.Maker(func(p *tree.Plan) {
			tree.Add(p, func(w *core.Button) {
				ctx.LinkButton(w, "https://github.com/cogentcore/lab")
				w.SetText("GitHub").SetIcon(icons.GitHub)
			})
			tree.Add(p, func(w *core.Button) {
				ctx.LinkButton(w, "https://youtube.com/@CogentCore")
				w.SetText("Videos").SetIcon(icons.VideoLibrary)
			})
			tree.Add(p, func(w *core.Button) {
				ctx.LinkButton(w, "https://cogentcore.org/blog")
				w.SetText("Blog").SetIcon(icons.RssFeed)
			})
			tree.Add(p, func(w *core.Button) {
				ctx.LinkButton(w, "https://cogentcore.org/community")
				w.SetText("Community").SetIcon(icons.Forum)
			})
			tree.Add(p, func(w *core.Button) {
				ctx.LinkButton(w, "https://github.com/sponsors/cogentcore")
				w.SetText("Sponsor").SetIcon(icons.Favorite)
			})
		})
	})
	b.RunMainWindow()
}
