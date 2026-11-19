// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotcore

import (
	"cogentcore.org/core/colors"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/tree"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/stats/cluster"
	"cogentcore.org/lab/stats/metric"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorcore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newPlotTable returns a table with an int X column, two float Y columns,
// and a string column, with the given name.
func newPlotTable(name string) *table.Table {
	dt := table.New(name)
	xs := dt.AddIntColumn("X")
	ys := dt.AddFloat64Column("Y1")
	zs := dt.AddFloat64Column("Y2")
	ls := dt.AddStringColumn("Label")
	dt.SetNumRows(5)
	for i := range 5 {
		xs.SetInt1D(i, i)
		ys.SetFloat1D(float64(i*i), i)
		zs.SetFloat1D(float64(10-i), i)
		ls.SetString1D(string(rune('a'+i)), i)
	}
	return dt
}

// columnStyle returns the style of the given column of the table of the
// given editor, from its stylers.
func columnStyle(pl *Editor, name string) *plot.Style {
	st := &plot.Style{}
	st.Defaults()
	sty := plot.GetStylers(pl.table.Column(name).Tensor)
	sty.Run(st)
	return st
}

// columnSwitch returns the switch of the column with the given name.
func columnSwitch(pl *Editor, name string) *core.Switch {
	return pl.columnsFrame.ChildByName(name).AsTree().Child(0).(*core.Switch)
}

// windowScene waits for a window with the given title to be open,
// and returns its main scene.
func windowScene(t *testing.T, title string) *core.Scene {
	t.Helper()
	var sc *core.Scene
	assert.Eventually(t, func() bool {
		if w := core.AllRenderWindows.FindName(title); w != nil {
			sc = w.MainScene()
		}
		return sc != nil
	}, 5*time.Second, 10*time.Millisecond, "the window %q", title)
	return sc
}

// findWidget returns the first widget of the given type in the given
// widget and its descendants.
func findWidget[T core.Widget](w core.Widget) T {
	var zero T
	for cw := range w.AsWidget().AllWidgets() {
		if x, ok := cw.(T); ok {
			return x
		}
	}
	return zero
}

// TestEditorColumns tests the default styles of the columns, by their
// types, the title from the name of the table, and turning the columns
// on and off with their switches, the Clear button, and by name.
func TestEditorColumns(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	pl.SetTable(newPlotTable("Squares"))
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, "Squares", pl.PlotStyle.Title)
		assert.Equal(t, plot.Off, pl.PlotStyle.PointsOn)
		x := columnStyle(pl, "X")
		assert.Equal(t, plot.X, x.Role)
		y1 := columnStyle(pl, "Y1")
		assert.Equal(t, plot.Y, y1.Role)
		assert.Equal(t, plot.PlotterName(""), y1.Plotter, "numbers use the default, the PlotStyle Plotter")
		assert.False(t, y1.On)
		lb := columnStyle(pl, "Label")
		assert.Equal(t, plot.Label, lb.Role)
		assert.Equal(t, plot.PlotterName(plots.LabelsType), lb.Plotter)
		assert.True(t, columnSwitch(pl, "X").StateIs(states.Disabled), "the x axis")

		for _, name := range []string{"Y1", "Y2"} {
			sw := columnSwitch(pl, name)
			sw.SetChecked(true)
			sw.SendChange()
		}
		assert.True(t, columnStyle(pl, "Y1").On)
		assert.True(t, columnStyle(pl, "Y2").On)
		require.NotNil(t, pl.plot)
		assert.Equal(t, 2, uniquePlotters(pl.plot))

		core.ButtonByText(pl, "Clear").Send(events.Click)
		assert.False(t, columnStyle(pl, "Y1").On, "the first column")
		assert.False(t, columnStyle(pl, "Y2").On)

		pl.setColumnsByName("2", true)
		assert.False(t, columnStyle(pl, "Y1").On)
		assert.True(t, columnStyle(pl, "Y2").On)
		assert.Equal(t, 1, uniquePlotters(pl.plot))
	}))
}

