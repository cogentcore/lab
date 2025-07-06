// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tensorcore

import (
	"image/color"
	"log"

	"cogentcore.org/core/base/slicesx"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/abilities"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/tensor"
)

// LabelSpace is space after label in dot pixels.
const LabelSpace = 8

// TensorGrid is a widget that displays tensor values as a grid
// of colored squares. Higher-dimensional data is projected into 2D
// using [tensor.Projection2DShape] and related functions.
type TensorGrid struct {
	core.WidgetBase

	// Tensor is the tensor that we view.
	Tensor tensor.Tensor `set:"-"`

	// GridStyle has grid display style properties.
	GridStyle GridStyle

	// ColorMap is the colormap displayed (based on)
	ColorMap *colormap.Map

	// RowLabels are optional labels for each row of the 2D shape.
	// Empty strings cause grouping with rendered lines.
	RowLabels []string

	// ColumnLabels are optional labels for each column of the 2D shape.
	// Empty strings cause grouping with rendered lines.
	ColumnLabels []string

	rowMaxSz    math32.Vector2 // maximum label size
	rowMinBlank int            // minimum number of blank rows
	rowNGps     int            // number of groups in row (non-blank after blank)
	colMaxSz    math32.Vector2 // maximum label size
	colMinBlank int            // minimum number of blank cols
	colNGps     int            // number of groups in col (non-blank after blank)
}

func (tg *TensorGrid) WidgetValue() any { return &tg.Tensor }

func (tg *TensorGrid) SetWidgetValue(value any) error {
	tg.SetTensor(value.(tensor.Tensor))
	return nil
}

func (tg *TensorGrid) Init() {
	tg.WidgetBase.Init()
	tg.GridStyle.Defaults()
	tg.Styler(func(s *styles.Style) {
		s.SetAbilities(true, abilities.DoubleClickable)
		s.Background = colors.Scheme.Surface
		s.Font.Size.Dp(tg.GridStyle.FontSize)
		s.Font.Size.ToDots(&s.UnitContext)
		ms := tg.MinSize()
		s.Min.Set(units.Dot(ms.X), units.Dot(ms.Y))
		s.Grow.Set(1, 1)
	})

	tg.OnDoubleClick(func(e events.Event) {
		tg.TensorEditor()
	})
	tg.AddContextMenu(func(m *core.Scene) {
		core.NewFuncButton(m).SetFunc(tg.TensorEditor).SetIcon(icons.Edit)
		core.NewFuncButton(m).SetFunc(tg.EditStyle).SetIcon(icons.Edit)
	})
}

func (tg *TensorGrid) MakeToolbar(p *tree.Plan) {
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(tg.TensorEditor).SetIcon(icons.Edit)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(tg.EditStyle).SetIcon(icons.Edit)
	})
}

// SetTensor sets the tensor.  Must call Update after this.
func (tg *TensorGrid) SetTensor(tsr tensor.Tensor) *TensorGrid {
	if _, ok := tsr.(*tensor.String); ok {
		log.Printf("TensorGrid: String tensors cannot be displayed using TensorGrid\n")
		return tg
	}
	tg.Tensor = tsr
	if tg.Tensor != nil {
		tg.GridStyle.ApplyStylersFrom(tg.Tensor)
	}
	return tg
}

// TensorEditor pulls up a TensorEditor of our tensor
func (tg *TensorGrid) TensorEditor() { //types:add
	d := core.NewBody("Tensor editor")
	tb := core.NewToolbar(d)
	te := NewTensorEditor(d).SetTensor(tg.Tensor)
	te.OnChange(func(e events.Event) {
		tg.NeedsRender()
	})
	tb.Maker(te.MakeToolbar)
	d.RunWindowDialog(tg)
}

func (tg *TensorGrid) EditStyle() { //types:add
	d := core.NewBody("Tensor grid style")
	core.NewForm(d).SetStruct(&tg.GridStyle).
		OnChange(func(e events.Event) {
			tg.Restyle()
		})
	d.RunWindowDialog(tg)
}

