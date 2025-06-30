# Plot

See the [Cogent Lab Docs](https://cogentcore.org/lab/plot) for full documentation.

## Design discussion

The `plot` package generates 2D plots of data using the Cogent Core `paint` rendering system.  The `plotcore` sub-package has Cogent Core Widgets that can be used in applications.  
* `Plot` is just a wrapper around a `plot.Plot`, for code-generated plots.
* `Editor` is an interactive plot viewer that supports selection of which data to plot, and GUI configuration of plot parameters.

`plot` is designed to work in two potentially-conflicting ways:
* Code-based creation of a specific plot with specific data.
* GUI-based configuration of plots based on a `tensor.Table` of data columns (via `Editor`).

The GUI constraint requires a more systematic, factorial organization of the space of possible plot data and how it is organized to create a plot, so that it can be configured with a relatively simple set of GUI settings. The overall logic is as follows:

* The overall plot has a single shared range of X, Y and optionally Z coordinate ranges (under the corresponding `Axis` field), that defines where a data value in any plot type is plotted. These ranges are set based on the DataRanger interface.

* Plot content is driven by `Plotter` elements that each consume one or more sets of data, which is provided by a `Valuer` interface that maps onto a minimal subset of the `tensor.Tensor` interface, so a tensor directly satisfies the interface.

* Each `Plotter` element can generally handle multiple different data elements, that are index-aligned. For example, the basic `XY` plotter requires a `Y` Valuer, and typically an `X`, but indexes will be used if it is not present. It optionally uses `Size` or `Color` Valuers that apply to the Point elements. A `Bar` gets at least a `Y` but also optionally a `High` Valuer for an error bar.  The `plot.Data` = `map[Roles]Valuer` is used to create new Plotter elements, allowing an unordered and explicit way of specifying the `Roles` of each `Valuer` item. Each Plotter also allows a single `Valuer` (i.e., Tensor) argument instead of the data, for a convenient minimal plot cse.  There are also shortcut methods for `NewXY` and `NewY`.

The table-driven plotting case uses a `Group` name along with the `Roles` type (`X`, `Y` etc) and Plotter type names to organize different plots based on `Style` settings.  Columns with the same Group name all provide data to the same plotter using their different Roles, making it easy to configure various statistical plots of multiple series of grouped data.

Different plotter types (including custom ones) are registered along with their accepted input roles, to allow any type of plot to be generated.

# History

The code is adapted from the [gonum plot](https://github.com/gonum/plot) package (which in turn was adapted from google's [plotinum](https://code.google.com/archive/p/plotinum/), to use the Cogent Core [styles](../styles) and [paint](../paint) rendering framework, which also supports SVG output of the rendering.

Here is the copyright notice for that package:
```go
// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
```

# TODO

* Min / Max not just for extending but also _limiting_ the range -- currently doesn't do

* tensor index
* Grid? in styling.

