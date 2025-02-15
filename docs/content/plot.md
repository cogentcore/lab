**Plots** allow you to graphically plot data.

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
