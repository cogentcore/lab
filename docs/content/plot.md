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

In a table plot, a string column with the `Title` role sets the title of the plot from the data: its value in the last row of the table, unless a `Plot.Title` style sets the title. For example, a table with the activity of a simulation over the cycles of a trial can show the name of the trial as its title, which is updated with the data:

```go
plot.SetStyler(dt.Column("TrialName"), func(s *plot.Style) { s.Role = plot.Title })
```

### Colors

The `Color` data role maps values onto colors, using the color axis of the plot, which is shown in a color bar at the right of the plot. The `ColorAxis` plot style sets the color map (`ColdHot` by default), the range of values mapped onto it (the range of the data by default, where either end can be fixed), and a label for the color bar, which can also be turned off:

```Goal
xs, ys, cs := make(plot.Values, 21), make(plot.Values, 21), make(plot.Values, 21)
for i := range 21 {
	xs[i] = float64(i)
	ys[i] = math.Sin(float64(i) / 3)
	cs[i] = 0.15 + 0.1*math.Cos(float64(i)/4)
}
plt := lab.NewPlot(b)
plots.NewScatter(plt, plot.Data{plot.X: xs, plot.Y: ys, plot.Color: cs}).Styler(func(s *plot.Style) {
	s.Point.Size.Pt(6)
	s.Plot.ColorAxis.ColorMap = "Viridis"
	s.Plot.ColorAxis.Label = "C"
})
```

### Grid

The `Grid` plot style draws grid lines behind the data at the major ticks of the `X` and / or `Y` axes, in its `Line` style (a thin light line by default):

```Goal
plt := lab.NewPlot(b)
plots.NewLine(plt, plot.Data{plot.X: xs, plot.Y: ys}).Styler(func(s *plot.Style) {
	s.Plot.Grid.X = true
	s.Plot.Grid.Y = true
})
```

### Reference lines and spans

`HLine` and `VLine` draw horizontal and vertical reference lines across the whole plot, at the given `Y` or `X` values (e.g., at chance level or zero, or at the time of an event), and `HSpan` and `VSpan` fill spans across the whole plot between pairs of `Low` and `High` values (e.g., an epoch of time). The values are included in the range of their axis. Lines are drawn in the `Line` style, which has a neutral color by default, and spans are filled with the `Line.Fill`, or the `Line.Color` at 15% opacity, with their edges drawn if `Line.On` is `On`.

```Goal
plt := lab.NewPlot(b)
plots.NewVSpan(plt, 6, 12)
plots.NewLine(plt, plot.Data{plot.X: xs, plot.Y: ys})
plots.NewHLine(plt, 0).Styler(func(s *plot.Style) {
	s.Line.Dashes = []float32{4, 2}
})
plots.NewVLine(plt, 3, 15)
```

In a table plot, the values of a line are those of its column, whatever its role (other than `X`, which is the X axis), so a column with the times of events (and `NaN` elsewhere) can be plotted as `VLine`s.

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

## Plot types

The builtin plot types, in the `plots` package, are described in these pages:

* [[Plot types]]: plots of `X`, `Y` data (lines, points, bars, error bars, labels, box plots, bands, and pie charts).
* [[Plot area]]: plots of the values of a 2D grid (heatmaps and contours).
* [[Plot stats]]: statistical plots of the columns of a [[table]] (mean and error, histograms, box and violin plots), computed by the `statsplot` package.

## Plot pages
