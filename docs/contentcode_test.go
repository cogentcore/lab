// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !noyaegi

//go:debug randseednop=0

package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"cogentcore.org/core/content"
	"cogentcore.org/core/core"
	"cogentcore.org/core/yaegicore/basesymbols"
	"cogentcore.org/core/yaegicore/coresymbols"
	"cogentcore.org/lab/yaegilab"
	"github.com/cogentcore/yaegi/interp"
)

var goalFence = regexp.MustCompile("(?m)^(\\s*)```Goal\\s*\n((?:.*\n)*?)\\s*```")

// goalSources returns the Goal code blocks of the given page of the docs.
func goalSources(t *testing.T, file string) []string {
	b, err := fs.ReadFile(econtent, "content/"+file)
	if err != nil {
		t.Fatal(err)
	}
	var srcs []string
	for _, m := range goalFence.FindAllStringSubmatch(string(b), -1) {
		var lines []string
		for _, l := range strings.Split(strings.TrimSuffix(m[2], "\n"), "\n") {
			lines = append(lines, strings.TrimPrefix(l, m[1]))
		}
		srcs = append(srcs, strings.Join(lines, "\n")+"\n")
	}
	return srcs
}

// TestCompiledGoal runs the Goal code of every page of the docs that is
// compiled into the app (contentcode_gen.go), block by block in order, as
// the docs do when they show the page, and the same code in the yaegi Goal
// interpreter, and compares what each block prints.
func TestCompiledGoal(t *testing.T) {
	b := core.NewBody("code blocks")
	parent := core.NewFrame(b)
	basesymbols.Load()
	coresymbols.Load()
	var files []string
	for f := range content.CompiledGoal {
		files = append(files, f)
	}
	sort.Strings(files)
	nblocks, same := 0, 0
	for _, file := range files {
		pg := content.CompiledGoal[file]
		srcs := goalSources(t, file)
		if len(srcs) < len(pg.Keys) {
			t.Errorf("%s: %d compiled blocks, %d in the page", file, len(pg.Keys), len(srcs))
			continue
		}

		// compiled
		rand.Seed(1)
		var compiled []string
		next := pg.Start()
		for i, key := range pg.Keys {
			if content.CodeKey(srcs[i]) != key {
				t.Errorf("%s: block %d: key %s, page source %s", file, i, key, content.CodeKey(srcs[i]))
			}
			var out bytes.Buffer
			err := runTimeout(func() {
				parent.DeleteChildren()
				core.ExternalParent = parent
				next = next(parent, &out)
				parent.Update()
			})
			if err != nil {
				t.Errorf("%s: block %d: %v", file, i, err)
				break
			}
			compiled = append(compiled, out.String())
			nblocks++
		}

		// interpreted
		var interpOut bytes.Buffer
		in := yaegilab.NewInterpreter(interp.Options{Stdout: &interpOut})
		ip := new(*core.Frame)
		coresymbols.Symbols["."]["b"] = reflect.ValueOf(ip).Elem()
		in.Use(basesymbols.Symbols)
		in.Use(coresymbols.Symbols)
		in.ImportUsed()
		rand.Seed(1)
		for i := range compiled {
			interpOut.Reset()
			err := runTimeout(func() {
				parent.DeleteChildren()
				core.ExternalParent = parent
				*ip = parent
				_, err := in.Eval(srcs[i])
				if err != nil {
					panic(err)
				}
				parent.Update()
			})
			if err != nil {
				t.Logf("%s: block %d: interpreter: %v", file, i, err)
			}
			if got := interpOut.String(); got != compiled[i] {
				t.Errorf("%s: block %d: output differs:\ncompiled:\n%s\ninterpreted:\n%s", file, i, compiled[i], got)
			} else {
				same++
			}
		}
	}
	t.Logf("compiled Goal blocks: %d ran, %d with the same output as the interpreter, %d pages", nblocks, same, len(files))
}

// runTimeout runs f, and returns its panic as an error.
func runTimeout(f func()) error {
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		f()
		done <- nil
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		return fmt.Errorf("still running after 10s")
	}
}
