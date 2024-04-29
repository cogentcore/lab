// Code generated by "core generate -add-types"; DO NOT EDIT.

package tensorview

import (
	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
)

// SimMatGridType is the [types.Type] for [SimMatGrid]
var SimMatGridType = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.SimMatGrid", IDName: "sim-mat-grid", Doc: "SimMatGrid is a widget that displays a similarity / distance matrix\nwith tensor values as a grid of colored squares, and labels for rows and columns.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Embeds: []types.Field{{Name: "TensorGrid"}}, Fields: []types.Field{{Name: "SimMat", Doc: "the similarity / distance matrix"}, {Name: "rowMaxSz"}, {Name: "rowMinBlank"}, {Name: "rowNGps"}, {Name: "colMaxSz"}, {Name: "colMinBlank"}, {Name: "colNGps"}}, Instance: &SimMatGrid{}})

// NewSimMatGrid adds a new [SimMatGrid] with the given name to the given parent:
// SimMatGrid is a widget that displays a similarity / distance matrix
// with tensor values as a grid of colored squares, and labels for rows and columns.
func NewSimMatGrid(parent tree.Node, name ...string) *SimMatGrid {
	return parent.NewChild(SimMatGridType, name...).(*SimMatGrid)
}

// NodeType returns the [*types.Type] of [SimMatGrid]
func (t *SimMatGrid) NodeType() *types.Type { return SimMatGridType }

// New returns a new [*SimMatGrid] value
func (t *SimMatGrid) New() tree.Node { return &SimMatGrid{} }

// SetTooltip sets the [SimMatGrid.Tooltip]
func (t *SimMatGrid) SetTooltip(v string) *SimMatGrid { t.Tooltip = v; return t }

// SetDisp sets the [SimMatGrid.Disp]
func (t *SimMatGrid) SetDisp(v TensorDisplay) *SimMatGrid { t.Disp = v; return t }

// SetColorMap sets the [SimMatGrid.ColorMap]
func (t *SimMatGrid) SetColorMap(v *colormap.Map) *SimMatGrid { t.ColorMap = v; return t }

// TableViewType is the [types.Type] for [TableView]
var TableViewType = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.TableView", IDName: "table-view", Doc: "TableView provides a GUI view for [table.Table] values.", Embeds: []types.Field{{Name: "SliceViewBase"}}, Fields: []types.Field{{Name: "Table", Doc: "the idx view of the table that we're a view of"}, {Name: "TensorDisplay", Doc: "overall display options for tensor display"}, {Name: "ColumnTensorDisplay", Doc: "per column tensor display params"}, {Name: "ColumnTensorBlank", Doc: "per column blank tensor values"}, {Name: "NCols", Doc: "number of columns in table (as of last update)"}, {Name: "SortIndex", Doc: "current sort index"}, {Name: "SortDesc", Doc: "whether current sort order is descending"}, {Name: "HeaderWidths", Doc: "HeaderWidths has number of characters in each header, per visfields"}, {Name: "ColMaxWidths", Doc: "ColMaxWidths records maximum width in chars of string type fields"}, {Name: "BlankString", Doc: "\tblank values for out-of-range rows"}, {Name: "BlankFloat"}}, Instance: &TableView{}})

// NewTableView adds a new [TableView] with the given name to the given parent:
// TableView provides a GUI view for [table.Table] values.
func NewTableView(parent tree.Node, name ...string) *TableView {
	return parent.NewChild(TableViewType, name...).(*TableView)
}

// NodeType returns the [*types.Type] of [TableView]
func (t *TableView) NodeType() *types.Type { return TableViewType }

// New returns a new [*TableView] value
func (t *TableView) New() tree.Node { return &TableView{} }

// SetNCols sets the [TableView.NCols]:
// number of columns in table (as of last update)
func (t *TableView) SetNCols(v int) *TableView { t.NCols = v; return t }

// SetSortIndex sets the [TableView.SortIndex]:
// current sort index
func (t *TableView) SetSortIndex(v int) *TableView { t.SortIndex = v; return t }

// SetSortDesc sets the [TableView.SortDesc]:
// whether current sort order is descending
func (t *TableView) SetSortDesc(v bool) *TableView { t.SortDesc = v; return t }

// SetHeaderWidths sets the [TableView.HeaderWidths]:
// HeaderWidths has number of characters in each header, per visfields
func (t *TableView) SetHeaderWidths(v ...int) *TableView { t.HeaderWidths = v; return t }

// SetBlankString sets the [TableView.BlankString]:
//
//	blank values for out-of-range rows
func (t *TableView) SetBlankString(v string) *TableView { t.BlankString = v; return t }

// SetBlankFloat sets the [TableView.BlankFloat]
func (t *TableView) SetBlankFloat(v float64) *TableView { t.BlankFloat = v; return t }

// SetTooltip sets the [TableView.Tooltip]
func (t *TableView) SetTooltip(v string) *TableView { t.Tooltip = v; return t }

// SetMinRows sets the [TableView.MinRows]
func (t *TableView) SetMinRows(v int) *TableView { t.MinRows = v; return t }

// SetViewPath sets the [TableView.ViewPath]
func (t *TableView) SetViewPath(v string) *TableView { t.ViewPath = v; return t }

