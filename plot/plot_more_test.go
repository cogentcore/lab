// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"encoding"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/enums"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/paint"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPlotter is a minimal [Plotter] that draws a point at each X, Y
// data value, used to test the plot package without the plots package.
type testPlotter struct {
	X, Y    Values
	Style   Style
	stylers Stylers
	pX, pY  []float32
}

func newTestPlotter(plt *Plot, data Data) Plotter {
	tp := &testPlotter{}
	if err := tp.SetData(data); err != nil {
		return nil
	}
	return tp
}

func (tp *testPlotter) SetData(data any) error {
	dt, err := DataOrValuer(data, Y)
	if err != nil {
		return err
	}
	tp.Y = MustCopyRole(dt, Y)
	if len(tp.Y) == 0 {
		return ErrNoData
	}
	tp.X = CopyRole(dt, X)
	if tp.X == nil {
		tp.X = make(Values, len(tp.Y))
		for i := range tp.X {
			tp.X[i] = float64(i)
		}
	}
	tp.stylers = GetStylersFromData(dt, Y)
	return nil
}

func (tp *testPlotter) Data() (Data, []float32, []float32) {
	return Data{X: tp.X, Y: tp.Y}, tp.pX, tp.pY
}

func (tp *testPlotter) Stylers() *Stylers { return &tp.stylers }

func (tp *testPlotter) ApplyStyle(ps *PlotStyle, idx int) {
	tp.Style.Defaults()
	tp.Style.Point.SpacedColor(idx)
	tp.Style.Line.SpacedColor(idx)
	ps.SetElementStyle(&tp.Style)
	tp.stylers.Run(&tp.Style)
}

func (tp *testPlotter) UpdateRange(plt *Plot) {
	Range(tp.X, &plt.X.Range)
	yax := &plt.Y
	if tp.Style.RightY {
		yax = &plt.YR
	}
	Range(tp.Y, &yax.Range)
	RangeLogic(plt.Style.OutOfRange, &yax.Range, &tp.Style.Range)
}

func (tp *testPlotter) Plot(plt *Plot) {
	tp.pX = PlotX(plt, tp.X)
	if tp.Style.RightY {
		tp.pY = PlotYR(plt, tp.Y)
	} else {
		tp.pY = PlotY(plt, tp.Y)
	}
	pc := plt.Painter
	if len(tp.pX) > 1 {
		tp.Style.Line.Draw(plt, math32.Vec2(tp.pX[0], tp.pY[0]), math32.Vec2(tp.pX[len(tp.pX)-1], tp.pY[len(tp.pY)-1]))
	}
	if !tp.Style.Point.SetStroke(plt) {
		return
	}
	for i := range tp.pX {
		tp.Style.Point.SetColorIndex(pc, i)
		tp.Style.Point.DrawShape(pc, math32.Vec2(tp.pX[i], tp.pY[i]))
	}
}

func (tp *testPlotter) Thumbnail(plt *Plot) {
	ptb := plt.CurBounds()
	midY := 0.5 * float32(ptb.Min.Y+ptb.Max.Y)
	tp.Style.Line.Draw(plt, math32.Vec2(float32(ptb.Min.X), midY), math32.Vec2(float32(ptb.Max.X), midY))
}

func init() {
	RegisterPlotter("TestPoints", "draws points for testing", []Roles{Y}, []Roles{X}, newTestPlotter)
}

// TestScales tests the axis scale normalizers.
func TestScales(t *testing.T) {
	assert.Equal(t, 0.25, LinearScale{}.Normalize(0, 4, 1))
	assert.InDelta(t, 0.5, LogScale{}.Normalize(1, 100, 10), 1.0e-12)
	assert.Panics(t, func() { LogScale{}.Normalize(0, 100, 10) })
	assert.Equal(t, 0.75, InvertedScale{LinearScale{}}.Normalize(0, 4, 1))

	assert.Equal(t, LinearScale{}, Linear.Normalizer())
	assert.Equal(t, LogScale{}, Log.Normalizer())
	assert.Equal(t, InvertedScale{LinearScale{}}, InverseLinear.Normalizer())
	assert.Equal(t, InvertedScale{LogScale{}}, InverseLog.Normalizer())
	assert.Equal(t, LinearScale{}, AxisScales(100).Normalizer())

	var ax Axis
	ax.Defaults(math32.X)
	ax.Range.Set(10, 20)
	assert.Equal(t, 0.5, ax.Norm(15))
	ax.Style.Scale = InverseLinear
	ax.drawConfig()
	assert.Equal(t, 0.25, ax.Norm(17.5))

	var va VirtualAxis
	va.Range.Set(1, 1000)
	va.Style.Scale = Log
	va.drawConfig()
	assert.InDelta(t, 2.0/3.0, va.Norm(100), 1.0e-12)
}

