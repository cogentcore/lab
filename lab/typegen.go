// Code generated by "core generate"; DO NOT EDIT.

package lab

import (
	"io/fs"

	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/lab.Basic", IDName: "basic", Doc: "Basic is a basic data browser with the files as the left panel,\nand the Tabber as the right panel.", Embeds: []types.Field{{Name: "Frame"}, {Name: "Browser"}}})

// NewBasic returns a new [Basic] with the given optional parent:
// Basic is a basic data browser with the files as the left panel,
// and the Tabber as the right panel.
func NewBasic(parent ...tree.Node) *Basic { return tree.New[Basic](parent...) }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/lab.Browser", IDName: "browser", Doc: "Browser holds all the elements of a data browser, for browsing data\neither on an OS filesystem or as a tensorfs virtual data filesystem.\nIt supports the automatic loading of [goal] scripts as toolbar actions to\nperform pre-programmed tasks on the data, to create app-like functionality.\nScripts are ordered alphabetically and any leading #- prefix is automatically\nremoved from the label, so you can use numbers to specify a custom order.\nIt is not a [core.Widget] itself, and is intended to be incorporated into\na [core.Frame] widget, potentially along with other custom elements.\nSee [Basic] for a basic implementation.", Directives: []types.Directive{{Tool: "types", Directive: "add", Args: []string{"-setters"}}}, Methods: []types.Method{{Name: "UpdateFiles", Doc: "UpdateFiles Updates the files list.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "UpdateScripts", Doc: "UpdateScripts updates the Scripts and updates the toolbar.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}}, Fields: []types.Field{{Name: "FS", Doc: "FS is the filesystem, if browsing an FS."}, {Name: "DataRoot", Doc: "DataRoot is the path to the root of the data to browse."}, {Name: "StartDir", Doc: "StartDir is the starting directory, where the app was originally started."}, {Name: "ScriptsDir", Doc: "ScriptsDir is the directory containing scripts for toolbar actions.\nIt defaults to DataRoot/dbscripts"}, {Name: "Scripts", Doc: "Scripts are interpreted goal scripts (via yaegi) to automate\nroutine tasks."}, {Name: "Interpreter", Doc: "Interpreter is the interpreter to use for running Browser scripts.\nis of type: *goal/interpreter.Interpreter but can't use that directly\nto avoid importing goal unless needed. Import [labscripts] if needed."}, {Name: "Files", Doc: "Files is the [DataTree] tree browser of the tensorfs or files."}, {Name: "Tabs", Doc: "Tabs is the [Tabs] element managing tabs of data views."}, {Name: "Toolbar", Doc: "Toolbar is the top-level toolbar for the browser, if used."}, {Name: "Splits", Doc: "Splits is the overall [core.Splits] for the browser."}}})

// SetFS sets the [Browser.FS]:
// FS is the filesystem, if browsing an FS.
func (t *Browser) SetFS(v fs.FS) *Browser { t.FS = v; return t }

// SetDataRoot sets the [Browser.DataRoot]:
// DataRoot is the path to the root of the data to browse.
func (t *Browser) SetDataRoot(v string) *Browser { t.DataRoot = v; return t }

// SetStartDir sets the [Browser.StartDir]:
// StartDir is the starting directory, where the app was originally started.
func (t *Browser) SetStartDir(v string) *Browser { t.StartDir = v; return t }

// SetScriptsDir sets the [Browser.ScriptsDir]:
// ScriptsDir is the directory containing scripts for toolbar actions.
// It defaults to DataRoot/dbscripts
func (t *Browser) SetScriptsDir(v string) *Browser { t.ScriptsDir = v; return t }

// SetFiles sets the [Browser.Files]:
// Files is the [DataTree] tree browser of the tensorfs or files.
func (t *Browser) SetFiles(v *DataTree) *Browser { t.Files = v; return t }

// SetTabs sets the [Browser.Tabs]:
// Tabs is the [Tabs] element managing tabs of data views.
func (t *Browser) SetTabs(v *Tabs) *Browser { t.Tabs = v; return t }

// SetToolbar sets the [Browser.Toolbar]:
// Toolbar is the top-level toolbar for the browser, if used.
func (t *Browser) SetToolbar(v *core.Toolbar) *Browser { t.Toolbar = v; return t }

// SetSplits sets the [Browser.Splits]:
// Splits is the overall [core.Splits] for the browser.
func (t *Browser) SetSplits(v *core.Splits) *Browser { t.Splits = v; return t }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/lab.DataTree", IDName: "data-tree", Doc: "DataTree is the databrowser version of [filetree.Tree],\nwhich provides the Tabber to show data editors.", Embeds: []types.Field{{Name: "Tree"}}, Fields: []types.Field{{Name: "Tabber", Doc: "Tabber is the [Tabber] for this tree."}}})

// NewDataTree returns a new [DataTree] with the given optional parent:
// DataTree is the databrowser version of [filetree.Tree],
// which provides the Tabber to show data editors.
func NewDataTree(parent ...tree.Node) *DataTree { return tree.New[DataTree](parent...) }

// SetTabber sets the [DataTree.Tabber]:
// Tabber is the [Tabber] for this tree.
func (t *DataTree) SetTabber(v Tabber) *DataTree { t.Tabber = v; return t }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/lab.FileNode", IDName: "file-node", Doc: "FileNode is databrowser version of FileNode for FileTree", Methods: []types.Method{{Name: "EditFiles", Doc: "EditFiles calls EditFile on selected files", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "PlotFiles", Doc: "PlotFiles calls PlotFile on selected files", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "GridFiles", Doc: "GridFiles calls GridFile on selected files", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "DiffDirs", Doc: "DiffDirs displays a browser with differences between two selected directories", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}}, Embeds: []types.Field{{Name: "Node"}}})

// NewFileNode returns a new [FileNode] with the given optional parent:
// FileNode is databrowser version of FileNode for FileTree
func NewFileNode(parent ...tree.Node) *FileNode { return tree.New[FileNode](parent...) }

var _ = types.AddType(&types.Type{Name: "cogentcore.org/lab/lab.Tabs", IDName: "tabs", Doc: "Tabs implements the [Tabber] interface.", Embeds: []types.Field{{Name: "Tabs"}}})

// NewTabs returns a new [Tabs] with the given optional parent:
// Tabs implements the [Tabber] interface.
func NewTabs(parent ...tree.Node) *Tabs { return tree.New[Tabs](parent...) }
