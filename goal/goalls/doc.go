// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goalls

import (
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"cogentcore.org/core/text/textpos"
	"cogentcore.org/lab/goal/goalfmt"
	"cogentcore.org/lab/goal/transpile"
)

// doc is an open Goal file, with its Go code, which gopls has open as
// the Go file of the Goal file.
type doc struct {

	// uri is the URI of the Goal file, and goURI that of its Go file.
	uri, goURI string

	// filename is the name of the Goal file.
	filename string

	// version is the version of the text, from the client.
	version int

	// text is the text of the Goal file, and lines its lines.
	text  string
	lines []string

	// code is the Go code of the text.
	code *transpile.GoCode

	// diags are the last diagnostics of gopls for the Go code,
	// in the Goal file, which are published with the errors of
	// transpiling it.
	diags []any

	// goCode is the Go code of the lines of the Goal file with errors,
	// for those whose Go code is not the same as their source, which is
	// shown as an inlay hint after them.
	goCode map[int]string
}

// newDoc returns a new open Goal file with the given URI and text.
func newDoc(uri string, version int, text string) *doc {
	d := &doc{uri: uri, filename: uriFilename(uri), version: version}
	d.goURI = filenameURI(goalfmt.GoFilename(d.filename))
	d.setText(text)
	return d
}

// setText sets the text of the Goal file, and transpiles it.
func (d *doc) setText(text string) {
	d.text = text
	d.lines = strings.Split(text, "\n")
	var prev []string
	if d.code != nil {
		prev = d.code.Imports
	}
	d.code = transpile.TranspileGo(text, goalfmt.GoFilename(d.filename), transpile.GoOptions{FindImports: true, Imports: prev})
}

// isGoal returns whether the given file is a Goal file.
func isGoal(uri string) bool {
	return strings.EqualFold(filepath.Ext(uri), ".goal")
}

// uriFilename returns the file name of the given file URI.
func uriFilename(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return strings.TrimPrefix(uri, "file://")
	}
	return filepath.FromSlash(u.Path)
}

// filenameURI returns the file URI of the given file name.
func filenameURI(fname string) string {
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(fname)}).String()
}

// encoding is the encoding of the character offsets of positions,
// which the client and server agree on in initialization.
type encoding int

const (
	utf16Encoding encoding = iota
	utf8Encoding
	utf32Encoding
)

// toRune returns the rune index in the given line of the given offset.
func (enc encoding) toRune(line string, off int) int {
	switch enc {
	case utf32Encoding:
		return min(off, utf8.RuneCountInString(line))
	case utf8Encoding:
		return utf8.RuneCountInString(line[:min(off, len(line))])
	}
	n, u := 0, 0
	for _, r := range line {
		if u >= off {
			break
		}
		u += utf16.RuneLen(r)
		n++
	}
	return n
}

// fromRune returns the offset in the given line of the given rune index.
func (enc encoding) fromRune(line string, ri int) int {
	switch enc {
	case utf32Encoding:
		return ri
	case utf8Encoding:
		n := 0
		for i := range line {
			if n == ri {
				return i
			}
			n++
		}
		return len(line)
	}
	n, u := 0, 0
	for _, r := range line {
		if n == ri {
			break
		}
		u += utf16.RuneLen(r)
		n++
	}
	return u
}

// line returns the given line of the given lines, or "".
func line(lines []string, ln int) string {
	if ln < 0 || ln >= len(lines) {
		return ""
	}
	return lines[ln]
}

// toGo returns the position in the Go code of the given position in the
// Goal file, in the given encoding, and whether it has one.
func (d *doc) toGo(enc encoding, ln, ch int) (int, int, bool) {
	p, ok := d.code.ToGo(textpos.Pos{Line: ln, Char: enc.toRune(line(d.lines, ln), ch)})
	if !ok {
		return 0, 0, false
	}
	return p.Line, enc.fromRune(line(d.code.Lines, p.Line), p.Char), true
}

// toGoal returns the position in the Goal file of the given position in
// its Go code, in the given encoding, and whether it has one.
func (d *doc) toGoal(enc encoding, ln, ch int) (int, int, bool) {
	p, ok := d.code.ToGoal(textpos.Pos{Line: ln, Char: enc.toRune(line(d.code.Lines, ln), ch)})
	if !ok {
		return 0, 0, false
	}
	return p.Line, enc.fromRune(line(d.lines, p.Line), p.Char), true
}

// applyChange applies the given change to the given text, replacing the
// given range of it in the given encoding with the given text.
func applyChange(text string, enc encoding, sl, sc, el, ec int, repl string) string {
	lines := strings.Split(text, "\n")
	off := func(ln, ch int) int {
		o := 0
		for i := 0; i < ln && i < len(lines); i++ {
			o += len(lines[i]) + 1
		}
		if ln >= len(lines) {
			return len(text)
		}
		l := lines[ln]
		return min(o+len(string([]rune(l)[:enc.toRune(l, ch)])), len(text))
	}
	st, ed := off(sl, sc), off(el, ec)
	if ed < st {
		ed = st
	}
	return text[:st] + repl + text[ed:]
}