// majorTicks returns the values and labels of the major ticks,
// and the number of minor ticks.
func majorTicks(ticks []Tick) (vals []float64, lbls []string, nminor int) {
	for _, tk := range ticks {
		if tk.IsMinor() {
			nminor++
			continue
		}
		vals = append(vals, tk.Value)
		lbls = append(lbls, tk.Label)
	}
	return
}

// TestDefaultTicks tests the default tick marks for known ranges,
// including the minor ticks between them.
func TestDefaultTicks(t *testing.T) {
	vals, lbls, nminor := majorTicks(DefaultTicks{}.Ticks(-3.7, 12.2, 5))
	assert.Equal(t, []float64{-3, 0, 3, 6, 9, 12}, vals)
	assert.Equal(t, []string{"-3", "0", "3", "6", "9", "12"}, lbls)
	assert.Equal(t, 10, nminor) // every 1, from -3 to 12, plus none outside

	vals, lbls, nminor = majorTicks(DefaultTicks{}.Ticks(1e6, 5e7, 5))
	assert.Equal(t, []float64{1e7, 2e7, 3e7, 4e7, 5e7}, vals)
	assert.Equal(t, "1e+07", lbls[0])
	assert.Equal(t, 20, nminor) // every 2e6 from 2e6, excluding majors

	// all ticks are within range
	for _, tk := range (DefaultTicks{}).Ticks(0.13, 0.87, 5) {
		assert.GreaterOrEqual(t, tk.Value, 0.13-1e-9)
		assert.LessOrEqual(t, tk.Value, 0.87+1e-9)
	}

	assert.Nil(t, DefaultTicks{}.Ticks(0, 1, 1))
	assert.Panics(t, func() { DefaultTicks{}.Ticks(1, 1, 5) })
}

// TestLogTicks tests log-scale ticks, with labels on the powers of 10,
// including ranges that start below 1.
func TestLogTicks(t *testing.T) {
	vals, lbls, nminor := majorTicks(LogTicks{Prec: -1}.Ticks(1, 1000, 5))
	assert.Equal(t, []float64{1, 10, 100, 1000}, vals)
	assert.Equal(t, []string{"1", "10", "100", "1000"}, lbls)
	assert.Equal(t, 3*9, nminor) // includes a minor tick at each labeled power of 10

	vals, _, _ = majorTicks(LogTicks{Prec: 2}.Ticks(0.05, 5, 5))
	require.NotEmpty(t, vals)
	assert.LessOrEqual(t, vals[0], 0.05, "first tick must cover the minimum")
	assert.GreaterOrEqual(t, vals[len(vals)-1], 5.0)

	assert.Nil(t, LogTicks{}.Ticks(1, 10, 1))
	assert.Panics(t, func() { LogTicks{}.Ticks(0, 10, 5) })
}

// TestConstantTimeTicks tests ConstantTicks and TimeTicks.
func TestConstantTimeTicks(t *testing.T) {
	ct := ConstantTicks{{Value: 1, Label: "one"}, {Value: 2}}
	assert.Equal(t, []Tick(ct), ct.Ticks(0, 10, 5))

	tt := TimeTicks{Format: "2006-01-02"}
	day := 24 * 3600.0
	ticks := tt.Ticks(0, 10*day, 5)
	vals, lbls, _ := majorTicks(ticks)
	require.NotEmpty(t, vals)
	for i, v := range vals {
		assert.Equal(t, time.Unix(int64(v), 0).UTC().Format("2006-01-02"), lbls[i])
	}
	assert.Equal(t, "1970-01-01", lbls[0])
	assert.Nil(t, tt.Ticks(0, 1, 1))

	ticks = TimeTicks{Time: UnixTimeIn(time.UTC)}.Ticks(0, 100, 3)
	_, lbls, _ = majorTicks(ticks)
	assert.Equal(t, "1970-01-01T00:00:00Z", lbls[0])
}

