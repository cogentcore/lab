// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package labcoresymbols contains yaegi symbols for lab core packages.
package labcoresymbols

//go:generate ./make

import "reflect"

var Symbols = map[string]map[string]reflect.Value{}
