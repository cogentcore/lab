+++
Categories = ["Stats"]
+++

**Histogram** computes histograms. Go docs: [[doc:stats/histogram]]

## Kernel density estimate

A kernel density estimate (KDE) is a smooth version of a histogram, as used in violin plots: it adds up a gaussian for each value. `histogram.KDE` computes it at the values of a grid, such as one from `histogram.KDEGrid`, with a bandwidth (the standard deviation of the gaussians) from `histogram.KDEBandwidth` (Silverman's rule of thumb) by default:

```Goal
plt := lab.NewPlot(b)
vals := tensor.NewFloat64(500)
for i := range 500 {
	if i < 300 {
		vals.SetFloat1D(rand.NormFloat64(), i)
	} else {
		vals.SetFloat1D(4+0.7*rand.NormFloat64(), i)
	}
}
grid := histogram.KDEGrid(vals, 100, 0, 2)
dens := histogram.KDE(vals, grid, 0)
plots.NewBand(plt, plot.Data{plot.X: grid, plot.High: dens})
plots.NewLine(plt, plot.Data{plot.X: grid, plot.Y: histogram.KDE(vals, grid, 0.2)})
```

The filled area is the density with the default bandwidth, and the line has a smaller bandwidth, which shows more detail, and more noise.