// TestAxisRange tests SanitizeRange with and without AutoRescale,
// and SetTickLabel.
func TestAxisRange(t *testing.T) {
	var ax Axis
	ax.Defaults(math32.Y)
	assert.Equal(t, float32(-90), ax.Label.Style.Rotation)
	ax.Range.Set(0.13, 0.87)
	ax.SanitizeRange()
	assert.Equal(t, minmax.F64{Min: 0.13, Max: 0.87}, ax.Range)
	ax.AutoRescale = true
	ax.Ticker = ConstantTicks{{Value: 0}, {Value: 1}}
	ax.SanitizeRange()
	assert.Equal(t, minmax.F64{Min: 0, Max: 1}, ax.Range)

	ax.ticks = []Tick{{Value: 0, Label: "a"}}
	ax.SetTickLabel(0, "b")
	ax.SetTickLabel(3, "c") // out of range: ignored
	assert.Equal(t, "b", ax.ticks[0].Label)

	var yr Axis
	yr.RightY = true
	yr.Defaults(math32.Y)
	assert.NotEqual(t, ax.Style.TickText.Align, yr.Style.TickText.Align)
}

// TestDataHelpers tests the data checking, copying and range functions.
func TestDataHelpers(t *testing.T) {
	vs := Values{1, 2, 3}
	d, err := DataOrValuer(vs, Y)
	require.NoError(t, err)
	assert.Equal(t, Data{Y: vs}, d)
	xy := NewXY(vs, vs)
	d, err = DataOrValuer(xy, Y)
	require.NoError(t, err)
	assert.Equal(t, xy, d)
	d, err = DataOrValuer(&xy, Y)
	require.NoError(t, err)
	assert.Equal(t, xy, d)
	_, err = DataOrValuer(3, Y)
	assert.Error(t, err)
	assert.Equal(t, Data{Y: vs}, NewY(vs))

	assert.ErrorIs(t, CheckFloats(1, math.Inf(-1)), ErrInfinity)
	assert.ErrorIs(t, CheckFloats(), ErrNoData)
	assert.NoError(t, CheckFloats(1, math.NaN()))
	assert.True(t, CheckNaNs(1, math.NaN()))
	assert.False(t, CheckNaNs(1, 2))

	var rng minmax.F64
	rng.SetInfinity()
	Range(Values{3, math.NaN(), -1, 2}, &rng)
	assert.Equal(t, minmax.F64{Min: -1, Max: 3}, rng)

	assert.NoError(t, NewXY(vs, vs).CheckLengths())
	assert.Error(t, NewXY(vs, Values{1}).CheckLengths())

	assert.Equal(t, "2.5", Values{2.5}.String1D(0))
	lb := Labels{"a", "b"}
	assert.Equal(t, 2, lb.Len())
	assert.Equal(t, 0.0, lb.Float1D(1))
	assert.Equal(t, "b", lb.String1D(1))

	cp, err := CopyValues(tensor.NewFloat64FromValues(1, 2))
	require.NoError(t, err)
	assert.Equal(t, Values{1, 2}, cp)
	_, err = CopyValues(Values{1, math.Inf(1)})
	assert.ErrorIs(t, err, ErrInfinity)
	var nilv *tensor.Float64
	_, err = CopyValues(nilv)
	assert.ErrorIs(t, err, ErrNoData)

	data := Data{Y: vs, Label: tensor.NewStringFromValues("x", "y", "z")}
	assert.Equal(t, vs, MustCopyRole(data, Y))
	assert.Nil(t, MustCopyRole(data, X))
	assert.Equal(t, vs, CopyRole(data, Y))
	assert.Nil(t, CopyRole(data, X))
	assert.Equal(t, Labels{"x", "y", "z"}, CopyRoleLabels(data, Label))
	assert.Nil(t, CopyRoleLabels(data, X))
}

