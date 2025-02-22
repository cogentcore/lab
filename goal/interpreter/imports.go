// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"reflect"

	"github.com/cogentcore/yaegi/interp"
)

var Symbols = map[string]map[string]reflect.Value{}

// ImportGoal makes the methods of goal object available in goalrun package.
func (in *Interpreter) ImportGoal() {
	in.Interp.Use(interp.Exports{
		"cogentcore.org/lab/goalrun/goalrun": map[string]reflect.Value{
			"Run":         reflect.ValueOf(in.Goal.Run),
			"RunErrOK":    reflect.ValueOf(in.Goal.RunErrOK),
			"Output":      reflect.ValueOf(in.Goal.Output),
			"OutputErrOK": reflect.ValueOf(in.Goal.OutputErrOK),
			"Start":       reflect.ValueOf(in.Goal.Start),
			"AddCommand":  reflect.ValueOf(in.Goal.AddCommand),
			"RunCommands": reflect.ValueOf(in.Goal.RunCommands),
			"Args":        reflect.ValueOf(in.Goal.Args),
		},
	})
}