// uniquePlotters returns the number of different plotters of the given
// plot. (NewTablePlot adds each plotter twice: the New function of its
// type adds it, and then NewTablePlot does.)
func uniquePlotters(plt *plot.Plot) int {
	seen := map[plot.Plotter]bool{}
	for _, p := range plt.Plotters {
		seen[p] = true
	}
	return len(seen)
}

// TestEditorClearFirst tests that the Clear button turns off the first
// column, when it is not the X axis.
func TestEditorClearFirst(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	dt := table.New()
	dt.AddFloat64Column("A").SetNumRows(3)
	dt.AddFloat64Column("B").SetNumRows(3)
	dt.SetNumRows(3)
	plot.SetStyler(dt.Column("A"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		assert.True(t, columnStyle(pl, "A").On)
		core.ButtonByText(pl, "Clear").Send(events.Click)
		assert.False(t, columnStyle(pl, "A").On)
		assert.False(t, columnSwitch(pl, "A").IsChecked())
	}))
}

// TestEditorStylers tests the styles of the columns from their stylers:
// a bar plot, which is filled with the color of the column, and a split
// column, which keeps the default colors of the others.
func TestEditorStylers(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	dt := newPlotTable("")
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) {
		s.On = true
		s.Plotter = plots.BarType
	})
	plot.SetStyler(dt.Column("Label"), func(s *plot.Style) {
		s.On = true
		s.Role = plot.Split
	})
	pl.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, "", pl.PlotStyle.Title)
		y1 := columnStyle(pl, "Y1")
		assert.Equal(t, plot.PlotterName(plots.BarType), y1.Plotter)
		assert.Equal(t, plot.Split, columnStyle(pl, "Label").Role)
		sw := columnSwitch(pl, "Label")
		sw.SetChecked(false)
		sw.SendChange()
		sw.SetChecked(true)
		sw.SendChange()
		pl.Update()
	}))
}

// TestEditorToolbar tests the toolbar of the editor: toggling pan and zoom,
// updating, filtering, and the style, table, and column dialogs.
func TestEditorToolbar(t *testing.T) {
	defer core.TrackWindows()()
	b := core.NewBody()
	pl := NewEditor(b)
	pl.SetName("squares")
	dt := newPlotTable("Squares")
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	var tb *core.Toolbar
	b.AddTopBar(func(bar *core.Frame) {
		tb = core.NewToolbar(bar)
		tb.Maker(pl.MakeToolbar)
	})
	core.RenderImage(b, core.Locked(b, func() {
		var pan *core.Button
		for _, c := range tb.Children {
			if bt, ok := c.(*core.Button); ok && bt.Icon == icons.PanTool {
				pan = bt
			}
		}
		require.NotNil(t, pan)
		assert.False(t, pl.plotWidget.IsReadOnly())
		pan.Send(events.Click)
		assert.True(t, pl.plotWidget.IsReadOnly())
		pan.Send(events.Click)
		assert.False(t, pl.plotWidget.IsReadOnly())

		core.ButtonByText(tb, "Update").Send(events.Click)
		require.NoError(t, pl.table.FilterString("Label", "c", tensor.StringMatch{}))
		pl.UpdatePlot()
		assert.Equal(t, 1, pl.table.NumRows())
		core.ButtonByText(tb, "Unfilter").Send(events.Click)
		assert.Equal(t, 5, pl.table.NumRows())

		core.ButtonByText(tb, "Style").Send(events.Click)
		core.ButtonByText(tb, "Table").Send(events.Click)
		core.ButtonByText(pl.columnsFrame, "Y1").Send(events.Click)
	}), func() {
		sc := windowScene(t, "Plot style")
		require.NotNil(t, sc)
		core.Locked(sc, func() { findWidget[*core.Form](sc).SendChange() })()
		sc = windowScene(t, "squares Data")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			tv := findWidget[*tensorcore.Table](sc)
			require.NotNil(t, tv)
			assert.Equal(t, 5, tv.SliceSize)
		})()
		sc = windowScene(t, "Y1 style properties")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			fm := findWidget[*core.Form](sc)
			fm.Struct.(*plot.Style).Line.Width.Dp(5)
			fm.Modified["Line • Width"] = true // as the form does
			fm.SendChange()
		})()
	}, core.Locked(b, func() {
		assert.Equal(t, float32(5), columnStyle(pl, "Y1").Line.Width.Value)
	}))
}

