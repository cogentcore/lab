# basic

This is an example of a basic lab.Browser with the files as the left panel, and the Tabber as the right panel.

You must run this from the command line, to get the interactive interpreter prompt.

## Make some data

```Go
# x := rand(10)
```

## Make a new plot

```Go
plt := plot.New()
plt.Add(plots.NewLine(plot.NewY(x)))
lab.Lab.Plot("plot", plt)
```

## Attach some styling to the plot data

```Go
plot.SetStyler(x, func(s *plot.Style) {
	s.Plot.Title = "Test Line"
	s.Plot.XAxis.Label = "X Axis"
	s.Line.Color = colors.Uniform(colors.Red)
})
```

Note: you have to regenerate the plot to get these styles to take effect.

