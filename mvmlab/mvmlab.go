// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mvmlab provides functions connecting the mvm Go interpreter
// (https://github.com/cogentcore/mvm) to Cogent Lab: it registers the Goal
// interpreter in [mvmcore.Interpreters], with the symbols of the Cogent Core
// and Cogent Lab packages, so that Goal code blocks of content can be
// edited and run.
package mvmlab

import (
	"fmt"
	"io"
	"reflect"

	"cogentcore.org/core/core"
	"cogentcore.org/core/mvmcore"
	"cogentcore.org/core/mvmcore/coresymbols"
	"cogentcore.org/core/text/textcore"
	"cogentcore.org/lab/goal/interpreter"
	"cogentcore.org/lab/mvmlab/labsymbols"
	"cogentcore.org/lab/tensorfs"
)

func init() {
	mvmcore.Interpreters["Goal"] = func(out io.Writer) (mvmcore.Interpreter, error) {
		return NewInterpreter(out)
	}
}

// BindTextEditor is [mvmcore.BindTextEditor], with the Goal interpreter
// of this package registered. Content apps whose code blocks are compiled
// into them (see cogentcore.org/core/content/cmd/contentcode) set
// content.BindTextEditor to it to run edited code, and can list it in the
// Web.DeferredFunctions of their core.toml, so that js builds load mvm
// and its symbols only when code is first edited.
func BindTextEditor(ed *textcore.Editor, parent *core.Frame, language string) {
	mvmcore.BindTextEditor(ed, parent, language)
}

// Interpreter implements [mvmcore.Interpreter] using the [interpreter.Interpreter] for Goal.
type Interpreter struct {
	*interpreter.Interpreter
}

// NewInterpreter returns a new [Interpreter] writing its output to out, with
// the symbols of the GUI packages of Cogent Core and Cogent Lab imported, in
// addition to those of [interpreter.NewInterpreter], and the names of
// [mvmcore.DotPackage], such as b, the parent widget of the code.
func NewInterpreter(out io.Writer) (*Interpreter, error) {
	in := &Interpreter{interpreter.NewInterpreter(interpreter.Options{Stdout: out})}
	UseGUI(in.Interpreter, map[string]map[string]reflect.Value{mvmcore.DotPackage: mvmcore.DotValues()})
	in.Config()
	if _, err := in.Interp.Eval("goal", fmt.Sprintf("import . %q\n", mvmcore.DotPackage)); err != nil {
		return nil, err
	}
	return in, nil
}

// UseGUI imports into the given Goal interpreter the symbols of the GUI
// packages of Cogent Core and Cogent Lab ([coresymbols] and [labsymbols]),
// and those of the given packages, which replace them, such that they can
// be used without an import. Apps with a GUI that run Goal call it before
// [interpreter.Interpreter.Config].
func UseGUI(in *interpreter.Interpreter, values ...map[string]map[string]reflect.Value) {
	mvmcore.Load()
	labsymbols.Load()
	in.Use(append([]map[string]map[string]reflect.Value{coresymbols.Values, labsymbols.Values}, values...)...)
}

// Eval runs the given Goal code.
func (in *Interpreter) Eval(src string) (res reflect.Value, err error) {
	tensorfs.ListOutput = in.Goal.Config.StdIO.Out
	in.Interpreter.Goal.TrState.MathRecord = true
	res, _, err = in.Interpreter.Eval(src)
	tensorfs.ListOutput = nil
	return
}
