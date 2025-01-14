// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package yaegilab provides functions connecting
// https://github.com/cogentcore/yaegi to Cogent Lab.
package yaegilab

import (
	"reflect"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/yaegicore"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/yaegilab/gui"
	"cogentcore.org/lab/yaegilab/nogui"
	"github.com/cogentcore/yaegi/interp"
)

func init() {
	yaegicore.Interpreters["Goal"] = func(options interp.Options) yaegicore.Interpreter {
		return NewInterpreter(options)
	}
}

// Interpreter implements [yaegicore.Interpreter] using the [interpreter.Interpreter] for Goal.
type Interpreter struct {
	*interpreter.Interpreter
}

// NewInterpreter returns a new [Interpreter] initialized with the given options.
func NewInterpreter(options interp.Options) *Interpreter {
	return &Interpreter{interpreter.NewInterpreter(options)}
}

func (in *Interpreter) Use(values interp.Exports) error {
	return in.Interp.Use(values)
}

func (in *Interpreter) ImportUsed() {
	errors.Log(in.Use(nogui.Symbols))
	errors.Log(in.Use(gui.Symbols))
	in.Config()
}

func (in *Interpreter) Eval(src string) (res reflect.Value, err error) {
	res, _, err = in.Interpreter.Eval(src)
	return
}
