// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package plotcore provides Cogent Core widgets for viewing and editing plots.
package plotcore

//go:generate core generate

import (
	"fmt"
	"image"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/metadata"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/system"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorcore"
	"cogentcore.org/lab/tensorfs"
	"golang.org/x/exp/maps"
)

// Editor is a widget that provides an interactive 2D plot
// of selected columns of tabular data, represented by a [table.Table] into
// a [table.Table]. Other types of tabular data can be converted into this format.
// The user can change various options for the plot and also modify the underlying data.
type Editor struct { //types:add
	core.Frame

	// table is the table of data being plotted.
	table *table.Table

	// PlotStyle has the overall plot style parameters.
	PlotStyle plot.PlotStyle

	// plot is the plot object.
	plot *plot.Plot

	// current svg file
	svgFile core.Filename

	// current csv data file
	dataFile core.Filename

	// currently doing a plot
	inPlot bool

	columnsFrame      *core.Frame
	plotWidget        *Plot
	plotStyleModified map[string]bool

	// columns has the style of each column, by name.
	columns map[string]*editorColumn

	// PlotTableFunc, if set, shows the table of a new plot made from this
	// one, with the given name, such as a statistical plot of the Stats
	// menu (e.g., in a new tab), instead of in a new window.
	PlotTableFunc func(name string, dt *table.Table) `display:"-"`

	// StatsDir is the tensorfs directory of the data of the statistical
	// plots of the Stats menu.
	StatsDir *tensorfs.Node `display:"-"`
}

// editorColumn has the style of a column in the [Editor],
// which is applied to the column on top of any of its own stylers.
type editorColumn struct {

	// style is the style of the column.
	style *plot.Style

	// mods are the modified fields of style, which are applied.
	mods map[string]bool

	// isString is whether the column holds strings.
	isString bool
}

func (pl *Editor) CopyFieldsFrom(frm tree.Node) {
	fr := frm.(*Editor)
	pl.Frame.CopyFieldsFrom(&fr.Frame)
	pl.PlotStyle = fr.PlotStyle
	pl.setTable(fr.table)
}

// NewSubPlot returns a [Editor] with its own separate [core.Toolbar],
// suitable for a tab or other element that is not the main plot.
func NewSubPlot(parent ...tree.Node) *Editor {
	fr := core.NewFrame(parent...)
	tb := core.NewToolbar(fr)
	pl := NewEditor(fr)
	fr.Styler(func(s *styles.Style) {
		s.Direction = styles.Column
		s.Grow.Set(1, 1)
	})
	tb.Maker(pl.MakeToolbar)
	return pl
}

// ToolbarPlace specifies where the toolbar of an editor made by
// [NewEditorWithBar] goes: [TopBar] or [InlineBar].
type ToolbarPlace bool

const (
	// TopBar puts the toolbar of the editor in the top bar of its [core.Body].
	TopBar ToolbarPlace = true

	// InlineBar puts the toolbar of the editor in its parent, before it.
	InlineBar ToolbarPlace = false
)

// NewEditorWithBar returns a new [Editor] in the given parent, with its
// toolbar: in the top bar of the parent if topBar is [TopBar] and the parent
// is a [core.Body], or else ([InlineBar]) in the parent, before the editor.
func NewEditorWithBar(parent tree.Node, topBar ToolbarPlace) *Editor {
	if b, ok := parent.(*core.Body); ok && topBar == TopBar {
		pl := NewEditor(b)
		b.AddTopBar(func(bar *core.Frame) {
			core.NewToolbar(bar).Maker(pl.MakeToolbar)
		})
		return pl
	}
	tb := core.NewToolbar(parent)
	pl := NewEditor(parent)
	tb.Maker(pl.MakeToolbar)
	return pl
}

