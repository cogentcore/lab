// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/colors"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

// sinBandData returns a sine mean (Y) with Low, High deviations of
// varying width around it.
func sinBandData() plot.Data {
	n := 21
	xd, yd := make(plot.Values, n), make(plot.Values, n)
	low, high := make(plot.Values, n), make(plot.Values, n)
	for i := range n {
		x := float64(i)
		xd[i] = x * 5
		yd[i] = 50 + 40*math.Sin((x/8)*math.Pi)
		w := 5 + 4*math.Cos((x/8)*math.Pi)
		low[i] = w
		high[i] = w
	}
	return plot.Data{plot.X: xd, plot.Y: yd, plot.Low: low, plot.High: high}
}

// gaussBand returns a violin-like band at X position pos, with a
// gaussian density of given mean and sd along Y values from 0 to 10,
// as the High deviation around X, or as the absolute Low, High edges
// (without X) if edges is true.
func gaussBand(pos, mean, sd float64, edges bool) plot.Data {
	n := 41
	xd, yd := make(plot.Values, n), make(plot.Values, n)
	low, high := make(plot.Values, n), make(plot.Values, n)
	for i := range n {
		y := 10 * float64(i) / float64(n-1)
		d := 0.4 * math.Exp(-0.5*(y-mean)*(y-mean)/(sd*sd))
		yd[i] = y
		xd[i] = pos
		low[i] = pos - d
		high[i] = pos + d
	}
	if edges {
		return plot.Data{plot.Y: yd, plot.Low: low, plot.High: high}
	}
	for i := range n {
		high[i] -= pos
	}
	return plot.Data{plot.X: xd, plot.Y: yd, plot.High: high}
}

func TestBand(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Error band"
	data := sinBandData()

	bd := NewBand(plt, data)
	assert.NotNil(t, bd)
	ln := NewLine(plt, data)
	assert.NotNil(t, ln)
	// same color for the band and the line
	bd.Styler(func(s *plot.Style) {
		s.Line.Color = colors.Uniform(colors.Blue)
	})
	ln.Styler(func(s *plot.Style) {
		s.Line.Color = colors.Uniform(colors.Blue)
	})
	plt.Legend.Add("Band", bd)
	plt.Legend.Add("Mean", ln)

	assert.NotNil(t, bd.Center)
	pd, px, py := bd.Data()
	assert.Equal(t, data[plot.X], pd[plot.X])
	assert.Equal(t, data[plot.Y], pd[plot.Y])
	assert.Equal(t, data[plot.High], pd[plot.High])

	imagex.Assert(t, plt.RenderImage(), "band")
	_, px, py = bd.Data()
	assert.Equal(t, 21, len(px))
	assert.Equal(t, 21, len(py))
	// a larger High is higher on the screen (smaller Y pixel)
	assert.Less(t, py[4], py[0])

	// range includes the band
	assert.InDelta(t, 50+40+5, plt.Y.Range.Max, 1)
}

func TestBandEdges(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Band with edges, fill, and missing data"
	data := sinBandData()
	data[plot.High].(plot.Values)[10] = math.NaN()

	bd := NewBand(plt, data)
	assert.NotNil(t, bd)
	bd.Styler(func(s *plot.Style) {
		s.Line.On = plot.On
		s.Line.Color = colors.Uniform(colors.Red)
		s.Line.Fill = colors.Uniform(colors.Yellow)
	})
	assert.Equal(t, [][2]int{{0, 10}, {11, 21}}, bd.segments())
	imagex.Assert(t, plt.RenderImage(), "band-edges")
}

func TestBandArea(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Stacked areas"
	n := 21
	xd := make(plot.Values, n)
	a, b := make(plot.Values, n), make(plot.Values, n)
	for i := range n {
		xd[i] = float64(i)
		a[i] = 2 + math.Sin(float64(i)/4)
		b[i] = a[i] + 1 + 0.5*math.Cos(float64(i)/3)
	}
	// High only: fills down to 0
	b1 := NewBand(plt, plot.Data{plot.X: xd, plot.High: a})
	assert.NotNil(t, b1)
	b2 := NewBand(plt, plot.Data{plot.X: xd, plot.Low: a, plot.High: b})
	assert.NotNil(t, b2)
	plt.Legend.Add("A", b1)
	plt.Legend.Add("B", b2)
	imagex.Assert(t, plt.RenderImage(), "band-area")
	assert.Equal(t, 0.0, plt.Y.Range.Min)

	// a Valuer alone is the High values, with index positions
	plt = plot.New()
	b3 := NewBand(plt, tensor.NewFloat64FromValues(1, 3, 2))
	assert.NotNil(t, b3)
	assert.Equal(t, plot.Values{0, 1, 2}, b3.Pos)
	assert.Equal(t, plot.Values{0, 0, 0}, b3.Low)

	assert.Nil(t, NewBand(plot.New(), plot.Data{plot.Low: a}))
}

