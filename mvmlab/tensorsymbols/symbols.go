// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tensorsymbols has the mvm symbols of the packages of Cogent Lab
// that do not use the GUI, which interpreted code, such as that of Goal, can
// use: the tensor, table, stats and other math packages (see labsymbols
// for the GUI packages).
package tensorsymbols

//go:generate ./make

import (
	"reflect"
	"sync"
)

// Values are the symbols of the packages, by import path and name, in the
// form that mvm's ImportPackageValues takes. They are added by [Load], which
// is called at init time except on the web (see [Load]).
var Values = map[string]map[string]reflect.Value{}

var (
	// loaders are the functions that add the symbols of each package
	// to [Values], which are registered at init time.
	loaders []func()

	// loadOnce makes [Load] add the symbols once.
	loadOnce sync.Once
)

// register adds the given function that adds symbols to [Values]
// to the functions that [Load] calls.
func register(loader func()) bool {
	loaders = append(loaders, loader)
	return true
}

// add adds the symbols of the given packages to [Values].
func add(pkgs map[string]map[string]reflect.Value) {
	for path, values := range pkgs {
		Values[path] = values
	}
}

// Load adds the symbols of all packages to [Values], once. Natively, it is
// called at init time. On the web, making the symbols at init time is a
// noticeable part of the startup time of an app, so it is instead called
// by the interpreter when it is first made; other code using [Values]
// on the web must call it first.
func Load() {
	loadOnce.Do(func() {
		for _, loader := range loaders {
			loader()
		}
		loaders = nil
	})
}

func init() {
	if !lazyLoad {
		Load()
	}
}
