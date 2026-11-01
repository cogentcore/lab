// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"unicode/utf8"

	"cogentcore.org/core/base/stringsx"
	"cogentcore.org/core/text/textpos"
	"golang.org/x/tools/imports"
)

// GoCode is the Go code of a Goal file, as `goal build` makes it, but
// without formatting it, so that the Go code of each line of the Goal
// source is on its own line, after a header, with a map between the
// positions of the two. Tools for Go code, such as gofmt, goimports, and
// the Go language server (gopls), can then be used for Goal code, with
// their results mapped back to the Goal source, without parsing Goal.
type GoCode struct {

	// Code is the Go code.
	Code string

	// Lines are the lines of the Go code.
	Lines []string

	// Offset is the line of the Go code of the first line of the Goal
	// source: the number of lines of the header.
	Offset int

	// Source are the lines of the Goal source, with the map from the
	// columns of their Go code back to their own columns.
	Source []SourceLine

	// HasPackage is whether the Goal source has a package declaration.
	// Otherwise, it is a script, whose code is in a main function,
	// after the imports.
	HasPackage bool

	// Imports are the imports of the Go code of a script, as goimports
	// finds them, with a name before the path if it has one, such as _.
	Imports []string

	// Errors are the errors of transpiling the Goal source.
	Errors []*Error
}

// GoOptions are the options for [TranspileGo].
type GoOptions struct {

	// Header is the first line of the Go code, which is a comment.
	// The default is a comment that it is the Go code of a Goal file.
	Header string

	// FindImports is whether to find the imports of a script with
	// goimports, as `goal build` does. Otherwise, Imports are used.
	FindImports bool

	// Imports are the imports of a script that are used if FindImports
	// is not set, or if goimports fails, which it can for code with
	// errors. The default is [ScriptImports].
	Imports []string
}

// ScriptImports are the imports of the Go code of a Goal script, before
// goimports adds the others that it uses, and removes those that it does
// not use.
var ScriptImports = []string{
	`"cogentcore.org/lab/goal"`,
	`"cogentcore.org/lab/goal/goalib"`,
	`"cogentcore.org/lab/tensor"`,
	`_ "cogentcore.org/lab/tensor/tmath"`,
	`_ "cogentcore.org/lab/stats/stats"`,
	`_ "cogentcore.org/lab/stats/metric"`,
}

// scriptHeader returns the lines of the Go code of a Goal script before
// its code, with the given imports, up to its main function.
func scriptHeader(imps []string) []string {
	hdr := []string{"package main", "import ("}
	for _, imp := range imps {
		hdr = append(hdr, "\t"+imp)
	}
	return append(hdr, ")", "", "func main() {", "\tgoalrun := goal.NewGoal()", "\t_ = goalrun")
}

// TranspileGo returns the Go code of the given Goal source, for the Go
// file of the given name (see [GoCode]).
func TranspileGo(src, gofile string, opts GoOptions) *GoCode {
	lns := stringsx.SplitLines(src)
	gc := &GoCode{}
	for _, ln := range lns {
		if strings.HasPrefix(ln, "package ") {
			gc.HasPackage = true
			break
		}
	}
	st := NewState()
	st.FuncToVar = !gc.HasPackage
	st.TranspileCode(src)
	st.AddChunk()
	for _, err := range st.Errors {
		if te, ok := err.(*Error); ok {
			gc.Errors = append(gc.Errors, te)
		}
	}

	// The Go code of each source line, which is one line of Go code.
	var body []string
	for ci, ch := range st.Chunks {
		srcs := st.ChunkSources[ci]
		for i, cl := range strings.Split(ch, "\n") {
			body = append(body, cl)
			if i < len(srcs) {
				gc.Source = append(gc.Source, srcs[i])
			} else {
				gc.Source = append(gc.Source, SourceLine{})
			}
		}
	}
	for len(body) < len(lns) { // empty lines at the end
		body = append(body, "")
		gc.Source = append(gc.Source, SourceLine{Src: lns[len(gc.Source)]})
	}

	hdr := opts.Header
	if hdr == "" {
		hdr = "// The Go code of a Goal file."
	}
	if gc.HasPackage {
		gc.Offset = 1
		gc.setCode([]string{hdr}, body, nil)
		return gc
	}
	imps := opts.Imports
	if imps == nil {
		imps = ScriptImports
	}
	if opts.FindImports {
		full := append(append([]string{hdr}, scriptHeader(ScriptImports)...), body...)
		full = append(full, "}", "")
		if found := goImports(gofile, strings.Join(full, "\n")); found != nil {
			imps = found
		}
	}
	gc.Imports = imps
	head := append([]string{hdr}, scriptHeader(imps)...)
	gc.Offset = len(head)
	gc.setCode(head, body, []string{"}"})
	return gc
}

// setCode sets the code from the given lines of the header, the body,
// and after it.
func (gc *GoCode) setCode(header, body, after []string) {
	gc.Code = strings.Join(append(append(header, body...), after...), "\n") + "\n"
	gc.Lines = strings.Split(gc.Code, "\n")
}

