// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotosl

import (
	"fmt"
	"reflect"
	"strings"
)

// GenKernelHeader returns the novel generated WGSL kernel code
// for given kernel, which goes at the top of the resulting file.
func (st *State) GenKernelHeader(sy *System, kn *Kernel) string {
	var b strings.Builder
	b.WriteString("// Code generated by \"gosl\"; DO NOT EDIT\n")
	b.WriteString("// kernel: " + kn.Name + "\n\n")

	for gi, gp := range sy.Groups {
		if gp.Doc != "" {
			b.WriteString("// " + gp.Doc + "\n")
		}
		str := "storage, read_write"
		if gp.Uniform {
			str = "uniform"
		}
		for vi, vr := range gp.Vars {
			if vr.Doc != "" {
				b.WriteString("// " + vr.Doc + "\n")
			}
			b.WriteString(fmt.Sprintf("@group(%d) @binding(%d)\n", gi, vi))
			b.WriteString(fmt.Sprintf("var<%s> %s: ", str, vr.Name))
			b.WriteString(fmt.Sprintf("array<%s>;\n", vr.SLType()))
		}
	}

	b.WriteString("\n")
	b.WriteString("@compute @workgroup_size(64, 1, 1)\n")
	// todo: conditional on different index dimensionality
	b.WriteString("fn main(@builtin(global_invocation_id) idx: vec3<u32>) {\n")
	b.WriteString(fmt.Sprintf("\t%s(idx.x);\n", kn.Name))
	b.WriteString("}\n")
	b.WriteString(st.GenTensorFuncs(sy))
	return b.String()
}

// GenTensorFuncs returns the generated WGSL code
// for indexing the tensors in given system.
func (st *State) GenTensorFuncs(sy *System) string {
	var b strings.Builder

	done := make(map[string]bool)

	for _, gp := range sy.Groups {
		for _, vr := range gp.Vars {
			if !vr.Tensor {
				continue
			}
			typ := vr.SLType()
			fn := vr.IndexFunc()
			if _, ok := done[fn]; ok {
				continue
			}
			done[fn] = true
			tconv := ""
			switch vr.TensorKind {
			case reflect.Float32:
				tconv = "bitcast<u32>("
			case reflect.Int32:
				tconv = "u32("
			}
			tend := ""
			if tconv != "" {
				tend = ")"
			}
			b.WriteString("\nfn " + fn + "(")
			nd := vr.TensorDims
			for d := range nd {
				b.WriteString(fmt.Sprintf("s%d: %s, ", d, typ))
			}
			for d := range nd {
				b.WriteString(fmt.Sprintf("i%d: u32", d))
				if d < nd-1 {
					b.WriteString(", ")
				}
			}
			b.WriteString(") -> u32 {\n\treturn ")
			b.WriteString(fmt.Sprintf("u32(%d)", vr.TensorDims))
			for d := range nd {
				b.WriteString(fmt.Sprintf(" + %ss%d%s * i%d", tconv, d, tend, d))
			}
			b.WriteString(";\n}\n")
		}
	}
	return b.String()
}