// MinSize returns minimum size based on tensor and display settings
func (tg *TensorGrid) MinSize() math32.Vector2 {
	if tg.Tensor == nil || tg.Tensor.Len() == 0 {
		return math32.Vector2{}
	}
	if tg.GridStyle.Image {
		return math32.Vec2(float32(tg.Tensor.DimSize(1)), float32(tg.Tensor.DimSize(0)))
	}
	rows, cols, rowEx, colEx := tensor.Projection2DShape(tg.Tensor.Shape(), tg.GridStyle.OddRow)
	frw := float32(rows) + float32(rowEx)*tg.GridStyle.DimExtra // extra spacing
	fcl := float32(cols) + float32(colEx)*tg.GridStyle.DimExtra // extra spacing
	mx := float32(max(frw, fcl))
	gsz := tg.GridStyle.TotalSize / mx
	gsz = tg.GridStyle.Size.ClampValue(gsz)
	gsz = max(gsz, 2)
	sz := math32.Vec2(gsz*float32(fcl), gsz*float32(frw))

	if len(tg.RowLabels) > 0 {
		tg.RowLabels = slicesx.SetLength(tg.RowLabels, rows)
	}
	if len(tg.ColumnLabels) > 0 {
		tg.ColumnLabels = slicesx.SetLength(tg.ColumnLabels, rows)
	}
	tg.rowMinBlank, tg.rowNGps, tg.rowMaxSz = tg.SizeLabel(tg.RowLabels, false)
	tg.colMinBlank, tg.colNGps, tg.colMaxSz = tg.SizeLabel(tg.ColumnLabels, true)
	// tg.colMaxSz.Y += tg.rowMaxSz.Y // needs one more for some reason
	if tg.rowMaxSz.X > 0 {
		sz.X += tg.rowMaxSz.X + LabelSpace
	}
	if tg.colMaxSz.Y > 0 {
		sz.Y += tg.colMaxSz.Y + LabelSpace
	}
	return sz
}

func (tg *TensorGrid) SizeLabel(lbs []string, col bool) (minBlank, ngps int, sz math32.Vector2) {
	minBlank = len(lbs)
	if minBlank == 0 {
		return
	}
	mx := 0
	mxi := 0
	curblk := 0
	ngps = 0
	for i, lb := range lbs {
		l := len(lb)
		if l == 0 {
			curblk++
			continue
		}
		if curblk > 0 {
			ngps++
		}
		if i > 0 {
			minBlank = min(minBlank, curblk)
		}
		curblk = 0
		if l > mx {
			mx = l
			mxi = i
		}
	}
	minBlank = min(minBlank, curblk)
	ts := tg.Scene.TextShaper()
	if ts != nil {
		sty, tsty := tg.Styles.NewRichText()
		tx := rich.NewText(sty, []rune(lbs[mxi]))
		lns := ts.WrapLines(tx, sty, tsty, &rich.DefaultSettings, math32.Vec2(10000, 1000))
		sz = lns.Bounds.Size().Ceil()
		if col {
			sz.X, sz.Y = sz.Y, sz.X
		}
	}
	return
}

// EnsureColorMap makes sure there is a valid color map that matches specified name
func (tg *TensorGrid) EnsureColorMap() {
	if tg.ColorMap != nil && tg.ColorMap.Name != string(tg.GridStyle.ColorMap) {
		tg.ColorMap = nil
	}
	if tg.ColorMap == nil {
		ok := false
		tg.ColorMap, ok = colormap.AvailableMaps[string(tg.GridStyle.ColorMap)]
		if !ok {
			tg.GridStyle.ColorMap = ""
			tg.GridStyle.Defaults()
		}
		tg.ColorMap = colormap.AvailableMaps[string(tg.GridStyle.ColorMap)]
	}
}

func (tg *TensorGrid) Color(val float64) (norm float64, clr color.Color) {
	if tg.ColorMap.Indexed {
		clr = tg.ColorMap.MapIndex(int(val))
	} else {
		norm = tg.GridStyle.Range.ClipNormValue(val)
		clr = tg.ColorMap.Map(float32(norm))
	}
	return
}

func (tg *TensorGrid) UpdateRange() {
	if !tg.GridStyle.Range.FixMin || !tg.GridStyle.Range.FixMax {
		min, max, _, _ := tensor.Range(tg.Tensor.AsValues())
		if !tg.GridStyle.Range.FixMin {
			nmin := minmax.NiceRoundNumber(min, true) // true = below #
			tg.GridStyle.Range.Min = nmin
		}
		if !tg.GridStyle.Range.FixMax {
			nmax := minmax.NiceRoundNumber(max, false) // false = above #
			tg.GridStyle.Range.Max = nmax
		}
	}
}

