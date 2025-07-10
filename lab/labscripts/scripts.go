// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package labscripts

import (
	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/logx"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/lab"
	"github.com/cogentcore/yaegi/interp"
)

func init() {
	lab.RunScriptCode = RunScriptCode
}

// br.Interpreter = in
// if br.Interpreter == nil {
// 	br.InitInterp()
// 	in = br.Interpreter
// }
// in.Interp.Use(coresymbols.Symbols) // gui imports

// Interpreter returns the interpreter for given browser,
// or nil and an error message if not set.
func Interpreter(br *lab.Browser) (*interpreter.Interpreter, error) {
	if br.Interpreter == nil {
		return nil, errors.New("No interpreter has been set for the Browser, cannot run script")
	}
	return br.Interpreter.(*interpreter.Interpreter), nil
}

// InitInterpreter initializes a new interpreter if not already set.
func InitInterpreter(br *lab.Browser) {
	if br.Interpreter == nil {
		br.Interpreter = interpreter.NewInterpreter(interp.Options{})
	}
	// logx.UserLevel = slog.LevelDebug // for debugging of init loading
}

// RunScript runs given script from list of Scripts in the Browser.
func RunScript(br *lab.Browser, scriptName string) error {
	in, err := Interpreter(br)
	if err != nil {
		return errors.Log(err)
	}
	sc, ok := br.Scripts[scriptName]
	if !ok {
		err := errors.New("script name not found: " + scriptName)
		return errors.Log(err)
	}
	logx.PrintlnDebug("\n################\nrunning script:\n", sc, "\n")
	_, _, err = in.Eval(sc)
	if err == nil {
		err = in.Goal.TrState.DepthError()
	}
	in.Goal.TrState.ResetDepth()
	return err
}

// RunScriptCode runs given script code string in Browser's interpreter.
func RunScriptCode(br *lab.Browser, code string) error {
	in, err := Interpreter(br)
	if err != nil {
		return err
	}
	_, _, err = in.Eval(code)
	return err
}