// TestEditorSave tests saving the plot as SVG, PDF and PNG files, and its
// data as a TSV file, opening it again, and opening it from a file system.
func TestEditorSave(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	dt := newPlotTable("Squares")
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	var tb *core.Toolbar
	b.AddTopBar(func(bar *core.Frame) {
		tb = core.NewToolbar(bar)
		tb.Maker(pl.MakeToolbar)
	})
	dir := t.TempDir()
	core.RenderImage(b, core.Locked(b, func() {
		save := core.ButtonByText(tb, "Save")
		require.NotNil(t, save)
		m := core.NewScene()
		save.Menu(m, image.Point{})
		assert.NotNil(t, core.ButtonByText(m, "Save all"))

		pl.SaveAll(core.Filename(filepath.Join(dir, "plot.ext")))
		for _, ext := range []string{".tsv", ".png", ".svg", ".pdf"} {
			fi, err := os.Stat(filepath.Join(dir, "plot"+ext))
			if assert.NoError(t, err, ext) {
				assert.Greater(t, fi.Size(), int64(0), ext)
			}
		}
		assert.Equal(t, core.Filename(filepath.Join(dir, "plot.svg")), pl.svgFile)
		assert.Equal(t, core.Filename(filepath.Join(dir, "plot.tsv")), pl.dataFile)
		svg, err := os.ReadFile(filepath.Join(dir, "plot.svg"))
		require.NoError(t, err)
		assert.True(t, strings.Contains(string(svg), "<svg"))

		pl.OpenCSV(core.Filename(filepath.Join(dir, "plot.tsv")), tensor.Tab)
		assert.Equal(t, 5, pl.table.NumRows())
		assert.Equal(t, []string{"X", "Y1", "Y2", "Label"}, pl.table.Columns.Keys)

		data, err := os.ReadFile(filepath.Join(dir, "plot.tsv"))
		require.NoError(t, err)
		pl.OpenFS(fstest.MapFS{"d.tsv": {Data: data}}, "d.tsv", tensor.Tab)
		assert.Equal(t, 5, pl.table.NumRows())
	}))
}

// TestEditorGoUpdatePlot tests updating the plot from another goroutine,
// after rows are added to the table.
func TestEditorGoUpdatePlot(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	dt := newPlotTable("Squares")
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	core.RenderImage(b, func() {
		b.AsyncLock()
		pl.table.SetNumRows(8)
		b.AsyncUnlock()
		pl.GoUpdatePlot()
		b.AsyncLock()
		defer b.AsyncUnlock()
		assert.Equal(t, 8, pl.table.NumRows())
	})
	var nilEditor *Editor
	assert.NotPanics(t, func() {
		nilEditor.GoUpdatePlot()
		nilEditor.UpdatePlot()
	})
}

// TestSubPlot tests a sub plot with its own toolbar, and copying an
// editor, which shares the table.
func TestSubPlot(t *testing.T) {
	b := core.NewBody()
	pl := NewSubPlot(b)
	dt := newPlotTable("Squares")
	pl.SetTable(dt)
	other := NewEditor(b)
	core.RenderImage(b, core.Locked(b, func() {
		tb := pl.Parent.AsTree().Child(0).(*core.Toolbar)
		assert.NotNil(t, core.ButtonByText(tb, "Update"))
		other.CopyFieldsFrom(pl)
		assert.Same(t, pl.table, other.table)
		assert.Equal(t, pl.PlotStyle.Title, other.PlotStyle.Title)
	}))
}

