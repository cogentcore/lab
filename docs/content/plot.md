**Plots** allow you to graphically plot data.

You can plot a [[vector]]:

```Goal
# x := rand(10)
plt := plot.New()
plt.Add(plots.NewLine(plot.NewY(x)))
plotcore.NewPlot(b).SetPlot(plt)
```

## Styles

You can style a plot line:

```Goal
# x := rand(10)
plot.SetStylersTo(x, plot.Stylers{func(s *plot.Style) {
    s.Line.Color = colors.Scheme.Primary.Base
}})
plt := plot.New()
plt.Add(plots.NewLine(plot.NewY(x)))
plotcore.NewPlot(b).SetPlot(plt)
```
