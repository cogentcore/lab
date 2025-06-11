// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"cogentcore.org/core/base/errors"
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

// TarFiles writes a tar file to given writer, from given source directory.
// Tar file names are as listed here so it will unpack directly to those files.
// If gz is true, then tar is gzipped.
func TarFiles(w io.Writer, dir string, gz bool, files ...string) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("TarFiles: directory not accessible: %s", err.Error())
	}
	ow := w
	if gz {
		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		ow = gzw
	}
	tw := tar.NewWriter(ow)
	defer tw.Close()

	var errs []error
	for _, fn := range files {
		fname := filepath.Join(dir, fn)
		fi, err := os.Stat(fname)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		hdr, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		hdr.Name = fn
		if err := tw.WriteHeader(hdr); err != nil {
			errs = append(errs, err)
			break
		}
		f, err := os.Open(fname)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if _, err := io.Copy(tw, f); err != nil {
			errs = append(errs, err)
			break
		}
		f.Close()
	}
	return errors.Join(errs...)
}

// Untar extracts a tar file from given reader, into given source directory.
// If gz is true, then tar is gzipped.
func Untar(r io.Reader, dir string, gz bool) error {
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
		fn := filepath.Join(dir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(fn, 0755)
			if allErr := addErr(err); allErr != nil {
				return allErr
			}
		case tar.TypeReg:
			f, err := os.OpenFile(fn, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(hdr.Mode))
			if allErr := addErr(err); allErr != nil {
				return allErr
			}
			_, err = io.Copy(f, tr)
			f.Close()
			if allErr := addErr(err); allErr != nil {
				return allErr
			}
		}
	}
	return errors.Join(errs...)
}
