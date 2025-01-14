This [[tutorial]] provides a simple example of **compressing** a small [[vector]] using a precomputed orthogonal [[matrix]] as a new basis. Through that simple example, this tutorial demonstrates the capacities of Cogent Lab.

We start with a data vector, **x**:

```Goal
# x := [0.53766714, 1.83388501, -2.25884686, 0.86217332, 0.31876524, -1.3076883, -0.43359202, 0.34262447]
```

We can plot **x**:

```Goal
# x := [0.53766714, 1.83388501, -2.25884686, 0.86217332, 0.31876524, -1.3076883, -0.43359202, 0.34262447]
plt := plot.New()
plt.Add(plots.NewLine(plot.NewY(x)))
plt.Add(plots.NewLine(plot.NewY(#zeros(8)#)))
plotcore.NewPlot(b).SetPlot(plt)
```
