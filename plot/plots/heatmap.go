// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plots

import (
	"fmt"
	"math"
	"strconv"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/tensor"
)

// HeatmapType is used for specifying the type name of a [Heatmap].
const HeatmapType = "Heatmap"

func init() {
	plot.RegisterPlotter(HeatmapType, "draws a grid of cells colored by the values of a 2D tensor in the Color data role (rows by columns, e.g., a correlation matrix), via the color map of the plot, with optional X and Y labels for the columns and rows. For a higher-dimensional tensor (e.g., a table column of matrices), the last row of the outer dimension is drawn.", []plot.Roles{plot.Color}, []plot.Roles{plot.X, plot.Y}, func(plt *plot.Plot, data plot.Data) plot.Plotter {
		return NewHeatmap(plt, data)
	})
}

// Heatmap draws a grid of cells colored by the values of a 2D tensor
// (rows by columns), via the [plot.ColorAxis] of the plot, which is
// shown in a color bar (as imshow, matshow, or pcolormesh do in
// matplotlib). Column j is at X = j, and row i at Y = i, or at the top
// for row 0 if TopZero is set (the default, as for a matrix or image).
// NaN values are not drawn.
//
// The data is a tensor in the Color role: a 2D tensor is used as is,
// and a 1D tensor as one row. For a higher-dimensional tensor, one row
// of the outer dimension (Row, the last one by default) is drawn, with
// its innermost dimension as the columns and the rest as rows, which
// shows the last matrix of a column of matrices in a table, for example.
// Optional X and Y labels (e.g., strings) are the labels of the columns
// and rows, if they have the same number of values.
type Heatmap struct {
	// Values are the values of the cells, row by row.
	Values plot.Values

	// Rows, Columns are the number of rows and columns.
	Rows, Columns int

	// XLabels, YLabels are the optional labels of the columns and rows.
	XLabels, YLabels plot.Labels

	// PX, PY are the pixel coordinates of the cell centers,
	// in the order of the Values.
	PX, PY []float32

	// TopZero draws row 0 at the top, as for a matrix or image.
	// It is true by default.
	TopZero bool

	// Row is the row of the outer dimension of a higher-dimensional
	// tensor that is drawn; -1 (the default) is the last row.
	// It is used by SetData.
	Row int

	// Style is the style for plotting.
	Style plot.Style

	data    plot.Data
	stylers plot.Stylers
}

// NewHeatmap adds a new [Heatmap] plotter to the given plot, for the
// given data: a [plot.Data] with the Color role and the optional X and
// Y labels, or a [tensor.Tensor] with the values (see [Heatmap]).
// Styler functions are obtained from the Color metadata if present.
func NewHeatmap(plt *plot.Plot, data any) *Heatmap {
	hm := &Heatmap{TopZero: true, Row: -1}
	err := hm.SetData(data)
	if err != nil {
		return nil
	}
	hm.Defaults()
	plt.Add(hm)
	return hm
}

func (hm *Heatmap) Defaults() {
	hm.Style.Defaults()
}

// SetData sets the plot data.
func (hm *Heatmap) SetData(data any) error {
	dt, err := plot.DataOrValuer(data, plot.Color)
	if err != nil {
		return err
	}
	hm.data = dt
	return hm.setValues()
}

// setValues sets the values from the data, for the current Row.
func (hm *Heatmap) setValues() error {
	var err error
	hm.Values, hm.Rows, hm.Columns, err = gridValues(HeatmapType, hm.data, hm.Row)
	if err != nil {
		return err
	}
	hm.XLabels = hm.labels(plot.X, hm.Columns)
	hm.YLabels = hm.labels(plot.Y, hm.Rows)
	hm.stylers = plot.GetStylersFromData(hm.data, plot.Color)
	return nil
}

// gridValues returns the values of the grid of the tensor in the Color
// role of the given data, with its number of rows and columns, for a
// plotter of the given type: a 2D tensor is used as is, and a 1D tensor
// as one row. For a higher-dimensional tensor, the given row of the outer
// dimension is used (the last one if it is out of range), with its
// innermost dimension as the columns and the rest as rows.
func gridValues(typ string, data plot.Data, row int) (vals plot.Values, rows, cols int, err error) {
	tsr, ok := data[plot.Color].(tensor.Tensor)
	if !ok {
		return nil, 0, 0, fmt.Errorf("%s: Color data must be a tensor.Tensor", typ)
	}
	sizes := tsr.ShapeSizes()
	nd := len(sizes)
	var cell tensor.Tensor = tsr
	switch {
	case nd == 0 || tsr.Len() == 0:
		return nil, 0, 0, plot.ErrNoData
	case nd == 1:
		rows, cols = 1, sizes[0]
	case nd == 2:
		rows, cols = sizes[0], sizes[1]
	default:
		if row < 0 || row >= sizes[0] {
			row = sizes[0] - 1
		}
		cell = tensor.Cells1D(tsr, row)
		cols = sizes[nd-1]
		rows = cell.Len() / cols
	}
	vals, err = plot.CopyValues(cell)
	return
}

// labels returns the labels for the given role if it has n values.
func (hm *Heatmap) labels(role plot.Roles, n int) plot.Labels {
	lb := plot.CopyRoleLabels(hm.data, role)
	if len(lb) != n {
		return nil
	}
	return lb
}

// Styler adds a style function to set style parameters.
func (hm *Heatmap) Styler(f func(s *plot.Style)) *Heatmap {
	hm.stylers.Add(f)
	return hm
}

