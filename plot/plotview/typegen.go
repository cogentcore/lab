// Code generated by "core generate"; DO NOT EDIT.

package plotview

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/plot/plotview.PlotParams", IDName: "plot-params", Doc: "PlotParams are parameters for overall plot", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Title", Doc: "optional title at top of plot"}, {Name: "Type", Doc: "type of plot to generate.  For a Bar plot, items are plotted ordinally by row and the XAxis is optional"}, {Name: "Lines", Doc: "whether to plot lines"}, {Name: "Points", Doc: "whether to plot points with symbols"}, {Name: "LineWidth", Doc: "width of lines"}, {Name: "PointSize", Doc: "size of points"}, {Name: "PointShape", Doc: "the shape used to draw points"}, {Name: "BarWidth", Doc: "width of bars for bar plot, as fraction of available space (1 = no gaps)"}, {Name: "NegXDraw", Doc: "draw lines that connect points with a negative X-axis direction -- otherwise these are treated as breaks between repeated series and not drawn"}, {Name: "Scale", Doc: "overall scaling factor -- the larger the number, the larger the fonts are relative to the graph"}, {Name: "XAxisCol", Doc: "what column to use for the common X axis -- if empty or not found, the row number is used.  This optional for Bar plots -- if present and LegendCol is also present, then an extra space will be put between X values."}, {Name: "LegendCol", Doc: "optional column for adding a separate colored / styled line or bar according to this value -- acts just like a separate Y variable, crossed with Y variables"}, {Name: "XAxisRot", Doc: "rotation of the X Axis labels, in degrees"}, {Name: "XAxisLabel", Doc: "optional label to use for XAxis instead of column name"}, {Name: "YAxisLabel", Doc: "optional label to use for YAxis -- if empty, first column name is used"}, {Name: "Plot", Doc: "our plot, for update method"}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/plot/plotview.ColumnParams", IDName: "column-params", Doc: "ColumnParams are parameters for plotting one column of data", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "On", Doc: "whether to plot this column"}, {Name: "Column", Doc: "name of column we're plotting"}, {Name: "Lines", Doc: "whether to plot lines; uses the overall plot option if unset"}, {Name: "Points", Doc: "whether to plot points with symbols; uses the overall plot option if unset"}, {Name: "LineWidth", Doc: "the width of lines; uses the overall plot option if unset"}, {Name: "PointSize", Doc: "the size of points; uses the overall plot option if unset"}, {Name: "PointShape", Doc: "the shape used to draw points; uses the overall plot option if unset"}, {Name: "Range", Doc: "effective range of data to plot -- either end can be fixed"}, {Name: "FullRange", Doc: "full actual range of data -- only valid if specifically computed"}, {Name: "Color", Doc: "color to use when plotting the line / column"}, {Name: "NTicks", Doc: "desired number of ticks"}, {Name: "Lbl", Doc: "if non-empty, this is an alternative label to use in plotting"}, {Name: "TensorIndex", Doc: "if column has n-dimensional tensor cells in each row, this is the index within each cell to plot -- use -1 to plot *all* indexes as separate lines"}, {Name: "ErrCol", Doc: "specifies a column containing error bars for this column"}, {Name: "IsString", Doc: "if true this is a string column -- plots as labels"}, {Name: "Plot", Doc: "our plot, for update method"}}})

