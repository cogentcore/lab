// Code generated by "goal build"; DO NOT EDIT.
//line sim.goal:1
// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate

import (
	"math/rand/v2"

	"cogentcore.org/core/core"
	"cogentcore.org/core/plot/plotcore"
	"cogentcore.org/core/tensor"
	"cogentcore.org/core/tensor/databrowser"
	"cogentcore.org/core/tensor/datafs"
	"cogentcore.org/core/tensor/stats/stats"
)

type Times int32 //enums:enum

const (
	Trial Times = iota
	Epoch
	Run
)

type Sim struct {
	Root      *datafs.Data
	Config    *datafs.Data
	Stats     *datafs.Data
	StatFuncs []func(tm Times)
	Counters  [TimesN]int
}

// ConfigAll configures the sim
func (ss *Sim) ConfigAll() {
	ss.Root, _ = datafs.NewDir("Root")
	ss.Config, _ = ss.Root.Mkdir("Config")
	mx := datafs.Value[int](ss.Config, "Max", int(TimesN)).(*tensor.Int)
	mx.Set1D(5, int(Trial))
	mx.Set1D(4, int(Epoch))
	mx.Set1D(3, int(Run))
	// todo: failing - assigns 3 to all
	// # mx[Trial] = 5
	// # mx[Epoch] = 4
	// # mx[Run] = 3
	ss.ConfigStats()
}

func (ss *Sim) AddStat(f func(tm Times)) {
	ss.StatFuncs = append(ss.StatFuncs, f)
}

func (ss *Sim) RunStats(tm Times) {
	for _, sf := range ss.StatFuncs {
		sf(tm)
	}
}

func (ss *Sim) ConfigStats() {
	ss.Stats, _ = ss.Root.Mkdir("Stats")
	mx := ss.Config.Value("Max").(*tensor.Int)
	ctrs := []Times{Run, Epoch, Trial}
	for _, ctr := range ctrs {
		ss.AddStat(func(tm Times) {
			name := ctr.String()
			if tm > ctr {
				return
			}
			ctv := ss.Counters[ctr]
			mxi := mx.Value1D(int(tm))
			td := ss.Stats.RecycleDir(tm.String())
			cd := ss.Stats.RecycleDir("Current")
			datafs.Scalar[int](cd, name).SetInt1D(ctv, 0)
			tv := datafs.Value[int](td, name, mxi)
			tv.SetInt1D(ctv, ss.Counters[tm])
			if ps := plotcore.GetPlotStylers(tv.Metadata()); ps == nil {
				ps = plotcore.NewPlotStylers()
				ps.ColumnStyler(name, func(co *plotcore.ColumnOptions) {
					co.Range.FixMin = true
					co.On = false
				})
				if tm == ctr {
					ps.PlotStyler(func(po *plotcore.PlotOptions) {
						po.XAxis = name
					})
				}
				plotcore.SetPlotStylers(tv.Metadata(), ps)
			}
		})
	}
	ss.AddStat(func(tm Times) {
		sseName := "SSE"
		errName := "Err"
		td := ss.Stats.RecycleDir(tm.String())
		switch tm {
		case Trial:
			ctv := ss.Counters[tm]
			mxi := mx.Value1D(int(tm))
			cd := ss.Stats.RecycleDir("Current")
			sse := rand.Float32()
			terr := float32(1)
			if sse < 0.5 {
				terr = 0
			}
			datafs.Scalar[float32](cd, sseName).SetFloat(float64(sse), 0)
			datafs.Scalar[float32](cd, errName).SetFloat(float64(terr), 0)
			datafs.Value[float32](td, sseName, mxi).SetFloat1D(float64(sse), ctv)
			datafs.Value[float32](td, errName, mxi).SetFloat1D(float64(terr), ctv)
		case Epoch:
			ctv := ss.Counters[tm]
			mxi := mx.Value1D(int(tm))
			trld, _ := ss.Stats.Mkdir(Trial.String())
			sse := stats.StatMean.Call(trld.Value(sseName)).Float1D(0)
			terr := stats.StatMean.Call(trld.Value(errName)).Float1D(0)
			datafs.Value[float32](td, sseName, mxi).SetFloat1D(float64(sse), ctv)
			datafs.Value[float32](td, errName, mxi).SetFloat1D(float64(terr), ctv)
		}
	})
}