// SetSelectedValue sets the [TableView.SelectedValue]
func (t *TableView) SetSelectedValue(v any) *TableView { t.SelectedValue = v; return t }

// SetSelectedIndex sets the [TableView.SelectedIndex]
func (t *TableView) SetSelectedIndex(v int) *TableView { t.SelectedIndex = v; return t }

// SetInitSelectedIndex sets the [TableView.InitSelectedIndex]
func (t *TableView) SetInitSelectedIndex(v int) *TableView { t.InitSelectedIndex = v; return t }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.TensorLayout", IDName: "tensor-layout", Doc: "TensorLayout are layout options for displaying tensors", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "OddRow", Doc: "even-numbered dimensions are displayed as Y*X rectangles -- this determines along which dimension to display any remaining odd dimension: OddRow = true = organize vertically along row dimension, false = organize horizontally across column dimension"}, {Name: "TopZero", Doc: "if true, then the Y=0 coordinate is displayed from the top-down; otherwise the Y=0 coordinate is displayed from the bottom up, which is typical for emergent network patterns."}, {Name: "Image", Doc: "display the data as a bitmap image.  if a 2D tensor, then it will be a greyscale image.  if a 3D tensor with size of either the first or last dim = either 3 or 4, then it is a RGB(A) color image"}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.TensorDisplay", IDName: "tensor-display", Doc: "TensorDisplay are options for displaying tensors", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Embeds: []types.Field{{Name: "TensorLayout"}}, Fields: []types.Field{{Name: "Range", Doc: "range to plot"}, {Name: "MinMax", Doc: "if not using fixed range, this is the actual range of data"}, {Name: "ColorMap", Doc: "the name of the color map to use in translating values to colors"}, {Name: "GridFill", Doc: "what proportion of grid square should be filled by color block -- 1 = all, .5 = half, etc"}, {Name: "DimExtra", Doc: "amount of extra space to add at dimension boundaries, as a proportion of total grid size"}, {Name: "GridMinSize", Doc: "minimum size for grid squares -- they will never be smaller than this"}, {Name: "GridMaxSize", Doc: "maximum size for grid squares -- they will never be larger than this"}, {Name: "TotPrefSize", Doc: "total preferred display size along largest dimension.\ngrid squares will be sized to fit within this size,\nsubject to harder GridMin / Max size constraints"}, {Name: "FontSize", Doc: "font size in standard point units for labels (e.g., SimMat)"}, {Name: "GridView", Doc: "our gridview, for update method"}}})

// TensorGridType is the [types.Type] for [TensorGrid]
var TensorGridType = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.TensorGrid", IDName: "tensor-grid", Doc: "TensorGrid is a widget that displays tensor values as a grid of colored squares.", Methods: []types.Method{{Name: "EditSettings", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}}, Embeds: []types.Field{{Name: "WidgetBase"}}, Fields: []types.Field{{Name: "Tensor", Doc: "the tensor that we view"}, {Name: "Disp", Doc: "display options"}, {Name: "ColorMap", Doc: "the actual colormap"}}, Instance: &TensorGrid{}})

// NewTensorGrid adds a new [TensorGrid] with the given name to the given parent:
// TensorGrid is a widget that displays tensor values as a grid of colored squares.
func NewTensorGrid(parent tree.Node, name ...string) *TensorGrid {
	return parent.NewChild(TensorGridType, name...).(*TensorGrid)
}

// NodeType returns the [*types.Type] of [TensorGrid]
func (t *TensorGrid) NodeType() *types.Type { return TensorGridType }

// New returns a new [*TensorGrid] value
func (t *TensorGrid) New() tree.Node { return &TensorGrid{} }

// SetDisp sets the [TensorGrid.Disp]:
// display options
func (t *TensorGrid) SetDisp(v TensorDisplay) *TensorGrid { t.Disp = v; return t }

// SetColorMap sets the [TensorGrid.ColorMap]:
// the actual colormap
func (t *TensorGrid) SetColorMap(v *colormap.Map) *TensorGrid { t.ColorMap = v; return t }

// SetTooltip sets the [TensorGrid.Tooltip]
func (t *TensorGrid) SetTooltip(v string) *TensorGrid { t.Tooltip = v; return t }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.TensorGridValue", IDName: "tensor-grid-value", Doc: "TensorGridValue manages a [TensorGrid] view of an [tensor.Tensor].", Embeds: []types.Field{{Name: "ValueBase"}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.TensorValue", IDName: "tensor-value", Doc: "TensorValue presents a button that pulls up the [TensorView] viewer for an [tensor.Tensor].", Embeds: []types.Field{{Name: "ValueBase"}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.TableValue", IDName: "table-value", Doc: "TableValue presents a button that pulls up the [TableView] viewer for a [table.Table].", Embeds: []types.Field{{Name: "ValueBase"}}})

var _ = types.AddType(&types.Type{Name: "cogentcore.org/core/tensor/tensorview.SimMatValue", IDName: "sim-mat-value", Doc: "SimMatValue presents a button that pulls up the [SimMatGridView] viewer for a [table.Table].", Embeds: []types.Field{{Name: "ValueBase"}}})