// TestRangeLogic tests the stretch, clamp and fixed range logic.
func TestRangeLogic(t *testing.T) {
	sty := minmax.Range64{}
	sty.SetMin(0).SetMax(10)

	rng := minmax.F64{Min: 2, Max: 8}
	assert.True(t, RangeLogic(Stretch, &rng, &sty))
	assert.Equal(t, minmax.F64{Min: 0, Max: 10}, rng)

	rng = minmax.F64{Min: -2, Max: 12}
	assert.True(t, RangeLogic(Stretch, &rng, &sty)) // stretch never shrinks
	assert.Equal(t, minmax.F64{Min: -2, Max: 12}, rng)

	rng = minmax.F64{Min: -2, Max: 8}
	assert.False(t, RangeLogic(Mark, &rng, &sty))
	assert.Equal(t, minmax.F64{Min: 0, Max: 10}, rng)

	rng = minmax.F64{Min: 2, Max: 12}
	assert.False(t, RangeSet(&rng, &sty))
	rng = minmax.F64{Min: 2, Max: 8}
	assert.True(t, RangeSet(&rng, &sty))

	free := minmax.Range64{}
	rng = minmax.F64{Min: 2, Max: 8}
	RangeClamp(&rng, &free)
	assert.Equal(t, minmax.F64{Min: 2, Max: 8}, rng)

	assert.True(t, Mark.HasMark())
	assert.True(t, BreakMark.HasMark())
	assert.False(t, Break.HasMark())
	assert.True(t, Break.HasBreak())
	assert.True(t, BreakMark.HasBreak())
	assert.False(t, Stretch.HasBreak())
}

// TestStylers tests adding, running and storing stylers on data.
func TestStylers(t *testing.T) {
	var st Stylers
	st.Add(func(s *Style) { s.Label = "a" })
	st.Add(func(s *Style) { s.Label += "b" })
	s := NewStyle()
	st.Run(s)
	assert.Equal(t, "ab", s.Label)
	var nst *Stylers
	nst.Run(s) // nil is a no-op

	ps := &PlotStyle{}
	ps.Defaults()
	ps.LinesOn = Off
	ps.PointsOn = On
	ps.BarWidth = 0.5
	es := st.NewStyle(ps)
	assert.Equal(t, "ab", es.Label)
	assert.Equal(t, Off, es.Line.On)
	assert.Equal(t, On, es.Point.On)
	assert.Equal(t, 0.5, es.Width.Width)

	y := tensor.NewFloat64(3)
	assert.Nil(t, GetStylers(y))
	SetFirstStyler(y, func(s *Style) { s.Label = "first" })
	SetFirstStyler(y, func(s *Style) { s.Label = "second" }) // ignored
	Styler(y, func(s *Style) { s.Label += "+" })
	assert.Len(t, GetStylers(y), 2)
	s = NewStyle()
	ys := GetStylers(y)
	ys.Run(s)
	assert.Equal(t, "first+", s.Label)

	SetStyler(y, func(s *Style) { s.Role = Y })
	x := tensor.NewFloat64(3)
	SetStyler(x, func(s *Style) { s.Role = X })
	all := GetStylersFromData(Data{X: x, Y: y, Size: tensor.NewFloat64(3)}, X, Y, Size, Label)
	assert.Len(t, all, 2)

	bs := BasicStylers()
	s = NewStyle()
	bs.Run(s)
	assert.Equal(t, Y, s.Role)
	assert.True(t, s.On)

	dt := table.New()
	dt.AddFloat64Column("X")
	dt.AddFloat64Column("Y")
	SetBasicStylers(dt)
	s = NewStyle()
	st0 := GetStylers(dt.Columns.Values[0])
	st0.Run(s)
	assert.Equal(t, X, s.Role)
	s = NewStyle()
	st1 := GetStylers(dt.Columns.Values[1])
	st1.Run(s)
	assert.Equal(t, Y, s.Role)
}

