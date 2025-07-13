// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotosl

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/stack"
)

// System represents a ComputeSystem, and its kernels and variables.
type System struct {
	Name string

	// Kernels are the kernels using this compute system.
	Kernels map[string]*Kernel

	// Groups are the variables for this compute system.
	Groups []*Group

	// NTensors is the number of tensor vars.
	NTensors int
}

func NewSystem(name string) *System {
	sy := &System{Name: name}
	sy.Kernels = make(map[string]*Kernel)
	return sy
}

// Kernel represents a kernel function, which is the basis for
// each wgsl generated code file.
type Kernel struct {
	Name string

	Args string

	// Filename is the name of the kernel shader file, e.g., shaders/Compute.wgsl
	Filename string

	// function code
	FuncCode string

	// Lines is full shader code
	Lines [][]byte

	// ReadWriteVars are variables marked as read-write for current kernel.
	ReadWriteVars map[string]bool

	Atomics  map[string]*Var
	VarsUsed map[string]*Var
}

// Var represents one global system buffer variable.
type Var struct {
	Name string

	// Group number that we are in.
	Group int

	// Binding number within group.
	Binding int

	// comment docs about this var.
	Doc string

	// Type of variable: either []Type or F32, U32 for tensors
	Type string

	// ReadOnly indicates that this variable is never read back from GPU,
	// specified by the gosl:read-only property in the variable comments.
	// It is important to optimize GPU memory usage to indicate this.
	ReadOnly bool

	// ReadOrWrite indicates that this variable defaults to ReadOnly
	// but is also flagged as read-write in some cases. It is registered
	// as read_write in the gpu ComputeSystem, but processed as ReadOnly
	// by default except for kernels that declare it as read-write.
	ReadOrWrite bool

	// True if a tensor type
	Tensor bool

	// Number of dimensions
	TensorDims int

	// data kind of the tensor
	TensorKind reflect.Kind

	// index of tensor in list of tensor variables, for indexing.
	TensorIndex int

	// NBuffs is the number of buffers to allocate to this variable; default is 1,
	// which provides direct access. Otherwise, a wrapper function is generated
	// that allows > max buffer size total storage.
	// The index still has to fit in a uint32 variable, so 4g max value.
	// Assuming 4 bytes per element, that means a total of 16g max total storage.
	// The Config.MaxBufferSize (set at compile time, defaults to 2g) determines
	// how many buffers: if 2g, then 16 / 2 = 8 max buffers.
	NBuffs int
}

func (vr *Var) SetTensorKind() {
	kindStr := strings.TrimPrefix(vr.Type, "tensor.")
	kind := reflect.Float32
	switch kindStr {
	case "Float32":
		kind = reflect.Float32
	case "Uint32":
		kind = reflect.Uint32
	case "Int32":
		kind = reflect.Int32
	default:
		errors.Log(fmt.Errorf("gosl: variable %q type is not supported: %q", vr.Name, kindStr))
	}
	vr.TensorKind = kind
}

// SLType returns the WGSL type string
func (vr *Var) SLType() string {
	if vr.Tensor {
		switch vr.TensorKind {
		case reflect.Float32:
			return "f32"
		case reflect.Int32:
			return "i32"
		case reflect.Uint32:
			return "u32"
		}
	} else {
		return vr.Type[2:]
	}
	return ""
}

// GoType returns the Go type string for tensors
func (vr *Var) GoType() string {
	if vr.Tensor {
		switch vr.TensorKind {
		case reflect.Float32:
			return "float32"
		case reflect.Int32:
			return "int32"
		case reflect.Uint32:
			return "uint32"
		}
	}
	return ""
}

// IndexFunc returns the tensor index function name
func (vr *Var) IndexFunc() string {
	return fmt.Sprintf("Index%dD", vr.TensorDims)
}

// IndexStride returns the tensor stride variable reference
func (vr *Var) IndexStride(dim int) string {
	return fmt.Sprintf("TensorStrides[%d]", vr.TensorIndex*10+dim)
}

// Group represents one variable group.
type Group struct {
	Name string

	// comment docs about this group
	Doc string

	// Uniform indicates a uniform group; else default is Storage.
	Uniform bool

	Vars []*Var
}

// File has contents of a file as lines of bytes.
type File struct {
	Name  string
	Lines [][]byte
}

// GetGlobalVar holds GetVar expression, to Set variable back when done.
type GetGlobalVar struct {
	// global variable
	Var *Var

	// name of temporary variable
	TmpVar string

	// index passed to the Get function
	IdxExpr ast.Expr

	// rw override
	ReadWrite bool
}

