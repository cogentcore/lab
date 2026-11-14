// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs

import (
	"errors"
	"io/fs"
	"path"
	"strings"
)

// write.go has the operations that change the filesystem by paths, as the
// hackpadfs interfaces of those operations (MkdirAllFS, RenameFS,
// RemoveAllFS) and filetree.CopyFS, so that its files can be changed in a
// file tree: made, renamed (moved), removed, and copied.

// pathNames returns the names of the nodes of the given path within this
// directory, without that of this directory itself, if it starts with it
// (as the paths of a file tree of it do), or ".".
func (dir *Node) pathNames(p string) []string {
	p = path.Clean(p)
	if p == "." || p == "/" {
		return nil
	}
	names := strings.Split(strings.TrimPrefix(p, "/"), "/")
	if names[0] == "." || (names[0] == dir.name && dir.nodes.At(dir.name) == nil) {
		names = names[1:]
	}
	return names
}

// parentAt returns the directory of the given path within this directory,
// which must exist, and the name of the node in it.
func (dir *Node) parentAt(op, p string) (*Node, string, error) {
	names := dir.pathNames(p)
	if len(names) == 0 {
		return nil, "", &fs.PathError{Op: op, Path: p, Err: errors.New("cannot change the root directory")}
	}
	cur := dir
	for _, nm := range names[:len(names)-1] {
		nd, ok := cur.nodes.AtTry(nm)
		if !ok {
			return nil, "", &fs.PathError{Op: op, Path: p, Err: fs.ErrNotExist}
		}
		if !nd.IsDir() {
			return nil, "", &fs.PathError{Op: op, Path: p, Err: errors.New("not a directory: " + nm)}
		}
		cur = nd
	}
	return cur, names[len(names)-1], nil
}

// nodeAt returns the node of the given path within this directory, which
// is the directory itself for "." or its own name.
func (dir *Node) nodeAt(op, p string) (*Node, error) {
	if len(dir.pathNames(p)) == 0 {
		return dir, nil
	}
	pd, name, err := dir.parentAt(op, p)
	if err != nil {
		return nil, err
	}
	nd, ok := pd.nodes.AtTry(name)
	if !ok {
		return nil, &fs.PathError{Op: op, Path: p, Err: fs.ErrNotExist}
	}
	return nd, nil
}

// MkdirAll makes the directory at the given path within this directory,
// with any of its parent directories that do not exist yet. The
// permissions are not used.
func (dir *Node) MkdirAll(p string, perm fs.FileMode) error {
	if err := dir.mustDir("MkdirAll", p); err != nil {
		return err
	}
	cur := dir
	for _, nm := range dir.pathNames(p) {
		if nd, ok := cur.nodes.AtTry(nm); ok {
			if !nd.IsDir() {
				return &fs.PathError{Op: "MkdirAll", Path: p, Err: errors.New("not a directory: " + nm)}
			}
			cur = nd
			continue
		}
		nd, err := NewDir(nm, cur)
		if err != nil {
			return err
		}
		cur = nd
	}
	return nil
}

// Rename renames (moves) the node at the old path within this directory
// to the new path, which does not exist, keeping its time (and so its
// order in the directory by time).
func (dir *Node) Rename(oldpath, newpath string) error {
	nd, err := dir.nodeAt("Rename", oldpath)
	if err != nil {
		return err
	}
	if nd == dir {
		return &fs.PathError{Op: "Rename", Path: oldpath, Err: errors.New("cannot rename the root directory")}
	}
	pd, name, err := dir.parentAt("Rename", newpath)
	if err != nil {
		return err
	}
	if _, ok := pd.nodes.AtTry(name); ok {
		return &fs.PathError{Op: "Rename", Path: newpath, Err: fs.ErrExist}
	}
	for d := pd; d != nil; d = d.Parent {
		if d == nd {
			return &fs.PathError{Op: "Rename", Path: newpath, Err: errors.New("cannot move a directory into itself")}
		}
	}
	nd.Parent.nodes.DeleteByKey(nd.name)
	nd.name = name
	return pd.Add(nd)
}

// RemoveAll removes the node at the given path within this directory,
// with all that it contains, if it exists.
func (dir *Node) RemoveAll(p string) error {
	nd, err := dir.nodeAt("RemoveAll", p)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if nd == dir {
		return &fs.PathError{Op: "RemoveAll", Path: p, Err: errors.New("cannot remove the root directory")}
	}
	nd.Parent.nodes.DeleteByKey(nd.name)
	nd.Parent = nil
	return nil
}

// CopyPath copies the node at the source path within this directory to
// the destination path, as a clone (see [Node.Clone]), replacing anything
// there, which is newer than the other nodes of its directory.
func (dir *Node) CopyPath(dst, src string) error {
	nd, err := dir.nodeAt("CopyPath", src)
	if err != nil {
		return err
	}
	pd, name, err := dir.parentAt("CopyPath", dst)
	if err != nil {
		return err
	}
	for d := pd; d != nil; d = d.Parent {
		if d == nd {
			return &fs.PathError{Op: "CopyPath", Path: dst, Err: errors.New("cannot copy a directory into itself")}
		}
	}
	cp := nd.Clone()
	cp.name = name
	pd.nodes.DeleteByKey(name)
	return pd.Add(cp)
}
