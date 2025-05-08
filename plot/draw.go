// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted from gonum/plot:
// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"image"

	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint/render"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/sides"
)

// drawConfig configures everything for drawing, applying styles etc.
func (pt *Plot) drawConfig() {
	pt.applyStyle()
	pt.X.drawConfig()
	pt.Y.drawConfig()
	pt.YR.drawConfig()
	pt.Z.drawConfig()
	pt.Painter.ToDots()
}

// Draw draws the plot to image.
// Plotters are drawn in the order in which they were
// added to the plot.
func (pt *Plot) Draw() {
	pt.drawConfig()
	pc := pt.Painter

	ptb := pt.PaintBox
	off := math32.FromPoint(pt.PaintBox.Min)
	sz := pt.PaintBox.Size()
	ptw := float32(sz.X)
	pth := float32(sz.Y)

	pc.PushContext(nil, render.NewBoundsRect(pt.PaintBox, sides.Floats{}))

	if pt.Style.Background != nil {
		pc.BlitBox(off, math32.FromPoint(sz), pt.Style.Background)
	}

	if pt.Title.Text != "" {
		pt.Title.Config(pt)
		pos := pt.Title.PosX(ptw)
		pad := pt.Title.Style.Padding.Dots
		pos.Y = pad
		pt.Title.Draw(pt, pos.Add(off))
		rsz := pt.Title.PaintText.Bounds.Size().Ceil()
		th := rsz.Y + 2*pad
		pth -= th
		ptb.Min.Y += int(math32.Ceil(th))
	}

	pt.X.SanitizeRange()
	pt.Y.SanitizeRange()
	pt.YR.SanitizeRange()

	ywidth, tickWidth, tpad, bpad := pt.Y.sizeY(pt, ptb.Min.Y)
	yrwidth, yrtickWidth, yrtpad, yrbpad := pt.YR.sizeY(pt, ptb.Min.Y)
	xheight, lpad, rpad := pt.X.sizeX(pt, float32(ywidth), float32(yrwidth), float32(sz.X-int(ywidth+yrwidth)))

	tb := ptb
	tb.Min.X += ywidth
	tb.Max.X -= yrwidth
	pt.X.drawX(pt, tb, lpad, rpad)

	tb = ptb
	tb.Max.Y -= xheight
	pt.Y.drawY(pt, tb, tickWidth, tpad, bpad)
	pt.YR.drawY(pt, tb, yrtickWidth, yrtpad, yrbpad)

	tb = ptb
	tb.Min.X += ywidth + lpad
	tb.Max.X -= yrwidth + rpad
	tb.Max.Y -= xheight + bpad
	tb.Min.Y += tpad
	pt.PlotBox.SetFromRect(tb)

	// don't cut off lines
	tb.Min.X -= 2
	tb.Min.Y -= 2
	tb.Max.X += 2
	tb.Max.Y += 2
	pt.PushBounds(tb)

	for _, plt := range pt.Plotters {
		plt.Plot(pt)
	}

	pt.Legend.draw(pt)
	pc.PopContext()
	pc.PopContext() // global
	pc.RenderToImage()
}

////////	Axis

// drawTicks returns true if the tick marks should be drawn.
func (ax *Axis) drawTicks() bool {
	return ax.Style.TickLine.Width.Value > 0 && ax.Style.TickLength.Value > 0
}

