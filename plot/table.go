// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"fmt"
	"image"
	"reflect"
	"slices"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/colors"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
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
//
// Returns nil if no valid plot elements were present.
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
		stl := GetStylers(cl)
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

	type pitem struct {
		ptyp string
		pt   *PlotterType
		data Data
		lbl  string
		ci   int
		clr  image.Image // if set, set styler
	}
	var ptrs []*pitem // accumulate in case of grouping

	doneGps := map[string]bool{}
	var split tensor.Values
	nLegends := 0

	for ci, cl := range dt.Columns.Values {
		cnm := dt.Columns.Keys[ci]
		st := csty[ci]
		if !st.On || st.Role == X {
			continue
		}
		if st.Role == Split {
			if split != nil {
				errs = append(errs, errors.New("NewTablePlot: Only 1 Split role can be defined, using the first one"))
			}
			split = cl
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
		for _, rl := range pt.Optional {
			if rl == st.Role || (rl == X && globalX) {
				continue
			}
			got := false
			for _, gi := range gcols {
				gst := csty[gi]
				if gst.Role == rl {
					data[rl] = dt.Columns.Values[gi]
					got = true
					if rl == X {
						gotX = gi // fallthrough so we get the last X
					} else {
						break
					}
				}
			}
			if !got && rl == X && xi >= 0 {
				gotX = xi
				data[rl] = dt.Columns.Values[xi]
			}
		}
		if gotX >= 0 {
			xidxs[gotX] = true
		}
		ptrs = append(ptrs, &pitem{pt: pt, data: data, lbl: lbl, ci: ci})
		if !st.NoLegend {
			nLegends++
		}
	}

	if len(ptrs) == 0 {
		return nil, errors.Join(errs...)
	}

	plt := New()

	// do splits here, make a new list of ptrs
	if split != nil {
		spnm := metadata.Name(split)
		if spnm == "" {
			spnm = "0"
		}

		dir := errors.Log1(tensorfs.NewDir("TablePlot"))
		err := stats.Groups(dir, split)
		if err != nil {
			errs = append(errs, err) // todo maybe bail here
		}
		sdir := dir.Dir("Groups").Dir(spnm)
		gps := errors.Log1(sdir.Values())

		// generate tensor.Rows indexed views of the original data
		// for each unique element in pt.data.* -- the x axis is shared
		// so we need a map to just do this once.
		// [gp][pt.data.*]sliced
		subd := make(map[tensor.Tensor]map[tensor.Values]*tensor.Rows)
		for _, gp := range gps {
			sv := make(map[tensor.Values]*tensor.Rows)
			idxs := slices.Clone(gp.(*tensor.Int).Values)
			for _, pt := range ptrs {
				for _, dd := range pt.data {
					dv := dd.(tensor.Values)
					rv, ok := sv[dv]
					if !ok {
						rv = tensor.NewRows(dv, idxs...)
					}
					sv[dv] = rv
				}
			}
			subd[gp] = sv
		}

		// now go in plotter item order, then groups within, and make the new
		// plot items
		nptrs := make([]*pitem, 0, len(gps)*len(ptrs))
		nLegends = len(gps) * nLegends
		idx := 0
		for _, pt := range ptrs {
			for _, gp := range gps {
				nd := Data{}
				for rl, dd := range pt.data {
					dv := dd.(tensor.Values)
					rv := subd[gp][dv]
					nd[rl] = rv
				}
				npt := *pt
				pt.clr = colors.Uniform(colors.Spaced(idx))
				npt.data = nd
				npt.lbl = metadata.Name(gp) + " " + pt.lbl
				nptrs = append(nptrs, &npt)
				idx++
			}
		}
		ptrs = nptrs
	}

	var barCols []int  // column indexes of bar plots
	var barPlots []int // plotter indexes of bar plots
	for _, pt := range ptrs {
		pl := pt.pt.New(plt, pt.data)
		if reflectx.IsNil(reflect.ValueOf(pl)) {
			err := fmt.Errorf("plot.NewTablePlot: error in creating plotter type: %q", pt.ptyp)
			errs = append(errs, err)
			continue
		}
		if pt.clr != nil {
			pl.Stylers().Add(func(s *Style) {
				s.Line.Color = pt.clr
				s.Point.Color = pt.clr
				s.Point.Fill = pt.clr
			})
		}
		plt.Add(pl)
		st := csty[pt.ci]
		if !st.NoLegend && nLegends > 1 {
			if tn, ok := pl.(Thumbnailer); ok {
				plt.Legend.Add(pt.lbl, tn)
			}
		}
		if pt.ptyp == "Bar" {
			barCols = append(barCols, pt.ci)
			barPlots = append(barPlots, len(plt.Plotters)-1)
		}
	}

	// Get XAxis label from actual x axis.
	// todo: probably range from here too.
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

	// Set bar spacing based on total number of bars present.
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
