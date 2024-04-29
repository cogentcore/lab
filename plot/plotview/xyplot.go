// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotview

import (
	"fmt"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/plot"
	"cogentcore.org/core/plot/plots"
	"cogentcore.org/core/styles"
)

// GenPlotXY generates an XY (lines, points) plot, setting Plot variable
func (pl *PlotView) GenPlotXY() {
	plt := plot.New() // todo: not clear how to re-use, due to newtablexynames
	plt.Title.Text = pl.Params.Title
	plt.X.Label.Text = pl.XLabel()
	plt.Y.Label.Text = pl.YLabel()

	plt.Background = colors.Scheme.Surface

	clr := colors.C(colors.Scheme.OnSurface)

	plt.Title.Style.Color = clr
	plt.Legend.TextStyle.Color = clr
	plt.X.Line.Color = clr
	plt.Y.Line.Color = clr
	plt.X.Label.Style.Color = clr
	plt.Y.Label.Style.Color = clr
	plt.X.TickLine.Color = clr
	plt.Y.TickLine.Color = clr
	plt.X.TickText.Style.Color = clr
	plt.Y.TickText.Style.Color = clr

	// process xaxis first
	xi, err := pl.PlotXAxis(plt)
	if err != nil {
		return
	}
	xp := pl.Columns[xi]

	/*
		var lsplit *etable.Splits
		nleg := 1
		if pl.Params.LegendCol != "" {
			_, err = pl.Table.Table.ColIndexTry(pl.Params.LegendCol)
			if err != nil {
				slog.Error("plot.LegendCol", "err", err.Error())
			} else {
				errors.Log(xview.SortStableColNames([]string{pl.Params.LegendCol, xp.Col}, etable.Ascending))
				lsplit = split.GroupBy(xview, []string{pl.Params.LegendCol})
				nleg = max(lsplit.Len(), 1)
			}
		}
	*/
	nleg := 1

	var firstXY *plots.TableXYer
	var strCols []*ColumnParams
	nys := 0
	for _, cp := range pl.Columns {
		if !cp.On {
			continue
		}
		if cp.IsString {
			strCols = append(strCols, cp)
			continue
		}
		if cp.TensorIndex < 0 {
			// yc := pl.Table.Table.ColByName(cp.Col)
			// _, sz := yc.RowCellSize()
			// nys += sz
		} else {
			nys++
		}
		if cp.Range.FixMin {
			plt.Y.Min = math32.Min(plt.Y.Min, float32(cp.Range.Min))
		}
		if cp.Range.FixMax {
			plt.Y.Max = math32.Max(plt.Y.Max, float32(cp.Range.Max))
		}
	}

	if nys == 0 {
		return
	}

	firstXY = nil
	yidx := 0
	for ci, cp := range pl.Columns {
		if !cp.On || cp == xp {
			continue
		}
		if cp.IsString {
			continue
		}
		for li := 0; li < nleg; li++ {
			// lview := xview
			leg := ""
			// if lsplit != nil && len(lsplit.Values) > li {
			// 	leg = lsplit.Values[li][0]
			// 	lview = lsplit.Splits[li]
			// 	_, _, xbreaks, _ = pl.PlotXAxis(plt, lview)
			// }
			// stRow := 0
			nidx := 1
			stidx := cp.TensorIndex
			// if cp.TensorIndex < 0 { // do all
			// 	yc := pl.Table.Table.ColByName(cp.Col)
			// 	_, sz := yc.RowCellSize()
			// 	nidx = sz
			// 	stidx = 0
			// }
			for ii := 0; ii < nidx; ii++ {
				idx := stidx + ii
				// tix := lview.Clone()
				// tix.Indexes = tix.Indexes[stRow:edRow]
				// xy, _ := NewTableXYName(tix, xi, xp.TensorIndex, cp.Col, idx, cp.Range)
				xy := plots.NewTableXYer(pl.Table, xi, ci)
				// if xy == nil {
				// 	continue
				// }
				if firstXY == nil {
					firstXY = xy
				}
				var pts *plots.Scatter
				var lns *plots.Line
				lbl := cp.Label()
				clr := cp.Color
				if leg != "" {
					lbl = leg + " " + lbl
				}
				if nleg > 1 {
					cidx := yidx*nleg + li
					clr = colors.Spaced(cidx)
				}
				if nidx > 1 {
					clr = colors.Spaced(idx)
					lbl = fmt.Sprintf("%s_%02d", lbl, idx)
				}
				if cp.Lines.Or(pl.Params.Lines) && cp.Points.Or(pl.Params.Points) {
					lns, pts, _ = plots.NewLinePoints(xy)
				} else if cp.Points.Or(pl.Params.Points) {
					pts, _ = plots.NewScatter(xy)
				} else {
					lns, _ = plots.NewLine(xy)
				}
				if lns != nil {
					lns.LineStyle.Width.Pt(float32(cp.LineWidth.Or(pl.Params.LineWidth)))
					lns.LineStyle.Color = colors.C(clr)
					lns.NegativeXDraw = pl.Params.NegativeXDraw
					plt.Add(lns)
					// if bi == 0 {
					// 	plt.Legend.Add(lbl, lns)
					// }
				}
				if pts != nil {
					pts.LineStyle.Color = colors.C(clr)
					pts.LineStyle.Width.Pt(float32(cp.LineWidth.Or(pl.Params.LineWidth)))
					pts.PointSize.Pt(float32(cp.PointSize.Or(pl.Params.PointSize)))
					pts.PointShape = cp.PointShape.Or(pl.Params.PointShape)
					plt.Add(pts)
					// if lns == nil && bi == 0 {
					// 	plt.Legend.Add(lbl, pts)
					// }
				}
				// if cp.ErrCol != "" {
				// 	ec := pl.Table.Table.ColIndex(cp.ErrCol)
				// 	if ec >= 0 {
				// 		xy.ErrCol = ec
				// 		eb, _ := plots.NewYErrorBars(xy)
				// 		eb.LineStyle.Color = clr
				// 		plt.Add(eb)
				// 	}
				// }
			}
		}
		yidx++
	}
	/*
			if firstXY != nil && len(strCols) > 0 {
				for _, cp := range strCols {
					xy, _ := NewTableXYName(xview, xi, xp.TensorIndex, cp.Col, cp.TensorIndex, firstXY.YRange)
					xy.LblCol = xy.YCol
					xy.YCol = firstXY.YCol
					xy.YIndex = firstXY.YIndex
					lbls, _ := plots.NewLabels(xy)
					if lbls != nil {
						plt.Add(lbls)
					}
				}
			}

		// Use string labels for X axis if X is a string
		xc := pl.Table.Table.Columns[xi]
		if xc.DataType() == etensor.STRING {
			xcs := xc.(*etensor.String)
			vals := make([]string, pl.Table.Len())
			for i, dx := range pl.Table.Indexes {
				vals[i] = xcs.Values[dx]
			}
			plt.NominalX(vals...)
		}
	*/

	plt.Legend.Top = true
	plt.X.TickText.Style.Rotation = float32(pl.Params.XAxisRot)
	if pl.Params.XAxisRot > 10 {
		plt.X.TickText.Style.Align = styles.Center
		// plt.X.Tick.Label.Style.Align = draw.XRight
	}
	pl.Plot = plt
	if pl.ConfigPlotFunc != nil {
		pl.ConfigPlotFunc()
	}
}