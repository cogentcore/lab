**Plots** allow you to graphically plot data.

You can plot a [[vector]]:

```Goal
# x := rand(10)
plt := plot.New()
plt.Add(plots.NewLine(plot.NewY(x)))
plotcore.NewPlot(b).SetPlot(plt)
```
