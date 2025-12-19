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
	"cogentcore.org/core/text/csl"
	_ "cogentcore.org/core/text/tex" // include this to get math
	"cogentcore.org/core/tree"
	_ "cogentcore.org/lab/yaegilab"
)

// NOTE: you must make a symbolic link to the zotero CCNLab CSL file as ccnlab.json
// in this directory, to generate references and have the generated reference links
// use the official APA style. https://www.zotero.org/groups/340666/ccnlab
// Must configure using BetterBibTeX for zotero: https://retorque.re/zotero-better-bibtex/

//go:generate mdcite -vv -refs ./ccnlab.json -d ./content

//go:embed content citedrefs.json
var econtent embed.FS

func main() {
	b := core.NewBody("Cogent Lab")
	ct := content.NewContent(b).SetContent(econtent)
	ctx := ct.Context
	content.OfflineURL = "https://cogentcore.org/lab"
	refs, err := csl.OpenFS(econtent, "citedrefs.json")
	if err == nil {
		ct.References = csl.NewKeyList(refs)
	}
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
