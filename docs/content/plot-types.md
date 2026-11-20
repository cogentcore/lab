+++
Categories = ["Plot"]
+++

These are the builtin standard [[plot]] types for `X`, `Y` data, in the `plots` package (see [[plot area]] for plots of the values of 2D grids, and [[plot stats]] for statistical plots). The examples use this data: a sine wave `y` along `x`, with a standard error `sem` of each value:

```Goal
x, y, sem := make(plot.Values, 21), make(plot.Values, 21), make(plot.Values, 21)
for i := range 21 {
	x[i] = float64(i)
	y[i] = math.Sin(float64(i) / 3)
	sem[i] = 0.15 + 0.1*math.Cos(float64(i)/4)
}
```

## XY

`XY` is the workhorse standard Plotter, taking at least `X` and `Y` inputs, and plotting lines and / or points at each X, Y point (`plots.NewLine`, `plots.NewScatter`, and `plots.NewPointLine` set which ones by default). 

Optionally, a `Size` input can be provided, which sets the size of the points, for a bubble plot, and a `Color` input sets the colors of the points via the color map of the plot (see Colors in [[plot]]). The `Point.ColorFunc` and `Point.FillFunc` styles can also set the color of each point.

```Goal
plt := lab.NewPlot(b)
plots.NewScatter(plt, plot.Data{plot.X: x, plot.Y: y, plot.Size: sem}).Styler(func(s *plot.Style) {
	s.Point.Shape = plot.Ring
	s.Point.Size.Pt(8)
})
```

## Bar

`Bar` takes `Y` inputs, and draws bars of corresponding height. An optional `X` input provides labels for the bars, and an optional `High` input plots error bars above each bar, with an optional `Low` input for the error below the top of each bar (as for `ErrorBar`).

```Goal
plt := lab.NewPlot(b)
plots.NewBar(plt, plot.Data{plot.X: tensor.NewStringFromValues("A", "B", "C", "D"), plot.Y: plot.Values{3, 5, 2, 4}, plot.Low: plot.Values{0.3, 0.6, 0.3, 0.4}, plot.High: plot.Values{0.5, 0.8, 0.3, 0.6}})
```

`XBar` draws horizontal bars, with the values along the X axis (as `NewXBar` does, or the `Horizontal` field of a `Bar`):

```Goal
plt := lab.NewPlot(b)
plots.NewXBar(plt, plot.Data{plot.X: tensor.NewStringFromValues("A", "B", "C", "D"), plot.Y: plot.Values{3, 5, 2, 4}, plot.High: plot.Values{0.5, 0.8, 0.3, 0.6}})
```

The `Base` field of a `Bar` sets the bottom of each bar, for floating bars, such as a waterfall plot of a running total:

```Goal
plt := lab.NewPlot(b)
steps := plot.Values{5, 3, -2, 4, -3}
base, hts := make(plot.Values, 5), make(plot.Values, 5)
total := 0.0
for i, s := range steps {
	base[i] = min(total, total+s)
	hts[i] = math.Abs(s)
	total += s
}
plots.NewBar(plt, plot.Data{plot.X: tensor.NewStringFromValues("a", "b", "c", "d", "e"), plot.Y: hts}).Base = base
```

Bars can be stacked on top of each other:

```Goal
plt := lab.NewPlot(b)
b1 := plots.NewBar(plt, plot.Values{3, 5, 2, 4})
plots.NewBar(plt, plot.Values{1, 2, 2, 1}).StackOn(b1)
```

To create a plot with multiple sets of bars next to each other, multiple Bar Plotters are created, with `Style.Width` parameters that have a shared `Stride` and an `Offset` that increments for each set of bars. Table plots (and the [[plot editor]]) do this automatically for multiple bar columns, or a `Split` column:

```Goal
plt := lab.NewPlot(b)
plots.NewBar(plt, plot.Values{3, 5, 2, 4}).Styler(func(s *plot.Style) {
	s.Width.Width = 0.4
	s.Width.Offset = 0.8
})
plots.NewBar(plt, plot.Values{2, 4, 3, 3}).Styler(func(s *plot.Style) {
	s.Width.Width = 0.4
	s.Width.Offset = 1.2
})
```

## ErrorBar

`XErrorBars` and `YErrorBars` take `X`, `Y`, `Low`, and `High` inputs, and draw an `I` shaped error bar at the X, Y coordinate with the error "handles" around it. `Low` and `High` are the deviations below and above the value, and `Low` is the same as `High` if it is not given.

```Goal
plt := lab.NewPlot(b)
data := plot.Data{plot.X: x, plot.Y: y, plot.High: sem}
plots.NewLine(plt, data)
plots.NewYErrorBars(plt, data)
```

## Labels

`Labels` takes `X`, `Y` and `Label` string inputs and plots labels at the given coordinates.

```Goal
plt := lab.NewPlot(b)
data := plot.Data{plot.X: plot.Values{1, 2, 3}, plot.Y: plot.Values{2, 4, 3}, plot.Label: tensor.NewStringFromValues("one", "two", "three")}
plots.NewScatter(plt, data)
plots.NewLabels(plt, data).Styler(func(s *plot.Style) {
	s.Text.Offset.X.Dp(6)
	s.Text.Offset.Y.Dp(-6)
})
```

## Box