// TestClusterPlot tests the cluster plot of the patterns of a table.
func TestClusterPlot(t *testing.T) {
	dt := table.New()
	names := dt.AddStringColumn("Name")
	in := dt.AddFloat64Column("Input", 4)
	dt.SetNumRows(4)
	for i := range 4 {
		names.SetString1D(string(rune('a'+i)), i)
		for j := range 4 {
			v := 0.0
			if j <= i {
				v = 1
			}
			in.SetFloat(v, i, j)
		}
	}
	b := core.NewBody()
	pl := NewEditor(b)
	ClusterPlot(pl, dt, metric.MetricL2Norm, cluster.Min, "Input", "Name")
	core.RenderImage(b, core.Locked(b, func() {
		require.NotNil(t, pl.table)
		assert.Greater(t, pl.table.NumRows(), 4)
	}))
}

// TestPlotterChooser tests the chooser of the plotter of a style, which
// has the names of all of the plotters.
func TestPlotterChooser(t *testing.T) {
	b := core.NewBody()
	name := plot.PlotterName(plots.BarType)
	w := core.NewValue(&name, "", b)
	pc, ok := w.(*PlotterChooser)
	require.True(t, ok)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Len(t, pc.Items, len(plot.Plotters)+1, "the default and all of the plotters")
		assert.Equal(t, plot.PlotterName(""), pc.Items[0].Value)
		assert.Equal(t, "Default", pc.Items[0].Text)
		assert.Equal(t, plot.PlotterName(plots.BarType), pc.CurrentItem.Value)
		assert.Equal(t, plot.PlotterName(plots.BarType), pc.Items[pc.CurrentIndex].Value)
		pc.SetCurrentIndex(len(pc.Items) - 1)
		pc.SendChange()
		assert.Equal(t, plot.PlotterName(plots.YErrorBarsType), name)
	}))
}

// TestColorMapName tests that the color map of the color axis style
// is edited with a color map button.
func TestColorMapName(t *testing.T) {
	b := core.NewBody()
	name := plot.ColorMapName("ColdHot")
	w := core.NewValue(&name, "", b)
	cb, ok := w.(*core.ColorMapButton)
	require.True(t, ok)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, "ColdHot", cb.MapName)
		cb.MapName = "Viridis"
		cb.SendChange()
		assert.Equal(t, plot.ColorMapName("Viridis"), name)
	}))
}

// TestPlotPanZoom tests panning the plot with the mouse, and zooming it
// with the scroll wheel, with the modifier keys that keep one axis.
func TestPlotPanZoom(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	dt := newPlotTable("Squares")
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	var pw *Plot
	ranges := 0
	core.RenderImage(b, core.Locked(b, func() {
		pw = pl.plotWidget
		pw.SetRangesFunc = func() { ranges++ }
		pz := &pw.Plot.PanZoom
		scroll := func(mods ...key.Modifiers) {
			var km key.Modifiers
			for _, m := range mods {
				km.SetFlag(true, m)
			}
			se := events.NewScroll(image.Pt(10, 10), math32.Vec2(0, 100), km)
			pw.HandleEvent(se)
		}
		scroll()
		assert.InDelta(t, 1.2, pz.XScale, 1e-6)
		assert.InDelta(t, 1.2, pz.YScale, 1e-6)
		scroll(key.Shift)
		assert.InDelta(t, 1.44, pz.XScale, 1e-6)
		assert.InDelta(t, 1.2, pz.YScale, 1e-6)
		scroll(key.Alt)
		assert.InDelta(t, 1.44, pz.XScale, 1e-6)
		assert.InDelta(t, 1.44, pz.YScale, 1e-6)
	}), func() {
		b.AsyncLock()
		bb := pw.Geom.ContentBBox
		b.AsyncUnlock()
		mid := bb.Min.Add(bb.Size().Div(2))
		core.DragMouse(b, mid, mid.Add(image.Pt(50, 30)))
	}, core.Locked(b, func() {
		pz := &pw.Plot.PanZoom
		assert.Less(t, pz.XOffset, 0.0, "dragged right")
		assert.Greater(t, pz.YOffset, 0.0, "dragged down")
		assert.Greater(t, ranges, 0)
	}))
}

