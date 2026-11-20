+++
Categories = ["Plot"]
+++

The `statsplot` package computes the data of common statistical [[plot]]s from the columns of a [[table]], and returns a table of the results with the plot styles set, which can be plotted directly, or in a [[plot editor]] (whose **Stats** menu makes these plots of its data). The results are also stored in a [[tensorfs]] directory. The values of an optional `Split` column (e.g., the condition of a simulation) are plotted as separate series, in their own columns and colors.

The examples use this table of simulated results: 8 runs of 20 epochs in two conditions (`Cond`), with an error (`Err`) that decreases faster in condition B, and a reaction time (`RT`) in three categories (`Cat`):

```Goal
results := table.New("Results")
epoch := results.AddIntColumn("Epoch")
cond := results.AddStringColumn("Cond")
cat := results.AddStringColumn("Cat")
errs := results.AddFloat64Column("Err")
rts := results.AddFloat64Column("RT")
for ci, cnd := range []string{"A", "B"} {
	for run := range 8 {
		for e := range 20 {
			results.AddRows(1)
			i := results.NumRows() - 1
			epoch.SetInt1D(e, i)
			cond.SetString1D(cnd, i)
			cg := (run + e) % 3
			cat.SetString1D([]string{"Easy", "Medium", "Hard"}[cg], i)
			errs.SetFloat1D(math.Exp(-(0.1+0.1*float64(ci))*float64(e))+0.05*rand.NormFloat64(), i)
			rts.SetFloat1D(300+100*float64(cg)+50*float64(ci)+math.Exp(4+0.5*rand.NormFloat64()), i)
		}
	}
}
statsDir, _ := tensorfs.NewDir("Stats")
```

## Mean and error

`MeanError` plots the mean of the values of columns for each value of the `X` column (e.g., across the runs of a simulation at each epoch), with the error of the mean (`Sem`, `Std`, or `CI95`) as a band or error bars. The means are plotted as bars for an `X` column of strings (categories).

```Goal
me := &statsplot.MeanError{X: "Epoch", Values: []statsplot.Column{"Err"}, Split: "Cond"}
mt, _ := me.Table(statsDir, results)
plt, _ := plot.NewTablePlot(mt)
lab.NewPlotFrom(plt, b)
```

In the result table, the error of each mean is in a column with the `High` role and a `Band` (or `YErrorBars`) plotter, in the same group as the column of the mean: a table plot group makes another plotter for an On column with a different plotter set.

## Histogram

`Histogram` plots the number of values within bins of equal width, which are the same for all of the values of the `Split` column, so they can be compared. `Density` normalizes the counts to a probability density.

```Goal
hs := &statsplot.Histogram{Value: "RT", Split: "Cond", Bins: 20}
ht, _ := hs.Table(statsDir, results)
plt, _ := plot.NewTablePlot(ht)
lab.NewPlotFrom(plt, b)
```

## Box plot

`Box` plots a box plot of the values for each value of the `Category` column, with whiskers out to the most extreme values within `Whisker` times the interquartile range, and the outliers beyond them as points.

```Goal
bx := &statsplot.Box{Category: "Cat", Value: "RT", Split: "Cond", Whisker: 1.5, Outliers: true}
bt, _ := bx.Table(statsDir, results)
plt, _ := plot.NewTablePlot(bt)
lab.NewPlotFrom(plt, b)
```

## Violin plot

`Violin` plots the distribution of the values for each category as a kernel density estimate (see [[histogram]]), mirrored on both sides, with a box plot inside it.

```Goal
vl := &statsplot.Violin{Category: "Cat", Value: "RT", Split: "Cond", Points: 50, Box: true, Whisker: 1.5}
vt, _ := vl.Table(statsDir, results)
plt, _ := plot.NewTablePlot(vt)
lab.NewPlotFrom(plt, b)
```

## In the plot editor

The **Stats** menu of a [[plot editor]] makes these plots of its table, with a dialog of their options, which starts from the columns of the current plot: the X axis, the columns that are on, and the Split column. The new plot is shown in a new tab in Cogent Lab, or a new window.

For the following plot, click on the Stats menu in the toolbar, and select `Mean and error` -- use the default options, and you'll see a nice Band summary of the data.

```Goal
plot.SetStyler(results.Column("Epoch"), func(s *plot.Style) { s.Role = plot.X })
plot.SetStyler(results.Column("Err"), func(s *plot.Style) { s.On = true })
plot.SetStyler(results.Column("Cond"), func(s *plot.Style) {
	s.On = true
	s.Role = plot.Split
})
plotcore.NewEditorWithBar(b, plotcore.InlineBar).SetTable(results)
```
