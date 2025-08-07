// Copyright (c) 2020, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/cli"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

//go:generate core generate -add-types -add-funcs

type Config struct {
	// Name of the column to compute stats on.
	Column string `posarg:"0" required:"+"`

	// Files to compute stats on.
	Files []string `posarg:"leftover" required:"+"`
}

func Run(c *Config) error {
	var errs []error
	fmt.Printf("| %44s ", "File")
	for _, st := range stats.DescriptiveStats {
		fmt.Printf("| %12s ", st.String())
	}
	fmt.Printf("|\n")

	for _, f := range c.Files {
		if ok, err := fsx.FileExists(f); !ok {
			errs = append(errs, err, fmt.Errorf("file %q not found", f))
			continue
		}
		dt := table.New()
		dt.OpenCSV(fsx.Filename(f), tensor.Detect)
		cl, err := dt.ColumnTry(c.Column)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		dir, _ := tensorfs.NewDir("Desc")
		stats.Describe(dir, cl)
		ds := dir.Dir("Describe/" + c.Column)
		fmt.Printf("| %44s ", f)
		for _, st := range stats.DescriptiveStats {
			v := ds.Float64(st.String())
			fmt.Printf("| %12.2f ", v.Float(0))
		}
		fmt.Printf("|\n")
	}
	return errors.Join(errs...)
}

func main() {
	opts := cli.DefaultOptions("tstats", "tstats computes standard descriptive statistics on a column of data in a CSV / TSV file.")
	cli.Run(opts, &Config{}, Run)
}
