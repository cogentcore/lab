// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"testing"

	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	dt := table.New().SetNumRows(4)
	dt.AddStringColumn("Name")
	dt.AddFloat32Column("Value")
	for i := range dt.NumRows() {
		gp := "A"
		if i >= 2 {
			gp = "B"
		}
		dt.Column("Name").SetStringRow(gp, i, 0)
		dt.Column("Value").SetFloatRow(float64(i), i, 0)
	}
	dir, _ := tensorfs.NewDir("Group")
	err := TableGroups(dir, dt, "Name")
	assert.NoError(t, err)

	ixs := dir.ValuesFunc(nil)
	assert.Equal(t, []int{0, 1}, tensor.AsInt(ixs[0]).Values)
	assert.Equal(t, []int{2, 3}, tensor.AsInt(ixs[1]).Values)

	err = TableGroupStats(dir, StatMean, dt, "Value")
	assert.NoError(t, err)

	// the group values are in the value directory, with the stat
	vdt := tensorfs.DirTable(dir.Dir("Stats/Name/Value"), nil)
	assert.Equal(t, []string{"Name", "Mean"}, vdt.Columns.Keys)

	err = TableGroupStats(dir, StatMax, dt, "Value")
	assert.NoError(t, err)
	sdt := GroupStatsDirTable(dir.Dir("Stats/Name"))
	assert.Equal(t, []string{"Name", "Value/Mean", "Value/Max"}, sdt.Columns.Keys)
	assert.Equal(t, 3.0, sdt.Column("Value/Max").Float1D(1))

	gdt := GroupStatsAsTableNoStatName(dir)
	assert.Equal(t, []string{"Name", "Value", "Value/Max"}, gdt.Columns.Keys)
	assert.Equal(t, 0.5, gdt.Column("Value").Float1D(0))
	assert.Equal(t, 2.5, gdt.Column("Value").Float1D(1))
	assert.Equal(t, "A", gdt.Column("Name").String1D(0))
	assert.Equal(t, "B", gdt.Column("Name").String1D(1))
}

func TestGroupDescribe(t *testing.T) {
	dt := table.New().SetNumRows(4)
	dt.AddStringColumn("Name")
	dt.AddIntColumn("Value")
	dt.AddFloat64Column("Other")
	for i := range dt.NumRows() {
		gp := "A"
		if i >= 2 {
			gp = "B"
		}
		dt.Column("Name").SetStringRow(gp, i, 0)
		dt.Column("Value").SetFloatRow(float64(i), i, 0)
		dt.Column("Other").SetFloatRow(float64(10*i), i, 0)
	}
	dir, _ := tensorfs.NewDir("Group")
	assert.NoError(t, TableGroups(dir, dt, "Name"))
	assert.NoError(t, TableGroupDescribe(dir, dt, "Value", "Other"))

	// the group values are the first column of each value directory
	nd := dir.Dir("Stats").Dir("Name")
	for _, vn := range []string{"Value", "Other"} {
		vdt := tensorfs.DirTable(nd.Dir(vn), nil)
		assert.Equal(t, "Name", vdt.Columns.Keys[0])
		assert.Equal(t, len(DescriptiveStats)+1, vdt.NumColumns())
		assert.Equal(t, "A", vdt.Column("Name").String1D(0))
		assert.Equal(t, "B", vdt.Column("Name").String1D(1))
	}
	vdt := tensorfs.DirTable(nd.Dir("Value"), nil)
	assert.Equal(t, 0.5, vdt.Column("Mean").Float1D(0)) // not truncated
	assert.Equal(t, 2.5, vdt.Column("Mean").Float1D(1))
}

/*
func TestAggEmpty(t *testing.T) {
	dt := table.New().SetNumRows(4)
	dt.AddStringColumn("Group")
	dt.AddFloat32Column("Value")
	for i := 0; i < dt.Rows; i++ {
		gp := "A"
		if i >= 2 {
			gp = "B"
		}
		dt.SetString("Group", i, gp)
		dt.SetFloat("Value", i, float64(i))
	}
	ix := table.NewIndexed(dt)
	ix.Filter(func(et *table.Table, row int) bool {
		return false // exclude all
	})
	spl := GroupBy(ix, "Group")
	assert.Equal(t, 1, len(spl.Splits))

	AggColumn(spl, "Value", stats.Mean)

	st := spl.AggsToTable(table.ColumnNameOnly)
	if st == nil {
		t.Error("AggsToTable should not be nil!")
	}
}
*/

func TestQuantilesOneValue(t *testing.T) {
	one := tensor.NewFloat64FromValues(2007)
	for _, st := range []Stats{StatQ1, StatMedian, StatQ3} {
		assert.Equal(t, 2007.0, st.Call(one).Float1D(0), st.String())
	}
}
