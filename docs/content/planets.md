+++
Categories = ["Tutorials"]
+++

This [[tutorial]] analyzes data about the planets that have been discovered around other stars (exoplanets), following the [aggregation and grouping](https://jakevdp.github.io/PythonDataScienceHandbook/03.08-aggregation-and-grouping.html) chapter of the Python Data Science Handbook, which uses pandas, so you can see how the same analyses are done here, and then plots the results. The `examples/planets` directory of the lab repository has a full app with this data and an interactive [[Goal]] shell.

## The data

The data is one of the standard datasets in `lab.Datasets`, which we open into a [[table]]:

```Goal
planets := table.New("planets")
planets.OpenFS(lab.Datasets, "planets.csv", tensor.Comma)
tensorcore.NewTable(b).SetTable(planets)
```

Each row is a planet: the `method` used to discover it, the `number` of planets around its star, its `orbital_period` (in days), its `mass` (relative to Jupiter), its `distance` from us (in parsecs), and the `year` it was discovered. Values that are not known are `NaN` (missing data).

## Simple aggregation

In [[math]] mode, the functions of NumPy such as `mean` and `std` reduce all of the values of a column to one value, ignoring `NaN` values as missing data:

```Goal
##
mass := planets.Column("mass")
n := stats.Count(mass)
avg := mean(mass)
sd := std(mass)
lo := min(mass)
hi := max(mass)
##
fmt.Println("mass: count:", n, "mean:", avg, "std:", sd, "min:", lo, "max:", hi)
```

The `stats.DescribeTableAll` function computes the standard descriptive statistics of all of the numeric columns into a [[tensorfs]] directory, which `tensorfs.DirTableMatrix` turns into a table, as `describe` does in pandas:

```Goal
dir, _ := tensorfs.NewDir("Planets")
stats.DescribeTableAll(dir, planets)
tensorcore.NewTable(b).SetTable(tensorfs.DirTableMatrix(dir.Dir("Describe"), nil))
```

## Computing new columns

Math mode computes new columns from others, element by element. Here, the decade of each discovery, and the logarithms of the orbital periods and masses, which span many orders of magnitude:

```Goal
##
decade := floor(planets.Column("year") / 10) * 10
logPeriod := log10(planets.Column("orbital_period"))
logMass := log10(planets.Column("mass"))
##
planets.AddColumn("decade", tensor.AsFloat64(decade))
planets.AddColumn("log period", tensor.AsFloat64(logPeriod))
planets.AddColumn("log mass", tensor.AsFloat64(logMass))
```

A scatter [[plot]] of the masses and orbital periods, colored by the year of discovery, shows that the first planets found were massive, and that lighter planets with shorter orbital periods were found as the methods improved (only the planets with a known mass are shown):

```Goal
plt := lab.NewPlot(b)
plots.NewScatter(plt, plot.Data{plot.X: planets.Column("log period"), plot.Y: planets.Column("log mass"), plot.Color: planets.Column("year")}).Styler(func(s *plot.Style) {
	s.Point.Size.Pt(1.5)
	s.Plot.Title = "Exoplanets"
	s.Plot.XAxis.Label = "log10 orbital period (days)"
	s.Plot.YAxisLabel = "log10 mass (Jupiter masses)"
	s.Plot.ColorAxis.Label = "Year"
	s.Plot.ColorAxis.ColorMap = "Viridis"
})
```

## Grouping

The `stats.TableGroups` function groups the rows by the values of the `method` and `decade` columns, and `stats.TableGroupStats` computes a statistic of a column for each group: here, the number of planets discovered by each method, and in each decade:

```Goal
stats.TableGroups(dir, planets, "method", "decade")
stats.TableGroupStats(dir, stats.StatCount, planets, "number")
byMethod := stats.GroupStatsDirTable(dir.Dir("Stats").Dir("method"))
plt := lab.NewPlot(b)
plots.NewXBar(plt, plot.Data{plot.X: byMethod.Column("method"), plot.Y: byMethod.Column("number/Count")}).Styler(func(s *plot.Style) {
	s.Plot.Title = "Discoveries by method"
	s.Plot.XAxis.Label = "Number of planets"
})
```

Two methods found most of the planets: the motion of their stars (radial velocity), and their passing in front of their stars (transits). Their shares of the discoveries, as a pie chart:

```Goal
shares := map[string]float64{}
for i := range byMethod.NumRows() {
	m := byMethod.Column("method").String1D(i)
	if m != "Radial Velocity" && m != "Transit" {
		m = "Other"
	}
	shares[m] += byMethod.Column("number/Count").Float1D(i)
}
plt := lab.NewPlot(b)
pe := plots.NewPie(plt, plot.Data{plot.Y: plot.Values{shares["Radial Velocity"], shares["Transit"], shares["Other"]}, plot.Label: tensor.NewStringFromValues("Radial Velocity", "Transit", "Other")})
pe.Percent = true
pe.Hole = 0.4
```

## Statistical plots

The two main methods find very different planets, as the statistical plots of the `statsplot` package show (see [[plot stats]]). First, we select the rows of these methods:

```Goal
twoMethods := table.NewView(planets)
twoMethods.Filter(func(dt *table.Table, row int) bool {
	m := dt.Column("method").StringRow(row, 0)
	return m == "Radial Velocity" || m == "Transit"
})
statsDir, _ := tensorfs.NewDir("Stats")
```

The histograms of the orbital periods of the two methods: transits find planets with short orbits, which pass in front of their stars often:

```Goal
hs := &statsplot.Histogram{Value: "log period", Split: "method", Bins: 30}
ht, _ := hs.Table(statsDir, twoMethods)
plt, _ := plot.NewTablePlot(ht)
lab.NewPlotFrom(plt, b)
```

The distributions of the orbital periods of the two methods in each decade, as violin plots:

```Goal
vl := &statsplot.Violin{Category: "decade", Value: "log period", Split: "method", Points: 50, Box: true, Whisker: 1.5}
vt, _ := vl.Table(statsDir, twoMethods)
plt, _ := plot.NewTablePlot(vt)
lab.NewPlotFrom(plt, b)
```

The mean mass of the planets discovered in each decade, with its standard error, of all the planets with a known mass:

```Goal
me := &statsplot.MeanError{X: "decade", Values: []statsplot.Column{"log mass"}}
mt, _ := me.Table(statsDir, planets)
plt, _ := plot.NewTablePlot(mt)
lab.NewPlotFrom(plt, b)
```

## Pivot table

A pivot table of the number of planets discovered by each method in each decade, as `pivot_table` does in pandas, is a matrix of counts, which a heatmap shows:

```Goal
methods := byMethod.Column("method")
byDecade := stats.GroupStatsDirTable(dir.Dir("Stats").Dir("decade"))
decades := byDecade.Column("decade")
// the index of each method and decade, by its name
methodIndex, decadeIndex := map[string]int{}, map[string]int{}
for i := range methods.Len() {
	methodIndex[methods.String1D(i)] = i
}
for i := range decades.Len() {
	decadeIndex[decades.String1D(i)] = i
}
counts := tensor.NewFloat64(methods.Len(), decades.Len())
for row := range planets.NumRows() {
	mi := methodIndex[planets.Column("method").StringRow(row, 0)]
	di := decadeIndex[planets.Column("decade").StringRow(row, 0)]
	counts.Set(counts.Value(mi, di)+1, mi, di)
}
plt := lab.NewPlot(b)
plots.NewHeatmap(plt, plot.Data{plot.Color: counts, plot.X: decades, plot.Y: methods}).Styler(func(s *plot.Style) {
	s.Line.On = plot.On
	s.Plot.Title = "Discoveries by method and decade"
	s.Plot.ColorAxis.Label = "Number of planets"
})
```