// TestPlotTooltip tests the tooltip of the plot at a data point, which
// highlights it, and away from the data.
func TestPlotTooltip(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	dt := newPlotTable("Squares")
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		pw := pl.plotWidget
		pw.Tooltip = "plot"
		pw.renderPlot() // the pixels of the data
		require.NotEmpty(t, pw.Plot.Plotters)
		_, xs, ys := pw.Plot.Plotters[0].Data()
		require.Len(t, xs, 5)
		tt, pos := pw.WidgetTooltip(image.Pt(int(xs[2]), int(ys[2])))
		assert.Equal(t, "[2]: (2, 4)", tt, "no legend for one plotter")
		assert.Equal(t, image.Pt(int(xs[2]), int(ys[2])), pos)
		assert.NotNil(t, pw.Plot.HighlightPlotter)

		tt, _ = pw.WidgetTooltip(image.Pt(-1000, -1000))
		assert.Equal(t, "plot", tt)
		assert.Nil(t, pw.Plot.HighlightPlotter)
		tt, _ = pw.WidgetTooltip(image.Pt(-1, -1))
		assert.Equal(t, "_", tt)

		empty := NewPlot(b)
		empty.Tooltip = "empty"
		tt, _ = empty.WidgetTooltip(image.Pt(5, 5))
		assert.Equal(t, "empty", tt)
	}))
}

// TestEditorTypeXAxis tests setting the type of plot of all the columns,
// and the Set X button of the style dialog of a column.
func TestEditorTypeXAxis(t *testing.T) {
	defer core.TrackWindows()()
	b := core.NewBody()
	pl := NewEditor(b)
	dt := newPlotTable("Squares")
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) { s.On = true })
	plot.SetStyler(dt.Column("Y2"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	var tb *core.Toolbar
	b.AddTopBar(func(bar *core.Frame) {
		tb = core.NewToolbar(bar)
		tb.Maker(pl.MakeToolbar)
	})
	setX := func(sc *core.Scene) *core.Button {
		return core.ButtonByText(sc, "Set X")
	}
	core.AssertRender(t, b, "type-xaxis", core.Locked(b, func() {
		require.NotNil(t, core.ButtonByText(tb, "Type"))
		// a column set to another plotter also gets the type
		pl.columns["Y2"].style.Plotter = plots.XErrorBarsType
		pl.columns["Y2"].mods["Plotter"] = true
		pl.SetPlotType(plots.BarType)
		assert.Equal(t, plot.PlotterName(plots.BarType), pl.PlotStyle.Plotter)
		for _, name := range []string{"Y1", "Y2"} {
			assert.Equal(t, plot.PlotterName(""), columnStyle(pl, name).Plotter, name)
		}
		assert.Equal(t, plot.PlotterName(plots.LabelsType), columnStyle(pl, "Label").Plotter)
		require.NotNil(t, pl.plot)
		require.Len(t, pl.plot.Plotters, 2)
		for _, pt := range pl.plot.Plotters {
			_, ok := pt.(*plots.Bar)
			assert.True(t, ok, "a bar plot")
		}
		core.ButtonByText(pl.columnsFrame, "X").Send(events.Click)
	}), func() {
		core.Locked(b, func() { // filled with the color of the column
			for i, pt := range pl.plot.Plotters {
				bar := pt.(*plots.Bar)
				bar.ApplyStyle(&pl.plot.Style, i)
				assert.NotEqual(t, colors.Scheme.OnSurface, bar.Style.Line.Color, "the color of the column")
				assert.Equal(t, bar.Style.Line.Color, bar.Style.Line.Fill)
			}
		})()
		sc := windowScene(t, "X style properties")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			assert.True(t, setX(sc).IsDisabled(), "the x axis")
		})()
		core.Locked(b, func() {
			core.ButtonByText(pl.columnsFrame, "Y2").Send(events.Click)
		})()
	}, func() {
		sc := windowScene(t, "Y2 style properties")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			bt := setX(sc)
			assert.False(t, bt.IsDisabled())
			bt.Send(events.Click)
			assert.True(t, bt.IsDisabled())
			assert.Equal(t, plot.X, findWidget[*core.Form](sc).Struct.(*plot.Style).Role)
		})()
	}, core.Locked(b, func() {
		assert.Equal(t, "Y2", pl.PlotStyle.XAxis.Column)
		y2 := columnStyle(pl, "Y2")
		assert.Equal(t, plot.X, y2.Role)
		assert.Equal(t, "Y2", y2.Plot.XAxis.Column)
		assert.Equal(t, plot.Y, columnStyle(pl, "X").Role, "the former x axis")
		assert.True(t, columnSwitch(pl, "Y2").StateIs(states.Disabled))
		assert.False(t, columnSwitch(pl, "X").StateIs(states.Disabled))
		assert.Equal(t, 1, uniquePlotters(pl.plot))

		menuSetX := func(name string) *core.Button { return menuSetX(pl, name) }
		assert.True(t, menuSetX("Y2").IsDisabled(), "the x axis")
		mx := menuSetX("X")
		require.False(t, mx.IsDisabled())
		mx.Send(events.Click)
		assert.Equal(t, "X", pl.PlotStyle.XAxis.Column)
		assert.Equal(t, plot.X, columnStyle(pl, "X").Role)
		y2 = columnStyle(pl, "Y2")
		assert.Equal(t, plot.Y, y2.Role)
		assert.False(t, y2.On, "as its switch shows")

		pl.SetPlotType(plots.XYType)
		assert.Equal(t, plot.PlotterName(plots.XYType), pl.PlotStyle.Plotter)
		for _, pt := range pl.plot.Plotters {
			_, ok := pt.(*plots.XY)
			assert.True(t, ok, "back to an XY plot")
		}
	}))
}

