You can graphically **Plot** data using the [[doc:plot]] package ([[doc:plot.Data]], [[doc:plot.PlotStyle]]). See [[plot editor]] for interactive customization of plots.

You can plot a [[vector]]:

```Goal
plt := lab.NewPlot(b)
plots.NewLine(plt, #rand(10)#)
```

You can plot multiple vectors:

```Goal
plt := lab.NewPlot(b)
plots.NewLine(plt, #rand(10)#)
plots.NewLine(plt, #-rand(10)#)
```

## Styles

You can style a plot line:

```Goal
plt := lab.NewPlot(b)
# x := rand(10)
plot.Styler(x, func(s *plot.Style) {
    s.Line.Color = colors.Scheme.Success.Base
})
plots.NewLine(plt, x)
```


<!--- TODO: s.Plot.Title = "My Plot" // overall Plot styles -->
<!--- plots.NewLine(plt, plot.Data{plot.X: xd, plot.Y: yd, plot.Low: low, plot.High: high}) -->
    
    
### Tensor metadata

Styler functions can be attached directly to a `tensor.Tensor` via its metadata, and the `Plotter` elements will automatically grab these functions from any data source that has such metadata set. This allows the data generator to directly set default styling parameters, which can always be overridden later by adding more styler functions. Tying the plot styling directly to the source data allows all of the relevant logic to be put in one place, instead of spreading this logic across different places in the code.

Here is an example of how this works:

```Goal
tx, ty := tensor.NewFloat64(21), tensor.NewFloat64(21)
for i := range 21 {
	tx.SetFloat1D(float64(i*5), i)
	ty.SetFloat1D(50.0+40*math.Sin((float64(i)/8)*math.Pi), i)
}
// attach stylers to the Y axis data: that is where plotter looks for it
plot.SetStyler(ty, func(s *plot.Style) {
	s.Plot.Title = "Test Line"
	s.Plot.XAxis.Label = "X Axis"
	s.Plot.YAxisLabel = "Y Axis"
	s.Plot.Scale = 2
	s.Plot.XAxis.Range.SetMax(105)
	s.Plot.SetLinesOn(plot.On).SetPointsOn(plot.On)
	s.Line.Color = colors.Uniform(colors.Red)
	s.Point.Color = colors.Uniform(colors.Blue)
	s.Range.SetMin(0).SetMax(100)
})

// somewhere else in the code:

plt := lab.NewPlot(b)
// NewLine automatically gets stylers from ty tensor metadata
plots.NewLine(plt, plot.Data{plot.X: tx, plot.Y: ty})
```

## Missing data: NaN

Consistent with the [[tensor]] convention, missing data points are marked by a `NaN` value (i.e., `math.NaN()`), and are processed appropriately by all plot types.

<!--- TODO: need a goal-defined NaN value.. -->

```Goal
data := #rand(10)#
data.Set(math.NaN(), 0)
data.Set(math.NaN(), 4)
plot.SetStyler(data, func(s *plot.Style) {
	s.Plot.SetPointsOn(plot.On)
})
plt := lab.NewPlot(b)
plots.NewLine(plt, data)
```

## Out-of-Range data

You can set fixed Min / Max ranges to plot, for either axis. By default, the out-of-range data will show up as a break in the plot line, with red X's marking the out-of-range points. This behavior can be adjusted by changing the `OutOfRange` property in the [[doc:plot.PlotStyle]].

```Goal
data := #rand(10)#
plot.SetStyler(data, func(s *plot.Style) {
	s.Range.SetMin(0.2) // sets a FixMin flag to indicate a fixed min has been set.
	s.Range.SetMax(0.8)
	s.Plot.OutOfRange = plot.BreakMark // default
})
plt := lab.NewPlot(b)
plots.NewLine(plt, data)
```

## Plot Types

The following are the builtin standard plot types, in the `plots` package. The examples use this data: a sine wave `y` along `x`, with a standard error `sem` of each value:

```Goal
x, y, sem := make(plot.Values, 21), make(plot.Values, 21), make(plot.Values, 21)
for i := range 21 {
	x[i] = float64(i)
	y[i] = math.Sin(float64(i) / 3)
	sem[i] = 0.15 + 0.1*math.Cos(float64(i)/4)
}
```

## 1D and 2D XY Data

### XY

`XY` is the workhorse standard Plotter, taking at least `X` and `Y` inputs, and plotting lines and / or points at each X, Y point (`plots.NewLine`, `plots.NewScatter`, and `plots.NewPointLine` set which ones by default). 

