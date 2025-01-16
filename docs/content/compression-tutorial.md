This [[tutorial]] provides a simple example of **compressing** a small [[vector]] using a precomputed orthogonal [[matrix]] as a new basis. Through that simple example, this tutorial demonstrates the capacities of Cogent Lab.

We start with a data vector, **x**:

```Goal
# x := [0.53766714, 1.83388501, -2.25884686, 0.86217332, 0.31876524, -1.3076883, -0.43359202, 0.34262447]
```

We can [[plot]] **x**:

```Goal
fig1 := lab.NewPlot(b)
fig1.Add(plots.NewLine(plot.NewY(x)))
fig1.Add(plots.NewLine(plot.NewY(#zeros(8)#)))
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
    s.Gap.Zero()
})
for i := range 8 {
    # v := U[:, i]
    fig2 := plot.New()
    pw := plotcore.NewPlot(fr).SetPlot(fig2)
    pw.Styler(func(s *styles.Style) {
        s.Min.Y.Dp(100)
    })
    fig2.Y.Range.Set(-0.5, 0.5)
    fig2.Add(plots.NewLine(plot.NewY(v)).Styler(func(s *plot.Style) {
        s.Plot.Axis.NTicks = 0
        s.Point.On = plot.On
    }))
    fig2.Add(plots.NewLine(plot.NewY(#zeros(8)#)))
}
```

Next, we can compute the vector **a**, which represents **x** in terms of the **U** basis (such that $x = Ua$). This is just $a = U^{-1}x$, as computed below:

```Goal
# a := matrix.Inverse(U) @ x
core.NewText(b).SetText(a.String())
```

To compress the data, we will define a function that zeroes all but the *n* elements of **a** with the highest absolute values:

```Goal
// TODO: implement using sort
compress := func(n int) tensor.Tensor {
    # res := zeros(8)
    for i := range 8 {
        if i < n {
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
plt := lab.NewPlotFrom(fig1, b)
plt.Add(plots.NewLine(plot.NewY(x2)))
```
