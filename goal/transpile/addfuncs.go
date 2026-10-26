// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"path"
	"reflect"
	"strings"

	"cogentcore.org/lab/mvmlab/tensorsymbols"
	"cogentcore.org/lab/tensor"
)

func init() {
	AddTensorFuncs()
}

var tensorPackages = []string{"/lab/tensor", "/lab/stats", "/lab/vector", "/lab/matrix"}

// AddTensorFuncs grabs all tensor* package functions of the interpreter
// symbols (see [tensorsymbols]) and adds them to the `tensor.Funcs` map so we can
// properly convert symbols to either tensors or basic literals,
// depending on the arg types for the current function.
func AddTensorFuncs() {
	tensorsymbols.Load()
	for pth, symap := range tensorsymbols.Values {
		has := false
		for _, p := range tensorPackages {
			if strings.Contains(pth, p) {
				has = true
				break
			}
		}
		if !has {
			continue
		}
		_, pkg := path.Split(pth)
		for name, val := range symap {
			if val.Kind() != reflect.Func {
				continue
			}
			pnm := pkg + "." + name
			tensor.AddFunc(pnm, val.Interface())
		}
	}
}
