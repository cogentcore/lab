// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/text/highlighting"
	_ "cogentcore.org/core/text/parse/languages/markdown"
)

// Markdown files, such as README.md, are highlighted with the markdown
// language of parse, as chroma lexers other than Go are not included
// (see [highlighting.Lexers]). This does not change a program that
// already uses parse, with its own languages.
func init() {
	if highlighting.UseParse {
		return
	}
	highlighting.UseParse = true
	highlighting.ParseFilter = func(known fileinfo.Known) bool {
		return known == fileinfo.Markdown
	}
}