// PlotType is the [types.Type] for [Plot]
var PlotType = types.AddType(&types.Type{Name: "cogentcore.org/core/plot/plotview.Plot", IDName: "plot", Doc: "Plot is a Widget that renders a [plot.Plot] object.\nIf it is not [states.ReadOnly], the user can pan and zoom the display.\nBy default, it is [states.ReadOnly]. See [ConfigPlotToolbar] for a\ntoolbar with panning, selecting, and I/O buttons,\nand PlotView for an interactive interface for selecting columns to view.", Methods: []types.Method{{Name: "SavePlot", Doc: "SaveSVG saves the current Plot to an SVG file", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename"}, Returns: []string{"error"}}, {Name: "SavePNG", Doc: "SavePNG saves the current rendered Plot image to an PNG image file.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename"}, Returns: []string{"error"}}}, Embeds: []types.Field{{Name: "WidgetBase"}}, Fields: []types.Field{{Name: "Plot", Doc: "Plot is the Plot object associated with the element."}}, Instance: &Plot{}})

// NewPlot adds a new [Plot] with the given name to the given parent:
// Plot is a Widget that renders a [plot.Plot] object.
// If it is not [states.ReadOnly], the user can pan and zoom the display.
// By default, it is [states.ReadOnly]. See [ConfigPlotToolbar] for a
// toolbar with panning, selecting, and I/O buttons,
// and PlotView for an interactive interface for selecting columns to view.
func NewPlot(parent tree.Node, name ...string) *Plot {
	return parent.NewChild(PlotType, name...).(*Plot)
}

// NodeType returns the [*types.Type] of [Plot]
func (t *Plot) NodeType() *types.Type { return PlotType }

// New returns a new [*Plot] value
func (t *Plot) New() tree.Node { return &Plot{} }

// SetTooltip sets the [Plot.Tooltip]
func (t *Plot) SetTooltip(v string) *Plot { t.Tooltip = v; return t }

// PlotViewType is the [types.Type] for [PlotView]
var PlotViewType = types.AddType(&types.Type{Name: "cogentcore.org/core/plot/plotview.PlotView", IDName: "plot-view", Doc: "PlotView is a Cogent Core Widget that provides a 2D plot of selected columns of Table data", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "SaveSVG", Doc: "SaveSVG saves the plot to an svg -- first updates to ensure that plot is current", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"fname"}}, {Name: "SavePNG", Doc: "SavePNG saves the current plot to a png, capturing current render", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"fname"}}, {Name: "SaveAll", Doc: "SaveAll saves the current plot to a png, svg, and the data to a tsv -- full save\nAny extension is removed and appropriate extensions are added", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"fname"}}, {Name: "SetColumnsByName", Doc: "SetColumnsByName turns cols On or Off if their name contains given string", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"nameContains", "on"}}}, Embeds: []types.Field{{Name: "Layout"}}, Fields: []types.Field{{Name: "Table", Doc: "the table of data being plotted"}, {Name: "Params", Doc: "the overall plot parameters"}, {Name: "Columns", Doc: "the parameters for each column of the table"}, {Name: "Plot", Doc: "the plot object"}, {Name: "ConfigPlotFunc", Doc: "ConfigPlotFunc is a function to call to configure [PlotView.Plot], the gonum plot that\nactually does the plotting. It is called after [Plot] is generated, and properties\nof [Plot] can be modified in it. Properties of [Plot] should not be modified outside\nof this function, as doing so will have no effect."}, {Name: "SVGFile", Doc: "current svg file"}, {Name: "InPlot", Doc: "currently doing a plot"}}, Instance: &PlotView{}})

// NewPlotView adds a new [PlotView] with the given name to the given parent:
// PlotView is a Cogent Core Widget that provides a 2D plot of selected columns of Table data
func NewPlotView(parent tree.Node, name ...string) *PlotView {
	return parent.NewChild(PlotViewType, name...).(*PlotView)
}

// NodeType returns the [*types.Type] of [PlotView]
func (t *PlotView) NodeType() *types.Type { return PlotViewType }

// New returns a new [*PlotView] value
func (t *PlotView) New() tree.Node { return &PlotView{} }

// SetParams sets the [PlotView.Params]:
// the overall plot parameters
func (t *PlotView) SetParams(v PlotParams) *PlotView { t.Params = v; return t }

// SetConfigPlotFunc sets the [PlotView.ConfigPlotFunc]:
// ConfigPlotFunc is a function to call to configure [PlotView.Plot], the gonum plot that
// actually does the plotting. It is called after [Plot] is generated, and properties
// of [Plot] can be modified in it. Properties of [Plot] should not be modified outside
// of this function, as doing so will have no effect.
func (t *PlotView) SetConfigPlotFunc(v func()) *PlotView { t.ConfigPlotFunc = v; return t }

// SetSVGFile sets the [PlotView.SVGFile]:
// current svg file
func (t *PlotView) SetSVGFile(v core.Filename) *PlotView { t.SVGFile = v; return t }

// SetTooltip sets the [PlotView.Tooltip]
func (t *PlotView) SetTooltip(v string) *PlotView { t.Tooltip = v; return t }
