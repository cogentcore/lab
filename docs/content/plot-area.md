+++
Categories = ["Plot"]
+++

These [[plot]] types draw the values of a 2D grid (e.g., a matrix, or a function of two variables), with colors via the color axis of the plot (see Colors in [[plot]]), in the `plots` package.

## Heatmap

`Heatmap` draws a grid of cells colored by the values of a 2D tensor in the `Color` role (rows by columns), using the color axis of the plot, with optional `X` and `Y` labels for the columns and rows. Row 0 is at the top by default, as for a matrix or image (`TopZero`), and the cell edges are drawn if `Line.On` is `On`. For a higher-dimensional tensor, such as a table column of matrices, one row of the outer dimension (`Row`, the last one by default) is drawn. Here is the correlation matrix of four random variables:

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

## Contour

`Contour` draws the contour lines of the values of a 2D tensor in the `Color` role (rows by columns, as for a `Heatmap`): the lines along which the values are equal to each of a number of levels, colored by their level via the color axis of the plot. `FilledContour` fills the regions between the levels instead. The levels are about `Levels` (10 by default) round values across the range of the values, as for axis ticks, or the given `LevelValues`. Optional `X` and `Y` data are the coordinates of the columns and rows; otherwise column j is at X = j and row i at Y = i, with row 0 at the bottom. The values are interpolated linearly between the grid points, and NaN values leave holes. `Labels` draws the value of each level on its lines, in a gap along the middle of each line that is long enough. For `FilledContour`, the color bar shows the colors of the regions between the levels, as steps.

Here are the contours of two gaussian peaks, of opposite signs, with labels:

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
ct := plots.NewContour(plt, plot.Data{plot.Color: z, plot.X: x, plot.Y: y})
ct.Labels = true
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

## Vector

X,Y,U,V

Quiver?

## 3D 

TODO: use math32 3D projection math and you can just take each 3d point and reduce to 2D. For stuff you want to actually be able to use in SVG, it needs to ultimately be 2D, so it makes sense to support basic versions here, including XYZ (points, lines), Bar3D, wireframe.

Could also have a separate plot3d package based on `xyz` that is true 3D for interactive 3D plots of surfaces or things that don't make sense in this more limited 2D world.