func (pl *Editor) Init() {
	pl.Frame.Init()

	pl.PlotStyle.Defaults()

	pl.Styler(func(s *styles.Style) {
		s.Grow.Set(1, 1)
		if pl.SizeClass() == core.SizeCompact {
			s.Direction = styles.Column
		}
	})

	pl.OnShow(func(e events.Event) {
		pl.UpdatePlot()
	})

	pl.Updater(func() {
		if pl.table != nil {
			pl.plotStyleFromTable()
		}
	})
	// the buttons for all the columns are in a header that stays in place,
	// above the columns, which scroll
	tree.AddChildAt(pl, "columns-panel", func(w *core.Frame) {
		w.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Background = colors.Scheme.SurfaceContainerLow
			if w.SizeClass() == core.SizeCompact {
				s.Grow.Set(1, 0)
			} else {
				s.Grow.Set(0, 1)
			}
		})
		tree.AddChildAt(w, "columns-header", func(w *core.Frame) {
			pl.makeColumnsHeader(w)
		})
		tree.AddChildAt(w, "separator", func(w *core.Separator) {})
		tree.AddChildAt(w, "columns", func(w *core.Frame) {
			pl.columnsFrame = w
			w.Styler(func(s *styles.Style) {
				s.Direction = styles.Column
				if w.SizeClass() != core.SizeCompact {
					s.Grow.Set(1, 1) // the width of the header too
					s.Overflow.Y = styles.OverflowAuto
				}
			})
			w.Maker(pl.makeColumns)
		})
	})
	tree.AddChildAt(pl, "plot", func(w *Plot) {
		pl.plotWidget = w
		w.Plot = pl.plot
		w.Styler(func(s *styles.Style) {
			s.Grow.Set(1, 1)
		})
	})
}

// setTable sets the table to view and does UpdatePlot.
func (pl *Editor) setTable(tab *table.Table) *Editor {
	pl.table = tab
	pl.UpdatePlot()
	return pl
}

// SetTable sets the table to a new view of given table,
// and does UpdatePlot.
func (pl *Editor) SetTable(tab *table.Table) *Editor {
	pl.table = nil
	pl.Update() // reset
	pl.table = table.NewView(tab)
	pl.UpdatePlot()
	pl.Update() // update to new table
	return pl
}

// SetSlice sets the table to a [table.NewSliceTable] from the given slice.
// Optional styler functions are used for each struct field in sequence,
// and any can contain global plot style. See [BasicStylers] for example.
func (pl *Editor) SetSlice(sl any, stylers ...func(s *plot.Style)) *Editor {
	dt, err := table.NewSliceTable(sl)
	errors.Log(err)
	if dt == nil {
		return nil
	}
	mx := min(dt.NumColumns(), len(stylers))
	for i := range mx {
		plot.SetStyler(dt.Columns.Values[i], stylers[i])
	}
	return pl.SetTable(dt)
}

// SaveSVG saves the plot to an SVG file.
func (pl *Editor) SaveSVG(fname core.Filename) { //types:add
	plt := pl.plotWidget.Plot
	mp := plt.PaintBox.Min
	plt.PaintBox = plt.PaintBox.Sub(mp)
	ptr := paint.NewPainter(math32.FromPoint(plt.PaintBox.Size()))
	ptr.UnitContext = pl.Styles.UnitContext // preserve DPI from current
	sv := paint.RenderToSVG(plt.Draw(ptr))
	err := os.WriteFile(string(fname), sv, 0666)
	plt.PaintBox = plt.PaintBox.Add(mp)
	if err != nil {
		core.ErrorSnackbar(pl, err)
	}
	pl.svgFile = fname
}

// SavePDF saves the plot to a PDF file.
func (pl *Editor) SavePDF(fname core.Filename) { //types:add
	plt := pl.plotWidget.Plot
	mp := plt.PaintBox.Min
	plt.PaintBox = plt.PaintBox.Sub(mp)
	ptr := paint.NewPainter(math32.FromPoint(plt.PaintBox.Size()))
	ptr.UnitContext = pl.Styles.UnitContext // preserve DPI from current
	pd := paint.RenderToPDF(plt.Draw(ptr))
	err := os.WriteFile(string(fname), pd, 0666)
	plt.PaintBox = plt.PaintBox.Add(mp)
	if err != nil {
		core.ErrorSnackbar(pl, err)
	}
}

