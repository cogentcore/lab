// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotosl

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
)

// Function represents the call graph of functions
type Function struct {
	Name     string
	Funcs    map[string]*Function
	Atomics  map[string]*Var // variables that have atomic operations in this function
	VarsUsed map[string]*Var // all global variables referenced by this function.
}

func NewFunction(name string) *Function {
	return &Function{Name: name, Funcs: make(map[string]*Function)}
}

func (fn *Function) AddAtomic(vr *Var) {
	if fn.Atomics == nil {
		fn.Atomics = make(map[string]*Var)
	}
	fn.Atomics[vr.Name] = vr
}

func (fn *Function) AddVarUsed(vr *Var) {
	if fn.VarsUsed == nil {
		fn.VarsUsed = make(map[string]*Var)
	}
	fn.VarsUsed[vr.Name] = vr
}

// get or add a function of given name
func (st *State) RecycleFunc(name string) *Function {
	fn, ok := st.FuncGraph[name]
	if !ok {
		fn = NewFunction(name)
		st.FuncGraph[name] = fn
	}
	return fn
}

func getAllFuncs(f *Function, all map[string]*Function) {
	for fnm, fn := range f.Funcs {
		_, ok := all[fnm]
		if ok {
			continue
		}
		all[fnm] = fn
		getAllFuncs(fn, all)
	}
}

// AllFuncs returns aggregated list of all functions called be given function
func (st *State) AllFuncs(name string) map[string]*Function {
	fn, ok := st.FuncGraph[name]
	if !ok {
		fmt.Printf("gosl: ERROR kernel function named: %q not found\n", name)
		return nil
	}
	all := make(map[string]*Function)
	all[name] = fn
	getAllFuncs(fn, all)
	// cfs := maps.Keys(all)
	// sort.Strings(cfs)
	// for _, cfnm := range cfs {
	// 	fmt.Println("\t" + cfnm)
	// }
	return all
}

// VarsUsed returns all the atomic and used global variables
// used by the list of functions. Also the total number of used vars
// that includes the NBuffs counts.
func (st *State) VarsUsed(funcs map[string]*Function) (avars, uvars map[string]*Var, nvars int) {
	avars = make(map[string]*Var)
	uvars = make(map[string]*Var)
	for _, fn := range funcs {
		for vn, v := range fn.Atomics {
			avars[vn] = v
		}
		for vn, v := range fn.VarsUsed {
			uvars[vn] = v
		}
	}
	nvars = 1 // assume TensorStrides always
	for _, vr := range uvars {
		if vr.NBuffs > 1 {
			nvars += vr.NBuffs
		} else {
			nvars++
		}
	}
	return
}

func (st *State) PrintFuncGraph() {
	funs := maps.Keys(st.FuncGraph)
	sort.Strings(funs)
	for _, fname := range funs {
		fmt.Println(fname)
		fn := st.FuncGraph[fname]
		cfs := maps.Keys(fn.Funcs)
		sort.Strings(cfs)
		for _, cfnm := range cfs {
			fmt.Println("\t" + cfnm)
		}
	}
}
