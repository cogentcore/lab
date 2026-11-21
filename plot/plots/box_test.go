// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"math"
	"math/rand"
	"slices"
	"testing"

	"cogentcore.org/core/base/iox/imagex"
	"cogentcore.org/core/colors"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
)

// boxSamples returns sorted gaussian samples for each of the given means.
func boxSamples(rng *rand.Rand, n int, means ...float64) [][]float64 {
	smp := make([][]float64, len(means))
	for g, mn := range means {
		vals := make([]float64, n)
		for i := range vals {
			vals[i] = mn + (1+0.5*float64(g))*rng.NormFloat64()
		}
		slices.Sort(vals)
		smp[g] = vals
	}
	return smp
}

// boxStats returns the box plot data for the given sorted samples:
// the quartiles, and the whiskers at the furthest values within 1.5 times
// the interquartile range, along with the outliers beyond them.
func boxStats(smp [][]float64) (data plot.Data, outliers plot.Data) {
	ng := len(smp)
	med, q1, q3 := make(plot.Values, ng), make(plot.Values, ng), make(plot.Values, ng)
	low, high := make(plot.Values, ng), make(plot.Values, ng)
	var ox, oy plot.Values
	for g, vals := range smp {
		q := func(p float64) float64 { return vals[int(p*float64(len(vals)-1)+0.5)] }
		med[g], q1[g], q3[g] = q(0.5), q(0.25), q(0.75)
		iqr := q3[g] - q1[g]
		low[g], high[g] = q3[g], q1[g]
		for _, v := range vals {
			if v < q1[g]-1.5*iqr || v > q3[g]+1.5*iqr {
				ox = append(ox, float64(g+1)) // Box positions start at 1
				oy = append(oy, v)
				continue
			}
			low[g] = min(low[g], v)
			high[g] = max(high[g], v)
		}
	}
	data = plot.Data{plot.Y: med, plot.U: q1, plot.V: q3, plot.Low: low, plot.High: high}
	outliers = plot.Data{plot.X: ox, plot.Y: oy}
	return
}

// violinData returns the XBand data for a violin of the given sorted
// samples at X position pos: a gaussian kernel density over the range
// of the samples, scaled to a maximum half width of 0.4.
func violinData(vals []float64, pos float64) plot.Data {
	n := 50
	bw := 0.5
	xd, yd, hd := make(plot.Values, n), make(plot.Values, n), make(plot.Values, n)
	mn, mx := vals[0], vals[len(vals)-1]
	dmax := 0.0
	for i := range n {
		y := mn + (mx-mn)*float64(i)/float64(n-1)
		d := 0.0
		for _, v := range vals {
			z := (y - v) / bw
			d += math.Exp(-0.5 * z * z)
		}
		xd[i], yd[i], hd[i] = pos, y, d
		dmax = max(dmax, d)
	}
	for i := range hd {
		hd[i] *= 0.4 / dmax
	}
	return plot.Data{plot.X: xd, plot.Y: yd, plot.High: hd}
}

func TestBox(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	smp := boxSamples(rng, 100, 2, 5, 3)
	data, outs := boxStats(smp)
	data[plot.X] = tensor.NewStringFromValues("A", "B", "C")

	plt := plot.New()
	plt.Title.Text = "Box plots with violins and outliers"
	for g, vals := range smp {
		vb := NewXBand(plt, violinData(vals, float64(g+1)))
		assert.NotNil(t, vb)
		vb.Styler(func(s *plot.Style) {
			s.Line.Color = colors.Uniform(colors.Gray)
		})
	}
	bx := NewBox(plt, data)
	assert.NotNil(t, bx)
	bx.Styler(func(s *plot.Style) {
		s.Width.Width = 0.2
		s.Line.Fill = colors.Uniform(colors.White)
	})
	sc := NewScatter(plt, outs)
	assert.NotNil(t, sc)

	imagex.Assert(t, plt.RenderImage(), "box")
	pd, px, _ := bx.Data()
	assert.Equal(t, 3, len(px))
	assert.Less(t, px[0], px[1])
	assert.Equal(t, plot.Labels{"A", "B", "C"}, pd[plot.X])
	ticks, ok := plt.X.Ticker.(plot.ConstantTicks)
	assert.True(t, ok)
	assert.Equal(t, "B", ticks[1].Label)
	assert.Equal(t, 2.0, ticks[1].Value)
}