// SaveImage saves the current plot as an image (e.g., png).
func (pl *Editor) SaveImage(fname core.Filename) { //types:add
	plt := pl.plotWidget.Plot
	mp := plt.PaintBox.Min
	plt.PaintBox = plt.PaintBox.Sub(mp)
	err := pl.plotWidget.Plot.SaveImage(string(fname))
	plt.PaintBox = plt.PaintBox.Add(mp)
	if err != nil {
		core.ErrorSnackbar(pl, err)
	}
}

// SaveCSV saves the Table data to a csv (comma-separated values) file with headers (any delim)
func (pl *Editor) SaveCSV(fname core.Filename, delim tensor.Delims) { //types:add
	pl.table.SaveCSV(fsx.Filename(fname), delim, table.Headers)
	pl.dataFile = fname
}

// SaveAll saves the current plot to a png, svg, and the data to a tsv -- full save
// Any extension is removed and appropriate extensions are added
func (pl *Editor) SaveAll(fname core.Filename) { //types:add
	fn := string(fname)
	fn = strings.TrimSuffix(fn, filepath.Ext(fn))
	pl.SaveCSV(core.Filename(fn+".tsv"), tensor.Tab)
	pl.SaveImage(core.Filename(fn + ".png"))
	pl.SaveSVG(core.Filename(fn + ".svg"))
	pl.SavePDF(core.Filename(fn + ".pdf"))
}

// OpenCSV opens the Table data from a csv (comma-separated values) file (or any delim)
func (pl *Editor) OpenCSV(filename core.Filename, delim tensor.Delims) { //types:add
	dt := table.New()
	dt.OpenCSV(fsx.Filename(filename), delim)
	pl.dataFile = filename
	pl.SetTable(dt)
}

// OpenFS opens the Table data from a csv (comma-separated values) file (or any delim)
// from the given filesystem.
func (pl *Editor) OpenFS(fsys fs.FS, filename core.Filename, delim tensor.Delims) {
	dt := table.New()
	dt.OpenFS(fsys, string(filename), delim)
	pl.SetTable(dt)
}

// GoUpdatePlot updates the display based on current Indexed view into table.
// This version can be called from goroutines. It does Sequential() on
// the [table.Table], under the assumption that it is used for tracking a
// the latest updates of a running process.
func (pl *Editor) GoUpdatePlot() {
	if pl == nil || pl.This == nil {
		return
	}
	if core.TheApp.Platform() == system.Web {
		time.Sleep(time.Millisecond) // critical to prevent hanging!
	}
	if !pl.IsVisible() || pl.table == nil || pl.inPlot {
		return
	}
	pl.Scene.AsyncLock()
	pl.table.Sequential()
	pl.genPlot()
	pl.NeedsRender()
	pl.Scene.AsyncUnlock()
}

// UpdatePlot updates the display based on current Indexed view into table.
// It does not automatically update the [table.Table] unless it is
// nil or out date.
func (pl *Editor) UpdatePlot() {
	if pl == nil || pl.This == nil {
		return
	}
	if pl.table == nil || pl.inPlot {
		return
	}
	if len(pl.Children) != 2 { // || len(pl.Columns) != pl.table.NumColumns() { // todo:
		pl.Update()
	}
	if pl.table.NumRows() == 0 {
		pl.table.Sequential()
	}
	pl.genPlot()
}

