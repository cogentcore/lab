// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command goalls is a language server for Goal files, for editors, which
// is a proxy for the Go language server, gopls, for the Go code of the
// Goal files (see package goalls). It is used instead of gopls, for both
// Go and Goal files: its arguments are passed to gopls, except for
// -nowrite, which stops it writing the Go file of a Goal file when it
// is saved, as `goal build` does.
//
//	goalls [-nowrite] [gopls arguments]
package main

import (
	"fmt"
	"os"

	"cogentcore.org/lab/goal/goalls"
)

func main() {
	s := &goalls.Server{Gopls: []string{"gopls"}}
	for _, arg := range os.Args[1:] {
		if arg == "-nowrite" {
			s.NoWrite = true
		} else {
			s.Gopls = append(s.Gopls, arg)
		}
	}
	// The standard output is for the messages to the client: anything
	// else printed goes to the standard error.
	out := os.Stdout
	os.Stdout = os.Stderr
	if err := s.Run(os.Stdin, out); err != nil {
		fmt.Fprintln(os.Stderr, "goalls:", err)
		os.Exit(1)
	}
}
