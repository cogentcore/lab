// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command goalimports formats Goal code as gofmt formats Go code, and
// fixes its imports as goimports does, using gofmt and goimports on the
// Go code that it is transpiled to (see package goalfmt). Its usage is
// that of goimports:
//
//	goalimports [flags] [path ...]
//
// The paths are Goal files, or directories, in which all .goal files are
// processed, recursively. Without paths, it processes standard input.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"cogentcore.org/lab/goal/goalfmt"
	"github.com/pmezard/go-difflib/difflib"
)

var (
	list       = flag.Bool("l", false, "list files whose formatting differs from goalimports'")
	write      = flag.Bool("w", false, "write the result to the (source) file instead of standard output")
	diff       = flag.Bool("d", false, "display diffs instead of rewriting files")
	formatOnly = flag.Bool("format-only", false, "only format the code, as gofmt does, without fixing its imports")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: goalimports [flags] [path ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	opt := &goalfmt.Options{FormatOnly: *formatOnly}
	failed := false
	report := func(err error) {
		fmt.Fprintln(os.Stderr, err)
		failed = true
	}
	if flag.NArg() == 0 {
		if *write {
			report(fmt.Errorf("goalimports: cannot use -w with standard input"))
		} else if err := process("<standard input>", os.Stdin, os.Stdout, opt); err != nil {
			report(err)
		}
	}
	for _, path := range flag.Args() {
		err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			switch {
			case err != nil:
				return err
			case d.IsDir():
				if p != path && (strings.HasPrefix(d.Name(), ".") || d.Name() == "testdata") {
					return filepath.SkipDir
				}
				return nil
			case p != path && filepath.Ext(p) != ".goal":
				return nil
			}
			if err := process(p, nil, os.Stdout, opt); err != nil {
				report(err)
			}
			return nil
		})
		if err != nil {
			report(err)
		}
	}
	if failed {
		os.Exit(2)
	}
}

// process processes the Goal file of the given name, reading it from
// in if it is non-nil, and writing the result to out, or to the file,
// or listing it, or its diffs, according to the flags.
func process(filename string, in io.Reader, out io.Writer, opt *goalfmt.Options) error {
	var src []byte
	var err error
	if in != nil {
		src, err = io.ReadAll(in)
	} else {
		src, err = os.ReadFile(filename)
	}
	if err != nil {
		return err
	}
	res, err := goalfmt.Source(filename, src, opt)
	if err != nil {
		return err
	}
	if bytes.Equal(src, res) {
		if !*list && !*write && !*diff {
			_, err = out.Write(res)
		}
		return err
	}
	if *list {
		fmt.Fprintln(out, filename)
	}
	if *write {
		if in != nil {
			return fmt.Errorf("goalimports: cannot use -w with standard input")
		}
		if err := os.WriteFile(filename, res, 0666); err != nil {
			return err
		}
	}
	if *diff {
		ud := difflib.UnifiedDiff{A: difflib.SplitLines(string(src)), B: difflib.SplitLines(string(res)),
			FromFile: filename + ".orig", ToFile: filename, Context: 3}
		d, err := difflib.GetUnifiedDiffString(ud)
		if err != nil {
			return err
		}
		fmt.Fprint(out, d)
	}
	if !*list && !*write && !*diff {
		_, err = out.Write(res)
	}
	return err
}