// sizeX returns the total height of the axis, left and right padding
func (ax *Axis) sizeX(pt *Plot, yw, yrw, axw float32) (ht, lpad, rpad int) {
	if !ax.Style.On {
		return
	}
	uc := pt.UnitContext()
	ax.Style.TickLength.ToDots(uc)
	ax.ticks = ax.Ticker.Ticks(ax.Range.Min, ax.Range.Max, ax.Style.NTicks)
	h := float32(0)
	if ax.Label.Text != "" { // We assume that the label isn't rotated.
		ax.Label.Config(pt)
		h += ax.Label.Size().Y
		h += ax.Label.Style.Padding.Dots
	}
	lw := ax.Style.Line.Width.Dots
	lpad = int(math32.Ceil(lw)) + 4
	rpad = int(math32.Ceil(lw)) + 4
	tht := float32(0)
	if len(ax.ticks) > 0 {
		if ax.drawTicks() {
			h += ax.Style.TickLength.Dots
		}
		ftk := ax.firstTickLabel()
		if ftk.Label != "" {
			px, _ := ax.tickPosX(pt, ftk, axw)
			if px < -yw {
				lpad += int(math32.Ceil(-px - yw))
			}
			tht = max(tht, ax.TickText.Size().Y)
		}
		ltk := ax.lastTickLabel()
		if ltk.Label != "" {
			px, wd := ax.tickPosX(pt, ltk, axw)
			if px+wd > axw+yrw {
				rpad += int(math32.Ceil((px + wd) - (axw + yrw)))
			}
			tht = max(tht, ax.TickText.Size().Y)
		}
		ax.TickText.Text = ax.longestTickLabel()
		if ax.TickText.Text != "" {
			ax.TickText.Config(pt)
			tht = max(tht, ax.TickText.Size().Y)
		}
		h += ax.TickText.Style.Padding.Dots
	}
	h += tht + lw + ax.Style.Padding.Dots

	ht = int(math32.Ceil(h))
	return
}

// tickLabelPosX returns the relative position and width for given tick along X axis
// for given total axis width
func (ax *Axis) tickPosX(pt *Plot, t Tick, axw float32) (px, wd float32) {
	x := axw * float32(ax.Norm(t.Value))
	if x < 0 || x > axw {
		return
	}
	ax.TickText.Text = t.Label
	ax.TickText.Config(pt)
	pos := ax.TickText.PosX(0)
	px = pos.X + x
	wd = ax.TickText.Size().X
	return
}

func (ax *Axis) firstTickLabel() Tick {
	for _, tk := range ax.ticks {
		if tk.Label != "" {
			return tk
		}
	}
	return Tick{}
}

func (ax *Axis) lastTickLabel() Tick {
	n := len(ax.ticks)
	for i := n - 1; i >= 0; i-- {
		tk := ax.ticks[i]
		if tk.Label != "" {
			return tk
		}
	}
	return Tick{}
}

func (ax *Axis) longestTickLabel() string {
	lst := ""
	for _, tk := range ax.ticks {
		if len(tk.Label) > len(lst) {
			lst = tk.Label
		}
	}
	return lst
}

func (ax *Axis) sizeY(pt *Plot, theight int) (ywidth, tickWidth, tpad, bpad int) {
	if !ax.Style.On {
		return
	}
	uc := pt.UnitContext()
	ax.ticks = ax.Ticker.Ticks(ax.Range.Min, ax.Range.Max, ax.Style.NTicks)
	ax.Style.TickLength.ToDots(uc)

	w := float32(0)
	if ax.Label.Text != "" {
		ax.Label.Config(pt)
		w += ax.Label.Size().X
		w += ax.Label.Style.Padding.Dots
	}

	lw := ax.Style.Line.Width.Dots
	tpad = int(math32.Ceil(lw)) + 2
	bpad = int(math32.Ceil(lw)) + 2

	if len(ax.ticks) > 0 {
		if ax.drawTicks() {
			w += ax.Style.TickLength.Dots
		}
		ax.TickText.Text = ax.longestTickLabel()
		if ax.TickText.Text != "" {
			ax.TickText.Config(pt)
			tw := math32.Ceil(ax.TickText.Size().X + ax.TickText.Style.Padding.Dots)
			w += tw
			tickWidth = int(tw)
			tht := int(math32.Ceil(0.5 * ax.TickText.Size().X))
			if theight == 0 {
				tpad += tht
			}
		}
	}
	w += lw + ax.Style.Padding.Dots
	ywidth = int(math32.Ceil(w))
	return
}

