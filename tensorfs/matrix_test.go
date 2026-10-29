// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorfs_test

import (
	"math"
	"testing"

	"cogentcore.org/core/base/metadata"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
)

func TestDirTableMatrix(t *testing.T) {
	dir, _ := tensorfs.NewDir("root")
	a := tensor.NewFloat64FromValues(1, 2, 3, 4)
	metadata.SetName(a, "a")
	b := tensor.NewFloat64FromValues(10, 20, 30, 40)
	metadata.SetName(b, "b")
	stats.Describe(dir, a, b)
	dd := dir.Dir("Describe")
	tensorfs.Scalar[float64](dd, "Ignored") // not in a subdirectory
	tensorfs.Scalar[float64](dd.Dir("b"), "Extra").SetFloat1D(7, 0)

	dt := tensorfs.DirTableMatrix(dd, nil)
	var names []string
	for _, st := range stats.DescriptiveStats {
		names = append(names, st.String())
	}
	names = append(names, "Extra")
	assert.Equal(t, []string{"Name", "a", "b"}, dt.Columns.Keys)
	assert.Equal(t, len(names), dt.NumRows())
	for i, nm := range names {
		assert.Equal(t, nm, dt.Column("Name").String1D(i))
	}
	row := func(nm string) int {
		for i := range dt.NumRows() {
			if dt.Column("Name").String1D(i) == nm {
				return i
			}
		}
		return -1
	}
	assert.Equal(t, 4.0, dt.Column("a").Float1D(row("Count")))
	assert.Equal(t, 2.5, dt.Column("a").Float1D(row("Mean")))
	assert.Equal(t, 25.0, dt.Column("b").Float1D(row("Mean")))
	assert.Equal(t, 40.0, dt.Column("b").Float1D(row("Max")))
	assert.Equal(t, 7.0, dt.Column("b").Float1D(row("Extra")))
	assert.True(t, math.IsNaN(dt.Column("a").Float1D(row("Extra"))))

	// the filter selects subdirectories and elements
	dt = tensorfs.DirTableMatrix(dd, func(nd *tensorfs.Node) bool {
		return nd.IsDir() || nd.Name() == "Mean"
	})
	assert.Equal(t, []string{"Name", "a", "b"}, dt.Columns.Keys)
	assert.Equal(t, 1, dt.NumRows())
	assert.Equal(t, "Mean", dt.Column("Name").String1D(0))
}

// TestDirTableMatrixGroups tests a matrix of the results of
// stats.GroupDescribe, with multiple values and strings.
func TestDirTableMatrixGroups(t *testing.T) {
	dir, _ := tensorfs.NewDir("root")
	gd := dir.Dir("Stats").Dir("Group")
	for _, vn := range []string{"x", "y"} {
		vd := gd.Dir(vn)
		g := tensorfs.Value[string](vd, "Group", 2)
		g.SetString1D("a", 0)
		g.SetString1D("b", 1)
		m := tensorfs.Value[float64](vd, "Mean", 2)
		m.SetFloat1D(1.5, 0)
		m.SetFloat1D(2.5, 1)
	}
	tensorfs.Value[float64](gd.Dir("z"), "Mean", 2).SetFloat1D(7, 1) // no strings

	dt := tensorfs.DirTableMatrix(gd, nil)
	assert.Equal(t, []string{"Name", "x", "y", "z"}, dt.Columns.Keys)
	assert.Equal(t, 2, dt.NumRows())
	assert.Equal(t, "Group", dt.Column("Name").String1D(0))
	assert.Equal(t, "Mean", dt.Column("Name").String1D(1))
	x := dt.Column("x")
	assert.True(t, x.IsString())
	assert.Equal(t, []int{2, 2}, x.ShapeSizes())
	assert.Equal(t, "a", x.StringRow(0, 0))
	assert.Equal(t, "b", x.StringRow(0, 1))
	assert.Equal(t, "2.5", x.StringRow(1, 1))
	z := dt.Column("z")
	assert.False(t, z.IsString())
	assert.True(t, math.IsNaN(z.FloatRow(0, 0)))
	assert.Equal(t, 7.0, z.FloatRow(1, 1))
}

func TestHasOnly(t *testing.T) {
	dir, _ := tensorfs.NewDir("root")
	empty := dir.Dir("Empty")
	data := dir.Dir("Data")
	tensorfs.Scalar[float64](data, "X")
	dirs := dir.Dir("Dirs")
	dirs.Dir("Sub")
	mixed := dir.Dir("Mixed")
	mixed.Dir("Sub")
	tensorfs.Scalar[float64](mixed, "X")
	x := data.Node("X")

	for _, tc := range []struct {
		nd          *tensorfs.Node
		dirs, datas bool
	}{
		{empty, false, false},
		{data, false, true},
		{dirs, true, false},
		{mixed, false, false},
		{dir, true, false},
		{x, false, false},
	} {
		assert.Equal(t, tc.dirs, tc.nd.HasOnlyDirs(), "%s HasOnlyDirs", tc.nd.Name())
		assert.Equal(t, tc.datas, tc.nd.HasOnlyData(), "%s HasOnlyData", tc.nd.Name())
	}
}
