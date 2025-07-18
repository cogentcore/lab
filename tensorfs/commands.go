// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/lab/tensor"
)

var (
	// CurDir is the current working directory.
	CurDir *Node

	// CurRoot is the current root tensorfs system.
	// A default root tensorfs is created at startup.
	CurRoot *Node

	// ListOutput is where to send the output of List commands,
	// if non-nil (otherwise os.Stdout).
	ListOutput io.Writer
)

func init() {
	CurRoot, _ = NewDir("data")
	CurDir = CurRoot
}

// Record saves given tensor to current directory with given name.
func Record(tsr tensor.Tensor, name string) {
	if CurDir == nil {
		CurDir = CurRoot
	}
	SetTensor(CurDir, tsr, name)
}

// Chdir changes the current working tensorfs directory to the named directory.
func Chdir(dir string) error {
	if CurDir == nil {
		CurDir = CurRoot
	}
	if dir == "" {
		CurDir = CurRoot
		return nil
	}
	ndir, err := CurDir.DirAtPath(dir)
	if err != nil {
		return err
	}
	CurDir = ndir
	return nil
}

// Mkdir creates a new directory with the specified name in the current directory.
// It returns an existing directory of the same name without error.
func Mkdir(dir string) *Node {
	if CurDir == nil {
		CurDir = CurRoot
	}
	if dir == "" {
		err := &fs.PathError{Op: "Mkdir", Path: dir, Err: errors.New("path must not be empty")}
		errors.Log(err)
		return nil
	}
	return CurDir.Dir(dir)
}

// List lists files using arguments (options and path) from the current directory.
func List(opts ...string) error {
	if CurDir == nil {
		CurDir = CurRoot
	}

	long := false
	recursive := false
	if len(opts) > 0 && len(opts[0]) > 0 && opts[0][0] == '-' {
		op := opts[0]
		if strings.Contains(op, "l") {
			long = true
		}
		if strings.Contains(op, "r") {
			recursive = true
		}
		opts = opts[1:]
	}
	dir := CurDir
	if len(opts) > 0 {
		nd, err := CurDir.DirAtPath(opts[0])
		if err == nil {
			dir = nd
		}
	}
	ls := dir.List(long, recursive)
	if ListOutput != nil {
		fmt.Fprintln(ListOutput, ls)
	} else {
		fmt.Println(ls)
	}
	return nil
}

// Get returns the tensor value at given path relative to the
// current working directory.
// This is the direct pointer to the node, so changes
// to it will change the node. Clone the tensor to make
// a new copy disconnected from the original.
func Get(name string) tensor.Tensor {
	if CurDir == nil {
		CurDir = CurRoot
	}
	if name == "" {
		err := &fs.PathError{Op: "Get", Path: name, Err: errors.New("name must not be empty")}
		errors.Log(err)
		return nil
	}
	nd, err := CurDir.NodeAtPath(name)
	if errors.Log(err) != nil {
		return nil
	}
	if nd.IsDir() {
		err := &fs.PathError{Op: "Get", Path: name, Err: errors.New("node is a directory, not a data node")}
		errors.Log(err)
		return nil
	}
	return nd.Tensor
}

// Set sets tensor to given name or path relative to the
// current working directory.
// If the node already exists, its previous tensor is updated to the
// given one; if it doesn't, then a new node is created.
func Set(name string, tsr tensor.Tensor) error {
	if CurDir == nil {
		CurDir = CurRoot
	}
	if name == "" {
		err := &fs.PathError{Op: "Set", Path: name, Err: errors.New("name must not be empty")}
		return errors.Log(err)
	}
	itm, err := CurDir.NodeAtPath(name)
	if err == nil {
		if itm.IsDir() {
			err := &fs.PathError{Op: "Set", Path: name, Err: errors.New("existing node is a directory, not a data node")}
			return errors.Log(err)
		}
		itm.Tensor = tsr
		return nil
	}
	cd := CurDir
	dir, name := path.Split(name)
	if dir != "" {
		d, err := CurDir.DirAtPath(dir)
		if err != nil {
			return errors.Log(err)
		}
		cd = d
	}
	SetTensor(cd, tsr, name)
	return nil
}

// SetCopy sets tensor to given name or path relative to the
// current working directory.
// Unlike [Set], this version saves a [tensor.Clone] of the tensor,
// so future changes to the tensor do not affect this value.
func SetCopy(name string, tsr tensor.Tensor) error {
	return Set(name, tensor.Clone(tsr))
}