Optionally, a `Size` input can be provided, which sets the size of the points, for a bubble plot, and a `Color` input sets the colors of the points via the color map of the plot (see Colors below). The `Point.ColorFunc` and `Point.FillFunc` styles can also set the color of each point.

```Goal
plt := lab.NewPlot(b)
plots.NewScatter(plt, plot.Data{plot.X: x, plot.Y: y, plot.Size: sem}).Styler(func(s *plot.Style) {
	s.Point.Shape = plot.Ring
	s.Point.Size.Pt(8)
})
```

### Bar

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

### ErrorBar

`XErrorBars` and `YErrorBars` take `X`, `Y`, `Low`, and `High` inputs, and draw an `I` shaped error bar at the X, Y coordinate with the error "handles" around it. `Low` and `High` are the deviations below and above the value, and `Low` is the same as `High` if it is not given.

```Goal
plt := lab.NewPlot(b)
data := plot.Data{plot.X: x, plot.Y: y, plot.High: sem}
plots.NewLine(plt, data)
plots.NewYErrorBars(plt, data)
```

### Labels

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

### Box

`Box` takes `Y` (median), `U` (first quartile), and `V` (third quartile) inputs, and optional `Low`, `High` inputs for the ends of the whiskers (e.g., the min and max, or the furthest values within 1.5 times the interquartile range), and draws a box from `U` to `V` filled with `Line.Fill`, with a median line and whiskers with caps drawn in the `Line` style. The boxes are positioned like `Bar`s (`Style.Width` `Offset`, `Stride`, and `Width`), and an optional `X` input provides the category labels, so multiple box plots can be placed next to each other, and outliers can be added with an XY scatter. `XBox` draws horizontal boxes, with the values along the X axis.

```Goal
plt := lab.NewPlot(b)
plots.NewBox(plt, plot.Data{plot.X: tensor.NewStringFromValues("Control", "Test"), plot.Y: plot.Values{2.1, 3.4}, plot.U: plot.Values{1.5, 2.8}, plot.V: plot.Values{2.9, 4.3}, plot.Low: plot.Values{0.2, 1.6}, plot.High: plot.Values{4.0, 5.9}})
```

### Band, XBand

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

### Reference lines and spans

`HLine` and `VLine` draw horizontal and vertical reference lines across the whole plot, at the given `Y` or `X` values (e.g., at chance level or zero, or at the time of an event), and `HSpan` and `VSpan` fill spans across the whole plot between pairs of `Low` and `High` values (e.g., an epoch of time). The values are included in the range of their axis. Lines are drawn in the `Line` style, which has a neutral color by default, and spans are filled with the `Line.Fill`, or the `Line.Color` at 15% opacity, with their edges drawn if `Line.On` is `On`.

```Goal
plt := lab.NewPlot(b)
plots.NewVSpan(plt, 6, 12)
plots.NewLine(plt, plot.Data{plot.X: x, plot.Y: y})
plots.NewHLine(plt, 0).Styler(func(s *plot.Style) {
	s.Line.Dashes = []float32{4, 2}
})
plots.NewVLine(plt, 3, 15)
```

In a table plot, the values of a line are those of its column, whatever its role (other than `X`, which is the X axis), so a column with the times of events (and `NaN` elsewhere) can be plotted as `VLine`s.

### Colors

The `Color` data role maps values onto colors, using the color axis of the plot, which is shown in a color bar at the right of the plot. The `ColorAxis` plot style sets the color map (`ColdHot` by default), the range of values mapped onto it (the range of the data by default, where either end can be fixed), and a label for the color bar, which can also be turned off:

```Goal
plt := lab.NewPlot(b)
plots.NewScatter(plt, plot.Data{plot.X: x, plot.Y: y, plot.Color: sem}).Styler(func(s *plot.Style) {
	s.Point.Size.Pt(6)
	s.Plot.ColorAxis.ColorMap = "Viridis"
	s.Plot.ColorAxis.Label = "SEM"
})
```

### Grid

The `Grid` plot style draws grid lines behind the data at the major ticks of the `X` and / or `Y` axes, in its `Line` style (a thin light line by default):

```Goal
plt := lab.NewPlot(b)
plots.NewLine(plt, plot.Data{plot.X: x, plot.Y: y}).Styler(func(s *plot.Style) {
	s.Plot.Grid.X = true
	s.Plot.Grid.Y = true
})
```

### Pie

`Pie` takes a list of `Y` values that are plotted as the size of segments of a circular pie plot.  Y values are automatically normalized for plotting.

TODO: implement, details on mapping, 

## 2D Grid-based

### Heatmap