// genPlot generates a new plot from the current table.
// It surrounds operation with InPlot true / false to prevent multiple updates
func (pl *Editor) genPlot() {
	if pl.inPlot {
		slog.Error("plot: in plot already") // note: this never seems to happen -- could probably nuke
		return
	}
	pl.inPlot = true
	if pl.table == nil {
		pl.inPlot = false
		return
	}
	if len(pl.table.Indexes) == 0 {
		pl.table.Sequential()
	} else {
		lsti := pl.table.Indexes[pl.table.NumRows()-1]
		if lsti >= pl.table.Columns.Rows { // out of date
			pl.table.Sequential()
		}
	}
	var err error
	pl.plot, err = plot.NewTablePlot(pl.table)
	if pl.plot != nil && pl.plot.Style.ShowErrors && err != nil {
		core.ErrorSnackbar(pl, fmt.Errorf("%s: %w", pl.PlotStyle.Title, err))
	}
	if pl.plot != nil {
		pl.plotWidget.SetPlot(pl.plot)
		// } else {
		// errors.Log(fmt.Errorf("%s: nil plot: %w", pl.PlotStyle.Title, err))
	}
	// pl.plotWidget.updatePlot()
	pl.plotWidget.NeedsRender()
	pl.inPlot = false
}

// allColumnsOff turns all columns off.
func (pl *Editor) allColumnsOff() {
	for _, cli := range pl.columnsFrame.Children {
		cl := cli.(*core.Frame)
		sw := cl.Child(0).(*core.Switch)
		sw.SetChecked(false)
		sw.SendChange()
	}
	pl.Update()
}

// setColumnsByName turns columns on or off if their name contains
// the given string.
func (pl *Editor) setColumnsByName(nameContains string, on bool) { //types:add
	for _, cli := range pl.columnsFrame.Children {
		cl := cli.(*core.Frame)
		if !strings.Contains(cl.Name, nameContains) {
			continue
		}
		sw := cl.Child(0).(*core.Switch)
		sw.SetChecked(on)
		sw.SendChange()
	}
	pl.Update()
}

// makeColumnsHeader makes the buttons for all the columns, in the given
// header above them.
func (pl *Editor) makeColumnsHeader(w *core.Frame) {
	tree.AddChild(w, func(w *core.Button) {
		w.SetText("Clear").SetIcon(icons.ClearAll).SetType(core.ButtonAction)
		w.SetTooltip("Turn all columns off")
		w.OnClick(func(e events.Event) {
			pl.allColumnsOff()
		})
	})
	tree.AddChild(w, func(w *core.Button) {
		w.SetText("Search").SetIcon(icons.Search).SetType(core.ButtonAction)
		w.SetTooltip("Select columns by column name")
		w.OnClick(func(e events.Event) {
			core.CallFunc(pl, pl.setColumnsByName)
		})
	})
}

