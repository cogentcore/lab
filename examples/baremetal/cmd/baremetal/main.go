// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/logx"
	"cogentcore.org/core/cli"
	"cogentcore.org/lab/examples/baremetal"
	pb "cogentcore.org/lab/examples/baremetal/baremetal"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Config struct {
	// The server port number.
	Port int `default:"8585"`
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
	return &pb.Error{Error: err.Error()}, nil
}

// FetchResults gets job results back from server for given job id(s).
// Results are available as job.Results as a compressed tar file.
func (s *server) FetchResults(_ context.Context, in *pb.JobIDList) (*pb.JobList, error) {
	slog.Info("FetchResults")
	jobs, err := s.bm.FetchResults(baremetal.JobIDsFromPB(in.JobID)...)
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

func main() {
	logx.UserLevel = slog.LevelInfo
	opts := cli.DefaultOptions("baremetal", "Bare metal server for job running on bare servers over ssh")
	cfg := &Config{}
	cli.Run(opts, cfg, Run)
}

func Run(cfg *Config) error {
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
