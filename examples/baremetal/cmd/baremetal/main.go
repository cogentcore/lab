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
	log.Printf("Submitting: %v", in.Path)
	job := s.bm.Submit(in.Source, in.Path, in.Script, in.ResultsGlob, in.Files)
	return baremetal.JobToPB(job), nil
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
	return nil
}