// makeColumns makes the Plans for columns
func (pl *Editor) makeColumns(p *tree.Plan) {
	hasSplit := false // split uses different color styling
	colorIdx := 0     // index for color sequence -- skips various types
	if pl.table == nil {
		return
	}
	if pl.columns == nil {
		pl.columns = map[string]*editorColumn{}
	}
	maps.DeleteFunc(pl.columns, func(name string, ec *editorColumn) bool {
		return pl.table.Columns.IndexByKey(name) < 0
	})
	xcol, nfloat := pl.defaultXColumn()
	intColorIdx := nfloat // the colors of floats are the same as without ints
	for ci, cl := range pl.table.Columns.Values {
		cnm := pl.table.Columns.Keys[ci]
		tree.AddAt(p, cnm, func(w *core.Frame) {
			psty := plot.GetStylers(cl)
			cidx := &colorIdx
			if !reflectx.KindIsFloat(cl.DataType()) {
				cidx = &intColorIdx
			}
			cst, mods, clr := pl.defaultColumnStyle(cl, cnm == xcol, cidx, &hasSplit, psty)
			pl.columns[cnm] = &editorColumn{style: cst, mods: mods, isString: cl.IsString()}
			isSplit := cst.Role == plot.Split
			stys := psty
			stys.Add(func(s *plot.Style) {
				mf := modFields(mods)
				errors.Log(reflectx.CopyFields(s, cst, mf...))
				errors.Log(reflectx.CopyFields(&s.Plot, &pl.PlotStyle, modFields(pl.plotStyleModified)...))
				// bars are filled with the color of the column
				if !mods["Line.Fill"] && s.PlotterOrDefault(cl.IsString(), &pl.PlotStyle) == plots.BarType {
					s.Line.Fill = s.Line.Color
				}
			})
			plot.SetStyler(cl, stys...)

			w.Styler(func(s *styles.Style) {
				s.CenterAll()
			})
			tree.AddChild(w, func(w *core.Switch) {
				w.SetType(core.SwitchCheckbox).SetTooltip("Turn this column on or off")
				w.Styler(func(s *styles.Style) {
					s.Color = clr
				})
				tree.AddChildInit(w, "stack", func(w *core.Frame) {
					f := func(name string) {
						tree.AddChildInit(w, name, func(w *core.Icon) {
							w.Styler(func(s *styles.Style) {
								s.Color = clr
							})
						})
					}
					f("icon-on")
					f("icon-off")
					f("icon-indeterminate")
				})
				w.OnChange(func(e events.Event) {
					mods["On"] = true
					cst.On = w.IsChecked()
					pl.UpdatePlot()
				})
				w.Updater(func() {
					xaxis := cst.Role == plot.X //  || cp.Column == pl.Options.Legend
					w.SetState(xaxis, states.Disabled, states.Indeterminate)
					if xaxis {
						cst.On = false
					} else {
						w.SetChecked(cst.On)
					}
					if cst.Role == plot.Split {
						isSplit = true
						hasSplit = true // update global flag
					} else {
						if isSplit && cst.Role != plot.Split {
							isSplit = false
							hasSplit = false
						}
					}
				})
			})
			tree.AddChild(w, func(w *core.Button) {
				tt := "[Edit all styling options for this column] " + metadata.Doc(cl)
				w.SetText(cnm).SetType(core.ButtonAction).SetTooltip(tt)
				w.AddContextMenu(func(m *core.Scene, pos image.Point) {
					bt := core.NewButton(m).SetText("Set X")
					bt.SetTooltip("Use this column as the X axis of the plot")
					bt.SetEnabled(pl.xAxisColumn() != cnm)
					bt.OnClick(func(e events.Event) {
						pl.SetXAxis(cnm)
					})
				})
				w.OnClick(func(e events.Event) {
					update := func() {
						if core.TheApp.Platform().IsMobile() {
							pl.Update()
							return
						}
						// we must be async on multi-window platforms since
						// it is coming from a separate window
						pl.AsyncLock()
						pl.Update()
						pl.AsyncUnlock()
					}
					d := core.NewBody(cnm + " style properties")
					fm := core.NewForm(d).SetStruct(cst)
					fm.Modified = mods
					fm.OnChange(func(e events.Event) {
						update()
					})
					d.AddTopBar(func(bar *core.Frame) {
						core.NewToolbar(bar).Maker(func(p *tree.Plan) {
							tree.Add(p, func(w *core.Button) {
								w.SetText("Set X")
								w.SetTooltip("Use this column as the X axis of the plot")
								w.OnClick(func(e events.Event) {
									if core.TheApp.Platform().IsMobile() {
										pl.SetXAxis(cnm)
									} else { // the editor is in another window
										pl.AsyncLock()
										pl.SetXAxis(cnm)
										pl.AsyncUnlock()
									}
									fm.Update()
									w.Update()
								})
								w.Updater(func() {
									w.SetEnabled(pl.xAxisColumn() != cnm)
								})
							})
						})
					})
					d.RunWindowDialog(pl)
				})
			})
		})
	}
}

