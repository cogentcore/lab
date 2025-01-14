// Code generated by 'yaegi extract cogentcore.org/lab/lab'. DO NOT EDIT.

package gui

import (
	"cogentcore.org/core/core"
	"cogentcore.org/lab/lab"
	"reflect"
)

func init() {
	Symbols["cogentcore.org/lab/lab/lab"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"AsDataTree":         reflect.ValueOf(lab.AsDataTree),
		"FirstComment":       reflect.ValueOf(lab.FirstComment),
		"IsTableFile":        reflect.ValueOf(lab.IsTableFile),
		"Lab":                reflect.ValueOf(&lab.Lab).Elem(),
		"LabBrowser":         reflect.ValueOf(&lab.LabBrowser).Elem(),
		"NewBasic":           reflect.ValueOf(lab.NewBasic),
		"NewBasicWindow":     reflect.ValueOf(lab.NewBasicWindow),
		"NewDataTree":        reflect.ValueOf(lab.NewDataTree),
		"NewDiffBrowserDirs": reflect.ValueOf(lab.NewDiffBrowserDirs),
		"NewFileNode":        reflect.ValueOf(lab.NewFileNode),
		"NewPlot":            reflect.ValueOf(lab.NewPlot),
		"NewTabs":            reflect.ValueOf(lab.NewTabs),
		"PromptOKCancel":     reflect.ValueOf(lab.PromptOKCancel),
		"PromptString":       reflect.ValueOf(lab.PromptString),
		"PromptStruct":       reflect.ValueOf(lab.PromptStruct),
		"TensorFS":           reflect.ValueOf(lab.TensorFS),
		"TrimOrderPrefix":    reflect.ValueOf(lab.TrimOrderPrefix),

		// type definitions
		"Basic":    reflect.ValueOf((*lab.Basic)(nil)),
		"Browser":  reflect.ValueOf((*lab.Browser)(nil)),
		"DataTree": reflect.ValueOf((*lab.DataTree)(nil)),
		"FileNode": reflect.ValueOf((*lab.FileNode)(nil)),
		"Tabber":   reflect.ValueOf((*lab.Tabber)(nil)),
		"Tabs":     reflect.ValueOf((*lab.Tabs)(nil)),
		"Treer":    reflect.ValueOf((*lab.Treer)(nil)),

		// interface wrapper definitions
		"_Tabber": reflect.ValueOf((*_cogentcore_org_lab_lab_Tabber)(nil)),
		"_Treer":  reflect.ValueOf((*_cogentcore_org_lab_lab_Treer)(nil)),
	}
}

// _cogentcore_org_lab_lab_Tabber is an interface wrapper for Tabber type
type _cogentcore_org_lab_lab_Tabber struct {
	IValue      interface{}
	WAsCoreTabs func() *core.Tabs
	WAsLab      func() *lab.Tabs
}

func (W _cogentcore_org_lab_lab_Tabber) AsCoreTabs() *core.Tabs { return W.WAsCoreTabs() }
func (W _cogentcore_org_lab_lab_Tabber) AsLab() *lab.Tabs       { return W.WAsLab() }

// _cogentcore_org_lab_lab_Treer is an interface wrapper for Treer type
type _cogentcore_org_lab_lab_Treer struct {
	IValue      interface{}
	WAsDataTree func() *lab.DataTree
}

func (W _cogentcore_org_lab_lab_Treer) AsDataTree() *lab.DataTree { return W.WAsDataTree() }
