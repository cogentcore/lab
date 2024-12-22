// Code generated by "core generate"; DO NOT EDIT.

package plotcore

import (
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
	"cogentcore.org/lab/plot"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/plotcore.Plot", IDName: "plot", Doc: "Plot is a widget that renders a [plot.Plot] object.\nIf it is not [states.ReadOnly], the user can pan and zoom the graph.\nSee [PlotEditor] for an interactive interface for selecting columns to view.", Embeds: []types.Field{{Name: "WidgetBase"}}, Fields: []types.Field{{Name: "Plot", Doc: "Plot is the Plot to display in this widget"}, {Name: "SetRangesFunc", Doc: "SetRangesFunc, if set, is called to adjust the data ranges\nafter the point when these ranges are updated based on the plot data."}}})

// NewPlot returns a new [Plot] with the given optional parent:
// Plot is a widget that renders a [plot.Plot] object.
// If it is not [states.ReadOnly], the user can pan and zoom the graph.
// See [PlotEditor] for an interactive interface for selecting columns to view.
func NewPlot(parent ...tree.Node) *Plot { return tree.New[Plot](parent...) }

// SetSetRangesFunc sets the [Plot.SetRangesFunc]:
// SetRangesFunc, if set, is called to adjust the data ranges
// after the point when these ranges are updated based on the plot data.
func (t *Plot) SetSetRangesFunc(v func()) *Plot { t.SetRangesFunc = v; return t }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/plotcore.PlotEditor", IDName: "plot-editor", Doc: "PlotEditor is a widget that provides an interactive 2D plot\nof selected columns of tabular data, represented by a [table.Table] into\na [table.Table]. Other types of tabular data can be converted into this format.\nThe user can change various options for the plot and also modify the underlying data.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "SaveSVG", Doc: "SaveSVG saves the plot to an svg -- first updates to ensure that plot is current", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"fname"}}, {Name: "SavePNG", Doc: "SavePNG saves the current plot to a png, capturing current render", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"fname"}}, {Name: "SaveCSV", Doc: "SaveCSV saves the Table data to a csv (comma-separated values) file with headers (any delim)", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"fname", "delim"}}, {Name: "SaveAll", Doc: "SaveAll saves the current plot to a png, svg, and the data to a tsv -- full save\nAny extension is removed and appropriate extensions are added", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"fname"}}, {Name: "OpenCSV", Doc: "OpenCSV opens the Table data from a csv (comma-separated values) file (or any delim)", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename", "delim"}}, {Name: "setColumnsByName", Doc: "setColumnsByName turns columns on or off if their name contains\nthe given string.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"nameContains", "on"}}}, Embeds: []types.Field{{Name: "Frame"}}, Fields: []types.Field{{Name: "table", Doc: "table is the table of data being plotted."}, {Name: "PlotStyle", Doc: "PlotStyle has the overall plot style parameters."}, {Name: "plot", Doc: "plot is the plot object."}, {Name: "svgFile", Doc: "current svg file"}, {Name: "dataFile", Doc: "current csv data file"}, {Name: "inPlot", Doc: "currently doing a plot"}, {Name: "columnsFrame"}, {Name: "plotWidget"}, {Name: "plotStyleModified"}}})

// NewPlotEditor returns a new [PlotEditor] with the given optional parent:
// PlotEditor is a widget that provides an interactive 2D plot
// of selected columns of tabular data, represented by a [table.Table] into
// a [table.Table]. Other types of tabular data can be converted into this format.
// The user can change various options for the plot and also modify the underlying data.
func NewPlotEditor(parent ...tree.Node) *PlotEditor { return tree.New[PlotEditor](parent...) }

// SetPlotStyle sets the [PlotEditor.PlotStyle]:
// PlotStyle has the overall plot style parameters.
func (t *PlotEditor) SetPlotStyle(v plot.PlotStyle) *PlotEditor { t.PlotStyle = v; return t }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/plotcore.PlotterChooser", IDName: "plotter-chooser", Doc: "PlotterChooser represents a [Plottername] value with a [core.Chooser]\nfor selecting a plotter.", Embeds: []types.Field{{Name: "Chooser"}}})

// NewPlotterChooser returns a new [PlotterChooser] with the given optional parent:
// PlotterChooser represents a [Plottername] value with a [core.Chooser]
// for selecting a plotter.
func NewPlotterChooser(parent ...tree.Node) *PlotterChooser {
	return tree.New[PlotterChooser](parent...)
}