// xAxisColumn returns the name of the column that is the X axis of
// the plot: the [plot.XAxisStyle.Column] of [Editor.PlotStyle] if set,
// and otherwise the last column with the X role, as in [plot.NewTablePlot].
func (pl *Editor) xAxisColumn() string {
	if xcol := pl.PlotStyle.XAxis.Column; xcol != "" {
		return xcol
	}
	xcol := ""
	for _, name := range pl.table.Columns.Keys {
		if ec, ok := pl.columns[name]; ok && ec.style.Role == plot.X {
			xcol = name
		}
	}
	return xcol
}

// SetPlotType sets the type of plot of all the columns that hold numbers
// (not strings) to the given type, e.g., to switch the plot from XY to Bar:
// the [plot.PlotStyle.Plotter] of [Editor.PlotStyle], which is the default
// for the columns, whose own [plot.Style.Plotter] is cleared so that they
// use it. See [plot.SetPlotters] for the version for a table.
func (pl *Editor) SetPlotType(plotType plot.PlotterName) { //types:add
	pl.PlotStyle.Plotter = plotType
	if pl.plotStyleModified == nil {
		pl.plotStyleModified = map[string]bool{}
	}
	pl.plotStyleModified["Plotter"] = true
	for _, ec := range pl.columns {
		if ec.isString || ec.style.Plotter == "" {
			continue
		}
		ec.style.Plotter = ""
		ec.mods["Plotter"] = true
	}
	pl.UpdatePlot()
	pl.Update()
}

// SetXAxis sets the column of the given name to be the X axis of the
// plot: its role is set to X, and the [plot.XAxisStyle.Column] of
// [Editor.PlotStyle] is set to it. Any other column with the X role
// is changed to the Y role, and turned off. See [plot.SetXAxis] for
// the version for a table.
func (pl *Editor) SetXAxis(column string) { //types:add
	xc, ok := pl.columns[column]
	if !ok {
		core.ErrorSnackbar(pl, fmt.Errorf("SetXAxis: column %q not found", column))
		return
	}
	for name, ec := range pl.columns {
		if name != column && ec.style.Role == plot.X {
			ec.style.Role = plot.Y
			ec.mods["Role"] = true
			ec.style.On = false // as its switch shows; it can be turned on
			ec.mods["On"] = true
		}
	}
	xc.style.Role = plot.X
	xc.mods["Role"] = true
	pl.PlotStyle.XAxis.Column = column
	if pl.plotStyleModified == nil {
		pl.plotStyleModified = map[string]bool{}
	}
	pl.plotStyleModified["XAxis.Column"] = true
	pl.UpdatePlot()
	pl.Update()
}

// defaultXColumn returns the name of the column that is the X axis of
// the plot by default, which is the only column with the X role:
// the [plot.XAxisStyle.Column] of the plot style if set, and otherwise
// the last column with the X role set by its stylers, as in
// [plot.NewTablePlot], and otherwise the last column of integers
// with no role set. It also returns the number of columns of floats.
func (pl *Editor) defaultXColumn() (string, int) {
	xaxis := pl.PlotStyle.XAxis.Column
	xrole, xint := "", ""
	nfloat := 0
	for ci, cl := range pl.table.Columns.Values {
		cnm := pl.table.Columns.Keys[ci]
		st := plot.NewStyle()
		if sty := plot.GetStylers(cl); sty != nil {
			sty.Run(st)
		}
		if xaxis == "" {
			xaxis = st.Plot.XAxis.Column
		}
		isfloat := reflectx.KindIsFloat(cl.DataType())
		if isfloat {
			nfloat++
		}
		switch {
		case st.Role == plot.X:
			xrole = cnm
		case st.Role == plot.NoRole && !isfloat && !cl.IsString():
			xint = cnm
		}
	}
	switch {
	case xaxis != "" && pl.table.Columns.IndexByKey(xaxis) >= 0:
		return xaxis, nfloat
	case xrole != "":
		return xrole, nfloat
	}
	return xint, nfloat
}

