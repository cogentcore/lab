// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slrand provides random number generation functions that run
// identically on the GPU (in WGSL, from slrand.wgsl) and the CPU (in Go),
// based on counter-based Philox2x32 generators.
package slrand
