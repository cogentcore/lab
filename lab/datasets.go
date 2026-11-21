// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"embed"
	"io/fs"
)

//go:embed datasets
var datasets embed.FS

// Datasets has standard data files for examples and tutorials, which can
// be opened with [table.Table.OpenFS] and other functions that take an
// [fs.FS]:
//
//   - planets.csv: the planets discovered around other stars (exoplanets)
//     through 2014, from the seaborn datasets, with the columns method,
//     number, orbital_period, mass, distance, and year.
var Datasets, _ = fs.Sub(datasets, "datasets")