// defaultColumnStyle initializes the column style with any existing stylers
// plus additional general defaults, returning the initially modified field names.
// isX is whether the column is the X axis of the plot: only that column has
// the X role, and other columns of numbers have the Y role.
func (pl *Editor) defaultColumnStyle(cl tensor.Values, isX bool, colorIdx *int, hasSplit *bool, psty plot.Stylers) (*plot.Style, map[string]bool, image.Image) {
	cst := &plot.Style{}
	cst.Defaults()
	if psty != nil {
		psty.Run(cst)
	}
	if cst.On && cst.Role == plot.Split {
		*hasSplit = true
	}
	mods := map[string]bool{}
	isString := cl.IsString()
	// the plotter of numbers is the default, the PlotStyle Plotter
	if cst.Plotter == "" && isString {
		cst.Plotter = plot.PlotterName(plots.LabelsType)
		mods["Plotter"] = true
	}
	switch {
	case isX:
		if cst.Role != plot.X {
			cst.Role = plot.X
			mods["Role"] = true
		}
	case cst.Role == plot.NoRole || (cst.Role == plot.X && cst.Group == ""):
		// an X column of a group is the X data of its plotter
		mods["Role"] = true
		if isString {
			cst.Role = plot.Label
		} else {
			cst.Role = plot.Y
		}
	}
	clr := cst.Line.Color
	ptyp := cst.PlotterOrDefault(isString, &pl.PlotStyle)
	isBox := ptyp == plots.BoxType || ptyp == plots.XBoxType
	if isBox && cst.Line.Fill != nil {
		if cst.Line.Fill != colors.Scheme.Surface { // white boxes: the line color
			clr = cst.Line.Fill // the color of a box is its fill
		}
	} else if clr == colors.Scheme.OnSurface {
		if cst.Role == plot.Y && !isString {
			clr = colors.Uniform(colors.Spaced(*colorIdx))
			(*colorIdx)++
			switch {
			case *hasSplit:
			case isBox: // the lines of boxes stay in contrast with the fill
				cst.Line.Fill = clr
				mods["Line.Fill"] = true
			default:
				cst.Line.Color = clr
				mods["Line.Color"] = true
				cst.Point.Color = clr
				mods["Point.Color"] = true
				cst.Point.Fill = clr
				mods["Point.Fill"] = true
			}
		}
	}
	return cst, mods, clr
}

func (pl *Editor) plotStyleFromTable() {
	if pl.plotStyleModified != nil { // already set
		return
	}
	pst := &pl.PlotStyle
	mods := map[string]bool{}
	pl.plotStyleModified = mods
	tst := &plot.Style{}
	tst.Defaults()
	tst.Plot.Defaults()
	for _, cl := range pl.table.Columns.Values {
		stl := plot.GetStylers(cl)
		if stl == nil {
			continue
		}
		stl.Run(tst)
	}
	*pst = tst.Plot
	if pst.PointsOn == plot.Default {
		pst.PointsOn = plot.Off
		mods["PointsOn"] = true
	}
	if pst.Title == "" {
		pst.Title = metadata.Name(pl.table)
		if pst.Title != "" {
			mods["Title"] = true
		}
	}
}

// modFields returns the modified fields as field paths using . separators
func modFields(mods map[string]bool) []string {
	fns := maps.Keys(mods)
	rf := make([]string, 0, len(fns))
	for _, f := range fns {
		if !mods[f] {
			continue
		}
		fc := strings.ReplaceAll(f, " • ", ".")
		rf = append(rf, fc)
	}
	slices.Sort(rf)
	return rf
}

