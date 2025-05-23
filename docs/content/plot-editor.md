+++
Categories = ["Plots"]
+++

A **plot editor** allows you to create data plots that users can customize interactively.

Here is a simple plot editor:

```Go
dt := table.New("Data")
values := tensor.NewFloat32FromValues(1, 2.5, 7)
plot.SetStyler(values, func(s *plot.Style) {
    s.On = true
})
errors.Log(dt.AddColumn("Values", values.AsValues()))
plotcore.NewEditor(b).SetTable(dt)
```
