// Code generated by "core generate"; DO NOT EDIT.

package main

import (
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "main.Random", IDName: "random", Doc: "Random is the random distribution plotter widget.", Methods: []types.Method{{Name: "Plot", Doc: "Plot generates the data and plots a histogram of results.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}}, Embeds: []types.Field{{Name: "Frame"}, {Name: "Data"}}})

// NewRandom returns a new [Random] with the given optional parent:
// Random is the random distribution plotter widget.
func NewRandom(parent ...tree.Node) *Random { return tree.New[Random](parent...) }

var _ = types.AddType(&types.Type{Name: "main.Data", IDName: "data", Doc: "Data contains the random distribution plotter data and options.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Dist", Doc: "random params"}, {Name: "NumSamples", Doc: "number of samples"}, {Name: "NumBins", Doc: "number of bins in the histogram"}, {Name: "Range", Doc: "range for histogram"}, {Name: "Table", Doc: "table for raw data"}, {Name: "Histogram", Doc: "histogram of data"}, {Name: "plot", Doc: "the plot"}}})
