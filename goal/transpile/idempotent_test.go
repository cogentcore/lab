// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goTokens returns the tokens of the given Go source, with their literals,
// one per line, including the comments, as the source is transpiled:
// with the automatic semicolons all as ";".
func goTokens(src []byte) (string, error) {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	var errs scanner.ErrorList
	s.Init(file, src, func(pos token.Position, msg string) { errs.Add(pos, msg) }, scanner.ScanComments)
	var b strings.Builder
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			lit = ";"
		}
		if tok == token.COMMENT {
			lit = strings.TrimSpace(lit)
		}
		fmt.Fprintf(&b, "%d: %s %s\n", fset.Position(pos).Line, tok, lit)
	}
	return b.String(), errs.Err()
}

// TestGoIdempotent transpiles each Go file of the lab repository as Goal
// code, and checks that the result has the same Go tokens: Go code is
// Goal code that transpiles to itself (modulo whitespace).
func TestGoIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("transpiles all of the Go files of the repository")
	}
	root := filepath.Join("..", "..")
	nfail, nfiles := 0, 0
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if n := d.Name(); path != root && (strings.HasPrefix(n, ".") || n == "testdata" || n == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		want, err := goTokens(src)
		if err != nil {
			return nil // not valid Go, as with build tags of other syntax
		}
		nfiles++
		code, perr := transpileFile(string(src))
		if perr != nil {
			nfail++
			t.Errorf("%s: %v", path, perr)
			return nil
		}
		got, _ := goTokens([]byte(code))
		if got != want {
			nfail++
			ln, diff := firstDiff(want, got)
			lines := strings.Split(string(src), "\n")
			sl := ""
			if ln > 0 && ln <= len(lines) {
				sl = strings.TrimSpace(lines[ln-1])
			}
			t.Errorf("%s:%d: %s\n\t%s", path, ln, sl, diff)
		}
		return nil
	})
	if nfiles < 100 {
		t.Errorf("only %d Go files found in %s", nfiles, root)
	}
	t.Logf("%d of %d Go files are not idempotent", nfail, nfiles)
}

// transpileFile transpiles the given source of a Go file, returning a
// panic of it as an error.
func transpileFile(src string) (code string, err error) {
	st := NewState()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic at line %d: %q: %v", st.lineNumber(), strings.TrimSpace(st.curSrc), r)
		}
	}()
	st.FuncToVar = false
	st.MathRecord = false
	st.TranspileCode(src)
	return st.Code(), nil
}

// firstDiff returns the source line number of the first token that
// differs between the given token lists, of want and got, and the
// difference.
func firstDiff(want, got string) (int, string) {
	wl, gl := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := range wl {
		ln := 0
		fmt.Sscanf(wl[i], "%d:", &ln)
		if i >= len(gl) {
			return ln, "missing: " + wl[i]
		}
		if wl[i] != gl[i] {
			return ln, fmt.Sprintf("want %q, got %q", wl[i], gl[i])
		}
	}
	return 0, "extra: " + gl[len(wl)]
}
