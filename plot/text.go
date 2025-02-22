// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"image"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/htmltext"
	"cogentcore.org/core/text/rich"
	"cogentcore.org/core/text/shaped"
	"cogentcore.org/core/text/text"
)

// DefaultFontFamily specifies a default font family for plotting.
// if not set, the standard Cogent Core default font is used.
var DefaultFontFamily = rich.SansSerif

// TextStyle specifies styling parameters for Text elements.
type TextStyle struct { //types:add -setters
	// Size of font to render. Default is 16dp
	Size units.Value

	// Family name for font (inherited): ordered list of comma-separated names
	// from more general to more specific to use. Use split on, to parse.
	Family rich.Family

	// Color of text.
	Color image.Image

	// Align specifies how to align text along the relevant
	// dimension for the text element.
	Align styles.Aligns

	// Padding is used in a case-dependent manner to add
	// space around text elements.
	Padding units.Value

	// Rotation of the text, in degrees.
	Rotation float32

	// Offset is added directly to the final label location.
	Offset units.XY
}

func (ts *TextStyle) Defaults() {
	ts.Size.Dp(16)
	ts.Color = colors.Scheme.OnSurface
	ts.Align = styles.Center
	ts.Family = DefaultFontFamily
}

// Text specifies a single text element in a plot
type Text struct {

	// text string, which can use HTML formatting
	Text string

	// styling for this text element
	Style TextStyle

	// font has the font rendering styles.
	font rich.Style

	// textStyle has the text rendering styles.
	textStyle text.Style

	// PaintText is the [shaped.Lines] for painting the text.
	PaintText *shaped.Lines
}

func (tx *Text) Defaults() {
	tx.Style.Defaults()
}

// config is called during the layout of the plot, prior to drawing
func (tx *Text) Config(pt *Plot) {
	uc := &pt.Painter.UnitContext
	ts := &tx.textStyle
	fs := &tx.font
	fs.Defaults()
	ts.Defaults()
	ts.FontSize = tx.Style.Size
	ts.WhiteSpace = text.WrapNever
	fs.Family = tx.Style.Family
	if tx.Style.Color != colors.Scheme.OnSurface {
		fs.SetFillColor(colors.ToUniform(tx.Style.Color))
	}
	if math32.Abs(tx.Style.Rotation) > 10 {
		tx.Style.Align = styles.End
	}
	ts.ToDots(uc)
	tx.Style.Padding.ToDots(uc)
	txln := float32(len(tx.Text))
	fht := tx.textStyle.FontSize.Dots
	hsz := float32(12) * txln
	// txs := &pt.StandardTextStyle

	rt := errors.Log1(htmltext.HTMLToRich([]byte(tx.Text), fs, nil))
	tx.PaintText = pt.TextShaper.WrapLines(rt, fs, ts, &core.AppearanceSettings.Text, math32.Vec2(hsz, fht))
	if tx.Style.Rotation != 0 {
		// todo:
		// rotx := math32.Rotate2D(math32.DegToRad(tx.Style.Rotation))
		// tx.PaintText.Transform(rotx, fs, uc)
	}
}

func (tx *Text) ToDots(uc *units.Context) {
	tx.textStyle.ToDots(uc)
	tx.Style.Padding.ToDots(uc)
}

// Size returns the actual render size of the text.
func (tx *Text) Size() math32.Vector2 {
	return tx.PaintText.Bounds.Size().Ceil()
}

// PosX returns the starting position for a horizontally-aligned text element,
// based on given width.  Text must have been config'd already.
func (tx *Text) PosX(width float32) math32.Vector2 {
	rsz := tx.Size()
	pos := math32.Vector2{}
	pos.X = styles.AlignFactor(tx.Style.Align) * width
	switch tx.Style.Align {
	case styles.Center:
		pos.X -= 0.5 * rsz.X
	case styles.End:
		pos.X -= rsz.X
	}
	if math32.Abs(tx.Style.Rotation) > 10 {
		pos.Y += 0.5 * rsz.Y
	}
	return pos
}

// PosY returns the starting position for a vertically-rotated text element,
// based on given height.  Text must have been config'd already.
func (tx *Text) PosY(height float32) math32.Vector2 {
	rsz := tx.PaintText.Bounds.Size().Ceil()
	pos := math32.Vector2{}
	pos.Y = styles.AlignFactor(tx.Style.Align) * height
	switch tx.Style.Align {
	case styles.Center:
		pos.Y -= 0.5 * rsz.Y
	case styles.End:
		pos.Y -= rsz.Y
	}
	return pos
}

// Draw renders the text at given upper left position
func (tx *Text) Draw(pt *Plot, pos math32.Vector2) {
	pt.Painter.TextLines(tx.PaintText, pos)
}
