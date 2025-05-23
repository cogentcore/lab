// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package labsymbols contains yaegi symbols for lab GUI packages.
package labsymbols

//go:generate ./make

import "reflect"

var Symbols = map[string]map[string]reflect.Value{}