// menuSetX returns the Set X button of the context menu of the name
// of the given column.
func menuSetX(pl *Editor, name string) *core.Button {
	bt := core.ButtonByText(pl.columnsFrame, name)
	m := core.NewScene()
	for _, cm := range bt.ContextMenus {
		cm(m, image.Point{})
	}
	return core.ButtonByText(m, "Set X")
}

// TestEditorSetXIntColumns tests Set X with more than one int column:
// only the last one is the X axis by default, with the X role.
func TestEditorSetXIntColumns(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	dt := newPlotTable("Squares")
	rn := dt.AddIntColumn("Run")
	for i := range 5 {
		rn.SetInt1D(10*i, i)
	}
	plot.SetStyler(dt.Column("Y1"), func(s *plot.Style) { s.On = true })
	pl.SetTable(dt)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, "Run", pl.xAxisColumn())
		assert.Equal(t, plot.X, columnStyle(pl, "Run").Role)
		assert.Equal(t, plot.Y, columnStyle(pl, "X").Role, "only the x axis has the X role")
		assert.False(t, columnSwitch(pl, "X").StateIs(states.Disabled))
		assert.True(t, columnSwitch(pl, "Run").StateIs(states.Disabled))
		assert.True(t, menuSetX(pl, "Run").IsDisabled(), "the x axis")
		mx := menuSetX(pl, "X")
		require.False(t, mx.IsDisabled(), "another int column")
		mx.Send(events.Click)
		assert.Equal(t, "X", pl.xAxisColumn())
		assert.Equal(t, plot.Y, columnStyle(pl, "Run").Role)
		d, _, _ := pl.plot.Plotters[0].Data()
		assert.Equal(t, 4.0, d[plot.X].Float1D(4))
	}))
}