// State holds the current Go -> WGSL processing state.
type State struct {
	// Config options.
	Config *Config

	// path to shaders/imports directory.
	ImportsDir string

	// name of the package
	Package string

	// GoFiles are all the files with gosl content in current directory.
	GoFiles map[string]*File

	// GoVarsFiles are all the files with gosl:vars content in current directory.
	// These must be processed first!  they are moved from GoFiles to here.
	GoVarsFiles map[string]*File

	// GoImports has all the imported files.
	GoImports map[string]map[string]*File

	// ImportPackages has short package names, to remove from go code
	// so everything lives in same main package.
	ImportPackages map[string]bool

	// Systems has the kernels and variables for each system.
	// There is an initial "Default" system when system is not specified.
	Systems map[string]*System

	// GetFuncs is a map of GetVar, SetVar function names for global vars.
	GetFuncs map[string]*Var

	// SLImportFiles are all the extracted and translated WGSL files in shaders/imports,
	// which are copied into the generated shader kernel files.
	SLImportFiles []*File

	// generated Go GPU gosl.go file contents
	GPUFile File

	// ExcludeMap is the compiled map of functions to exclude in Go -> WGSL translation.
	ExcludeMap map[string]bool

	// GetVarStack is a stack per function definition of GetVar variables
	// that need to be set at the end.
	GetVarStack stack.Stack[map[string]*GetGlobalVar]

	// GetFuncGraph is true if getting the function graph (first pass).
	GetFuncGraph bool

	// CurKernel is the current Kernel for second pass processing.
	CurKernel *Kernel

	// KernelFuncs are the list of functions to include for current kernel.
	KernelFuncs map[string]*Function

	// FuncGraph is the call graph of functions, for dead code elimination
	FuncGraph map[string]*Function
}

func (st *State) Init(cfg *Config) {
	st.Config = cfg
	st.GoImports = make(map[string]map[string]*File)
	st.Systems = make(map[string]*System)
	st.ExcludeMap = make(map[string]bool)
	ex := strings.Split(cfg.Exclude, ",")
	for _, fn := range ex {
		st.ExcludeMap[fn] = true
	}
	st.Systems["Default"] = NewSystem("Default")
}

func (st *State) Run() error {
	if gomod := os.Getenv("GO111MODULE"); gomod == "off" {
		err := errors.New("gosl only works in go modules mode, but GO111MODULE=off")
		return err
	}
	if st.Config.Output == "" {
		st.Config.Output = "shaders"
	}

	st.ProjectFiles() // get list of all files, recursively gets imports etc.
	if len(st.GoFiles) == 0 {
		return nil
	}

	st.ImportsDir = filepath.Join(st.Config.Output, "imports")
	os.MkdirAll(st.Config.Output, 0755)
	os.MkdirAll(st.ImportsDir, 0755)
	RemoveGenFiles(st.Config.Output)
	RemoveGenFiles(st.ImportsDir)

	st.ExtractFiles()   // get .go from project files
	st.ExtractImports() // get .go from imports
	st.TranslateDir("./" + st.ImportsDir)

	st.GenGPU()

	return nil
}

// System returns the given system by name, making if not made.
// if name is empty, "Default" is used.
func (st *State) System(sysname string) *System {
	if sysname == "" {
		sysname = "Default"
	}
	sy, ok := st.Systems[sysname]
	if ok {
		return sy
	}
	sy = NewSystem(sysname)
	st.Systems[sysname] = sy
	return sy
}

// GlobalVar returns global variable of given name, if found.
func (st *State) GlobalVar(vrnm string) *Var {
	if st == nil {
		return nil
	}
	if st.Systems == nil {
		return nil
	}
	for _, sy := range st.Systems {
		for _, gp := range sy.Groups {
			for _, vr := range gp.Vars {
				if vr.Name == vrnm {
					return vr
				}
			}
		}
	}
	return nil
}

// VarIsReadWrite returns true if var of name is set as read-write
// for current kernel.
func (st *State) VarIsReadWrite(vrnm string) bool {
	if st.CurKernel == nil {
		return false
	}
	if _, rw := st.CurKernel.ReadWriteVars[vrnm]; rw {
		return true
	}
	return false
}

// GetTempVar returns temp var for global variable of given name, if found.
func (st *State) GetTempVar(vrnm string) *GetGlobalVar {
	if st == nil || st.GetVarStack == nil {
		return nil
	}
	nv := len(st.GetVarStack)
	for i := nv - 1; i >= 0; i-- {
		gvars := st.GetVarStack[i]
		if gv, ok := gvars[vrnm]; ok {
			return gv
		}
	}
	return nil
}

// VarsAdded is called when a set of vars has been added; update relevant maps etc.
func (st *State) VarsAdded() {
	st.GetFuncs = make(map[string]*Var)
	for _, sy := range st.Systems {
		tensorIdx := 0
		for gi, gp := range sy.Groups {
			vn := 0
			if gi == 0 { // leave room for TensorStrides
				vn++
			}
			for _, vr := range gp.Vars {
				vr.Group = gi
				vr.Binding = vn
				if vr.Tensor {
					vr.TensorIndex = tensorIdx
					tensorIdx++
					if vr.NBuffs > 1 {
						vn += vr.NBuffs
					} else {
						vn++
					}
					continue
				}
				st.GetFuncs["Get"+vr.Name] = vr
				vn++
			}
		}
		sy.NTensors = tensorIdx
	}
}