func (pl *Editor) MakeToolbar(p *tree.Plan) {
	if pl.table == nil {
		return
	}
	tree.Add(p, func(w *core.Button) {
		w.SetIcon(icons.PanTool).
			SetTooltip("toggle the ability to zoom and pan the view").OnClick(func(e events.Event) {
			pw := pl.plotWidget
			pw.SetReadOnly(!pw.IsReadOnly())
			pw.Restyle()
		})
	})
	// tree.Add(p, func(w *core.Button) {
	// 	w.SetIcon(icons.ArrowForward).
	// 		SetTooltip("turn on select mode for selecting Plot elements").
	// 		OnClick(func(e events.Event) {
	// 			fmt.Println("this will select select mode")
	// 		})
	// })
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Button) {
		w.SetText("Update").SetIcon(icons.Update).
			SetTooltip("update fully redraws display, reflecting any new settings etc").
			OnClick(func(e events.Event) {
				pl.UpdatePlot()
				pl.Update()
			})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Style").SetIcon(icons.Settings).
			SetTooltip("Style for how the plot is rendered").
			OnClick(func(e events.Event) {
				d := core.NewBody("Plot style")
				fm := core.NewForm(d).SetStruct(&pl.PlotStyle)
				fm.Modified = pl.plotStyleModified
				fm.OnChange(func(e events.Event) {
					pl.GoUpdatePlot()
				})
				d.RunWindowDialog(pl)
			})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(pl.SetPlotType).SetText("Type").SetIcon(icons.BarChart)
		w.SetTooltip("Set the type of plot of all the columns that hold numbers, e.g., XY or Bar (the Plotter of the plot Style)")
		w.OnFirst(events.Click, func(e events.Event) { // start with the current type
			ptyp := pl.PlotStyle.Plotter
			if ptyp == "" {
				ptyp = plots.XYType
			}
			w.Args[0].SetValue(ptyp)
		})
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Stats").SetIcon(icons.QueryStats).SetMenu(pl.makeStatsMenu)
		w.SetTooltip("Make a statistical plot of the data in a new plot: the mean and error of columns, a histogram, a box plot, or a violin plot")
	})
	tree.Add(p, func(w *core.Button) {
		w.SetText("Table").SetIcon(icons.Edit).
			SetTooltip("open a Table window of the data").
			OnClick(func(e events.Event) {
				d := core.NewBody(pl.Name + " Data")
				tv := tensorcore.NewTable(d).SetTable(pl.table)
				d.AddTopBar(func(bar *core.Frame) {
					core.NewToolbar(bar).Maker(tv.MakeToolbar)
				})
				d.RunWindowDialog(pl)
			})
	})
	tree.Add(p, func(w *core.Separator) {})

	tree.Add(p, func(w *core.Button) {
		w.SetText("Save").SetIcon(icons.Save).SetMenu(func(m *core.Scene, pos image.Point) {
			core.NewFuncButton(m).SetFunc(pl.SaveSVG).SetIcon(icons.Save)
			core.NewFuncButton(m).SetFunc(pl.SavePDF).SetIcon(icons.Save)
			core.NewFuncButton(m).SetFunc(pl.SaveImage).SetIcon(icons.Save)
			core.NewFuncButton(m).SetFunc(pl.SaveCSV).SetIcon(icons.Save)
			core.NewSeparator(m)
			core.NewFuncButton(m).SetFunc(pl.SaveAll).SetIcon(icons.Save)
		})
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(pl.OpenCSV).SetIcon(icons.Open)
	})
	tree.Add(p, func(w *core.Separator) {})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(pl.table.FilterString).SetText("Filter").SetIcon(icons.FilterAlt)
		w.SetAfterFunc(pl.UpdatePlot)
	})
	tree.Add(p, func(w *core.FuncButton) {
		w.SetFunc(pl.table.Sequential).SetText("Unfilter").SetIcon(icons.FilterAltOff)
		w.SetAfterFunc(pl.UpdatePlot)
	})
}

func (pl *Editor) SizeFinal() {
	pl.Frame.SizeFinal()
	pl.UpdatePlot()
}
