// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"testing"

	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// the plots package is not available here, so the standard plotter
	// names used by NewTablePlot are registered as test plotters.
	RegisterPlotter("XY", "test XY", []Roles{Y}, []Roles{X, Size}, newTestPlotter)
	RegisterPlotter("Bar", "test Bar", []Roles{Y}, []Roles{High, X}, newTestPlotter)
	RegisterPlotter("TestXY", "requires X", []Roles{X, Y}, []Roles{Size}, newTestPlotter)
}

// testTable returns a table with X, X2, Y1, Y2 and Name columns.
func testTable() *table.Table {
	dt := table.New("Test")
	dt.AddFloat64Column("X")
	dt.AddFloat64Column("X2")
	dt.AddFloat64Column("Y1")
	dt.AddFloat64Column("Y2")
	dt.AddStringColumn("Name")
	dt.SetNumRows(6)
	for i := range 6 {
		dt.Column("X").SetFloatRow(float64(i), i, 0)
		dt.Column("X2").SetFloatRow(float64(10*i), i, 0)
		dt.Column("Y1").SetFloatRow(float64(i*i), i, 0)
		dt.Column("Y2").SetFloatRow(float64(-i), i, 0)
		nm := "A"
		if i >= 3 {
			nm = "B"
		}
		dt.Column("Name").SetStringRow(nm, i, 0)
	}
	return dt
}

// styleCol sets the stylers for the named column.
func styleCol(dt *table.Table, name string, f ...func(s *Style)) {
	SetStyler(dt.Columns.At(name), f...)
}

// TestTablePlot tests a basic table plot with two Y columns
// against one X column, with a legend.
func TestTablePlot(t *testing.T) {
	dt := testTable()
	styleCol(dt, "X", func(s *Style) { s.Role = X })
	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
		s.Plot.Title = "Table"
	})
	styleCol(dt, "Y2", func(s *Style) {
		s.On = true
		s.Role = Y
		s.Label = "Second"
	})
	plt, err := NewTablePlot(dt)
	require.NoError(t, err)
	require.Len(t, plt.Plotters, 2)
	require.Len(t, plt.Legend.Entries, 2)
	assert.Equal(t, "Y1", plt.Legend.Entries[0].Text)
	assert.Equal(t, "Second", plt.Legend.Entries[1].Text)

	d, _, _ := plt.Plotters[1].Data()
	assert.Equal(t, 5.0, d[X].Float1D(5))
	assert.Equal(t, -5.0, d[Y].Float1D(5))

	plt.RenderImage()
	assert.Equal(t, "Table", plt.Title.Text)
	assert.Equal(t, "X", plt.X.Label.Text)
	assert.Equal(t, minmax.F64{Min: 0, Max: 5}, plt.X.Range)
	assert.Equal(t, minmax.F64{Min: -5, Max: 25}, plt.Y.Range)
}

// TestTablePlotGlobalX tests the XAxis.Column plot style for selecting
// the X axis column, including an invalid column name.
func TestTablePlotGlobalX(t *testing.T) {
	dt := testTable()
	styleCol(dt, "X", func(s *Style) { s.Role = X })
	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
		s.Plot.XAxis.Column = "X2"
	})
	plt, err := NewTablePlot(dt)
	require.NoError(t, err)
	require.Len(t, plt.Plotters, 1)
	assert.Empty(t, plt.Legend.Entries) // only one
	d, _, _ := plt.Plotters[0].Data()
	assert.Equal(t, 50.0, d[X].Float1D(5))
	plt.applyStyle()
	assert.Equal(t, "X2", plt.X.Label.Text)

	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
		s.Plot.XAxis.Column = "Nope"
	})
	plt, err = NewTablePlot(dt)
	assert.Error(t, err)
	require.NotNil(t, plt)
	d, _, _ = plt.Plotters[0].Data()
	assert.Equal(t, 5.0, d[X].Float1D(5)) // falls back on the X role
}

// TestTablePlotGroups tests a group of columns providing the roles of one
// plotter, along with a required role that is missing.
func TestTablePlotGroups(t *testing.T) {
	dt := testTable()
	styleCol(dt, "X2", func(s *Style) {
		s.Role = X
		s.Group = "G"
	})
	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
		s.Group = "G"
		s.Plotter = "TestXY"
	})
	styleCol(dt, "Y2", func(s *Style) {
		s.On = true
		s.Role = Size
		s.Group = "G"
	})
	plt, err := NewTablePlot(dt)
	require.NoError(t, err)
	require.Len(t, plt.Plotters, 1) // the group is done once
	d, _, _ := plt.Plotters[0].Data()
	assert.Equal(t, 50.0, d[X].Float1D(5))

	// no X column for the required X role
	dt = testTable()
	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
		s.Plotter = "TestXY"
	})
	plt, err = NewTablePlot(dt)
	assert.Error(t, err)
	assert.Nil(t, plt)
}

