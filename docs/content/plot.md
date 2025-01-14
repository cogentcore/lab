**Plots** allow you to graphically plot data.

You can plot a [[vector]]:

```Goal
plt := lab.NewPlot(b)
plt.Add(plots.NewLine(plot.NewY(#rand(10)#)))
```

You can plot multiple vectors:

```Goal
plt := lab.NewPlot(b)
plt.Add(plots.NewLine(plot.NewY(#rand(10)#)))
plt.Add(plots.NewLine(plot.NewY(#-rand(10)#)))
```

## Styles

You can style a plot line:

```Goal
plt := lab.NewPlot(b)
plt.Add(plots.NewLine(plot.NewY(#rand(10)#)).Styler(func(s *plot.Style) {
    s.Line.Color = colors.Scheme.Primary.Base
}))
```