func TestXBand(t *testing.T) {
	plt := plot.New()
	plt.Title.Text = "Violin bodies"
	b1 := NewXBand(plt, gaussBand(1, 4, 1.5, false))
	assert.NotNil(t, b1)
	assert.Equal(t, b1.High, b1.Low) // Low is High around a Center
	b2 := NewXBand(plt, gaussBand(2, 6, 1, true))
	assert.Nil(t, b2.Center)
	assert.NotNil(t, b2)
	b2.Styler(func(s *plot.Style) {
		s.Line.On = plot.On
	})
	assert.True(t, b1.Horizontal)
	pd, _, _ := b1.Data()
	assert.NotNil(t, pd[plot.Y])
	assert.NotNil(t, pd[plot.X])
	imagex.Assert(t, plt.RenderImage(), "xband")
	assert.InDelta(t, 0.6, plt.X.Range.Min, 0.01)
	assert.InDelta(t, 2.4, plt.X.Range.Max, 0.01)
	assert.Equal(t, 10.0, plt.Y.Range.Max)
}

func TestTableBand(t *testing.T) {
	data := sinBandData()
	n := 2 * len(data[plot.X].(plot.Values))
	tx, ty := tensor.NewFloat64(n), tensor.NewFloat64(n)
	tl, th := tensor.NewFloat64(n), tensor.NewFloat64(n)
	sp := tensor.NewString(n)
	for i := range n {
		j := i % (n / 2)
		off := 0.0
		sp.SetString1D("A", i)
		if i >= n/2 {
			off = 30
			sp.SetString1D("B", i)
		}
		tx.SetFloat1D(data[plot.X].Float1D(j), i)
		ty.SetFloat1D(data[plot.Y].Float1D(j)+off, i)
		// absolute edges, as the Y column is not in the Band group
		tl.SetFloat1D(ty.Float1D(i)-data[plot.Low].Float1D(j), i)
		th.SetFloat1D(ty.Float1D(i)+data[plot.High].Float1D(j), i)
	}
	plot.SetStyler(tx, func(s *plot.Style) {
		s.Role = plot.X
	})
	plot.SetStyler(ty, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
		s.NoLegend = true
		s.Plot.Title = "Table bands with Split"
	})
	plot.SetStyler(tl, func(s *plot.Style) {
		s.Role = plot.Low
		s.Group = "Band"
	})
	plot.SetStyler(th, func(s *plot.Style) {
		s.On = true
		s.Role = plot.High
		s.Group = "Band"
		s.Plotter = BandType
	})
	plot.SetStyler(sp, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Split
	})
	dt := table.New("Bands")
	dt.AddColumn("X", tx)
	dt.AddColumn("Y", ty)
	dt.AddColumn("Low", tl)
	dt.AddColumn("High", th)
	dt.AddColumn("Split", sp)

	plt, err := plot.NewTablePlot(dt)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(plt.Plotters))
	imagex.Assert(t, plt.RenderImage(), "table-band")
}

// TestTableBandGroup tests an error band made by a second plotter in the
// group of the XY line, from the same columns, with the color of the line.
func TestTableBandGroup(t *testing.T) {
	for _, split := range []bool{false, true} {
		data := sinBandData()
		n := len(data[plot.X].(plot.Values))
		nr := n
		if split {
			nr = 2 * n
		}
		tx, ty, ts := tensor.NewFloat64(nr), tensor.NewFloat64(nr), tensor.NewFloat64(nr)
		sp := tensor.NewString(nr)
		for i := range nr {
			j := i % n
			off := 0.0
			sp.SetString1D("A", i)
			if i >= n {
				off = 30
				sp.SetString1D("B", i)
			}
			tx.SetFloat1D(data[plot.X].Float1D(j), i)
			ty.SetFloat1D(data[plot.Y].Float1D(j)+off, i)
			ts.SetFloat1D(data[plot.High].Float1D(j), i)
		}
		plot.SetStyler(tx, func(s *plot.Style) {
			s.Role = plot.X
		})
		plot.SetStyler(ty, func(s *plot.Style) {
			s.On = true
			s.Role = plot.Y
			s.Group = "Y"
			s.Plot.Title = "Table XY with an error Band in its group"
		})
		plot.SetStyler(ts, func(s *plot.Style) {
			s.On = true
			s.Role = plot.High
			s.Group = "Y"
			s.Plotter = BandType
		})
		dt := table.New("BandGroup")
		dt.AddColumn("X", tx)
		dt.AddColumn("Y", ty)
		dt.AddColumn("SEM", ts)
		if split {
			plot.SetStyler(sp, func(s *plot.Style) {
				s.On = true
				s.Role = plot.Split
			})
			dt.AddColumn("Split", sp)
		}
		plt, err := plot.NewTablePlot(dt)
		assert.NoError(t, err)
		nm := "table-band-group"
		np := 2
		if split {
			nm += "-split"
			np = 4
		}
		imagex.Assert(t, plt.RenderImage(), nm)
		assert.Equal(t, np, len(plt.Plotters))
		for i := 0; i < np/2; i++ {
			xy := plt.Plotters[i].(*XY)
			bd := plt.Plotters[np/2+i].(*Band)
			assert.Equal(t, xy.Style.Line.Color, bd.Style.Line.Color)
			assert.NotNil(t, bd.Center) // deviations around the Y of the group
		}
		nleg := 0 // no legend for one main plotter
		if split {
			nleg = 2 // the error bands are not in the legend
		}
		assert.Equal(t, nleg, len(plt.Legend.Entries))
	}
}
