# basic

This is an example of a basic lab.Browser with the files as the left panel, and the Tabber as the right panel.

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
plot.SetStylersTo(x, plot.Stylers{func(s *plot.Style) {
	s.Plot.Title = "Test Line"
	s.Plot.XAxis.Label = "X Axis"
	s.Line.Color = colors.Uniform(colors.Red)
}})
```

Note: you have to regenerate the plot to get these styles to take effect.

