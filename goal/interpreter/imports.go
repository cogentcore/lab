// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"reflect"
)

// goalrunValues returns the symbols of the goalrun package, which makes
// the methods of the goal object available to the code that Goal
// transpiles.
func (in *Interpreter) goalrunValues() map[string]map[string]reflect.Value {
	return map[string]map[string]reflect.Value{
		"cogentcore.org/lab/goalrun": {
			"Run":         reflect.ValueOf(in.Goal.Run),
			"RunErrOK":    reflect.ValueOf(in.Goal.RunErrOK),
			"Output":      reflect.ValueOf(in.Goal.Output),
			"OutputErrOK": reflect.ValueOf(in.Goal.OutputErrOK),
			"Start":       reflect.ValueOf(in.Goal.Start),
			"AddCommand":  reflect.ValueOf(in.Goal.AddCommand),
			"RunCommands": reflect.ValueOf(in.Goal.RunCommands),
			"Args":        reflect.ValueOf(in.Goal.Args),
		},
	}
}
