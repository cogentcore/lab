// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"time"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/lab/tensor"
)

// AllFiles returns all file names within given directory, including subdirectory,
// excluding those matching given glob expressions. Files are relative to dir,
// and do not include the full path.
func AllFiles(dir string, exclude ...string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		for _, ex := range exclude {
			if errors.Log1(filepath.Match(ex, path)) {
				return nil
			}
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// note: Tar code helped significantly by Steve Domino examples:
// https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07

// Tar writes a tar file to given writer, from given source directory,
// using given include function to select nodes to include (all if nil).
// If gz is true, then tar is gzipped.
// The tensor data is written using the [tensor.ToBinary] format, so the
// files are effectively opaque binary files.
func Tar(w io.Writer, dir *Node, gz bool, include func(nd *Node) bool) error {
	ow := w
	if gz {
		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		ow = gzw
	}
	tw := tar.NewWriter(ow)
	defer tw.Close()
	return tarWrite(tw, dir, "", include)
}

func tarWrite(w *tar.Writer, dir *Node, parPath string, include func(nd *Node) bool) error {
	var errs []error
	for _, it := range dir.nodes.Values {
		if include != nil && !include(it) {
			continue
		}
		if it.IsDir() {
			tarWrite(w, it, path.Join(parPath, it.name), include)
			continue
		}
		vtsr := it.Tensor.AsValues()
		b := tensor.ToBinary(vtsr)
		fname := path.Join(parPath, it.name)
		now := time.Now()
		hdr := &tar.Header{
			Name:       fname,
			Mode:       0666,
			Size:       int64(len(b)),
			Format:     tar.FormatPAX,
			ModTime:    now,
			AccessTime: now,
			ChangeTime: now,
		}
		if err := w.WriteHeader(hdr); err != nil {
			errs = append(errs, err)
			break
		}
		_, err := w.Write(b)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Untar extracts a tar file from given reader, into given directory node.
// If gz is true, then tar is gzipped.
func Untar(r io.Reader, dir *Node, gz bool) error {
	or := r
	if gz {
		gzr, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		or = gzr
		defer gzr.Close()
	}
	tr := tar.NewReader(or)
	var errs []error
	addErr := func(err error) error { // if != nil, return
		if err == nil {
			return nil
		}
		errs = append(errs, err)
		if len(errs) > 10 {
			return errors.Join(errs...)
		}
		return nil
	}
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return errors.Join(errs...)
		case err != nil:
			if allErr := addErr(err); allErr != nil {
				return allErr
			}
			continue
		case hdr == nil:
			continue
		}
		fname := hdr.Name
		switch hdr.Typeflag {
		case tar.TypeDir:
			dir.Dir(fname)
		case tar.TypeReg:
			b := make([]byte, hdr.Size)
			_, err := tr.Read(b)
			if err != nil && err != io.EOF {
				fmt.Println("err:", err)
				if allErr := addErr(err); allErr != nil {
					return allErr
				}
				continue
			}
			dr, fn := path.Split(fname)
			pdir := dir
			if dr != "" {
				dr = path.Dir(fname)
				pdir = dir.Dir(dr)
			}
			nd, err := newNode(pdir, fn)
			if err != nil && err != fs.ErrExist {
				if allErr := addErr(err); allErr != nil {
					return allErr
				}
				continue
			}
			nd.Tensor = tensor.FromBinary(b)
		}
	}
	return errors.Join(errs...)
}
