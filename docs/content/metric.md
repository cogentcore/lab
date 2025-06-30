+++
Categories = ["Stats"]
+++

**Metric** computes distance metrics for comparing [[tensor]]s. The different metrics supported are: [[doc:stats/metric.Metrics]].

```Goal
##
x := rand(12)
y := rand(12)

l2 := metric.L2Norm(x, y)
r := metric.Correlation(x, y)
##

fmt.Println("x:", x)
fmt.Println("y:", y)
fmt.Println("l2:", l2)
fmt.Println("r:", r)
```

As with statistics, n-dimensional data is treated in a row-based manner, computing a metric value over the data across rows:

```Goal
##
x := rand(12).reshape(3,4)
y := rand(12).reshape(3,4)

l2 := metric.L2Norm(x, y)
r := metric.Correlation(x, y)
##

fmt.Println("x:", x)
fmt.Println("y:", y)
fmt.Println("l2:", l2)
fmt.Println("r:", r)
```

To get a single value for each row representing the metric computed on the elements within that row, you need to iterate and slice:

<!--- TODO: can't use anything reasonable in the max on this damn for loop! -->
<!--- x.DimSize(0) or something grabbed in math mode.. -->

```Goal
##
x := rand(12).reshape(3,4)
y := rand(12).reshape(3,4)
l2 := zeros(3)
## 

for i := range 3 {
    ##
    l2[i] = metric.L2Norm(x[i], y[i])
    ##
}

fmt.Println("x:", x)
fmt.Println("y:", y)
fmt.Println("l2:", l2)

```