func (tg *TensorGrid) Render() {
	if tg.Tensor == nil || tg.Tensor.Len() == 0 {
		return
	}
	tg.EnsureColorMap()
	tg.UpdateRange()
	if tg.GridStyle.Image {
		tg.renderImage()
		return
	}

	dimEx := tg.GridStyle.DimExtra
	tsr := tg.Tensor
	pc := &tg.Scene.Painter
	ts := tg.Scene.TextShaper()
	sty, tsty := tg.Styles.NewRichText()

	pos := tg.Geom.Pos.Content
	sz := tg.Geom.Size.Actual.Content
	// sz.SetSubScalar(tg.Disp.BotRtSpace.Dots)

	effsz := sz
	if tg.rowMaxSz.X > 0 {
		effsz.X -= tg.rowMaxSz.X + LabelSpace
	}
	if tg.colMaxSz.Y > 0 {
		effsz.Y -= tg.colMaxSz.Y + LabelSpace
	}

	pc.FillBox(pos, sz, tg.Styles.Background)

	rows, cols, rowEx, colEx := tensor.Projection2DShape(tsr.Shape(), tg.GridStyle.OddRow)
	rowsInner := rows
	colsInner := cols
	if rowEx > 0 {
		rowsInner = rows / rowEx
	}
	if colEx > 0 {
		colsInner = cols / colEx
	}
	// group lines
	rowEx += tg.rowNGps
	colEx += tg.colNGps
	frw := float32(rows) + float32(rowEx)*dimEx // extra spacing
	fcl := float32(cols) + float32(colEx)*dimEx // extra spacing

	tsz := math32.Vec2(fcl, frw)
	gsz := effsz.Div(tsz)

	if len(tg.RowLabels) > 0 { // Render Rows
		epos := pos
		epos.Y += tg.colMaxSz.Y + LabelSpace
		nr := len(tg.RowLabels)
		mx := min(nr, rows)
		ygp := 0
		prvyblk := false
		for y := 0; y < mx; y++ {
			lb := tg.RowLabels[y]
			if len(lb) == 0 {
				prvyblk = true
				continue
			}
			if prvyblk {
				ygp++
				prvyblk = false
			}
			yex := float32(ygp) * dimEx
			tx := rich.NewText(sty, []rune(lb))
			lns := ts.WrapLines(tx, sty, tsty, &rich.DefaultSettings, math32.Vec2(10000, 1000))
			cr := math32.Vec2(0, float32(y)+yex)
			pr := epos.Add(cr.Mul(gsz))
			pc.DrawText(lns, pr)
		}
		pos.X += tg.rowMaxSz.X + LabelSpace
	}

	if len(tg.ColumnLabels) > 0 { // Render Cols
		epos := pos
		if tg.GridStyle.ColumnRotation > 0 {
			epos.X += tg.colMaxSz.X
		}
		nc := len(tg.ColumnLabels)
		mx := min(nc, cols)
		xgp := 0
		prvxblk := false
		for x := 0; x < mx; x++ {
			lb := tg.ColumnLabels[x]
			if len(lb) == 0 {
				prvxblk = true
				continue
			}
			if prvxblk {
				xgp++
				prvxblk = false
			}
			xex := float32(xgp) * dimEx
			tx := rich.NewText(sty, []rune(lb))
			lns := ts.WrapLines(tx, sty, tsty, &rich.DefaultSettings, math32.Vec2(10000, 1000))
			cr := math32.Vec2(float32(x)+xex, 0)
			pr := epos.Add(cr.Mul(gsz))
			rot := tg.GridStyle.ColumnRotation
			if rot < 0 {
				pr.Y += tg.colMaxSz.Y
			}
			rotx := math32.Rotate2DAround(math32.DegToRad(rot), pr)
			m := pc.Paint.Transform
			pc.Paint.Transform = m.Mul(rotx)
			pc.DrawText(lns, pr)
			pc.Paint.Transform = m
		}
		pos.Y += tg.colMaxSz.Y + LabelSpace
	}

	ssz := gsz.MulScalar(tg.GridStyle.GridFill) // smaller size with margin
	prvyblk := false
	ygp := 0
	for y := 0; y < rows; y++ {
		yex := float32(int(y/rowsInner)) * dimEx
		if len(tg.RowLabels) > 0 {
			ylb := tg.RowLabels[y]
			if len(ylb) > 0 && prvyblk {
				ygp++
				prvyblk = false
			} else if len(ylb) == 0 {
				prvyblk = true
			}
			yex += float32(ygp) * dimEx
		}
		prvxblk := false
		xgp := 0
		for x := 0; x < cols; x++ {
			xex := float32(int(x/colsInner)) * dimEx
			ey := y
			if !tg.GridStyle.TopZero {
				ey = (rows - 1) - y
			}
			if len(tg.ColumnLabels) > 0 {
				xlb := tg.ColumnLabels[x]
				if len(xlb) > 0 && prvxblk {
					xgp++
					prvxblk = false
				} else if len(xlb) == 0 {
					prvxblk = true
				}
				xex += float32(xgp) * dimEx
			}
			val := tensor.Projection2DValue(tsr, tg.GridStyle.OddRow, ey, x)
			cr := math32.Vec2(float32(x)+xex, float32(y)+yex)
			pr := pos.Add(cr.Mul(gsz))
			_, clr := tg.Color(val)
			pc.FillBox(pr, ssz, colors.Uniform(clr))
		}
	}
}