`Box` takes `Y` (median), `U` (first quartile), and `V` (third quartile) inputs, and optional `Low`, `High` inputs for the ends of the whiskers (e.g., the min and max, or the furthest values within 1.5 times the interquartile range), and draws a box from `U` to `V` filled with `Line.Fill`, with a median line and whiskers with caps drawn in the `Line` style. The boxes are positioned like `Bar`s (`Style.Width` `Offset`, `Stride`, and `Width`), and an optional `X` input provides the category labels, so multiple box plots can be placed next to each other, and outliers can be added with an XY scatter. `XBox` draws horizontal boxes, with the values along the X axis.

```Goal
plt := lab.NewPlot(b)
plots.NewBox(plt, plot.Data{plot.X: tensor.NewStringFromValues("Control", "Test"), plot.Y: plot.Values{2.1, 3.4}, plot.U: plot.Values{1.5, 2.8}, plot.V: plot.Values{2.9, 4.3}, plot.Low: plot.Values{0.2, 1.6}, plot.High: plot.Values{4.0, 5.9}})
```

## Band, XBand

`Band` fills the area between `Low` and `High` Y values along the X axis, and `XBand` fills the area between `Low` and `High` X values along the Y axis. If the center values (`Y` for `Band`, `X` for `XBand`) are given, `Low` and `High` are deviations below and above them, as for `ErrorBar` (`Low` is `High` if not given), so the same data can be plotted as error bars or as an error band. Otherwise, `Low` and `High` are the edges of the band, and `Low` is 0 if not given, to fill the area below `High`. NaN values break the band into separate parts.

The fill is the `Line.Fill` style if set, and otherwise the `Line.Color` at 30% opacity, so a band and its line can share one color. The edges are drawn only if `Line.On` is `On`.

`Band` along with an XY line is the equivalent of the [matplotlib fill_between](https://matplotlib.org/stable/plot_types/basic/fill_between.html#sphx-glr-plot-types-basic-fill-between-py) plot, for error bands:

```Goal
plt := lab.NewPlot(b)
data := plot.Data{plot.X: x, plot.Y: y, plot.High: sem}
blue := func(s *plot.Style) {
	s.Line.Color = colors.Scheme.Primary.Base
}
plots.NewBand(plt, data).Styler(blue)
plots.NewLine(plt, data).Styler(blue)
```

and stacked areas, with the edges of the bands:

```Goal
plt := lab.NewPlot(b)
lowr, upr := make(plot.Values, 21), make(plot.Values, 21)
for i := range 21 {
	lowr[i] = 2 + y[i]
	upr[i] = lowr[i] + 1 + 0.5*math.Cos(float64(i)/3)
}
plots.NewBand(plt, plot.Data{plot.X: x, plot.High: lowr})
plots.NewBand(plt, plot.Data{plot.X: x, plot.Low: lowr, plot.High: upr})
```

`XBand` is the equivalent of fill_betweenx, and draws the body of a [matplotlib violin plot](https://matplotlib.org/stable/plot_types/stats/violin.html#sphx-glr-plot-types-stats-violin-py): `X` is the position of the violin, `Y` the values, and `High` the density at each value. Here, the violins show a gaussian kernel density of random samples in three groups, with a box plot of the quartiles of each group on top of them, and the outliers beyond the whiskers (1.5 times the interquartile range) as points:

```Goal
plt := lab.NewPlot(b)
med, q1, q3 := make(plot.Values, 3), make(plot.Values, 3), make(plot.Values, 3)
low, high := make(plot.Values, 3), make(plot.Values, 3)
var ox, oy plot.Values
for g, mean := range []float64{2, 5, 3} {
	vals := make([]float64, 100)
	for i := range vals {
		vals[i] = mean + (1+0.5*float64(g))*rand.NormFloat64()
	}
	sort.Float64s(vals)
	q := func(p float64) float64 { return vals[int(p*99+0.5)] }
	med[g], q1[g], q3[g] = q(0.5), q(0.25), q(0.75)
	iqr := q3[g] - q1[g]
	low[g], high[g] = q3[g], q1[g]
	for _, v := range vals {
		if v < q1[g]-1.5*iqr || v > q3[g]+1.5*iqr {
			ox = append(ox, float64(g+1))
			oy = append(oy, v)
		} else {
			low[g] = min(low[g], v)
			high[g] = max(high[g], v)
		}
	}
	vx, vy, vd := make(plot.Values, 50), make(plot.Values, 50), make(plot.Values, 50)
	dmax := 0.0
	for j := range 50 {
		vx[j] = float64(g + 1)
		vy[j] = vals[0] + (vals[99]-vals[0])*float64(j)/49
		for _, v := range vals {
			z := (vy[j] - v) / 0.5
			vd[j] += math.Exp(-0.5 * z * z)
		}
		dmax = max(dmax, vd[j])
	}
	for j := range vd {
		vd[j] *= 0.4 / dmax
	}
	plots.NewXBand(plt, plot.Data{plot.X: vx, plot.Y: vy, plot.High: vd}).Styler(func(s *plot.Style) {
		s.Line.Color = colors.Uniform(colors.Gray)
	})
}
plots.NewBox(plt, plot.Data{plot.X: tensor.NewStringFromValues("A", "B", "C"), plot.Y: med, plot.U: q1, plot.V: q3, plot.Low: low, plot.High: high}).Styler(func(s *plot.Style) {
	s.Width.Width = 0.2
	s.Line.Fill = colors.Scheme.Surface
})
plots.NewScatter(plt, plot.Data{plot.X: ox, plot.Y: oy})
```

## Pie

`Pie` takes a list of `Y` values that are plotted as the size of segments of a circular pie plot.  Y values are automatically normalized for plotting.

TODO: implement, details on mapping, 