// drawX draws the horizontal axis
func (ax *Axis) drawX(pt *Plot, ab image.Rectangle, lpad, rpad int) {
	if !ax.Style.On {
		return
	}
	ab.Min.X += lpad
	ab.Max.X -= rpad
	axw := float32(ab.Size().X)
	// axh := float32(ab.Size().Y) // height of entire plot
	if ax.Label.Text != "" {
		ax.Label.Config(pt)
		pos := ax.Label.PosX(axw)
		pos.X += float32(ab.Min.X)
		th := ax.Label.Size().Y
		pos.Y = float32(ab.Max.Y) - th
		ax.Label.Draw(pt, pos)
		ab.Max.Y -= int(math32.Ceil(th + ax.Label.Style.Padding.Dots))
	}

	tickHt := float32(0)
	for _, t := range ax.ticks {
		x := axw * float32(ax.Norm(t.Value))
		if x < 0 || x > axw || t.IsMinor() {
			continue
		}
		ax.TickText.Text = t.Label
		ax.TickText.Config(pt)
		pos := ax.TickText.PosX(0)
		pos.X += x + float32(ab.Min.X)
		tickHt = ax.TickText.Size().Y + ax.TickText.Style.Padding.Dots
		pos.Y += float32(ab.Max.Y) - tickHt
		ax.TickText.Draw(pt, pos)
	}

	if len(ax.ticks) > 0 {
		ab.Max.Y -= int(math32.Ceil(tickHt))
		// } else {
		// 	y += ax.Width / 2
	}

	if len(ax.ticks) > 0 && ax.drawTicks() {
		ln := ax.Style.TickLength.Dots
		for _, t := range ax.ticks {
			yoff := float32(0)
			if t.IsMinor() {
				yoff = 0.5 * ln
			}
			x := axw * float32(ax.Norm(t.Value))
			if x < 0 || x > axw {
				continue
			}
			x += float32(ab.Min.X)
			ax.Style.TickLine.Draw(pt, math32.Vec2(x, float32(ab.Max.Y)-yoff), math32.Vec2(x, float32(ab.Max.Y)-ln))
		}
		ab.Max.Y -= int(ln - 0.5*ax.Style.Line.Width.Dots)
	}

	ax.Style.Line.Draw(pt, math32.Vec2(float32(ab.Min.X), float32(ab.Max.Y)), math32.Vec2(float32(ab.Min.X)+axw, float32(ab.Max.Y)))
}

// drawY draws the Y axis along the left side
func (ax *Axis) drawY(pt *Plot, ab image.Rectangle, tickWidth, tpad, bpad int) {
	if !ax.Style.On {
		return
	}
	ab.Min.Y += tpad
	ab.Max.Y -= bpad
	axh := float32(ab.Size().Y)
	xpos := float32(ab.Min.X)
	if ax.RightY {
		xpos = float32(ab.Max.X)
	}
	if ax.Label.Text != "" {
		ax.Label.Style.Align = styles.Center
		pos := ax.Label.PosY(axh)
		tw := math32.Ceil(ax.Label.Size().X + ax.Label.Style.Padding.Dots)
		if ax.RightY {
			pos.Y += float32(ab.Min.Y)
			pos.X = xpos
			xpos -= tw
		} else {
			pos.Y += float32(ab.Min.Y) + ax.Label.Size().Y
			pos.X = xpos
			xpos += tw
		}
		ax.Label.Draw(pt, pos)
	}

	if len(ax.ticks) > 0 && ax.RightY {
		xpos -= float32(tickWidth)
	}
	for _, t := range ax.ticks {
		y := axh * (1 - float32(ax.Norm(t.Value)))
		if y < 0 || y > axh || t.IsMinor() {
			continue
		}
		ax.TickText.Text = t.Label
		ax.TickText.Config(pt)
		pos := ax.TickText.PosX(float32(tickWidth))
		pos.X += xpos
		pos.Y = float32(ab.Min.Y) + y - 0.5*ax.TickText.Size().Y
		ax.TickText.Draw(pt, pos)
	}

	if len(ax.ticks) > 0 && !ax.RightY {
		xpos += float32(tickWidth)
	}

	if len(ax.ticks) > 0 && ax.drawTicks() {
		ln := ax.Style.TickLength.Dots
		if ax.RightY {
			xpos -= math32.Ceil(ln + 0.5*ax.Style.Line.Width.Dots)
		}
		for _, t := range ax.ticks {
			xoff := float32(0)
			eln := ln
			if t.IsMinor() {
				if ax.RightY {
					eln *= .5
				} else {
					xoff = 0.5 * ln
				}
			}
			y := axh * (1 - float32(ax.Norm(t.Value)))
			if y < 0 || y > axh {
				continue
			}
			y += float32(ab.Min.Y)
			ax.Style.TickLine.Draw(pt, math32.Vec2(xpos+xoff, y), math32.Vec2(xpos+eln, y))
		}
		if !ax.RightY {
			xpos += math32.Ceil(ln + 0.5*ax.Style.Line.Width.Dots)
		}
	}

	ax.Style.Line.Draw(pt, math32.Vec2(xpos, float32(ab.Min.Y)), math32.Vec2(xpos, float32(ab.Max.Y)))
}

