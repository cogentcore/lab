// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos)

package interpreter

import "golang.org/x/term"

// makeRaw puts the given terminal into raw mode,
// returning a function that restores it.
func makeRaw(fd int) (func() error, error) {
	st, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return func() error { return term.Restore(fd, st) }, nil
}
