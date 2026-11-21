In addition to the basic statistics functions described below, there are several packages for computing **statistics** on [[tensor]] and [[table]] data:

* [[metric]] computes similarity / distance metrics for comparing two tensors, and associated distance / similarity matrix functions.

* [[cluster]] implements agglomerative clustering of items based on metric distance / similarity matrix data.

* [[convolve]] convolves data (e.g., for smoothing).

* [[glm]] fits a general linear model for one or more dependent variables as a function of one or more independent variables. This encompasses all forms of regression.

* [[histogram]] bins data into groups and reports the frequency of elements in the bins.

## Stats

The standard statistics functions supported are enumerated in [[doc:stats/stats.Stats]], and include things like `Mean`, `Var`iance, etc.

```Goal
##
x := linspace(0., 12., 12., false)
d := x.reshape(3,2,2) // n-dimensional data is handled 

mean := stats.Mean(x)
meand := stats.Mean(d)
##

fmt.Println("x:", x)
fmt.Println("mean:", mean)
fmt.Println("d:", d)
fmt.Println("mean d:", meand)
```

You can see that the stats on n-dimensional data are automatically computed across the _row_ (outer-most) dimension. You can reshape your data and the results as needed to get the statistics you want.

## Quantiles and box plots

`stats.Quantiles` computes any quantiles of the values (e.g., 0.5 for the median), with linear interpolation between the sorted values, and `stats.BoxStats` computes the five values of a box plot: the low whisker, first quartile, median, third quartile, and high whisker, where the whiskers are the most extreme values within the given multiple (typically 1.5) of the interquartile range from the quartiles, and values beyond them are outliers. Both are computed across rows for n-dimensional data, as for the other stats:

```Goal
##
x := [1, 2, 3, 4, 5, 6, 7, 8, 9, 30]
##
fmt.Println("median:", stats.Median(x))
fmt.Println("quantiles:", stats.Quantiles(x, tensor.NewFloat64FromValues(0.1, 0.9)))
fmt.Println("box stats:", stats.BoxStats(x, 1.5))
```

## Grouping and stats

The `stats` package has functions that group values in a [[tensor]] or a [[table]] so that statistics can be computed across the groups. The grouping uses [[tensorfs]] to organize the groups and statistics, as in the following example:

```Goal
dt := table.New().SetNumRows(4)
dt.AddStringColumn("Name")
dt.AddFloat32Column("Value")
for i := range 4 {
	gp := "A"
	if i >= 2 {
		gp = "B"
	}
	dt.Column("Name").SetStringRow(gp, i, 0)
	dt.Column("Value").SetFloatRow(float64(i), i, 0)
}
dir, _ := tensorfs.NewDir("Group")
stats.TableGroups(dir, dt, "Name")
stats.TableGroupStats(dir, stats.StatMean, dt, "Value")
gdt := stats.GroupStatsAsTableNoStatName(dir)

fmt.Println("dt:", dt)
fmt.Println("tensorfs listing:")
fmt.Println(dir.ListAll())

fmt.Println("gdt:", gdt)
```

To group by the combinations of the values of several columns jointly, as `groupby` with a list of columns does in pandas, use `stats.TableGroupCombined` (or `stats.GroupCombined` for tensors): `GroupStats` then makes a table with a column for each of the grouped columns, followed by the statistics of each combination, which `tensor.Pivot` reshapes into a matrix (a pivot table, or crosstab), as in the [[planets]] tutorial.

## Stats pages

