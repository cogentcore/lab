// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

// Builder is the global container of [physics.Model] elements,
// organized into worlds that are independently updated.
type Builder struct {

	// Worlds are the independent world elements.
	Worlds []World
}
