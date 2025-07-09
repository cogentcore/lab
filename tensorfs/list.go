// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs

import (
	"strings"

	"cogentcore.org/core/base/indent"
)

const (
	// Short is used as a named arg for the [Node.List] method
	// for a short, name-only listing, vs. [Long].
	Short = false

	// Long is used as a named arg for the [Node.List] method
	// for a long, name and size listing, vs. [Short].
	Long = true

	// DirOnly is used as a named arg for the [Node.List] method
	// for only listing the current directory, vs. [Recursive].
	DirOnly = false

	// Recursive is used as a named arg for the [Node.List] method
	// for listing all directories recursively, vs. [DirOnly].
	Recursive = true
)

func (nd *Node) String() string {
	if !nd.IsDir() {
		lb := nd.Tensor.Label()
		if !strings.HasPrefix(lb, nd.name) {
			lb = nd.name + " " + lb
		}
		return lb
	}
	return nd.List(Short, DirOnly)
}

// ListAll returns a Long, Recursive listing of nodes in the given directory.
func (dir *Node) ListAll() string {
	return dir.listLong(true, 0)
}

// List returns a listing of nodes in the given directory.
//   - long = include detailed information about each node, vs just the name.
//   - recursive = descend into subdirectories.
func (dir *Node) List(long, recursive bool) string {
	if long {
		return dir.listLong(recursive, 0)
	}
	return dir.listShort(recursive, 0)
}

// listShort returns a name-only listing of given directory.
func (dir *Node) listShort(recursive bool, ident int) string {
	var b strings.Builder
	nodes, _ := dir.Nodes()
	for _, it := range nodes {
		b.WriteString(indent.Tabs(ident))
		if it.IsDir() {
			if recursive {
				b.WriteString("\n" + it.listShort(recursive, ident+1))
			} else {
				b.WriteString(it.name + "/ ")
			}
		} else {
			b.WriteString(it.name + " ")
		}
	}
	return b.String()
}

// listLong returns a detailed listing of given directory.
func (dir *Node) listLong(recursive bool, ident int) string {
	var b strings.Builder
	nodes, _ := dir.Nodes()
	for _, it := range nodes {
		b.WriteString(indent.Tabs(ident))
		if it.IsDir() {
			b.WriteString(it.name + "/\n")
			if recursive {
				b.WriteString(it.listLong(recursive, ident+1))
			}
		} else {
			b.WriteString(it.String() + "\n")
		}
	}
	return b.String()
}
