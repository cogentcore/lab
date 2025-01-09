// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"runtime"

	"log/slog"

	"cogentcore.org/core/base/timer"
	"cogentcore.org/lab/tensor"
)

//go:generate gosl

func init() {
	// must lock main thread for gpu!
	runtime.LockOSThread()
}

func main() {
	GPUInit()

	// n := 10
	n := 16_000_000 // max for macbook M*
	// n := 200_000

	UseGPU = false

	Seed = make([]Seeds, 1)

	dataCU := tensor.NewUint32(n, 2)
	dataGU := tensor.NewUint32(n, 2)

	dataCF := tensor.NewFloat32(n, NVars)
	dataGF := tensor.NewFloat32(n, NVars)

	Uints = dataCU
	Floats = dataCF

	cpuTmr := timer.Time{}
	cpuTmr.Start()
	RunOneCompute(n)
	cpuTmr.Stop()

	UseGPU = true
	Uints = dataGU
	Floats = dataGF

	gpuFullTmr := timer.Time{}
	gpuFullTmr.Start()

	ToGPUTensorStrides()
	ToGPU(SeedVar, FloatsVar, UintsVar)

	gpuTmr := timer.Time{}
	gpuTmr.Start()

	RunCompute(n)
	gpuTmr.Stop()

	RunDone(FloatsVar, UintsVar)
	gpuFullTmr.Stop()

	anyDiffEx := false
	anyDiffTol := false
	mx := min(n, 5)
	fmt.Printf("Index\tDif(Ex,Tol)\t   CPU   \t  then GPU\n")
	for i := 0; i < n; i++ {
		smEx, smTol := IsSame(dataCU, dataGU, dataCF, dataGF, i)
		if !smEx {
			anyDiffEx = true
		}
		if !smTol {
			anyDiffTol = true
		}
		if i > mx {
			continue
		}
		exS := " "
		if !smEx {
			exS = "*"
		}
		tolS := " "
		if !smTol {
			tolS = "*"
		}
		fmt.Printf("%d\t%s %s\t%s\n\t\t%s\n", i, exS, tolS, String(dataCU, dataCF, i), String(dataGU, dataGF, i))
	}
	fmt.Printf("\n")

	if anyDiffEx {
		slog.Error("Differences between CPU and GPU detected at Exact level (excludes Gauss)")
	}
	if anyDiffTol {
		slog.Error("Differences between CPU and GPU detected at Tolerance level", "tolerance", Tol)
	}

	cpu := cpuTmr.Total
	gpu := gpuTmr.Total
	fmt.Printf("N: %d\t CPU: %v\t GPU: %v\t Full: %v\t CPU/GPU: %6.4g\n", n, cpu, gpu, gpuFullTmr.Total, float64(cpu)/float64(gpu))

	GPURelease()
}