func (tg *TensorGrid) renderImage() {
	if tg.Tensor == nil || tg.Tensor.Len() == 0 {
		return
	}
	pc := &tg.Scene.Painter
	pos := tg.Geom.Pos.Content
	sz := tg.Geom.Size.Actual.Content
	pc.FillBox(pos, sz, tg.Styles.Background)
	tsr := tg.Tensor
	ysz := tsr.DimSize(0)
	xsz := tsr.DimSize(1)
	nclr := 1
	outclr := false // outer dimension is color
	if tsr.NumDims() == 3 {
		if tsr.DimSize(0) == 3 || tsr.DimSize(0) == 4 {
			outclr = true
			ysz = tsr.DimSize(1)
			xsz = tsr.DimSize(2)
			nclr = tsr.DimSize(0)
		} else {
			nclr = tsr.DimSize(2)
		}
	}
	tsz := math32.Vec2(float32(xsz), float32(ysz))
	gsz := sz.Div(tsz)
	for y := 0; y < ysz; y++ {
		for x := 0; x < xsz; x++ {
			ey := y
			if !tg.GridStyle.TopZero {
				ey = (ysz - 1) - y
			}
			switch {
			case outclr:
				var r, g, b, a float64
				a = 1
				r = tg.GridStyle.Range.ClipNormValue(tsr.Float(0, y, x))
				g = tg.GridStyle.Range.ClipNormValue(tsr.Float(1, y, x))
				b = tg.GridStyle.Range.ClipNormValue(tsr.Float(2, y, x))
				if nclr > 3 {
					a = tg.GridStyle.Range.ClipNormValue(tsr.Float(3, y, x))
				}
				cr := math32.Vec2(float32(x), float32(ey))
				pr := pos.Add(cr.Mul(gsz))
				pc.Stroke.Color = colors.Uniform(colors.FromFloat64(r, g, b, a))
				pc.FillBox(pr, gsz, pc.Stroke.Color)
			case nclr > 1:
				var r, g, b, a float64
				a = 1
				r = tg.GridStyle.Range.ClipNormValue(tsr.Float(y, x, 0))
				g = tg.GridStyle.Range.ClipNormValue(tsr.Float(y, x, 1))
				b = tg.GridStyle.Range.ClipNormValue(tsr.Float(y, x, 2))
				if nclr > 3 {
					a = tg.GridStyle.Range.ClipNormValue(tsr.Float(y, x, 3))
				}
				cr := math32.Vec2(float32(x), float32(ey))
				pr := pos.Add(cr.Mul(gsz))
				pc.Stroke.Color = colors.Uniform(colors.FromFloat64(r, g, b, a))
				pc.FillBox(pr, gsz, pc.Stroke.Color)
			default:
				val := tg.GridStyle.Range.ClipNormValue(tsr.Float(y, x))
				cr := math32.Vec2(float32(x), float32(ey))
				pr := pos.Add(cr.Mul(gsz))
				pc.Stroke.Color = colors.Uniform(colors.FromFloat64(val, val, val, 1))
				pc.FillBox(pr, gsz, pc.Stroke.Color)
			}
		}
	}
}