// TestLinePointStyles tests the spaced colors, fill and on-state logic
// of line and point styles.
func TestLinePointStyles(t *testing.T) {
	var ls LineStyle
	ls.Defaults()
	ls.SpacedColor(2)
	assert.Equal(t, colors.Uniform(colors.Spaced(2)), ls.Color)
	assert.False(t, ls.HasFill())
	ls.SpacedFill(3)
	assert.Equal(t, colors.Uniform(colors.Spaced(3)), ls.Fill)
	assert.True(t, ls.HasFill())
	ls.Fill = colors.Uniform(colors.Transparent)
	assert.False(t, ls.HasFill())

	pt := New()
	pt.Painter = paint.NewPainter(math32.Vec2(100, 100))
	defer func() { pt.Painter = nil }()
	ls.On = Off
	assert.False(t, ls.SetStroke(pt))
	assert.False(t, ls.Draw(pt, math32.Vec2(0, 0), math32.Vec2(1, 1)))
	ls.On = On
	ls.Width.Pt(0)
	assert.False(t, ls.SetStroke(pt))

	var ps PointStyle
	ps.Defaults()
	ps.SpacedColor(1)
	assert.Equal(t, colors.Uniform(colors.Spaced(1)), ps.Color)
	assert.Equal(t, colors.Uniform(colors.Spaced(1)), ps.Fill)
	assert.True(t, ps.IsOn(pt))
	ps.On = Off
	assert.False(t, ps.IsOn(pt))
	assert.False(t, ps.SetStroke(pt))
	ps.On = On
	ps.Size.Pt(0)
	assert.False(t, ps.IsOn(pt))
	ps.Size.Pt(3)

	ps.Shape = Plus
	require.True(t, ps.SetStroke(pt))
	assert.Nil(t, pt.Painter.Fill.Color) // unfilled shape
	red := colors.Uniform(colors.Red)
	ps.ColorFunc = func(i int) image.Image { return red }
	ps.FillFunc = func(i int) image.Image { return red }
	ps.SetColorIndex(pt.Painter, 0)
	assert.Equal(t, red, pt.Painter.Stroke.Color)
	assert.Nil(t, pt.Painter.Fill.Color)
	ps.Shape = Box
	ps.SetColorIndex(pt.Painter, 0)
	assert.Equal(t, red, pt.Painter.Fill.Color)
}

// TestPlotterRegistry tests registering and creating plotters by type name.
func TestPlotterRegistry(t *testing.T) {
	pt, err := PlotterByType("TestPoints")
	require.NoError(t, err)
	assert.Equal(t, []Roles{Y}, pt.Required)
	_, err = PlotterByType("NoSuchPlotter")
	assert.Error(t, err)

	plt := New()
	pl := NewPlotter(plt, "TestPoints", NewY(Values{1, 2}))
	require.NotNil(t, pl)
	assert.Nil(t, NewPlotter(plt, "NoSuchPlotter", NewY(Values{1, 2})))
}

// newTestPlot returns a plot with two test plotters, one on the right Y axis.
func newTestPlot() (*Plot, *testPlotter, *testPlotter) {
	plt := New()
	p1 := newTestPlotter(plt, NewXY(Values{0, 1, 2, 3, 4}, Values{0, 10, 20, 30, 40})).(*testPlotter)
	p2 := newTestPlotter(plt, NewY(Values{5, 4, 3, 2, 1})).(*testPlotter)
	p2.stylers.Add(func(s *Style) {
		s.RightY = true
		s.Label = "Right"
		s.Point.Shape = Triangle
	})
	plt.Add(p1, p2)
	plt.Legend.Add("first", p1)
	plt.Legend.Add("second", p2)
	return plt, p1, p2
}

