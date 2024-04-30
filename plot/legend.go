// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"cogentcore.org/core/styles/units"
)

// A Legend gives a description of the meaning of different
// data elements of the plot.  Each legend entry has a name
// and a thumbnail, where the thumbnail shows a small
// sample of the display style of the corresponding data.
type Legend struct {
	// TextStyle is the style given to the legend entry texts.
	TextStyle TextStyle

	// Top and Left specify the location of the legend.
	// If Top is true the legend is located along the top
	// edge of the plot, otherwise it is located along
	// the bottom edge.  If Left is true then the legend
	// is located along the left edge of the plot, and the
	// text is positioned after the icons, otherwise it is
	// located along the right edge and the text is
	// positioned before the icons.
	Top, Left bool

	// XOffs and YOffs are added to the legend's final position.
	XOffs, YOffs units.Value

	// ThumbnailWidth is the width of legend thumbnails.
	ThumbnailWidth units.Value

	// Entries are all of the LegendEntries described by this legend.
	Entries []LegendEntry
}

func (lg *Legend) Defaults() {
	lg.TextStyle.Defaults()
	lg.TextStyle.Padding.Dp(4)
	lg.Top = true
	lg.ThumbnailWidth.Pt(20)
}

// Add adds an entry to the legend with the given name.
// The entry's thumbnail is drawn as the composite of all of the
// thumbnails.
func (lg *Legend) Add(name string, thumbs ...Thumbnailer) {
	lg.Entries = append(lg.Entries, LegendEntry{Text: name, Thumbs: thumbs})
}

// Thumbnailer wraps the Thumbnail method, which
// draws the small image in a legend representing the
// style of data.
type Thumbnailer interface {
	// Thumbnail draws an thumbnail representing
	// a legend entry.  The thumbnail will usually show
	// a smaller representation of the style used
	// to plot the corresponding data.
	Thumbnail(pt *Plot)
}

// A LegendEntry represents a single line of a legend, it
// has a name and an icon.
type LegendEntry struct {
	// text is the text associated with this entry.
	Text string

	// thumbs is a slice of all of the thumbnails styles
	Thumbs []Thumbnailer
}
