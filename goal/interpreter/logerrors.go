// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"runtime"
	"strings"
)

// LogError is an error that the Go code that the Goal code calls logged
// (with cogentcore.org/core/base/errors.Log), such as an error of the
// shapes of tensors in a tensor function, which stops the Goal code at
// the call, as an error of it (see [logHook]).
type LogError struct {
	Err error
}

func (e *LogError) Error() string { return e.Err.Error() }

func (e *LogError) Unwrap() error { return e.Err }

// logHook is the hook of errors.SetLogHook while the Goal code runs: an
// error that the code logs, on the goroutine that runs it, panics as a
// [LogError] instead, which the interpreter turns into a panic of the
// Goal code at the call, which stops it. Other errors are logged.
func logHook(err error) bool {
	if !inInterpreter() {
		return false
	}
	panic(&LogError{Err: err})
}

// inInterpreter returns whether the current goroutine runs code of the
// mvm interpreter, which is in the stack of the calls of the code.
func inInterpreter() bool {
	pcs := make([]uintptr, 64)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		f, more := frames.Next()
		if strings.HasPrefix(f.Function, "github.com/cogentcore/mvm/vm.") {
			return true
		}
		if !more {
			return false
		}
	}
}
