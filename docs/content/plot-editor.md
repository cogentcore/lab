+++
Categories = ["Plots"]
+++

A **plot editor** allows you to create data plots that users can customize interactively.

Here is a simple plot editor:

```Goal
dt := table.New("Data")
values := tensor.NewFloat32FromValues(1, 2.5, 7)
plot.SetStyler(values, func(s *plot.Style) {
    s.On = true
})
errors.Log(dt.AddColumn("Values", values.AsValues()))
moreValues := tensor.NewFloat32FromValues(5, 3, 2)
errors.Log(dt.AddColumn("More values", moreValues.AsValues()))
plotcore.NewEditor(b).SetTable(dt)
```