`Heatmap` draws a grid of cells colored by the values of a 2D tensor in the `Color` role (rows by columns), using the color axis of the plot (see Colors above), with optional `X` and `Y` labels for the columns and rows. Row 0 is at the top by default, as for a matrix or image (`TopZero`), and the cell edges are drawn if `Line.On` is `On`. For a higher-dimensional tensor, such as a table column of matrices, one row of the outer dimension (`Row`, the last one by default) is drawn. Here is the correlation matrix of four random variables:

```Goal
plt := lab.NewPlot(b)
vars := tensor.NewFloat64(100, 4)
for i := range 100 {
	a := rand.NormFloat64()
	vars.Set(a, i, 0)
	vars.Set(a+0.5*rand.NormFloat64(), i, 1)
	vars.Set(-a+rand.NormFloat64(), i, 2)
	vars.Set(rand.NormFloat64(), i, 3)
}
names := tensor.NewStringFromValues("A", "B", "C", "D")
corr := metric.CovarianceMatrix(metric.Correlation, vars)
plots.NewHeatmap(plt, plot.Data{plot.Color: corr, plot.X: names, plot.Y: names}).Styler(func(s *plot.Style) {
	s.Line.On = plot.On
	s.Plot.ColorAxis.Range.SetMin(-1).SetMax(1)
	s.Plot.ColorAxis.Label = "Correlation"
})
```

The `tensorcore.TensorGrid` widget also shows tensors as colored grids, including higher-dimensional ones, interactively in the GUI, while a `Heatmap` is part of a plot, with axes, labels, and other plot elements, and can be saved as SVG or PDF.

### Contour

`Contour` draws the contour lines of the values of a 2D tensor in the `Color` role (rows by columns, as for a `Heatmap`): the lines along which the values are equal to each of a number of levels, colored by their level via the color axis of the plot. `FilledContour` fills the regions between the levels instead. The levels are about `Levels` (10 by default) round values across the range of the values, as for axis ticks, or the given `LevelValues`. Optional `X` and `Y` data are the coordinates of the columns and rows; otherwise column j is at X = j and row i at Y = i, with row 0 at the bottom. The values are interpolated linearly between the grid points, and NaN values leave holes.

Here are the contours of two gaussian peaks, of opposite signs:

```Goal
plt := lab.NewPlot(b)
n := 41
x := tensor.NewFloat64SpacedLinear(tensor.NewFloat64Scalar(-3), tensor.NewFloat64Scalar(3), n, tensor.IncludeEnd)
y := tensor.NewFloat64SpacedLinear(tensor.NewFloat64Scalar(-2), tensor.NewFloat64Scalar(2), n, tensor.IncludeEnd)
z := tensor.NewFloat64(n, n)
for r := range n {
	for c := range n {
		xv, yv := x.Float1D(c), y.Float1D(r)
		z.Set(math.Exp(-xv*xv-yv*yv)-math.Exp(-(xv-1)*(xv-1)-(yv-1)*(yv-1)), r, c)
	}
}
plots.NewContour(plt, plot.Data{plot.Color: z, plot.X: x, plot.Y: y})
```

The filled regions, with the contour lines between them, which are drawn if `Line.On` is `On`, in the `Line.Color`:

```Goal
plt := lab.NewPlot(b)
n := 41
x := tensor.NewFloat64SpacedLinear(tensor.NewFloat64Scalar(-3), tensor.NewFloat64Scalar(3), n, tensor.IncludeEnd)
y := tensor.NewFloat64SpacedLinear(tensor.NewFloat64Scalar(-2), tensor.NewFloat64Scalar(2), n, tensor.IncludeEnd)
z := tensor.NewFloat64(n, n)
for r := range n {
	for c := range n {
		xv, yv := x.Float1D(c), y.Float1D(r)
		z.Set(math.Exp(-xv*xv-yv*yv)-math.Exp(-(xv-1)*(xv-1)-(yv-1)*(yv-1)), r, c)
	}
}
ct := plots.NewFilledContour(plt, plot.Data{plot.Color: z, plot.X: x, plot.Y: y})
ct.Levels = 8
ct.Styler(func(s *plot.Style) {
	s.Line.On = plot.On
	s.Line.Width.Pt(0.5)
	s.Plot.ColorAxis.Label = "Z"
})
```

### Vector

X,Y,U,V

Quiver?

## 3D 

TODO: use math32 3D projection math and you can just take each 3d point and reduce to 2D. For stuff you want to actually be able to use in SVG, it needs to ultimately be 2D, so it makes sense to support basic versions here, including XYZ (points, lines), Bar3D, wireframe.

Could also have a separate plot3d package based on `xyz` that is true 3D for interactive 3D plots of surfaces or things that don't make sense in this more limited 2D world.

