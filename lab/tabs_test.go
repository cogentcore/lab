// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lab

import (
	"os"
	"path/filepath"
	"testing"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/text/textcore"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/plotcore"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorcore"
	"cogentcore.org/lab/tensorfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTabsTable returns a table with an X and a Y column.
func newTabsTable() *table.Table {
	dt := table.New("data")
	xs := dt.AddIntColumn("X")
	ys := dt.AddFloat64Column("Y")
	dt.SetNumRows(4)
	for i := range 4 {
		xs.SetInt1D(i, i)
		ys.SetFloat1D(float64(i*i), i)
	}
	return dt
}

// TestTabsEditors tests the tabs of the editors of tables, tensors, and
// slices, which are recycled by their labels, and getting them by label
// and type.
func TestTabsEditors(t *testing.T) {
	b := core.NewBody()
	ts := NewTabs(b)
	dt := newTabsTable()
	tsr := tensor.NewFloat64(3, 3)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Same(t, ts, ts.AsLab())
		assert.Equal(t, core.FunctionalTabs, ts.Type)

		tv := ts.TensorTable("table", dt)
		assert.Equal(t, 4, tv.SliceSize)
		assert.Same(t, tv, ts.TensorTable("table", dt), "recycled")
		assert.Same(t, tv, TabAt[*tensorcore.Table](ts, "table"))

		te := ts.TensorEditor("tensor", tsr)
		assert.Same(t, tsr, te.Tensor)
		assert.Same(t, te, ts.TensorEditor("tensor", tsr))

		tg := ts.TensorGrid("grid", tsr)
		assert.Same(t, tsr, tg.Tensor)

		type city struct {
			Name string
			Area float32
		}
		cities := []city{{"Davis", 1}}
		st := ts.SliceTable("slice", &cities) // (a core.Table needs a pointer)
		assert.Equal(t, 1, st.SliceSize)

		// the wrong type, and a missing tab
		assert.Nil(t, TabAt[*tensorcore.TensorGrid](ts, "table"))
		assert.Nil(t, TabAt[*tensorcore.Table](ts, "none"))
		assert.Nil(t, NewTab(ts, "table", func(tab *core.Frame) *tensorcore.TensorGrid {
			t.Error("not made: the tab has another type")
			return nil
		}))

		assert.Nil(t, ts.TabUpdateRender("none"))
		assert.NotNil(t, ts.TabUpdateRender("table"))
	}))
}

// TestTabsText tests the tabs of text editors of strings and files.
func TestTabsText(t *testing.T) {
	b := core.NewBody()
	ts := NewTabs(b)
	fn := filepath.Join(t.TempDir(), "notes.txt")
	require.NoError(t, os.WriteFile(fn, []byte("from a file\n"), 0o644))
	core.RenderImage(b, core.Locked(b, func() {
		ed := ts.EditorString("text", "hello")
		assert.Equal(t, "hello\n", ed.Lines.String())
		assert.Same(t, ed, ts.EditorString("text", ""), "recycled, keeping its text")
		assert.Equal(t, "hello\n", ed.Lines.String())

		fe := ts.EditorFile("file", fn)
		assert.Equal(t, "from a file\n", fe.Lines.String())
		assert.Same(t, fe, TabAt[*textcore.Editor](ts, "file"))
	}))
}

// TestTabsPlots tests the tabs of plots: of tables, tensorfs data and
// directories, and plots, and updating them by label.
func TestTabsPlots(t *testing.T) {
	b := core.NewBody()
	ts := NewTabs(b)
	dt := newTabsTable()
	plot.SetStyler(dt.Column("Y"), func(s *plot.Style) { s.On = true })
	root, _ := tensorfs.NewDir("root")
	vals := tensorfs.Value[float64](root.Dir("Run"), "Loss", 5)
	for i := range 5 {
		vals.SetFloat1D(float64(5-i), i)
	}
	tensorfs.Value[float64](root.Dir("Run"), "Acc", 5)
	core.RenderImage(b, core.Locked(b, func() {
		pl := ts.PlotTable("plot", dt)
		require.NotNil(t, pl)
		assert.Same(t, pl, ts.PlotTable("plot", dt))
		assert.Same(t, pl, ts.UpdatePlot("plot"))
		assert.Nil(t, ts.UpdatePlot("none"))
		assert.Nil(t, ts.GoUpdatePlot("none"))

		np := ts.NewPlot("new")
		assert.NotNil(t, np)
		assert.Same(t, np, ts.NewPlot("new"))

		run := root.Dir("Run")
		dp := ts.PlotTensorFS(run)
		require.NotNil(t, dp)
		assert.NotNil(t, ts.TabByName("root Run Plot"))

		vp := ts.PlotTensorFS(run.Node("Loss"))
		require.NotNil(t, vp)
		pt := ts.TabByName("Run Loss Plot")
		require.NotNil(t, pt)

		assert.Nil(t, ts.GridTensorFS(run), "not a directory")
		assert.NotNil(t, ts.GridTensorFS(run.Node("Loss")))
		assert.NotNil(t, ts.TabByName("Run Loss Grid"))

		plt := plot.New()
		plots.NewLine(plt, tensor.NewFloat64FromValues(1, 2, 3))
		pw := ts.Plot("raw", plt)
		require.NotNil(t, pw)
		assert.Same(t, plt, pw.Plot)
		assert.Same(t, pw, ts.Plot("raw", plt))
	}))
}