// goImports returns the imports of the given Go code of the given file
// after goimports processes it, adding the imports that it uses and
// removing those that it does not use, or nil if that fails.
func goImports(gofile, code string) []string {
	res, err := imports.Process(gofile, []byte(code), &imports.Options{Comments: true})
	if err != nil {
		return nil
	}
	f, err := parser.ParseFile(token.NewFileSet(), gofile, res, parser.ImportsOnly)
	if err != nil {
		return nil
	}
	imps := []string{}
	for _, is := range f.Imports {
		if _, err := strconv.Unquote(is.Path.Value); err != nil {
			continue
		}
		imp := is.Path.Value
		if is.Name != nil {
			imp = is.Name.Name + " " + imp
		}
		imps = append(imps, imp)
	}
	return imps
}

// NumSourceLines returns the number of lines of the Goal source.
func (gc *GoCode) NumSourceLines() int {
	return len(gc.Source)
}

// GoLine returns the line of the Go code of the given line of the
// Goal source, or -1 if there is none.
func (gc *GoCode) GoLine(ln int) int {
	if ln < 0 || ln >= len(gc.Source) || gc.Offset+ln >= len(gc.Lines) {
		return -1
	}
	return gc.Offset + ln
}

// SourceLine returns the line of the Goal source of the given line of the
// Go code, or -1 if there is none, as for the code that the transpiler
// adds before and after the code of the source, such as the imports.
func (gc *GoCode) SourceLine(ln int) int {
	sl := ln - gc.Offset
	if sl < 0 || sl >= len(gc.Source) || ln >= len(gc.Lines) {
		return -1
	}
	return sl
}

// ToGo returns the position in the Go code of the given position in
// the Goal source, in lines and runes, and whether it has one.
func (gc *GoCode) ToGo(pos textpos.Pos) (textpos.Pos, bool) {
	gl := gc.GoLine(pos.Line)
	if gl < 0 {
		return textpos.Pos{}, false
	}
	sl := &gc.Source[pos.Line]
	gtxt := gc.Lines[gl]
	if len(sl.Cols) == 0 { // the same code
		return textpos.Pos{Line: gl, Char: min(pos.Char, utf8.RuneCountInString(gtxt))}, true
	}
	b := runeToByte(sl.Src, pos.Char)
	// The Go code of the source token that the position is in, which
	// is the last one that starts at or before it, at its offset in it.
	best := -1
	for i, c := range sl.Cols {
		if c > 0 && int(c)-1 <= b && (best < 0 || c > sl.Cols[best]) {
			best = i
		}
	}
	if best < 0 {
		return textpos.Pos{Line: gl}, true
	}
	end := best
	for end < len(sl.Cols) && sl.Cols[end] == sl.Cols[best] {
		end++
	}
	gb := min(best+b-(int(sl.Cols[best])-1), end, len(gtxt))
	return textpos.Pos{Line: gl, Char: utf8.RuneCountInString(gtxt[:gb])}, true
}

// ToGoal returns the position in the Goal source of the given position
// in the Go code, in lines and runes, and whether it has one: the code
// that the transpiler adds before and after the code of the source, such
// as the imports, does not. A position at a space is after the token
// before it, as the spaces of the Go code are added around its tokens.
func (gc *GoCode) ToGoal(pos textpos.Pos) (textpos.Pos, bool) {
	ln := gc.SourceLine(pos.Line)
	if ln < 0 {
		return textpos.Pos{}, false
	}
	sl := &gc.Source[ln]
	gtxt := gc.Lines[pos.Line]
	nsrc := utf8.RuneCountInString(sl.Src)
	if len(sl.Cols) == 0 { // the same code
		return textpos.Pos{Line: ln, Char: min(pos.Char, nsrc)}, true
	}
	i := runeToByte(gtxt, pos.Char)
	if i >= len(sl.Cols) { // at the end
		return textpos.Pos{Line: ln, Char: nsrc}, true
	}
	if i > 0 && gtxt[i] == ' ' {
		p, ok := gc.ToGoal(textpos.Pos{Line: pos.Line, Char: pos.Char - 1})
		p.Char = min(p.Char+1, nsrc)
		return p, ok
	}
	c := sl.Cols[i]
	if c == 0 {
		return textpos.Pos{Line: ln}, true
	}
	// at the offset of the position in the Go code of its token
	i0 := i
	for i0 > 0 && sl.Cols[i0-1] == c {
		i0--
	}
	b := min(int(c)-1+i-i0, len(sl.Src))
	return textpos.Pos{Line: ln, Char: utf8.RuneCountInString(sl.Src[:b])}, true
}

// ToGoalRegion returns the region in the Goal source of the given region
// in the Go code, and whether it has one, which it does if its start does.
func (gc *GoCode) ToGoalRegion(reg textpos.Region) (textpos.Region, bool) {
	st, ok := gc.ToGoal(reg.Start)
	if !ok {
		return textpos.Region{}, false
	}
	ed, ok := gc.ToGoal(reg.End)
	if !ok || ed.IsLess(st) {
		ed = st
	}
	return textpos.Region{Start: st, End: ed}, true
}

// runeToByte returns the byte offset in s of the given rune index.
func runeToByte(s string, r int) int {
	if r <= 0 {
		return 0
	}
	n := 0
	for i := range s {
		if n == r {
			return i
		}
		n++
	}
	return len(s)
}
