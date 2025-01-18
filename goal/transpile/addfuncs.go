// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"path"
	"reflect"
	"strings"

	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/yaegilab/labsymbols"
)

func init() {
	AddYaegiTensorFuncs()
}

var yaegiTensorPackages = []string{"/lab/tensor", "/lab/stats", "/lab/vector", "/lab/matrix"}

// AddYaegiTensorFuncs grabs all tensor* package functions registered
// in yaegicore and adds them to the `tensor.Funcs` map so we can
// properly convert symbols to either tensors or basic literals,
// depending on the arg types for the current function.
func AddYaegiTensorFuncs() {
	for pth, symap := range labsymbols.Symbols {
		has := false
		for _, p := range yaegiTensorPackages {
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