// TestEditorPlotStylePlotter tests that the Plotter of the plot style of
// the editor (as set in its Style dialog) changes the type of plot of the
// columns of numbers, whose bars are filled with the color of the column.
func TestEditorPlotStylePlotter(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	pl.SetTable(newPlotTable("Bars"))
	core.RenderImage(b, core.Locked(b, func() {
		// only Y2 is on: its bar is the first plotter, but its color is
		// the second, so that the default fill of the bar would differ
		pl.columns["Y2"].style.On = true
		pl.columns["Y2"].mods["On"] = true
		pl.PlotStyle.Plotter = plots.BarType
		pl.plotStyleModified["Plotter"] = true
		pl.UpdatePlot()
		require.Len(t, pl.plot.Plotters, 1)
		bar, ok := pl.plot.Plotters[0].(*plots.Bar)
		require.True(t, ok, "a bar plot")
		bar.ApplyStyle(&pl.plot.Style, 0)
		assert.Equal(t, colors.Uniform(colors.Spaced(1)), bar.Style.Line.Color)
		assert.Equal(t, bar.Style.Line.Color, bar.Style.Line.Fill)
	}))
}

// TestEditorTypeMenu tests the menu of the type of plot that the Type
// button prompts for: it starts with the current type, and has the
// default and the plotters, with no blank item for an unset value.
func TestEditorTypeMenu(t *testing.T) {
	b := core.NewBody()
	pl := NewEditor(b)
	pl.SetTable(newPlotTable("Type menu"))
	var tb *core.Toolbar
	b.AddTopBar(func(bar *core.Frame) {
		tb = core.NewToolbar(bar)
		tb.Maker(pl.MakeToolbar)
	})
	var pc *PlotterChooser
	core.AssertRender(t, b, "type-menu", core.Locked(b, func() {
		core.ButtonByText(tb, "Type").Send(events.Click)
	}), func() {
		core.Locked(b, func() {
			pc = findWidget[*PlotterChooser](core.TopScene(b))
			require.NotNil(t, pc)
			assert.Len(t, pc.Items, len(plot.Plotters)+1)
			assert.Equal(t, plot.PlotterName(plots.XYType), pc.CurrentItem.Value, "the current type")
			pc.Send(events.Click)
		})()
	}, func() {
		core.Locked(b, func() {
			menu := core.TopScene(b)
			require.NotEqual(t, pc.Scene, menu, "the menu")
			var texts []string
			menu.WidgetWalkDown(func(cw core.Widget, cwb *core.WidgetBase) bool {
				if bt, ok := cw.(*core.Button); ok {
					texts = append(texts, bt.Text)
				}
				return tree.Continue
			})
			assert.Equal(t, []string{"Default", "Band", "Bar", "Box", "HLine", "HSpan", "Labels", "VLine", "VSpan", "XBand", "XBox", "XErrorBars", "XY", "YErrorBars"}, texts)
		})()
	})
}

// TestEditorColumnsHeader tests that the buttons for all the columns stay
// in place above the columns, which scroll.
func TestEditorColumnsHeader(t *testing.T) {
	dt := table.New("Many")
	for i := range 40 {
		dt.AddFloat64Column(fmt.Sprintf("Col%02d", i))
	}
	dt.SetNumRows(3)
	b := core.NewBody()
	pl := NewEditor(b)
	pl.Styler(func(s *styles.Style) {
		s.Min.X.Dp(600)
		s.Max.Y.Dp(400)
	})
	pl.SetTable(dt)
	var clearY, colY int
	core.AssertRender(t, b, "columns-header", core.Locked(b, func() {
		clear := core.ButtonByText(pl, "Clear")
		require.NotNil(t, clear)
		assert.NotEqual(t, pl.columnsFrame, clear.Parent, "not in the scrolling columns")
		clearY = clear.Geom.TotalBBox.Min.Y
		colY = core.ButtonByText(pl.columnsFrame, "Col00").Geom.TotalBBox.Min.Y
		assert.True(t, pl.columnsFrame.HasScroll[math32.Y], "the columns scroll")
		pl.columnsFrame.ScrollDimToContentEnd(math32.Y)
	}), core.Locked(b, func() {
		assert.Equal(t, clearY, core.ButtonByText(pl, "Clear").Geom.TotalBBox.Min.Y, "in place")
		assert.Less(t, core.ButtonByText(pl.columnsFrame, "Col00").Geom.TotalBBox.Min.Y, colY, "scrolled")
	}))
}
