This [[tutorial]] provides a simple example of **compressing** a small [[vector]] using a precomputed orthogonal [[matrix]] as a new basis. Through that simple example, this tutorial demonstrates the capacities of Cogent Lab.

We start with a data vector, **x**:

```Goal
# x := [0.53766714, 1.83388501, -2.25884686, 0.86217332, 0.31876524, -1.3076883, -0.43359202, 0.34262447]
```

We can [[plot]] **x**:

```Goal
plt := lab.NewPlot(b)
plt.Add(plots.NewLine(plot.NewY(x)))
plt.Add(plots.NewLine(plot.NewY(#zeros(8)#)))
```

We can add in our precomputed orthogonal matrix, **U**, which will serve as our new basis for **x**.

```Goal
# U := [[0.35355339, 0.49039264, 0.46193977, 0.41573481, 0.35355339, 0.27778512, 0.19134172, 0.09754516], [0.35355339, 0.41573481, 0.19134172, -0.09754516, -0.35355339, -0.49039264, -0.46193977, -0.27778512], [0.35355339, 0.27778512, -0.19134172, -0.49039264, -0.35355339, 0.09754516, 0.46193977, 0.41573481], [0.35355339, 0.09754516, -0.46193977, -0.27778512, 0.35355339, 0.41573481, -0.19134172, -0.49039264], [0.35355339, -0.09754516, -0.46193977, 0.27778512, 0.35355339, -0.41573481, -0.19134172, 0.49039264], [0.35355339, -0.27778512, -0.19134172, 0.49039264, -0.35355339, -0.09754516, 0.46193977, -0.41573481], [0.35355339, -0.41573481, 0.19134172, 0.09754516, -0.35355339, 0.49039264, -0.46193977, 0.27778512], [0.35355339, -0.49039264, 0.46193977, -0.41573481, 0.35355339, -0.27778512, 0.19134172, -0.09754516]]
```

We can plot each column vector of **U**:

```Goal
# U := 2*rand(8, 8)-1

fr := core.NewFrame(b)
fr.Styler(func(s *styles.Style) {
    s.Direction = styles.Column
    s.Gap.Zero()
})
for i := range 8 {
    # v := U[:, i]
    plt := plot.New()
    pw := plotcore.NewPlot(fr).SetPlot(plt)
    pw.Styler(func(s *styles.Style) {
        s.Min.Y.Dp(100)
    })
    plt.Add(plots.NewLine(plot.NewY(v)).Styler(func(s *plot.Style) {
        s.Plot.Axis.NTicks = 2
    }))
    plt.Add(plots.NewLine(plot.NewY(#zeros(8)#)))
}
```

Next, we can compute the vector **a**, which represents **x** in terms of the **U** basis (such that $x = Ua$). This is just $a = U^{-1}x$, as computed below:

```Goal
# a := matrix.Inverse(U) @ x
core.NewText(b).SetText(a.String())
```
