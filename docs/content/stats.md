+++
+++

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

## Stats pages

