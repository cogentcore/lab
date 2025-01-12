// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cogentcore.org/core/base/errors"
)

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