// TestPlotRender tests rendering a plot, the resulting axis ranges and
// pixel projections, and finding the data closest to a pixel.
func TestPlotRender(t *testing.T) {
	plt, p1, p2 := newTestPlot()
	plt.Title.Text = "Render"
	plt.Style.XAxis.Label = "X Label"
	plt.Style.XAxis.Range.SetMax(8)
	plt.Style.YAxisLabel = "unused"
	plt.Style.PointsOn = On
	img := plt.RenderImage()
	assert.Equal(t, image.Point{640, 480}, img.Bounds().Size())

	assert.Equal(t, "X Label", plt.X.Label.Text)
	assert.Equal(t, "Right", plt.YR.Label.Text)
	assert.True(t, plt.YR.Style.On)
	assert.Equal(t, minmax.F64{Min: 0, Max: 8}, plt.X.Range)
	assert.Equal(t, minmax.F64{Min: 0, Max: 40}, plt.Y.Range)
	assert.Equal(t, minmax.F64{Min: 1, Max: 5}, plt.YR.Range)

	pb := plt.PlotBox
	assert.InDelta(t, pb.Min.X, plt.PX(0), 1.0e-3)
	assert.InDelta(t, pb.Max.X, plt.PX(8), 1.0e-3)
	assert.InDelta(t, pb.Max.Y, plt.PY(0), 1.0e-3)
	assert.InDelta(t, pb.Min.Y, plt.PY(40), 1.0e-3)
	assert.InDelta(t, pb.Min.Y, plt.PYR(5), 1.0e-3)
	assert.Less(t, pb.Min.Y, float32(100)) // title and padding only
	assert.Greater(t, pb.Min.X, float32(20))

	_, px, py := p1.Data()
	require.Len(t, px, 5)
	pl, pi, idx, dist, pix, data, lbl := plt.ClosestDataToPixel(int(px[2]), int(py[2]))
	assert.Equal(t, Plotter(p1), pl)
	assert.Equal(t, 0, pi)
	assert.Equal(t, 2, idx)
	assert.Less(t, dist, float32(1))
	assert.Equal(t, px[2], pix.X)
	assert.Equal(t, 20.0, data[Y].Float1D(idx))
	assert.Equal(t, "first", lbl)

	_, px, py = p2.Data()
	pl, pi, idx, _, _, _, lbl = plt.ClosestDataToPixel(int(px[4]), int(py[4]))
	assert.Equal(t, Plotter(p2), pl)
	assert.Equal(t, 1, pi)
	assert.Equal(t, 4, idx)
	assert.Equal(t, "second", lbl)
	assert.Equal(t, "", plt.Legend.LegendForPlotter(&testPlotter{}))
}

// TestPanZoom tests that pan and zoom scale and offset the axis ranges.
func TestPanZoom(t *testing.T) {
	plt, _, _ := newTestPlot()
	plt.PanZoom.XScale = 2
	plt.PanZoom.XOffset = 1
	plt.PanZoom.YScale = 0.5
	plt.PanZoom.YOffset = -1
	plt.applyStyle()
	assert.Equal(t, minmax.F64{Min: 1, Max: 9}, plt.X.Range)
	assert.Equal(t, minmax.F64{Min: -1, Max: 19}, plt.Y.Range)
	assert.Equal(t, minmax.F64{Min: -0.5, Max: 1.5}, plt.YR.Range)
}

// TestNominalHideAxes tests nominal axis labels and hiding axes.
func TestNominalHideAxes(t *testing.T) {
	plt, _, _ := newTestPlot()
	plt.NominalX("a", "b", "c")
	assert.Equal(t, []Tick{{0, "a"}, {1, "b"}, {2, "c"}}, plt.X.Ticker.Ticks(0, 1, 5))
	plt.NominalY("y0", "y1")
	assert.Equal(t, []Tick{{0, "y0"}, {1, "y1"}}, plt.Y.Ticker.Ticks(0, 1, 5))
	assert.Equal(t, float32(0), plt.X.Style.TickLength.Value)

	plt.HideAxes()
	assert.Empty(t, plt.X.Ticker.Ticks(0, 1, 5))
	assert.Empty(t, plt.Y.Ticker.Ticks(0, 1, 5))
	assert.Empty(t, plt.YR.Ticker.Ticks(0, 1, 5))
	assert.Equal(t, float32(0), plt.YR.Style.Line.Width.Value)

	plt.SetSize(image.Point{400, 200})
	assert.Equal(t, image.Point{300, 150}, plt.DataBox())
	img := plt.RenderImage()
	assert.Equal(t, image.Point{400, 200}, img.Bounds().Size())
}

