// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package goalfmt formats Goal code as gofmt formats Go code, and fixes
// its imports as goimports does. It does this with gofmt and goimports
// themselves, on the Go code that the Goal code is transpiled to (see
// [transpile.GoCode]), mapping their changes back to the Goal code, so
// that the Goal code is never parsed as Go.
package goalfmt

import (
	"errors"
	"fmt"
	"go/format"
	"go/scanner"
	"path/filepath"
	"strings"
	"unicode"

	"cogentcore.org/core/base/stringsx"
	"cogentcore.org/core/text/textpos"
	"cogentcore.org/lab/goal/transpile"
	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/tools/imports"
)

// Options are the options for [Source].
type Options struct {

	// FormatOnly only formats the code, as gofmt does, without fixing
	// its imports, as goimports does.
	FormatOnly bool
}

// GoFilename returns the name of the Go file of the given Goal file,
// which `goal build` writes.
func GoFilename(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename)) + ".go"
}

// Source returns the given source of the Goal file of the given name,
// formatted as gofmt formats Go code, with its imports fixed as goimports
// does, unless [Options.FormatOnly] is set. The name is used to find the
// imports of the package of the file, as for goimports. The Goal code is
// transpiled to Go code, which gofmt or goimports processes, and their
// changes to its lines are mapped back to the lines of the Goal code:
//   - A line whose Go code is the same as its source, apart from spaces,
//     is replaced with the formatted Go code, with its spaces.
//   - Other lines, such as commands and math, which the transpiler changes,
//     only get the indentation of their formatted Go code.
//   - The lines that gofmt or goimports removes, which are blank lines and
//     imports, are removed, and those they add, such as imports, are added.
//
// The imports of a Goal script, without a package declaration, are not
// in its source, as `goal build` adds them.
func Source(filename string, src []byte, opt *Options) ([]byte, error) {
	if opt == nil {
		opt = &Options{}
	}
	gofile := GoFilename(filename)
	gc := transpile.TranspileGo(string(src), gofile, transpile.GoOptions{})
	if len(gc.Errors) > 0 {
		errs := make([]error, len(gc.Errors))
		for i, e := range gc.Errors {
			errs[i] = fmt.Errorf("%s:%d:%d: %s", filename, e.Line, e.Col, e.Msg)
		}
		return nil, errors.Join(errs...)
	}
	var out []byte
	var err error
	if opt.FormatOnly {
		out, err = format.Source([]byte(gc.Code))
	} else {
		out, err = imports.Process(gofile, []byte(gc.Code), &imports.Options{Comments: true, TabIndent: true, TabWidth: 8})
	}
	if err != nil {
		return nil, sourceError(filename, gc, err)
	}
	lines := mapLines(gc, stringsx.SplitLines(string(src)), strings.Split(string(out), "\n"))
	res := strings.Join(lines, "\n")
	if res != "" && !strings.HasSuffix(res, "\n") {
		res += "\n"
	}
	return []byte(res), nil
}

// sourceError returns the given error of gofmt or goimports for the Go
// code with the positions of its syntax errors in the Goal file.
func sourceError(filename string, gc *transpile.GoCode, err error) error {
	var el scanner.ErrorList
	if !errors.As(err, &el) {
		return err
	}
	errs := make([]error, 0, len(el))
	for _, e := range el {
		pos := textpos.Pos{Line: e.Pos.Line - 1, Char: e.Pos.Column - 1}
		if p, ok := gc.ToGoal(pos); ok {
			errs = append(errs, fmt.Errorf("%s:%d:%d: %s (in the Go code: %s)", filename, p.Line+1, p.Char+1, e.Msg, strings.TrimSpace(gc.Lines[pos.Line])))
		} else {
			errs = append(errs, fmt.Errorf("%s: %s (in the Go code: %s)", filename, e.Msg, strings.TrimSpace(gc.Lines[min(pos.Line, len(gc.Lines)-1)])))
		}
	}
	return errors.Join(errs...)
}

// mapLines returns the given lines of the Goal source of the given Go
// code, with the changes of the given formatted lines of the Go code
// (see [Source]).
func mapLines(gc *transpile.GoCode, src, formatted []string) []string {
	// Script code is in the main function, which indents it.
	base := ""
	if !gc.HasPackage {
		base = "\t"
	}
	// same returns whether the Go code of the given source line is
	// the same as its source, apart from spaces.
	same := func(sl int) bool {
		return stripSpaces(src[sl]) == stripSpaces(gc.Lines[gc.GoLine(sl)])
	}
	// line returns the given source line with the changes of the
	// given formatted Go code of it.
	line := func(sl int, f string) string {
		f = strings.TrimPrefix(f, base)
		if same(sl) {
			return f
		}
		indent := f[:len(f)-len(strings.TrimLeft(f, " \t"))]
		return indent + strings.TrimLeft(src[sl], " \t")
	}
	// added returns the given formatted lines that a tool adds after
	// the given line of the Go code, if it is a line of the source.
	added := func(gl int, fs []string) []string {
		if gc.SourceLine(gl) < 0 {
			return nil
		}
		out := make([]string, len(fs))
		for i, f := range fs {
			out[i] = strings.TrimPrefix(f, base)
		}
		return out
	}
	// removed returns whether the given source line, whose Go code a
	// tool removes, is removed: it is if it is blank or the same as
	// its Go code, such as an import.
	removed := func(sl int) bool {
		return strings.TrimSpace(src[sl]) == "" || same(sl)
	}

	a := make([]string, len(gc.Lines))
	for i, l := range gc.Lines {
		a[i] = stripSpaces(l)
	}
	b := make([]string, len(formatted))
	for i, l := range formatted {
		b[i] = stripSpaces(l)
	}
	var res []string
	for _, op := range difflib.NewMatcher(a, b).GetOpCodes() {
		switch op.Tag {
		case 'e':
			for k := range op.I2 - op.I1 {
				if sl := gc.SourceLine(op.I1 + k); sl >= 0 {
					res = append(res, line(sl, formatted[op.J1+k]))
				}
			}
		case 'd':
			for gl := op.I1; gl < op.I2; gl++ {
				if sl := gc.SourceLine(gl); sl >= 0 && !removed(sl) {
					res = append(res, src[sl])
				}
			}
		case 'i':
			res = append(res, added(op.I1-1, formatted[op.J1:op.J2])...)
		case 'r':
			// Lines changed by a tool, such as imports, are replaced if
			// they are the same as their Go code, and otherwise kept.
			keep := false
			for gl := op.I1; gl < op.I2; gl++ {
				if sl := gc.SourceLine(gl); sl >= 0 && !removed(sl) {
					keep = true
				}
			}
			if keep {
				for gl := op.I1; gl < op.I2; gl++ {
					if sl := gc.SourceLine(gl); sl >= 0 {
						res = append(res, src[sl])
					}
				}
				continue
			}
			if gc.SourceLine(op.I1) >= 0 {
				res = append(res, added(op.I1, formatted[op.J1:op.J2])...)
			} else {
				res = append(res, added(op.I1-1, formatted[op.J1:op.J2])...)
			}
		}
	}
	return res
}

// stripSpaces returns the given string without its spaces.
func stripSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
