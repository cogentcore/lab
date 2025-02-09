// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"fmt"
	"reflect"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/lab/table"
	"golang.org/x/exp/maps"
)

// NewTablePlot returns a new Plot with all configuration based on given
// [table.Table] set of columns and associated metadata, which must have
// [Stylers] functions set (e.g., [SetStylersTo]) that at least set basic
// table parameters, including:
//   - On: Set the main (typically Role = Y) column On to include in plot.
//   - Role: Set the appropriate [Roles] role for this column (Y, X, etc).
//   - Group: Multiple columns used for a given Plotter type must be grouped
//     together with a common name (typically the name of the main Y axis),
//     e.g., for Low, High error bars, Size, Color, etc. If only one On column,
//     then Group can be empty and all other such columns will be grouped.
//   - Plotter: Determines the type of Plotter element to use, which in turn
//     determines the additional Roles that can be used within a Group.
func NewTablePlot(dt *table.Table) (*Plot, error) {
	nc := len(dt.Columns.Values)
	if nc == 0 {
		return nil, errors.New("plot.NewTablePlot: no columns in data table")
	}
	csty := make([]*Style, nc)
	gps := make(map[string][]int, nc)
	xi := -1 // get the _last_ role = X column -- most specific counter
	var errs []error
	var pstySt Style // overall PlotStyle accumulator
	pstySt.Defaults()
	for ci, cl := range dt.Columns.Values {
		st := &Style{}
		st.Defaults()
		stl := GetStylersFrom(cl)
		if stl != nil {
			stl.Run(st)
		}
		csty[ci] = st
		stl.Run(&pstySt)
		gps[st.Group] = append(gps[st.Group], ci)
		if st.Role == X {
			xi = ci
		}
	}
	psty := pstySt.Plot
	globalX := false
	xidxs := map[int]bool{} // map of all the _unique_ x indexes used
	if psty.XAxis.Column != "" {
		xc := dt.Columns.IndexByKey(psty.XAxis.Column)
		if xc >= 0 {
			xi = xc
			globalX = true
			xidxs[xi] = true
		} else {
			errs = append(errs, errors.New("XAxis.Column name not found: "+psty.XAxis.Column))
		}
	}
	doneGps := map[string]bool{}
	plt := New()
	var legends []Thumbnailer // candidates for legend adding -- only add if > 1
	var legLabels []string
	var barCols []int  // column indexes of bar plots
	var barPlots []int // plotter indexes of bar plots
	for ci, cl := range dt.Columns.Values {
		cnm := dt.Columns.Keys[ci]
		st := csty[ci]
		if !st.On || st.Role == X {
			continue
		}
		lbl := cnm
		if st.Label != "" {
			lbl = st.Label
		}
		gp := st.Group
		if doneGps[gp] {
			continue
		}
		if gp != "" {
			doneGps[gp] = true
		}
		ptyp := "XY"
		if st.Plotter != "" {
			ptyp = string(st.Plotter)
		}
		pt, err := PlotterByType(ptyp)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		data := Data{st.Role: cl}
		gcols := gps[gp]
		gotReq := true
		gotX := -1
		if globalX {
			data[X] = dt.Columns.Values[xi]
			gotX = xi
		}
		for _, rl := range pt.Required {
			if rl == st.Role || (rl == X && globalX) {
				continue
			}
			got := false
			for _, gi := range gcols {
				gst := csty[gi]
				if gst.Role == rl {
					if rl == Y {
						if !gst.On {
							continue
						}
					}
					data[rl] = dt.Columns.Values[gi]
					got = true
					if rl == X {
						gotX = gi // fallthrough so we get the last X
					} else {
						break
					}
				}
			}
			if !got {
				if rl == X && xi >= 0 {
					gotX = xi
					data[rl] = dt.Columns.Values[xi]
				} else {
					err = fmt.Errorf("plot.NewTablePlot: Required Role %q not found in Group %q, Plotter %q not added for Column: %q", rl.String(), gp, ptyp, cnm)
					errs = append(errs, err)
					gotReq = false
				}
			}
		}
		if !gotReq {
			continue
		}
		if gotX >= 0 {
			xidxs[gotX] = true
		}
		for _, rl := range pt.Optional {
			if rl == st.Role { // should not happen
				continue
			}
			for _, gi := range gcols {
				gst := csty[gi]
				if gst.Role == rl {
					data[rl] = dt.Columns.Values[gi]
					break
				}
			}
		}
		pl := pt.New(data)
		if reflectx.IsNil(reflect.ValueOf(pl)) {
			err = fmt.Errorf("plot.NewTablePlot: error in creating plotter type: %q", ptyp)
			errs = append(errs, err)
			continue
		}
		plt.Add(pl)
		if !st.NoLegend {
			if tn, ok := pl.(Thumbnailer); ok {
				legends = append(legends, tn)
				legLabels = append(legLabels, lbl)
			}
		}
		if ptyp == "Bar" {
			barCols = append(barCols, ci)
			barPlots = append(barPlots, len(plt.Plotters)-1)
		}
	}
	if len(legends) > 1 {
		for i, l := range legends {
			plt.Legend.Add(legLabels[i], l)
		}
	}
	if psty.XAxis.Label == "" && len(xidxs) == 1 {
		xi := maps.Keys(xidxs)[0]
		lbl := dt.Columns.Keys[xi]
		if csty[xi].Label != "" {
			lbl = csty[xi].Label
		}
		if len(plt.Plotters) > 0 {
			pl0 := plt.Plotters[0]
			if pl0 != nil {
				pl0.Stylers().Add(func(s *Style) {
					s.Plot.XAxis.Label = lbl
				})
			}
		}
	}
	nbar := len(barCols)
	if nbar > 1 {
		sz := 1.0 / (float64(nbar) + 0.5)
		for bi, bp := range barPlots {
			pl := plt.Plotters[bp]
			pl.Stylers().Add(func(s *Style) {
				s.Width.Stride = 1
				s.Width.Offset = float64(bi) * sz
				s.Width.Width = psty.BarWidth * sz
			})
		}
	}
	return plt, errors.Join(errs...)
}

// todo: bar chart rows, if needed
//
// netn := pl.table.NumRows() * stride
// xc := pl.table.ColumnByIndex(xi)
// vals := make([]string, netn)
// for i, dx := range pl.table.Indexes {
// 	pi := mid + i*stride
// 	if pi < netn && dx < xc.Len() {
// 		vals[pi] = xc.String1D(dx)
// 	}
// }
// plt.NominalX(vals...)

// todo:
// Use string labels for X axis if X is a string
// xc := pl.table.ColumnByIndex(xi)
// if xc.Tensor.IsString() {
// 	xcs := xc.Tensor.(*tensor.String)
// 	vals := make([]string, pl.table.NumRows())
// 	for i, dx := range pl.table.Indexes {
// 		vals[i] = xcs.Values[dx]
// 	}
// 	plt.NominalX(vals...)
// }