// // ConfigStats adds basic stats that we record for our simulation.
// func (ss *Sim) ConfigStats(dir *datafs.Data) *datafs.Data {
// 	stats, _ := dir.Mkdir("Stats")
// 	datafs.NewScalar[int](stats, "Run", "Epoch", "Trial") // counters
// 	datafs.NewScalar[string](stats, "TrialName")
// 	datafs.NewScalar[float32](stats, "SSE", "AvgSSE", "TrlErr")
// 	z1, key := plotcore.PlotColumnZeroOne()
// 	stats.SetMetaItems(key, z1, "AvgErr", "TrlErr")
// 	zmax, _ := plotcore.PlotColumnZeroOne()
// 	zmax.Range.FixMax = false
// 	stats.SetMetaItems(key, z1, "SSE")
// 	return stats
// }
//
// // ConfigLogs adds first-level logging of stats into tensors
// func (ss *Sim) ConfigLogs(dir *datafs.Data) *datafs.Data {
// 	logd, _ := dir.Mkdir("Log")
// 	trial := ss.ConfigTrialLog(logd)
// 	ss.ConfigAggLog(logd, "Epoch", trial, stats.StatMean, stats.StatSem, stats.StatMin)
// 	return logd
// }
//
// // ConfigTrialLog adds first-level logging of stats into tensors
// func (ss *Sim) ConfigTrialLog(dir *datafs.Data) *datafs.Data {
// 	logd, _ := dir.Mkdir("Trial")
// 	ntrial := ss.Config.Item("NTrial").AsInt()
// 	sitems := ss.Stats.ValuesFunc(nil)
// 	for _, st := range sitems {
// 		nm := st.Metadata().Name()
// 		lt := logd.NewOfType(nm, st.DataType(), ntrial)
// 		lt.Metadata().Copy(*st.Metadata()) // key affordance: we get meta data from source
// 		tensor.SetCalcFunc(lt, func() error {
// 			trl := ss.Stats.Item("Trial").AsInt()
// 			if st.IsString() {
// 				lt.SetStringRow(st.String1D(0), trl)
// 			} else {
// 				lt.SetFloatRow(st.Float1D(0), trl)
// 			}
// 			return nil
// 		})
// 	}
// 	alllogd, _ := dir.Mkdir("AllTrials")
// 	for _, st := range sitems {
// 		nm := st.Metadata().Name()
// 		// allocate full size
// 		lt := alllogd.NewOfType(nm, st.DataType(), ntrial*ss.Config.Item("NEpoch").AsInt()*ss.Config.Item("NRun").AsInt())
// 		lt.SetShapeSizes(0)                // then truncate to 0
// 		lt.Metadata().Copy(*st.Metadata()) // key affordance: we get meta data from source
// 		tensor.SetCalcFunc(lt, func() error {
// 			row := lt.DimSize(0)
// 			lt.SetShapeSizes(row + 1)
// 			if st.IsString() {
// 				lt.SetStringRow(st.String1D(0), row)
// 			} else {
// 				lt.SetFloatRow(st.Float1D(0), row)
// 			}
// 			return nil
// 		})
// 	}
// 	return logd
// }
//
// // ConfigAggLog adds a higher-level logging of lower-level into higher-level tensors
// func (ss *Sim) ConfigAggLog(dir *datafs.Data, level string, from *datafs.Data, aggs ...stats.Stats) *datafs.Data {
// 	logd, _ := dir.Mkdir(level)
// 	sitems := ss.Stats.ValuesFunc(nil)
// 	nctr := ss.Config.Item("N" + level).AsInt()
// 	for _, st := range sitems {
// 		if st.IsString() {
// 			continue
// 		}
// 		nm := st.Metadata().Name()
// 		src := from.Value(nm)
// 		if st.DataType() >= reflect.Float32 {
// 			// todo: pct correct etc
// 			dd, _ := logd.Mkdir(nm)
// 			for _, ag := range aggs { // key advantage of dir structure: multiple stats per item
// 				lt := dd.NewOfType(ag.String(), st.DataType(), nctr)
// 				lt.Metadata().Copy(*st.Metadata())
// 				tensor.SetCalcFunc(lt, func() error {
// 					stout := ag.Call(src)
// 					ctr := ss.Stats.Item(level).AsInt()
// 					lt.SetFloatRow(stout.FloatRow(0), ctr)
// 					return nil
// 				})
// 			}
// 		} else {
// 			lt := logd.NewOfType(nm, st.DataType(), nctr)
// 			lt.Metadata().Copy(*st.Metadata())
// 			tensor.SetCalcFunc(lt, func() error {
// 				v := st.Float1D(0)
// 				ctr := ss.Stats.Item(level).AsInt()
// 				lt.SetFloatRow(v, ctr)
// 				return nil
// 			})
// 		}
// 	}
// 	return logd
// }

func (ss *Sim) Run() {
	mx := ss.Config.Value("Max").(*tensor.Int)
	nrun := mx.Value1D(int(Run))
	nepc := mx.Value1D(int(Epoch))
	ntrl := mx.Value1D(int(Trial))
	for run := range nrun {
		ss.Counters[Run] = run
		for epc := range nepc {
			ss.Counters[Epoch] = epc
			for trl := range ntrl {
				ss.Counters[Trial] = trl
				ss.RunStats(Trial)
			}
			ss.RunStats(Epoch)
		}
		ss.RunStats(Run)
	}
	// alldt := ss.Logs.Item("AllTrials").GetDirTable(nil)
	// dir, _ := ss.Logs.Mkdir("Stats")
	// stats.TableGroups(dir, alldt, "Run", "Epoch", "Trial")
	// sts := []string{"SSE", "AvgSSE", "TrlErr"}
	// stats.TableGroupStats(dir, stats.StatMean, alldt, sts...)
	// stats.TableGroupStats(dir, stats.StatSem, alldt, sts...)
}

// func (ss *Sim) RunTrial(trl int) {
// 	ss.Stats.Item("TrialName").SetString("Trial_" + strconv.Itoa(trl))
// 	sse := rand.Float32()
// 	avgSSE := rand.Float32()
// 	ss.Stats.Item("SSE").SetFloat32(sse)
// 	ss.Stats.Item("AvgSSE").SetFloat32(avgSSE)
// 	trlErr := float32(1)
// 	if sse < 0.5 {
// 		trlErr = 0
// 	}
// 	ss.Stats.Item("TrlErr").SetFloat32(trlErr)
// 	ss.Logs.Item("Trial").CalcAll()
// 	ss.Logs.Item("AllTrials").CalcAll()
// }

// func (ss *Sim) EpochDone() {
// 	ss.Logs.Item("Epoch").CalcAll()
// }

func main() {
	ss := &Sim{}
	ss.ConfigAll()
	ss.Run()

	databrowser.NewBrowserWindow(ss.Root, "Root")
	core.Wait()
}
