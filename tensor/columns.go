// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensor

import "fmt"

// NamedColumns is implemented by types with named columns of tensors,
// such as a table.Table, which math mode indexes by name:
// dt["mass"] gets the mass column, and dt["mass"] = x sets it.
type NamedColumns interface {
	// Column returns the column of the given name, nil if there is none.
	Column(name string) *Rows

	// SetColumn sets the values of the column of the given name,
	// adding it if there is none.
	SetColumn(name string, values Tensor) error
}

// ColumnByName returns the column of the given name of x, which must be
// a [NamedColumns] such as a table.Table, for math mode indexing by name,
// as in dt["mass"]. It panics if x has no columns or no such column, as
// math mode indexing does for an index out of range.
func ColumnByName(x any, name string) Tensor {
	nc, ok := x.(NamedColumns)
	if !ok {
		panic(fmt.Sprintf("tensor: %T has no named columns, for the column %q", x, name))
	}
	cl := nc.Column(name)
	if cl == nil {
		panic(fmt.Sprintf("tensor: no column named %q", name))
	}
	return cl
}

// SetColumnByName sets the values of the column of the given name of x,
// which must be a [NamedColumns] such as a table.Table, to the given
// value (a tensor or a Go value, converted with [ToTensor]), adding the
// column if there is none, for math mode assignment by name, as in
// dt["mass"] = x. It panics if x has no columns, or the values do not fit.
func SetColumnByName(x any, name string, val any) {
	nc, ok := x.(NamedColumns)
	if !ok {
		panic(fmt.Sprintf("tensor: %T has no named columns, for the column %q", x, name))
	}
	if err := nc.SetColumn(name, ToTensor(val)); err != nil {
		panic(err)
	}
}
