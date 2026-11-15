// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorcore

import (
	"image"
	"path/filepath"
	"testing"

	"cogentcore.org/core/base/fileinfo"
	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestTensor returns a 3x2 tensor whose values are 10*row + column.
func newTestTensor() *tensor.Float64 {
	tsr := tensor.NewFloat64(3, 2)
	for r := range 3 {
		for c := range 2 {
			tsr.Set(float64(10*r+c), r, c)
		}
	}
	return tsr
}

// TestTensorEditorEdit tests the rows of a 2D tensor, with the Y=0 row at
// the bottom, editing a value, and the headers of the columns.
func TestTensorEditorEdit(t *testing.T) {
	b := core.NewBody()
	tsr := newTestTensor()
	te := NewTensorEditor(b)
	te.SetTensor(tsr)
	changes := 0
	te.OnChange(func(e events.Event) { changes++ })
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, 3, te.SliceSize)
		assert.Equal(t, 2, te.NCols)
		assert.Equal(t, "001", te.ColumnHeader(1))
		assert.False(t, te.Layout.TopZero)
		sp := valueWidget(&te.ListBase, 0, 1).(*core.Spinner)
		assert.Equal(t, float32(21), sp.Value, "the last row at the top")
		sp.SetValue(99)
		sp.SendChange()
		assert.Equal(t, 99.0, tsr.Value(2, 1))
		assert.Equal(t, 1, changes)

		w, ok := te.RowFirstVisWidget(1)
		assert.True(t, ok)
		assert.NotNil(t, w)
		_, ok = te.RowFirstVisWidget(-1)
		assert.False(t, ok)
		fw := te.RowGrabFocus(2)
		require.NotNil(t, fw)
		assert.Same(t, fw, te.RowGrabFocus(2), "already has the focus")
		assert.Nil(t, te.RowGrabFocus(100))

		assert.Nil(t, NewTensorEditor(b).SetTensor(nil))
	}))
}

// TestTensorEditorString tests a 1D string tensor, which is shown from
// the top down, with the widths of its strings, and editing it.
func TestTensorEditorString(t *testing.T) {
	b := core.NewBody()
	tsr := tensor.NewString(3)
	tsr.SetString1D("a longer string", 1)
	te := NewTensorEditor(b)
	te.ShowIndexes = true
	te.SetTensor(tsr)
	core.RenderImage(b, core.Locked(b, func() {
		assert.True(t, te.Layout.TopZero)
		assert.Equal(t, []int{len("a longer string")}, te.colMaxWidths)
		tf := valueWidget(&te.ListBase, 1, 0).(*core.TextField)
		assert.Equal(t, "a longer string", tf.Text())
		tf.SetText("x")
		tf.SendChange()
		assert.Equal(t, "x", tsr.String1D(1))

		ms, _ := te.Search("x", true)
		require.Len(t, ms, 1)
		assert.Equal(t, 1, ms[0].Region.Start.Line)
	}))
}

// TestTensorEditorSearch tests searching the values of a tensor, whose
// matches are at the rows of the display, in which the Y=0 row is at the
// bottom, and highlighting and selecting them.
func TestTensorEditorSearch(t *testing.T) {
	b := core.NewBody()
	te := NewTensorEditor(b)
	te.SetTensor(newTestTensor())
	core.RenderImage(b, core.Locked(b, func() {
		assert.Nil(t, te.TextRunes())
		ms, handled := te.Search("21", false)
		assert.True(t, handled)
		require.Len(t, ms, 1)
		assert.Equal(t, 0, ms[0].Region.Start.Line/te.NCols, "the top row")
		assert.Equal(t, 1, ms[0].Region.Start.Line%te.NCols)
		assert.True(t, te.HighlightMatches(ms))
		te.SelectMatch(ms, 0, true, false)
		te.SelectMatch(ms, 0, true, true)
		assert.True(t, te.HighlightMatches(nil))
	}))
}

// TestTensorEditorFiles tests saving and opening the tensor as a CSV file,
// with the toolbar, and that copy and paste do nothing yet.
func TestTensorEditorFiles(t *testing.T) {
	b := core.NewBody()
	tsr := newTestTensor()
	te := NewTensorEditor(b)
	te.SetTensor(tsr)
	var tb *core.Toolbar
	b.AddTopBar(func(bar *core.Frame) {
		tb = core.NewToolbar(bar)
		tb.Maker(te.MakeToolbar)
	})
	fn := filepath.Join(t.TempDir(), "t.tsv")
	core.RenderImage(b, core.Locked(b, func() {
		assert.GreaterOrEqual(t, tb.NumChildren(), 2)
		require.NoError(t, te.SaveCSV(core.Filename(fn)))
		tsr.SetZeros()
		require.NoError(t, te.OpenCSV(core.Filename(fn)))
		assert.Equal(t, 21.0, tsr.Value(2, 1))

		assert.Equal(t, fileinfo.DataCSV, te.MimeDataType())
		assert.Nil(t, te.CopySelectToMime())
		te.SelectIndexEvent(0, events.SelectOne)
		assert.Nil(t, te.CopySelectToMime())
		assert.Nil(t, te.FromMimeData(nil))
		te.PasteAssign(nil, 0)
		te.PasteAtIndex(nil, 0)
		assert.Equal(t, 21.0, tsr.Value(2, 1))
	}))

	empty := NewTensorEditor(core.NewBody())
	assert.NotPanics(t, func() { core.NewToolbar(core.NewBody()).Maker(empty.MakeToolbar) })
}

