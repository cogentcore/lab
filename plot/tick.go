// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"math"
	"strconv"
	"time"
)

// A Tick is a single tick mark on an axis.
type Tick struct {
	// Value is the data value marked by this Tick.
	Value float64

	// Label is the text to display at the tick mark.
	// If Label is an empty string then this is a minor tick mark.
	Label string
}

// IsMinor returns true if this is a minor tick mark.
func (tk *Tick) IsMinor() bool {
	return tk.Label == ""
}

// Ticker creates Ticks in a specified range
type Ticker interface {
	// Ticks returns Ticks in a specified range, with desired number of ticks,
	// which can be ignored depending on the ticker type.
	Ticks(mn, mx float64, nticks int) []Tick
}

// DefaultTicks is suitable for the Ticker field of an Axis,
// it returns a reasonable default set of tick marks.
type DefaultTicks struct{}

var _ Ticker = DefaultTicks{}

// Ticks returns Ticks in the specified range.
func (DefaultTicks) Ticks(mn, mx float64, nticks int) []Tick {
	if mx <= mn {
		panic("illegal range")
	}
	if nticks < 2 {
		return nil
	}

	labels, step, q, mag := talbotLinHanrahan(mn, mx, nticks, withinData, nil, nil, nil)
	majorDelta := step * math.Pow10(mag)
	if q == 0 {
		// Simple fall back was chosen, so
		// majorDelta is the label distance.
		majorDelta = labels[1] - labels[0]
	}

	// Choose a reasonable, but ad
	// hoc formatting for labels.
	fc := byte('f')
	var off int
	if mag < -1 || 6 < mag {
		off = 1
		fc = 'g'
	}
	mag10 := math.Pow10(mag)
	if math.Trunc(q*mag10) != q*mag10 {
		off += 2
	}
	prec := min(6, max(off, -mag))
	ticks := make([]Tick, len(labels))
	for i, v := range labels {
		ticks[i] = Tick{Value: v, Label: strconv.FormatFloat(float64(v), fc, prec, 64)}
	}

	var minorDelta float64
	// See talbotLinHanrahan for the values used here.
	switch step {
	case 1, 2.5:
		minorDelta = majorDelta / 5
	case 2, 3, 4, 5:
		minorDelta = majorDelta / step
	default:
		if majorDelta/2 < dlamchP {
			return ticks
		}
		minorDelta = majorDelta / 2
	}

	// Find the first minor tick not greater
	// than the lowest data value.
	var i float64
	for labels[0]+(i-1)*minorDelta > mn {
		i--
	}
	// Add ticks at minorDelta intervals when
	// they are not within minorDelta/2 of a
	// labelled tick.
	for {
		val := labels[0] + i*minorDelta
		if val > mx {
			break
		}
		found := false
		for _, t := range ticks {
			if math.Abs(t.Value-val) < minorDelta/2 {
				found = true
			}
		}
		if !found {
			ticks = append(ticks, Tick{Value: val})
		}
		i++
	}

	return ticks
}

// LogTicks is suitable for the Ticker field of an Axis,
// it returns tick marks suitable for a log-scale axis.
type LogTicks struct {
	// Prec specifies the precision of tick rendering
	// according to the documentation for strconv.FormatFloat.
	Prec int
}

var _ Ticker = LogTicks{}

// Ticks returns Ticks in a specified range
func (t LogTicks) Ticks(mn, mx float64, nticks int) []Tick {
	if mn <= 0 || mx <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}
	if nticks < 2 {
		return nil
	}

	val := math.Pow10(int(math.Log10(mn)))
	mx = math.Pow10(int(math.Ceil(math.Log10(mx))))
	var ticks []Tick
	for val < mx {
		for i := 1; i < 10; i++ {
			if i == 1 {
				ticks = append(ticks, Tick{Value: val, Label: formatFloatTick(val, t.Prec)})
			}
			ticks = append(ticks, Tick{Value: val * float64(i)})
		}
		val *= 10
	}
	ticks = append(ticks, Tick{Value: val, Label: formatFloatTick(val, t.Prec)})

	return ticks
}

// ConstantTicks is suitable for the Ticker field of an Axis.
// This function returns the given set of ticks.
type ConstantTicks []Tick

var _ Ticker = ConstantTicks{}

// Ticks returns Ticks in a specified range
func (ts ConstantTicks) Ticks(float64, float64, int) []Tick {
	return ts
}

// UnixTimeIn returns a time conversion function for the given location.
func UnixTimeIn(loc *time.Location) func(t float64) time.Time {
	return func(t float64) time.Time {
		return time.Unix(int64(t), 0).In(loc)
	}
}

// UTCUnixTime is the default time conversion for TimeTicks.
var UTCUnixTime = UnixTimeIn(time.UTC)

// TimeTicks is suitable for axes representing time values.
type TimeTicks struct {
	// Ticker is used to generate a set of ticks.
	// If nil, DefaultTicks will be used.
	Ticker Ticker

	// Format is the textual representation of the time value.
	// If empty, time.RFC3339 will be used
	Format string

	// Time takes a float32 value and converts it into a time.Time.
	// If nil, UTCUnixTime is used.
	Time func(t float64) time.Time
}

var _ Ticker = TimeTicks{}

// Ticks implements plot.Ticker.
func (t TimeTicks) Ticks(mn, mx float64, nticks int) []Tick {
	if t.Ticker == nil {
		t.Ticker = DefaultTicks{}
	}
	if t.Format == "" {
		t.Format = time.RFC3339
	}
	if t.Time == nil {
		t.Time = UTCUnixTime
	}
	if nticks < 2 {
		return nil
	}

	ticks := t.Ticker.Ticks(mn, mx, nticks)
	for i := range ticks {
		tick := &ticks[i]
		if tick.Label == "" {
			continue
		}
		tick.Label = t.Time(tick.Value).Format(t.Format)
	}
	return ticks
}

/*
// lengthOffset returns an offset that should be added to the
// tick mark's line to accout for its length.  I.e., the start of
// the line for a minor tick mark must be shifted by half of
// the length.
func (t Tick) lengthOffset(len vg.Length) vg.Length {
	if t.IsMinor() {
		return len / 2
	}
	return 0
}

// tickLabelHeight returns height of the tick mark labels.
func tickLabelHeight(sty text.Style, ticks []Tick) vg.Length {
	maxHeight := vg.Length(0)
	for _, t := range ticks {
		if t.IsMinor() {
			continue
		}
		r := sty.Rectangle(t.Label)
		h := r.Max.Y - r.Min.Y
		if h > maxHeight {
			maxHeight = h
		}
	}
	return maxHeight
}

// tickLabelWidth returns the width of the widest tick mark label.
func tickLabelWidth(sty text.Style, ticks []Tick) vg.Length {
	maxWidth := vg.Length(0)
	for _, t := range ticks {
		if t.IsMinor() {
			continue
		}
		r := sty.Rectangle(t.Label)
		w := r.Max.X - r.Min.X
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}
*/

// formatFloatTick returns a g-formated string representation of v
// to the specified precision.
func formatFloatTick(v float64, prec int) string {
	return strconv.FormatFloat(float64(v), 'g', prec, 64)
}

// // TickerFunc is suitable for the Ticker field of an Axis.
// // It is an adapter which allows to quickly setup a Ticker using a function with an appropriate signature.
// type TickerFunc func(min, max float64) []Tick
//
// var _ Ticker = TickerFunc(nil)
//
// // Ticks implements plot.Ticker.
// func (f TickerFunc) Ticks(min, max float64) []Tick {
// 	return f(min, max)
// }