func TestXBox(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	smp := boxSamples(rng, 50, 3, 6, 4, 5)
	data, _ := boxStats(smp)
	data[plot.Y].(plot.Values)[2] = math.NaN()    // missing box
	data[plot.High].(plot.Values)[3] = math.NaN() // missing whisker

	plt := plot.New()
	plt.Title.Text = "Horizontal boxes"
	bx := NewXBox(plt, data)
	assert.NotNil(t, bx)
	assert.True(t, bx.Horizontal)
	plt.Legend.Add("Boxes", bx)
	imagex.Assert(t, plt.RenderImage(), "xbox")
	assert.Equal(t, float32(0), bx.PX[2])
	assert.InDelta(t, 0, plt.Y.Range.Min, 0.01) // Offset 1 - Pad 1
	assert.InDelta(t, 5, plt.Y.Range.Max, 0.01)

	// all required roles
	delete(data, plot.U)
	assert.Nil(t, NewBox(plot.New(), data))
}

func TestTableBox(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	ng := 3
	dt := table.New("Boxes")
	gp := dt.AddStringColumn("Group")
	cond := dt.AddStringColumn("Cond")
	cols := map[plot.Roles]*tensor.Float64{}
	for _, rl := range []plot.Roles{plot.Y, plot.U, plot.V, plot.Low, plot.High} {
		cols[rl] = dt.AddFloat64Column(rl.String())
	}
	for ci, cnm := range []string{"Control", "Test"} {
		smp := boxSamples(rng, 50, 2+float64(ci), 4+2*float64(ci), 3)
		data, _ := boxStats(smp)
		for g := range ng {
			dt.AddRows(1)
			row := dt.NumRows() - 1
			gp.SetString1D(string(rune('A'+g)), row)
			cond.SetString1D(cnm, row)
			for rl, cl := range cols {
				cl.SetFloat1D(data[rl].Float1D(g), row)
			}
		}
	}
	plot.SetStyler(gp, func(s *plot.Style) {
		s.Role = plot.X
		s.Plot.Title = "Table box plots with Split"
	})
	plot.SetStyler(cond, func(s *plot.Style) {
		s.On = true
		s.Role = plot.Split
	})
	plot.SetStyler(cols[plot.Y], func(s *plot.Style) {
		s.On = true
		s.Role = plot.Y
		s.Plotter = BoxType
		s.Label = "Value"
	})
	for _, rl := range []plot.Roles{plot.U, plot.V, plot.Low, plot.High} {
		plot.SetStyler(cols[rl], func(s *plot.Style) {
			s.Role = rl
		})
	}
	plt, err := plot.NewTablePlot(dt)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(plt.Plotters))
	imagex.Assert(t, plt.RenderImage(), "table-box")
	b0 := plt.Plotters[0].(*Box)
	b1 := plt.Plotters[1].(*Box)
	assert.Less(t, b0.pos(0), b1.pos(0)) // side by side
	ticks, ok := plt.X.Ticker.(plot.ConstantTicks)
	assert.True(t, ok)
	assert.InDelta(t, 0.5*(b0.pos(1)+b1.pos(1)), ticks[1].Value, 1e-9) // centered
	assert.Equal(t, "B", ticks[1].Label)
	assert.Equal(t, 3, len(b0.Median))
}

// TestBoxSharedCategories tests that the categories of series of boxes
// next to each other all have ticks, when a series has no values for some
// of them (e.g., NaN medians at the start).
func TestBoxSharedCategories(t *testing.T) {
	plt := plot.New()
	labels := tensor.NewStringFromValues("A", "B", "C")
	nan := math.NaN()
	a := NewBox(plt, plot.Data{plot.X: labels, plot.Y: plot.Values{2, 3, 4}, plot.U: plot.Values{1, 2, 3}, plot.V: plot.Values{3, 4, 5}})
	b := NewBox(plt, plot.Data{plot.X: labels, plot.Y: plot.Values{nan, nan, 5}, plot.U: plot.Values{nan, nan, 4}, plot.V: plot.Values{nan, nan, 6}})
	for i, bx := range []*Box{a, b} {
		bx.Styler(func(s *plot.Style) {
			s.Width.Width = 0.4
			s.Width.Offset = 0.8 + 0.4*float64(i)
		})
	}
	plt.RenderImage()
	ticks := plt.X.Ticker.(plot.ConstantTicks)
	assert.Equal(t, 3, len(ticks), "the last series has all of the categories")
	assert.Equal(t, "A", ticks[0].Label)
}
