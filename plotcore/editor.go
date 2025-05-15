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
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/system"
	"cogentcore.org/core/tree"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/plot/plots"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorcore"
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
			pl.plotStyleFromTable(pl.table)
		}
	})
	tree.AddChildAt(pl, "columns", func(w *core.Frame) {
		pl.columnsFrame = w
		w.Styler(func(s *styles.Style) {
			s.Direction = styles.Column
			s.Background = colors.Scheme.SurfaceContainerLow
			if w.SizeClass() == core.SizeCompact {
				s.Grow.Set(1, 0)
			} else {
				s.Grow.Set(0, 1)
				s.Overflow.Y = styles.OverflowAuto
			}
		})
		w.Maker(pl.makeColumns)
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
// and any can contain global plot style.
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

// SaveSVG saves the plot to an svg -- first updates to ensure that plot is current
func (pl *Editor) SaveSVG(fname core.Filename) { //types:add
	plt := pl.plotWidget.Plot
	err := plt.SaveSVG(string(fname))
	if err != nil {
		core.ErrorSnackbar(pl, err)
	}
	pl.svgFile = fname
}

// SaveImage saves the current plot as an image (e.g., png).
func (pl *Editor) SaveImage(fname core.Filename) { //types:add
	err := pl.plotWidget.Plot.SaveImage(string(fname))
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
		if lsti >= pl.table.NumRows() { // out of date
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

const plotColumnsHeaderN = 3

// allColumnsOff turns all columns off.
func (pl *Editor) allColumnsOff() {
	fr := pl.columnsFrame
	for i, cli := range fr.Children {
		if i < plotColumnsHeaderN {
			continue
		}
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
	fr := pl.columnsFrame
	for i, cli := range fr.Children {
		if i < plotColumnsHeaderN {
			continue
		}
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

// makeColumns makes the Plans for columns
func (pl *Editor) makeColumns(p *tree.Plan) {
	tree.Add(p, func(w *core.Frame) {
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
	})
	hasSplit := false // split uses different color styling
	colorIdx := 0     // index for color sequence -- skips various types
	tree.Add(p, func(w *core.Separator) {})
	if pl.table == nil {
		return
	}
	for ci, cl := range pl.table.Columns.Values {
		cnm := pl.table.Columns.Keys[ci]
		tree.AddAt(p, cnm, func(w *core.Frame) {
			psty := plot.GetStylers(cl)
			cst, mods, clr := pl.defaultColumnStyle(cl, ci, &colorIdx, &hasSplit, psty)
			isSplit := cst.Role == plot.Split
			stys := psty
			stys.Add(func(s *plot.Style) {
				mf := modFields(mods)
				errors.Log(reflectx.CopyFields(s, cst, mf...))
				errors.Log(reflectx.CopyFields(&s.Plot, &pl.PlotStyle, modFields(pl.plotStyleModified)...))
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
					// d.AddTopBar(func(bar *core.Frame) {
					// 	core.NewToolbar(bar).Maker(func(p *tree.Plan) {
					// 		tree.Add(p, func(w *core.Button) {
					// 			w.SetText("Set x-axis").OnClick(func(e events.Event) {
					// 				pl.Options.XAxis = cp.Column
					// 				update()
					// 			})
					// 		})
					// 		tree.Add(p, func(w *core.Button) {
					// 			w.SetText("Set legend").OnClick(func(e events.Event) {
					// 				pl.Options.Legend = cp.Column
					// 				update()
					// 			})
					// 		})
					// 	})
					// })
					d.RunWindowDialog(pl)
				})
			})
		})
	}
}

// defaultColumnStyle initializes the column style with any existing stylers
// plus additional general defaults, returning the initially modified field names.
func (pl *Editor) defaultColumnStyle(cl tensor.Values, ci int, colorIdx *int, hasSplit *bool, psty plot.Stylers) (*plot.Style, map[string]bool, image.Image) {
	cst := &plot.Style{}
	cst.Defaults()
	if psty != nil {
		psty.Run(cst)
	}
	if cst.On && cst.Role == plot.Split {
		*hasSplit = true
	}
	mods := map[string]bool{}
	isfloat := reflectx.KindIsFloat(cl.DataType())
	if cst.Plotter == "" {
		if isfloat {
			cst.Plotter = plot.PlotterName(plots.XYType)
			mods["Plotter"] = true
		} else if cl.IsString() {
			cst.Plotter = plot.PlotterName(plots.LabelsType)
			mods["Plotter"] = true
		}
	}
	if cst.Role == plot.NoRole {
		mods["Role"] = true
		if isfloat {
			cst.Role = plot.Y
		} else if cl.IsString() {
			cst.Role = plot.Label
		} else {
			cst.Role = plot.X
		}
	}
	clr := cst.Line.Color
	if clr == colors.Scheme.OnSurface {
		if cst.Role == plot.Y && isfloat {
			clr = colors.Uniform(colors.Spaced(*colorIdx))
			(*colorIdx)++
			if !*hasSplit {
				cst.Line.Color = clr
				mods["Line.Color"] = true
				cst.Point.Color = clr
				mods["Point.Color"] = true
				cst.Point.Fill = clr
				mods["Point.Fill"] = true
				if cst.Plotter == plots.BarType {
					cst.Line.Fill = clr
					mods["Line.Fill"] = true
				}
			}
		}
	}
	return cst, mods, clr
}

func (pl *Editor) plotStyleFromTable(dt *table.Table) {
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
		if mods[f] == false {
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
		w.SetText("Save").SetIcon(icons.Save).SetMenu(func(m *core.Scene) {
			core.NewFuncButton(m).SetFunc(pl.SaveSVG).SetIcon(icons.Save)
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

func (pt *Editor) SizeFinal() {
	pt.Frame.SizeFinal()
	pt.UpdatePlot()
}
