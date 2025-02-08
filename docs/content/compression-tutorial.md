This [[tutorial]] provides a simple example of **compressing** a small [[vector]] using a precomputed orthogonal [[matrix]] as a new basis. Through that simple example, this tutorial demonstrates the capacities of Cogent Lab.

We start with a data vector, **x**:

```Goal
# x := [0.53766714, 1.83388501, -2.25884686, 0.86217332, 0.31876524, -1.3076883, -0.43359202, 0.34262447]
```

We can [[plot]] **x**:

```Goal
fig1 := lab.NewPlot(b)
plots.NewPointLine(fig1, x)
plots.NewLine(fig1, #zeros(8)#)
```

We can add in our precomputed orthogonal matrix, **U**, which will serve as our new basis for **x**.

```Goal
# U := [[0.35355339, 0.49039264, 0.46193977, 0.41573481, 0.35355339, 0.27778512, 0.19134172, 0.09754516], [0.35355339, 0.41573481, 0.19134172, -0.09754516, -0.35355339, -0.49039264, -0.46193977, -0.27778512], [0.35355339, 0.27778512, -0.19134172, -0.49039264, -0.35355339, 0.09754516, 0.46193977, 0.41573481], [0.35355339, 0.09754516, -0.46193977, -0.27778512, 0.35355339, 0.41573481, -0.19134172, -0.49039264], [0.35355339, -0.09754516, -0.46193977, 0.27778512, 0.35355339, -0.41573481, -0.19134172, 0.49039264], [0.35355339, -0.27778512, -0.19134172, 0.49039264, -0.35355339, -0.09754516, 0.46193977, -0.41573481], [0.35355339, -0.41573481, 0.19134172, 0.09754516, -0.35355339, 0.49039264, -0.46193977, 0.27778512], [0.35355339, -0.49039264, 0.46193977, -0.41573481, 0.35355339, -0.27778512, 0.19134172, -0.09754516]]
```

We can plot each column vector of **U**:

```Goal
fr := core.NewFrame(b)
fr.Styler(func(s *styles.Style) {
    s.Direction = styles.Column
    s.Gap.Y.Dp(2)
})
for i := range 8 {
    # v := U[:, i]
    fig2 := plot.New()
    pw := plotcore.NewPlot(fr).SetPlot(fig2)
    pw.Styler(func(s *styles.Style) {
        s.Min.Y.Dp(80)
    })
    plot.Styler(v, func(s *plot.Style) {
        s.Plot.Axis.On = false
        s.Range.SetMin(-0.5).SetMax(0.5)
    })
    plots.NewPointLine(fig2, v)
    plots.NewLine(fig2, #zeros(8)#)
}
```

Next, we can compute the vector **a**, which represents **x** in terms of the **U** basis (such that $x = Ua$). This is just $a = U^{-1}x$, as computed below:

```Goal
# a := matrix.Inverse(U) @ x
core.NewText(b).SetText(a.String())
```

To compress the data, we will define a function that zeroes all but the *n* elements of **a** with the highest absolute values:

```Goal
func compress(n int) tensor.Tensor {
    sorted := tensor.NewSliced(a)
    sorted.SortFunc(0, func(tsr tensor.Tensor, i, j int) int {
        return cmp.Compare(math.Abs(tsr.Float1D(j)), math.Abs(tsr.Float1D(i)))
    })
    # top := sorted[:n]
    # res := zeros(8)
    for i := range 8 {
        if tensor.ContainsFloat(top, #a[i]#) {
            # res[i] = a[i]
        } else {
            # res[i] = 0
        }
    }    
    return res
}
```

Then, we will make **a2**, the result of `compress(2)` (so the two most important elements are included):

```Goal
a2 := compress(2)
core.NewText(b).SetText(a2.String())
```

From **a2**, we can compute and plot **x2**, the approximation of **x** based on the two most important elements of **a**:

```Goal
# x2 := U @ a2
core.NewText(b).SetText(x2.String())
fig1a := lab.NewPlotFrom(fig1, b)
plot.Styler(x2, func(s *plot.Style) {
    s.Point.Shape = plot.Pyramid
})
plots.NewPointLine(fig1a, x2)
```

We can do the same thing for **a4** and **x4**, with the four most important elements of **a**:

```Goal
a4 := compress(4)
core.NewText(b).SetText(a4.String())
# x4 := U @ a4
core.NewText(b).SetText(x4.String())
fig1b := lab.NewPlotFrom(fig1a, b)
plot.Styler(x4, func(s *plot.Style) {
    s.Point.Shape = plot.Cross
})
plots.NewPointLine(fig1b, x4)
```

We can also compute **x8**, which just uses **a** without anything removed:

```Goal
# x8 := U @ a
core.NewText(b).SetText(x8.String())
```

We can find the error of **x8** relative to **x**, which is very small since nothing is compressed:

```Goal
func relativeError(y tensor.Tensor) tensor.Tensor {
    # return sqrt(stats.Sum((x-y)**2)/stats.Sum(x**2))
}
core.NewText(b).SetText("x8 error: "+relativeError(x8).String())
```

We can do the same for **x4** and **x2**:

```Goal
core.NewText(b).SetText("x4 error: "+relativeError(x4).String())
core.NewText(b).SetText("x2 error: "+relativeError(x2).String())
```