////////	Legend

// draw draws the legend
func (lg *Legend) draw(pt *Plot) {
	pc := pt.Painter
	uc := pt.UnitContext()
	ptb := pt.CurBounds()

	lg.Style.ThumbnailWidth.ToDots(uc)
	lg.Style.Position.XOffs.ToDots(uc)
	lg.Style.Position.YOffs.ToDots(uc)

	var ltxt Text
	ltxt.Defaults()
	ltxt.Style = lg.Style.Text
	ltxt.ToDots(uc)
	pad := math32.Ceil(ltxt.Style.Padding.Dots)
	em := ltxt.textStyle.FontHeight(&ltxt.font)
	var sz image.Point
	maxTht := 0
	for _, e := range lg.Entries {
		ltxt.Text = e.Text
		ltxt.Config(pt)
		sz.X = max(sz.X, int(math32.Ceil(ltxt.Size().X)))
		tht := int(math32.Ceil(ltxt.Size().Y + pad))
		maxTht = max(tht, maxTht)
	}
	sz.X += int(em)
	sz.Y = len(lg.Entries) * maxTht
	txsz := sz
	sz.X += int(lg.Style.ThumbnailWidth.Dots)

	pos := ptb.Min
	if lg.Style.Position.Left {
		pos.X += int(lg.Style.Position.XOffs.Dots)
	} else {
		pos.X = ptb.Max.X - sz.X - int(lg.Style.Position.XOffs.Dots)
	}
	if lg.Style.Position.Top {
		pos.Y += int(lg.Style.Position.YOffs.Dots)
	} else {
		pos.Y = ptb.Max.Y - sz.Y - int(lg.Style.Position.YOffs.Dots)
	}

	if lg.Style.Fill != nil {
		pc.FillBox(math32.FromPoint(pos), math32.FromPoint(sz), lg.Style.Fill)
	}
	cp := pos
	thsz := image.Point{X: int(lg.Style.ThumbnailWidth.Dots), Y: maxTht - 2*int(pad)}
	for _, e := range lg.Entries {
		tp := cp
		tp.X += int(txsz.X)
		tp.Y += int(pad)
		tb := image.Rectangle{Min: tp, Max: tp.Add(thsz)}
		pt.PushBounds(tb)
		for _, t := range e.Thumbs {
			t.Thumbnail(pt)
		}
		pc.PopContext()
		ltxt.Text = e.Text
		ltxt.Config(pt)
		ltxt.Draw(pt, math32.FromPoint(cp))
		cp.Y += maxTht
	}
}