// TestTabsPlotRows tests the plot of a tensorfs value that is a view of
// some of the rows of a tensor, which plots those rows.
func TestTabsPlotRows(t *testing.T) {
	b := core.NewBody()
	ts := NewTabs(b)
	root, _ := tensorfs.NewDir("root")
	vals := tensor.NewFloat64FromValues(1, 2, 3, 4)
	rows := tensor.NewRows(vals, 3, 1)
	nd := tensorfs.SetTensor(root, rows, "Some")
	core.RenderImage(b, core.Locked(b, func() {
		pl := ts.PlotTensorFS(nd)
		require.NotNil(t, pl)
		tv := TabAt[*plotcore.Editor](ts, "root Some Plot")
		assert.Same(t, pl, tv)
	}))
}

// TestPlotHelpers tests the helpers that make plots in plot widgets.
func TestPlotHelpers(t *testing.T) {
	b := core.NewBody()
	plt := NewPlot(b)
	require.NotNil(t, plt)
	plots.NewLine(plt, tensor.NewFloat64FromValues(1, 2, 3))
	p2, pw := NewPlotWidget(b)
	assert.Same(t, p2, pw.Plot)
	cp := NewPlotFrom(plt, b)
	assert.NotSame(t, plt, cp)
	assert.Len(t, cp.Plotters, 1)
	core.RenderImage(b)
}

// TestPrompts tests the dialogs that prompt the user, which call the
// function when the user clicks OK, and not when Cancel.
func TestPrompts(t *testing.T) {
	b := core.NewBody()
	ok := 0
	str := ""
	type opts struct{ N int }
	o := &opts{N: 1}
	click := func(text string) {
		bt := core.ButtonByText(core.TopScene(b), text)
		require.NotNil(t, bt, text)
		bt.Send(events.Click)
	}
	core.RenderImage(b, core.Locked(b, func() {
		PromptOKCancel(b, "Do it?", func() { ok++ })
	}), core.Locked(b, func() {
		click("OK")
	}), core.Locked(b, func() {
		assert.Equal(t, 1, ok)
		PromptOKCancel(b, "Do it again?", func() { ok++ })
	}), core.Locked(b, func() {
		click("Cancel")
	}), core.Locked(b, func() {
		assert.Equal(t, 1, ok)
		PromptString(b, "initial", "Name?", func(s string) { str = s })
	}), core.Locked(b, func() {
		tf := findText(core.TopScene(b))
		require.NotNil(t, tf)
		assert.Equal(t, "initial", tf.Text())
		tf.SetText("changed")
		click("OK")
	}), core.Locked(b, func() {
		assert.Equal(t, "changed", str)
		PromptStruct(b, o, "Options", func() { ok++ })
	}), core.Locked(b, func() {
		click("OK")
	}), core.Locked(b, func() {
		assert.Equal(t, 2, ok)
		PromptOKCancel(b, "nil", nil)
		click("OK")
		PromptString(b, "", "nil", nil)
		click("OK")
		PromptStruct(b, o, "nil", nil)
		click("OK")
	}))
}

// findText returns the first text field in the given widget.
func findText(w core.Widget) *core.TextField {
	for cw := range w.AsWidget().AllWidgets() {
		if tf, ok := cw.(*core.TextField); ok {
			return tf
		}
	}
	return nil
}

// TestScriptHelpers tests the helpers for the labels and tooltips of
// scripts, and the names of table files.
func TestScriptHelpers(t *testing.T) {
	assert.Equal(t, "Plot", TrimOrderPrefix("01-Plot"))
	assert.Equal(t, "Plot", TrimOrderPrefix("Plot"))
	assert.Equal(t, "my-Plot", TrimOrderPrefix("my-Plot"))
	assert.Equal(t, "a b ", FirstComment("// a\n// b\ncode()\n// c"))
	assert.Equal(t, "", FirstComment("code()"))
	assert.True(t, IsTableFile("a.tsv"))
	assert.True(t, IsTableFile("a.csv"))
	assert.False(t, IsTableFile("a.txt"))
	assert.Equal(t, "dir file", DirAndFileNoSlash(filepath.Join("root", "dir", "file")))
}

// TestBrowserScripts tests the scripts of the browser: the init scripts,
// with lower case names, which are run once, and the buttons of the
// others in the toolbar, ordered by name, with the first comments as
// their tooltips, which run them.
func TestBrowserScripts(t *testing.T) {
	dir := t.TempDir()
	write := func(name, code string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(code), 0o644))
	}
	write("init.goal", "x := 1\n")
	write("2-Second.goal", "// Does the second thing.\nfmt.Println(2)\n")
	write("1-First.goal", "fmt.Println(1)\n")

	oldRun, oldCode := RunScript, RunScriptCode
	defer func() { RunScript, RunScriptCode = oldRun, oldCode }()
	var ran, codes []string
	RunScript = func(br *Browser, name string) error { ran = append(ran, name); return nil }
	RunScriptCode = func(br *Browser, code string) error { codes = append(codes, code); return nil }

	b := core.NewBody()
	br := &Browser{ScriptsDir: dir}
	b.AddTopBar(func(bar *core.Frame) {
		br.Toolbar = core.NewToolbar(bar)
		br.Toolbar.Maker(br.MakeToolbar)
	})
	core.RenderImage(b, core.Locked(b, func() {
		br.UpdateScripts()
		assert.Equal(t, []string{"x := 1\n"}, codes, "the init script")
		assert.Len(t, br.Scripts, 2)
		br.UpdateScripts()
		assert.Len(t, codes, 1, "only run the first time")

		first := core.ButtonByText(br.Toolbar, "First")
		second := core.ButtonByText(br.Toolbar, "Second")
		require.NotNil(t, first)
		require.NotNil(t, second)
		assert.Less(t, first.IndexInParent(), second.IndexInParent())
		assert.Equal(t, "Does the second thing. ", second.Tooltip)
		assert.Contains(t, first.Tooltip, "Run Script")
		second.Send(events.Click)
		assert.Equal(t, []string{"2-Second"}, ran)

		br.UpdateFiles() // no files tree
	}))
}