// TestTablePlotSplit tests splitting the data by the values of a Split
// column, which makes a plotter and legend entry for each value.
func TestTablePlotSplit(t *testing.T) {
	dt := testTable()
	styleCol(dt, "X", func(s *Style) { s.Role = X })
	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
	})
	styleCol(dt, "Name", func(s *Style) {
		s.On = true
		s.Role = Split
	})
	plt, err := NewTablePlot(dt)
	require.NoError(t, err)
	require.Len(t, plt.Plotters, 2)
	require.Len(t, plt.Legend.Entries, 2)
	assert.Equal(t, "A Y1", plt.Legend.Entries[0].Text)
	assert.Equal(t, "B Y1", plt.Legend.Entries[1].Text)
	d, _, _ := plt.Plotters[1].Data()
	require.Equal(t, 3, d[Y].Len())
	assert.Equal(t, []float64{9, 16, 25}, []float64{d[Y].Float1D(0), d[Y].Float1D(1), d[Y].Float1D(2)})
	assert.Equal(t, 3.0, d[X].Float1D(0))

	plt.applyStyle()
	p0 := plt.Plotters[0].(*testPlotter)
	p1 := plt.Plotters[1].(*testPlotter)
	assert.NotEqual(t, p0.Style.Line.Color, p1.Style.Line.Color)

	// a second split column is an error, but the first one is used
	dt.AddStringColumn("Name2")
	styleCol(dt, "Name2", func(s *Style) {
		s.On = true
		s.Role = Split
	})
	plt, err = NewTablePlot(dt)
	assert.Error(t, err)
	require.NotNil(t, plt)
	assert.Len(t, plt.Plotters, 2)
}

// TestTablePlotBars tests that multiple bar plots are offset
// from each other so they are side by side.
func TestTablePlotBars(t *testing.T) {
	dt := testTable()
	for _, nm := range []string{"Y1", "Y2"} {
		styleCol(dt, nm, func(s *Style) {
			s.On = true
			s.Role = Y
			s.Plotter = "Bar"
			s.NoLegend = nm == "Y2"
		})
	}
	plt, err := NewTablePlot(dt)
	require.NoError(t, err)
	require.Len(t, plt.Plotters, 2)
	assert.Empty(t, plt.Legend.Entries) // only one has a legend
	plt.applyStyle()
	sz := 1.0 / 2.5
	p0 := plt.Plotters[0].(*testPlotter)
	p1 := plt.Plotters[1].(*testPlotter)
	assert.Equal(t, 0.0, p0.Style.Width.Offset)
	assert.InDelta(t, sz, p1.Style.Width.Offset, 1.0e-12)
	assert.InDelta(t, 0.8*sz, p1.Style.Width.Width, 1.0e-12)
}

// TestTablePlotErrors tests the errors for tables that cannot be plotted.
func TestTablePlotErrors(t *testing.T) {
	plt, err := NewTablePlot(table.New())
	assert.Error(t, err)
	assert.Nil(t, plt)

	// nothing is on
	plt, err = NewTablePlot(testTable())
	assert.NoError(t, err)
	assert.Nil(t, plt)

	dt := testTable()
	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
		s.Plotter = "NoSuchPlotter"
	})
	plt, err = NewTablePlot(dt)
	assert.Error(t, err)
	assert.Nil(t, plt)

	// the plotter can not be made from empty data
	dt = table.New()
	dt.AddColumn("Y", tensor.NewFloat64(0))
	styleCol(dt, "Y", func(s *Style) {
		s.On = true
		s.Role = Y
	})
	plt, err = NewTablePlot(dt)
	assert.Error(t, err)
	require.NotNil(t, plt)
	assert.Empty(t, plt.Plotters)
}

// TestSetPlottersXAxis tests switching the type of plot of all the
// columns that hold numbers, and setting the X axis column.
func TestSetPlottersXAxis(t *testing.T) {
	dt := testTable()
	styleCol(dt, "X", func(s *Style) { s.Role = X })
	styleCol(dt, "Y1", func(s *Style) {
		s.On = true
		s.Role = Y
	})
	styleCol(dt, "Y2", func(s *Style) {
		s.On = true
		s.Role = Y
	})
	SetPlotters(dt, "Bar")
	for _, name := range []string{"X", "X2", "Y1", "Y2"} {
		st := NewStyle()
		sty := GetStylers(dt.Columns.At(name))
		sty.Run(st)
		assert.Equal(t, PlotterName("Bar"), st.Plotter, name)
	}
	assert.Nil(t, GetStylers(dt.Columns.At("Name")), "a string column")

	assert.Error(t, SetXAxis(dt, "Nope"))
	require.NoError(t, SetXAxis(dt, "X2"))
	st := NewStyle()
	sty := GetStylers(dt.Columns.At("X"))
	sty.Run(st)
	assert.Equal(t, Y, st.Role, "the former x axis")
	plt, err := NewTablePlot(dt)
	require.NoError(t, err)
	d, _, _ := plt.Plotters[0].Data()
	assert.Equal(t, 50.0, d[X].Float1D(5))
	plt.applyStyle()
	assert.Equal(t, "X2", plt.X.Label.Text)
}