// TestRenderVariants tests rendering with the axis options that change
// the layout: rotated X tick labels, a log Y axis, a legend at the
// bottom left, and every point shape.
func TestRenderVariants(t *testing.T) {
	plt := New()
	var pls []*testPlotter
	for si, sh := range ShapesValues() {
		pl := newTestPlotter(plt, NewXY(Values{1, 10, 100}, Values{float64(si + 1), float64(si + 10), float64(si + 100)})).(*testPlotter)
		pl.stylers.Add(func(s *Style) {
			s.Point.On = On
			s.Point.Shape = sh
		})
		plt.Add(pl)
		plt.Legend.Add(sh.String(), pl)
		pls = append(pls, pl)
	}
	plt.Style.XAxis.Rotation = -45
	plt.Style.Legend.Position.Top = false
	plt.Style.Legend.Position.Left = true
	plt.Y.Ticker = LogTicks{}
	plt.Style.Axis.Scale = Log
	plt.RenderImage()
	assert.Equal(t, LogScale{}, plt.Y.Scale)

	// the higher value is drawn higher on the log axis
	for _, pl := range pls {
		_, _, py := pl.Data()
		require.Len(t, py, 3)
		assert.Greater(t, py[0], py[1])
		assert.Greater(t, py[1], py[2])
	}

	// positive rotation and no title
	plt.Style.XAxis.Rotation = 30
	plt.RenderImage()

	tx := &plt.X.TickText
	tx.Style.Rotation = -45
	tx.Text = "label"
	plt.Painter = paint.NewPainter(math32.Vec2(100, 100))
	plt.TextShaper = plotShaper
	tx.Config(plt)
	plt.Painter = nil
	pos, left, right, height := tx.TickPosX(50, 10)
	assert.Less(t, left, float32(0)) // text ends at the tick
	assert.Less(t, right, -left)
	assert.Greater(t, height, float32(0))
	assert.Less(t, pos.X, float32(50))
}

// TestSaveFiles tests saving the plot as an image and an SVG file.
func TestSaveFiles(t *testing.T) {
	plt, _, _ := newTestPlot()
	dir := t.TempDir()
	png := filepath.Join(dir, "plot.png")
	require.NoError(t, plt.SaveImage(png))
	fi, err := os.Stat(png)
	require.NoError(t, err)
	assert.Greater(t, fi.Size(), int64(0))

	svg := filepath.Join(dir, "plot.svg")
	require.NoError(t, plt.SaveSVG(svg))
	b, err := os.ReadFile(svg)
	require.NoError(t, err)
	assert.Contains(t, string(b), "<svg")
}

// enumRoundTrip checks that every value of an enum type converts to
// and from its string and text representations.
func enumRoundTrip[E interface {
	~int32
	enums.Enum
	encoding.TextMarshaler
}, P interface {
	*E
	enums.EnumSetter
	encoding.TextUnmarshaler
}](t *testing.T, vals []E) {
	require.NotEmpty(t, vals)
	assert.Len(t, vals[0].Values(), len(vals))
	for _, v := range vals {
		var e E
		require.NoError(t, P(&e).SetString(v.String()))
		assert.Equal(t, v, e)
		txt, err := v.MarshalText()
		require.NoError(t, err)
		e = 0
		require.NoError(t, P(&e).UnmarshalText(txt))
		assert.Equal(t, v, e)
		P(&e).SetInt64(v.Int64())
		assert.Equal(t, v, e)
		assert.NotEmpty(t, v.Desc())
	}
	var e E
	assert.Error(t, P(&e).SetString("not a valid value"))
}

// TestEnums tests the string conversions of the plot enum types.
func TestEnums(t *testing.T) {
	enumRoundTrip(t, AxisScalesValues())
	enumRoundTrip(t, RolesValues())
	enumRoundTrip(t, StepKindValues())
	enumRoundTrip(t, ShapesValues())
	enumRoundTrip(t, DefaultOffOnValues())
	enumRoundTrip(t, OutOfRangeValues())
}

// TestNegativeXOn tests the NegativeX of a line, with the default
// from the plot style.
func TestNegativeXOn(t *testing.T) {
	var ps PlotStyle
	ps.Defaults()
	var ls LineStyle
	ls.Defaults()
	assert.False(t, ls.NegativeXOn(&ps), "off by default")
	ps.NegativeX = true
	assert.True(t, ls.NegativeXOn(&ps), "from the plot")
	ls.NegativeX = Off
	assert.False(t, ls.NegativeXOn(&ps))
	ps.NegativeX = false
	ls.NegativeX = On
	assert.True(t, ls.NegativeXOn(&ps))
}