# Statistical plots

The `statsplot` package computes the data of common statistical plots from the columns of a [[table]], and returns a table of the results with the plot styles set, which can be plotted directly, or in a [[plot editor]] (whose **Stats** menu makes these plots of its data). The results are also stored in a [[tensorfs]] directory. The values of an optional `Split` column (e.g., the condition of a simulation) are plotted as separate series, in their own columns and colors.

The examples use this table of simulated results: 8 runs of 20 epochs in two conditions (`Cond`), with an error (`Err`) that decreases faster in condition B, and a reaction time (`RT`) in three categories (`Cat`):

```Goal
results := table.New("Results")
epoch := results.AddIntColumn("Epoch")
cond := results.AddStringColumn("Cond")
cat := results.AddStringColumn("Cat")
errs := results.AddFloat64Column("Err")
rts := results.AddFloat64Column("RT")
for ci, cnd := range []string{"A", "B"} {
	for run := range 8 {
		for e := range 20 {
			results.AddRows(1)
			i := results.NumRows() - 1
			epoch.SetInt1D(e, i)
			cond.SetString1D(cnd, i)
			cg := (run + e) % 3
			cat.SetString1D([]string{"Easy", "Medium", "Hard"}[cg], i)
			errs.SetFloat1D(math.Exp(-(0.1+0.1*float64(ci))*float64(e))+0.05*rand.NormFloat64(), i)
			rts.SetFloat1D(300+100*float64(cg)+50*float64(ci)+math.Exp(4+0.5*rand.NormFloat64()), i)
		}
	}
}
statsDir, _ := tensorfs.NewDir("Stats")
```

## Mean and error

`MeanError` plots the mean of the values of columns for each value of the `X` column (e.g., across the runs of a simulation at each epoch), with the error of the mean (`Sem`, `Std`, or `CI95`) as a band or error bars. The means are plotted as bars for an `X` column of strings (categories).

```Goal
me := &statsplot.MeanError{X: "Epoch", Values: []statsplot.Column{"Err"}, Split: "Cond"}
mt, _ := me.Table(statsDir, results)
plt, _ := plot.NewTablePlot(mt)
lab.NewPlotFrom(plt, b)
```

In the result table, the error of each mean is in a column with the `High` role and a `Band` (or `YErrorBars`) plotter, in the same group as the column of the mean: a table plot group makes another plotter for an On column with a different plotter set.

## Histogram

`Histogram` plots the number of values within bins of equal width, which are the same for all of the values of the `Split` column, so they can be compared. `Density` normalizes the counts to a probability density.

```Goal
hs := &statsplot.Histogram{Value: "RT", Split: "Cond", Bins: 20}
ht, _ := hs.Table(statsDir, results)
plt, _ := plot.NewTablePlot(ht)
lab.NewPlotFrom(plt, b)
```

## Box plot

`Box` plots a box plot of the values for each value of the `Category` column, with whiskers out to the most extreme values within `Whisker` times the interquartile range, and the outliers beyond them as points.

```Goal
bx := &statsplot.Box{Category: "Cat", Value: "RT", Split: "Cond", Whisker: 1.5, Outliers: true}
bt, _ := bx.Table(statsDir, results)
plt, _ := plot.NewTablePlot(bt)
lab.NewPlotFrom(plt, b)
```

## Violin plot

`Violin` plots the distribution of the values for each category as a kernel density estimate (see [[histogram]]), mirrored on both sides, with a box plot inside it.

```Goal
vl := &statsplot.Violin{Category: "Cat", Value: "RT", Split: "Cond", Points: 50, Box: true, Whisker: 1.5}
vt, _ := vl.Table(statsDir, results)
plt, _ := plot.NewTablePlot(vt)
lab.NewPlotFrom(plt, b)
```

## In the plot editor

The **Stats** menu of a [[plot editor]] makes these plots of its table, with a dialog of their options, which starts from the columns of the current plot: the X axis, the columns that are on, and the Split column. The new plot is shown in a new tab in Cogent Lab, or a new window.

For the following plot, click on the Stats menu in the toolbar, and select `Mean and error` -- use the default options, and you'll see a nice Band summary of the data.

```Goal
plot.SetStyler(results.Column("Epoch"), func(s *plot.Style) { s.Role = plot.X })
plot.SetStyler(results.Column("Err"), func(s *plot.Style) { s.On = true })
plot.SetStyler(results.Column("Cond"), func(s *plot.Style) {
	s.On = true
	s.Role = plot.Split
})
plotcore.NewEditorWithBar(b, plotcore.InlineBar).SetTable(results)
```

## Plot pages