// TestGridLabels tests a grid with labels of more columns than rows,
// grouped by blank labels.
func TestGridLabels(t *testing.T) {
	b := core.NewBody()
	tsr := tensor.NewFloat64(2, 5)
	tg := NewTensorGrid(b)
	tg.SetTensor(tsr)
	tg.RowLabels = []string{"R0", ""}
	tg.ColumnLabels = RepeatsToBlank([]string{"A", "A", "B", "B", "C"})
	assert.Equal(t, []string{"A", "", "B", "", "C"}, tg.ColumnLabels)
	img := core.RenderImage(b, core.Locked(b, func() {
		assert.Len(t, tg.ColumnLabels, 5)
		assert.Equal(t, 2, tg.colNGps)
		assert.Greater(t, tg.colMaxSz.Y, float32(0))
	}))
	assert.NotNil(t, img)
}

// TestGridImage tests the display of tensors as images: greyscale, and
// color with the colors as the outer or inner dimension.
func TestGridImage(t *testing.T) {
	for _, shape := range [][]int{{4, 6}, {3, 4, 6}, {4, 4, 6}, {4, 6, 3}, {4, 6, 4}} {
		b := core.NewBody()
		tsr := tensor.NewFloat64(shape...)
		for i := range tsr.Len() {
			tsr.SetFloat1D(float64(i%5)/4, i)
		}
		tg := NewTensorGrid(b)
		tg.GridStyle.Image = true
		tg.SetTensor(tsr)
		img := core.RenderImage(b, core.Locked(b, func() {
			assert.Equal(t, float32(shape[1]), tg.MinSize().X)
		}))
		assert.NotNil(t, img, "%v", shape)
	}
}

// TestGridColors tests the range of the values of a grid that is not
// fixed, which is set from the data, and the colors of indexed and
// unknown color maps.
func TestGridColors(t *testing.T) {
	b := core.NewBody()
	tsr := newTestTensor()
	tg := NewTensorGrid(b)
	tg.GridStyle.Range.FixMin = false
	tg.GridStyle.Range.FixMax = false
	tg.SetTensor(tsr)
	core.RenderImage(b, core.Locked(b, func() {
		assert.Equal(t, 0.0, tg.GridStyle.Range.Min)
		assert.GreaterOrEqual(t, tg.GridStyle.Range.Max, 21.0)
		norm, _ := tg.Color(tg.GridStyle.Range.Max)
		assert.Equal(t, 1.0, norm)

		src := colormap.AvailableMaps[string(tg.GridStyle.ColorMap)]
		cm := &colormap.Map{Name: src.Name, Indexed: true, Colors: src.Colors}
		tg.ColorMap = cm
		tg.EnsureColorMap()
		assert.Same(t, cm, tg.ColorMap)
		norm, clr := tg.Color(1)
		assert.Equal(t, 0.0, norm)
		assert.Equal(t, cm.MapIndex(1), clr)

		tg.GridStyle.ColorMap = "NoSuchMap"
		tg.GridStyle.GridFill = 0.5
		tg.EnsureColorMap()
		assert.Equal(t, core.ColorMapName("ColdHot"), tg.GridStyle.ColorMap)
		assert.NotNil(t, tg.ColorMap)
	}))
}

// TestGridValue tests the grid as the value of a tensor, which does not
// show string tensors, and an empty grid.
func TestGridValue(t *testing.T) {
	tg := NewTensorGrid(core.NewBody())
	tsr := newTestTensor()
	require.NoError(t, tg.SetWidgetValue(tensor.Tensor(tsr)))
	assert.Same(t, tsr, *tg.WidgetValue().(*tensor.Tensor))
	tg.SetTensor(tensor.NewString(2))
	assert.Same(t, tsr, tg.Tensor, "not a string tensor")

	empty := NewTensorGrid(core.NewBody())
	assert.Equal(t, float32(0), empty.MinSize().X)

	gs := NewGridStyle()
	assert.Equal(t, float32(0.9), gs.GridFill)
	assert.Equal(t, core.ColorMapName("ColdHot"), gs.ColorMap)
}

// TestGridDialogs tests the dialogs of a grid: the tensor editor that a
// double click opens, which updates the grid when it is edited, and the
// style editor of its context menu.
func TestGridDialogs(t *testing.T) {
	defer core.TrackWindows()()
	b := core.NewBody()
	tsr := newTestTensor()
	tg := NewTensorGrid(b)
	tg.SetTensor(tsr)
	var tb *core.Toolbar
	b.AddTopBar(func(bar *core.Frame) {
		tb = core.NewToolbar(bar)
		tb.Maker(tg.MakeToolbar)
	})
	core.RenderImage(b, core.Locked(b, func() {
		assert.GreaterOrEqual(t, tb.NumChildren(), 2)
		tg.Send(events.DoubleClick)
	}), func() {
		sc := windowScene(t, "Tensor editor")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			te := findWidget[*TensorEditor](sc)
			require.NotNil(t, te)
			assert.Same(t, tsr, te.Tensor)
			te.SendChange()
		})()
	}, core.Locked(b, func() {
		require.Len(t, tg.ContextMenus, 1)
		m := core.NewScene()
		tg.ContextMenus[0](m, image.Point{})
		core.ButtonByText(m, "Edit style").Send(events.Click)
	}), func() {
		sc := windowScene(t, "Tensor grid style")
		require.NotNil(t, sc)
		core.Locked(sc, func() {
			fm := findWidget[*core.Form](sc)
			require.NotNil(t, fm)
			fm.SendChange()
		})()
	})
}
