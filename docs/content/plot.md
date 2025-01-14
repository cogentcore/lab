**Plots** allow you to graphically plot data.

You can plot a [[vector]]:

```Goal
# x := rand(10)
plt := lab.NewPlot(b)
plt.Add(plots.NewLine(plot.NewY(x)))
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
# x := rand(10)
plot.SetStylersTo(x, func(s *plot.Style) {
    s.Line.Color = colors.Scheme.Primary.Base
})
plt := lab.NewPlot(b)
plt.Add(plots.NewLine(plot.NewY(x)))
```
