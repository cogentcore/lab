// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/logx"
	"cogentcore.org/core/cli"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/text/terminal"
	"cogentcore.org/lab/examples/baremetal"
	pb "cogentcore.org/lab/examples/baremetal/baremetal"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Config struct {
	// The server port number.
	Port int `default:"8585"`

	// Window runs the server in a terminal in its own window, which
	// stops the server when it is closed. Simmer starts it this way.
	Window bool
}

type server struct {
	pb.UnimplementedBareMetalServer
	bm *baremetal.BareMetal
}

// Submit adds a new Active job with given parameters.
func (s *server) Submit(_ context.Context, in *pb.Submission) (*pb.Job, error) {
	slog.Info("Submitting Job", "Source:", in.Source, "Path:", in.Path)
	job := s.bm.Submit(in.Source, in.Path, in.Script, in.ResultsGlob, in.Files)
	return baremetal.JobToPB(job), nil
}

// JobStatus gets current job data for given job id(s).
// An empty list returns all of the active jobs.
func (s *server) JobStatus(_ context.Context, in *pb.JobIDList) (*pb.JobList, error) {
	slog.Info("JobStatus")
	jobs := s.bm.JobStatus(baremetal.JobIDsFromPB(in.JobID)...)
	return baremetal.JobsToPB(jobs), nil
}

// CancelJobs cancels list of job IDs. Returns error for jobs not found.
func (s *server) CancelJobs(_ context.Context, in *pb.JobIDList) (*pb.Error, error) {
	slog.Info("CancelJobs")
	err := s.bm.CancelJobs(baremetal.JobIDsFromPB(in.JobID)...)
	estr := ""
	if err != nil {
		estr = err.Error()
	}
	return &pb.Error{Error: estr}, nil
}

// FetchResults gets job results back from server for given job id(s).
// Results are available as job.Results as a compressed tar file.
func (s *server) FetchResults(_ context.Context, in *pb.JobIDList) (*pb.JobList, error) {
	slog.Info("FetchResults")
	jobs, err := s.bm.FetchResults("", baremetal.JobIDsFromPB(in.JobID)...)
	errors.Log(err)
	return baremetal.JobsToPB(jobs), nil
}

// UpdateJobs pings the server to run its updates.
// This happens automatically very 10 seconds but this is for the impatient.
func (s *server) UpdateJobs(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	slog.Info("UpdateJobs")
	s.bm.UpdateJobs()
	return &emptypb.Empty{}, nil
}

// RecoverJob
func (s *server) RecoverJob(_ context.Context, in *pb.Job) (*pb.Job, error) {
	slog.Info("RecoverJob")
	job, err := s.bm.RecoverJob(baremetal.JobFromPB(in))
	errors.Log(err)
	return baremetal.JobToPB(job), err
}

func main() {
	opts := cli.DefaultOptions("baremetal", "Bare metal server for job running on bare servers over ssh")
	cfg := &Config{}
	cli.Run(opts, cfg, Run)
}

func Run(cfg *Config) error {
	if cfg.Window {
		return runWindow(cfg)
	}
	// the log is the output of the server, so it shows the info messages,
	// which cli does not by default (see [logx.LevelFromFlags]).
	logx.UserLevel = min(logx.UserLevel, slog.LevelInfo)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return errors.Log(fmt.Errorf("failed to listen: %v", err))
	}
	s := grpc.NewServer()
	bms := &server{}
	bms.bm = baremetal.NewBareMetal()
	bms.bm.Init()
	bms.bm.StartBGUpdates()
	pb.RegisterBareMetalServer(s, bms)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return errors.Log(fmt.Errorf("failed to serve: %v", err))
	}
	bms.bm.Interactive()
	return nil
}

// runWindow runs the server, as this executable without -window, in a
// terminal in its own window, unless the server is running already.
// The context menu of the terminal has items to edit the config of
// the servers and to restart the server.
func runWindow(cfg *Config) error {
	cl := baremetal.NewClient()
	cl.Host = fmt.Sprintf("localhost:%d", cfg.Port)
	if cl.Running() {
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return errors.Log(err)
	}
	b := core.NewBody("BareMetal")
	tm := terminal.NewTerminal(b).SetCommand(exe, "-port", strconv.Itoa(cfg.Port))
	var restarting atomic.Bool // no exit message while restarting
	tm.OnExit(func(err error) {
		if !restarting.Load() {
			core.MessageSnackbar(tm, "the baremetal server has exited: press a key to restart it")
		}
	})
	tm.AddContextMenu(func(m *core.Scene, pos image.Point) {
		core.NewButton(m).SetText("Config").SetIcon(icons.Settings).
			OnClick(func(e events.Event) {
				editConfig(tm)
			})
		core.NewButton(m).SetText("Restart").SetIcon(icons.Refresh).
			OnClick(func(e events.Event) {
				restarting.Store(true)
				restartServer(tm, cl, func() { restarting.Store(false) })
			})
	})
	b.RunMainWindow()
	return nil
}

// editConfig edits the config of the servers in a form,
// saving it to the config.toml file when the form is closed,
// if it has been changed.
func editConfig(tm *terminal.Terminal) {
	file := baremetal.NewBareMetal().ConfigFile()
	cf := &baremetal.Config{}
	if err := cf.Open(file); err != nil && !errors.Is(err, fs.ErrNotExist) {
		core.ErrorSnackbar(tm, err)
	}
	orig, _ := cf.Bytes() // to save only when changed
	d := core.NewBody("BareMetal config")
	core.NewText(d).SetType(core.TextSupporting).SetText(file + ": restart the server to use the changes")
	core.NewForm(d).SetStruct(cf)
	d.OnClose(func(e events.Event) {
		if b, err := cf.Bytes(); err == nil && bytes.Equal(b, orig) {
			return
		}
		if err := os.MkdirAll(filepath.Dir(file), 0777); err != nil {
			core.ErrorSnackbar(tm, err)
			return
		}
		if err := cf.Save(file); err != nil {
			core.ErrorSnackbar(tm, err)
			return
		}
		core.MessageSnackbar(tm, "Saved the config: restart the server to use it")
	})
	d.RunWindowDialog(tm)
}

// restartServer restarts the server in the given terminal, after the
// running one has ended and closed its port, calling done after.
func restartServer(tm *terminal.Terminal, cl *baremetal.Client, done func()) {
	tm.Stop()
	go func() {
		defer done()
		for range 50 {
			if !tm.IsRunning() && !cl.Running() {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		em := tm.Emulator() // after the exit message, as on a key
		em.Lock()
		em.WriteString("\r\n")
		em.Unlock()
		tm.AsyncLock()
		errors.Log(tm.Restart())
		tm.AsyncUnlock()
	}()
}
