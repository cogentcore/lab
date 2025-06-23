+++
Categories = ["Plots"]
+++

**Plots** allow you to graphically plot data. See [[plot editor]] for interactive customization of plots.

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

## Plot Types

The following are the builtin standard plot types, in the `plots` package:

## 1D and 2D XY Data

### XY

`XY` is the workhorse standard Plotter, taking at least `X` and `Y` inputs, and plotting lines and / or points at each X, Y point. 

Optionally `Size` and / or `Color` inputs can be provided, which apply to the points. Thus, by using a `Point.Shape` of `Ring` or `Circle`, you can create a bubble plot by providing Size and Color data.

### Bar

`Bar` takes `Y` inputs, and draws bars of corresponding height.

An optional `High` input can be provided to also plot error bars above each bar.

To create a plot with multiple error bars, multiple Bar Plotters are created, with `Style.Width` parameters that have a shared `Stride = 1 / number of bars` and `Offset` that increments for each bar added.  The `plots.NewBars` function handles this directly.

### ErrorBar

`XErrorBar` and `YErrorBar` take `X`, `Y`, `Low`, and `High` inputs, and draws an `I` shaped error bar at the X, Y coordinate with the error "handles" around it.

### Labels

`Labels` takes `X`, `Y` and `Labels` string inputs and plots labels at the given coordinates.

### Box

`Box` takes `X`, `Y` (median line), `U`, `V` (box first and 3rd quartile values), and `Low`, `High` (Min, Max) inputs, and renders a box plot with error bars.

### XFill, YFill

`XFill` and `YFill` are used to draw filled regions between pairs of X or Y points, using the `X`, `Y`, and `Low`, `High` values to specify the center point (X, Y) and the region below / left and above / right to fill around that central point.

XFill along with an XY line can be used to draw the equivalent of the [matplotlib fill_between](https://matplotlib.org/stable/plot_types/basic/fill_between.html#sphx-glr-plot-types-basic-fill-between-py) plot.

YFill can be used to draw the equivalent of the [matplotlib violin plot](https://matplotlib.org/stable/plot_types/stats/violin.html#sphx-glr-plot-types-stats-violin-py).

### Pie

`Pie` takes a list of `Y` values that are plotted as the size of segments of a circular pie plot.  Y values are automatically normalized for plotting.

TODO: implement, details on mapping, 

## 2D Grid-based

### ColorGrid

Input = Values and X, Y size

### Contour

??

### Vector

X,Y,U,V

Quiver?

## 3D 

TODO: use math32 3D projection math and you can just take each 3d point and reduce to 2D. For stuff you want to actually be able to use in SVG, it needs to ultimately be 2D, so it makes sense to support basic versions here, including XYZ (points, lines), Bar3D, wireframe.

Could also have a separate plot3d package based on `xyz` that is true 3D for interactive 3D plots of surfaces or things that don't make sense in this more limited 2D world.

# Statistical plots

The `statplot` package provides functions taking `tensor` data that produce statistical plots of the data, including Quartiles (Box with Median, Quartile, Min, Max), Histogram (Bar), Violin (YFill), Range (XFill), Cluster... 

TODO: add a Data scatter that plots points to overlay on top of Violin or Box.

## LegendGroups

* implements current legend grouping logic -- ends up being a multi-table output -- not sure how to interface.

## Histogram

## Quartiles

## Violin

## Range

## Cluster

