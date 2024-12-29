// Code generated by "goal build"; DO NOT EDIT.
//line sim.goal:1
// Copyright (c) 2024, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:generate core generate

import (
	"math/rand/v2"

	"cogentcore.org/core/core"
	"cogentcore.org/lab/lab"
	"cogentcore.org/lab/plot"
	"cogentcore.org/lab/stats/stats"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensorfs"
)

// Times are the looping time levels for running and statistics.
type Times int32 //enums:enum

const (
	Trial Times = iota
	Epoch
	Run
)

// LoopPhase is the phase of loop processing for given time.
type LoopPhase int32 //enums:enum

const (
	// Start is the start of the loop: resets accumulated stats, initializes.
	Start LoopPhase = iota

	// Step is each iteration of the loop.
	Step
)

type Sim struct {
	// Root is the root data dir.
	Root *tensorfs.Node

	// Config has config data.
	Config *tensorfs.Node

	// Stats has all stats data.
	Stats *tensorfs.Node

	// Current has current value of all stats
	Current *tensorfs.Node

	// StatFuncs are statistics functions, per stat, handles everything.
	StatFuncs []func(ltime Times, lphase LoopPhase)

	// Counters are current values of counters: normally in looper.
	Counters [TimesN]int
}

// ConfigAll configures the sim
func (ss *Sim) ConfigAll() {
	ss.Root, _ = tensorfs.NewDir("Root")
	ss.Config = ss.Root.Dir("Config")
	mx := tensorfs.Value[int](ss.Config, "Max", int(TimesN)).(*tensor.Int)
	mx.Set1D(5, int(Trial))
	mx.Set1D(4, int(Epoch))
	mx.Set1D(3, int(Run))
	// todo: failing - assigns 3 to all
	// # mx[Trial] = 5
	// # mx[Epoch] = 4
	// # mx[Run] = 3
	ss.ConfigStats()
}

func (ss *Sim) AddStat(f func(ltime Times, lphase LoopPhase)) {
	ss.StatFuncs = append(ss.StatFuncs, f)
}

func (ss *Sim) RunStats(ltime Times, lphase LoopPhase) {
	for _, sf := range ss.StatFuncs {
		sf(ltime, lphase)
	}
}

func (ss *Sim) ConfigStats() {
	ss.Stats = ss.Root.Dir("Stats")
	ss.Current = ss.Stats.Dir("Current")
	ctrs := []Times{Run, Epoch, Trial}
	for _, ctr := range ctrs {
		ss.AddStat(func(ltime Times, lphase LoopPhase) {
			if ltime > ctr { // don't record counter for time above it
				return
			}
			name := ctr.String() // name of stat = counter
			timeDir := ss.Stats.Dir(ltime.String())
			tsr := tensorfs.Value[int](timeDir, name)
			if lphase == Start {
				tsr.SetNumRows(0)
				if ps := plot.GetStylersFrom(tsr); ps == nil {
					ps.Add(func(s *plot.Style) {
						s.Range.SetMin(0)
					})
					plot.SetStylersTo(tsr, ps)
				}
				return
			}
			ctv := ss.Counters[ctr]
			tensorfs.Scalar[int](ss.Current, name).SetInt1D(ctv, 0)
			tsr.AppendRowInt(ctv)
		})
	}
	// note: it is essential to only have 1 per func
	// so generic names can be used for everything.
	ss.AddStat(func(ltime Times, lphase LoopPhase) {
		name := "SSE"
		timeDir := ss.Stats.Dir(ltime.String())
		tsr := timeDir.Float64(name)
		if lphase == Start {
			tsr.SetNumRows(0)
			if ps := plot.GetStylersFrom(tsr); ps == nil {
				ps.Add(func(s *plot.Style) {
					s.Range.SetMin(0).SetMax(1)
					s.On = true
				})
				plot.SetStylersTo(tsr, ps)
			}
			return
		}
		switch ltime {
		case Trial:
			stat := rand.Float64()
			tensorfs.Scalar[float64](ss.Current, name).SetFloat(stat, 0)
			tsr.AppendRowFloat(stat)
		case Epoch:
			subd := ss.Stats.Dir((ltime - 1).String())
			stat := stats.StatMean.Call(subd.Float64(name))
			tsr.AppendRow(stat)
		case Run:
			subd := ss.Stats.Dir((ltime - 1).String())
			stat := stats.StatMean.Call(subd.Float64(name))
			tsr.AppendRow(stat)
		}
	})
	ss.AddStat(func(ltime Times, lphase LoopPhase) {
		name := "Err"
		timeDir := ss.Stats.Dir(ltime.String())
		tsr := tensorfs.Value[float64](timeDir, name)
		if lphase == Start {
			tsr.SetNumRows(0)
			if ps := plot.GetStylersFrom(tsr); ps == nil {
				ps.Add(func(s *plot.Style) {
					s.Range.SetMin(0).SetMax(1)
					s.On = true
				})
				plot.SetStylersTo(tsr, ps)
			}
			return
		}
		switch ltime {
		case Trial:
			sse := tensorfs.Scalar[float64](ss.Current, "SSE").Float1D(0)
			stat := 1.0
			if sse < 0.5 {
				stat = 0
			}
			tensorfs.Scalar[float64](ss.Current, name).SetFloat(stat, 0)
			tsr.AppendRowFloat(stat)
		case Epoch:
			subd := ss.Stats.Dir((ltime - 1).String())
			stat := stats.StatMean.Call(subd.Value(name))
			tsr.AppendRow(stat)
		case Run:
			subd := ss.Stats.Dir((ltime - 1).String())
			stat := stats.StatMean.Call(subd.Value(name))
			tsr.AppendRow(stat)
		}
	})
}

func (ss *Sim) Run() {
	mx := ss.Config.Value("Max").(*tensor.Int)
	nrun := mx.Value1D(int(Run))
	nepc := mx.Value1D(int(Epoch))
	ntrl := mx.Value1D(int(Trial))
	ss.RunStats(Run, Start)
	for run := range nrun {
		ss.Counters[Run] = run
		ss.RunStats(Epoch, Start)
		for epc := range nepc {
			ss.Counters[Epoch] = epc
			ss.RunStats(Trial, Start)
			for trl := range ntrl {
				ss.Counters[Trial] = trl
				ss.RunStats(Trial, Step)
			}
			ss.RunStats(Epoch, Step)
		}
		ss.RunStats(Run, Step)
	}
	// todo: could do final analysis here
	// alldt := ss.Logs.Item("AllTrials").GetDirTable(nil)
	// dir := ss.Logs.Dir("Stats")
	// stats.TableGroups(dir, alldt, "Run", "Epoch", "Trial")
	// sts := []string{"SSE", "AvgSSE", "TrlErr"}
	// stats.TableGroupStats(dir, stats.StatMean, alldt, sts...)
	// stats.TableGroupStats(dir, stats.StatSem, alldt, sts...)
}

func main() {
	ss := &Sim{}
	ss.ConfigAll()
	ss.Run()

	b, _ := lab.NewBasicWindow(ss.Root, "Root")
	b.RunWindow()
	core.Wait()
}