func (hm *Heatmap) Stylers() *plot.Stylers { return &hm.stylers }

func (hm *Heatmap) ApplyStyle(ps *plot.PlotStyle, idx int) {
	ps.SetElementStyle(&hm.Style)
	hm.stylers.Run(&hm.Style)
}

func (hm *Heatmap) Data() (data plot.Data, pixX, pixY []float32) {
	pixX = hm.PX
	pixY = hm.PY
	data = plot.Data{plot.Color: hm.Values}
	if hm.XLabels != nil {
		data[plot.X] = hm.XLabels
	}
	if hm.YLabels != nil {
		data[plot.Y] = hm.YLabels
	}
	return
}

// rowY returns the Y position of the given row.
func (hm *Heatmap) rowY(row int) float64 {
	if hm.TopZero {
		return float64(hm.Rows - 1 - row)
	}
	return float64(row)
}

// Plot does the drawing, implementing the plot.Plotter interface.
func (hm *Heatmap) Plot(plt *plot.Plot) {
	pc := plt.Painter
	py := plt.PY
	if hm.Style.RightY {
		py = plt.PYR
	}
	n := len(hm.Values)
	hm.PX = make([]float32, n)
	hm.PY = make([]float32, n)
	for i, v := range hm.Values {
		r, c := i/hm.Columns, i%hm.Columns
		x, y := float64(c), hm.rowY(r)
		hm.PX[i], hm.PY[i] = plt.PX(x), py(y)
		if math.IsNaN(v) {
			continue
		}
		x0, x1 := plt.PX(x-0.5), plt.PX(x+0.5)
		y0, y1 := py(y-0.5), py(y+0.5)
		// overlap the cells by half a pixel to avoid seams
		pc.FillBox(math32.Vec2(min(x0, x1), min(y0, y1)), math32.Vec2(math32.Abs(x1-x0)+0.5, math32.Abs(y1-y0)+0.5), plt.ColorAxis.Uniform(v))
	}
	if hm.Style.Line.On == plot.On && hm.Style.Line.SetStroke(plt) {
		for r := range hm.Rows + 1 {
			y := py(float64(r) - 0.5)
			pc.MoveTo(plt.PX(-0.5), y)
			pc.LineTo(plt.PX(float64(hm.Columns)-0.5), y)
		}
		for c := range hm.Columns + 1 {
			x := plt.PX(float64(c) - 0.5)
			pc.MoveTo(x, py(-0.5))
			pc.LineTo(x, py(float64(hm.Rows)-0.5))
		}
		pc.Draw()
	}
}

// UpdateRange updates the given ranges.
func (hm *Heatmap) UpdateRange(plt *plot.Plot) {
	yax := &plt.Y
	if hm.Style.RightY {
		yax = &plt.YR
	}
	plt.X.Range.FitInRange(minmax.F64{Min: -0.5, Max: float64(hm.Columns) - 0.5})
	yax.Range.FitInRange(minmax.F64{Min: -0.5, Max: float64(hm.Rows) - 0.5})
	plt.ColorAxis.AddRange(hm.Values)

	nt := plt.Style.Axis.NTicks
	col := func(i int) float64 { return float64(i) }
	if hm.XLabels != nil {
		plt.X.Ticker = labelTicks(hm.XLabels, col)
	} else {
		plt.X.Ticker = indexTicks(hm.Columns, nt, col)
	}
	if hm.YLabels != nil {
		yax.Ticker = labelTicks(hm.YLabels, hm.rowY)
	} else {
		yax.Ticker = indexTicks(hm.Rows, nt, hm.rowY)
	}
	plot.RangeLogic(plt.Style.OutOfRange, &plt.X.Range, &plt.Style.XAxis.Range)
	plt.X.DataRange = plt.X.Range
	plot.RangeLogic(plt.Style.OutOfRange, &yax.Range, &hm.Style.Range)
	yax.DataRange = yax.Range
}

// indexTicks returns ticks at about nticks of the n integer indexes,
// labeled with the index, at the positions given by the pos function.
func indexTicks(n, nticks int, pos func(i int) float64) plot.ConstantTicks {
	var ticks plot.ConstantTicks
	for _, t := range (plot.DefaultTicks{}).Ticks(0, float64(n-1), nticks) {
		if t.IsMinor() || t.Value != math.Trunc(t.Value) {
			continue
		}
		i := int(t.Value)
		ticks = append(ticks, plot.Tick{Value: pos(i), Label: strconv.Itoa(i)})
	}
	return ticks
}

// labelTicks returns ticks with the given labels at the positions
// given by the pos function of the label index.
func labelTicks(labels plot.Labels, pos func(i int) float64) plot.ConstantTicks {
	ticks := make(plot.ConstantTicks, len(labels))
	for i, lb := range labels {
		ticks[i] = plot.Tick{Value: pos(i), Label: lb}
	}
	return ticks
}

// Thumbnail returns the thumbnail, implementing the plot.Thumbnailer interface.
func (hm *Heatmap) Thumbnail(plt *plot.Plot) {
	pc := plt.Painter
	ptb := plt.CurBounds()
	n := 4
	w := float32(ptb.Dx()) / float32(n)
	for i := range n {
		clr := plt.ColorAxis.Map.Map((float32(i) + 0.5) / float32(n))
		pc.FillBox(math32.Vec2(float32(ptb.Min.X)+float32(i)*w, float32(ptb.Min.Y)), math32.Vec2(w+0.5, float32(ptb.Dy())), colors.Uniform(clr))
	}
}
