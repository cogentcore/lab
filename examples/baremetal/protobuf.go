// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative baremetal/baremetal.proto

import (
	pb "cogentcore.org/lab/examples/baremetal/baremetal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JobToPB returns the protobuf version of given Job.
func JobToPB(job *Job) *pb.Job {
	pj := &pb.Job{ID: int64(job.ID), Status: pb.Status(job.Status), Source: job.Source, Path: job.Path, Script: job.Script, Files: job.Files, ResultsGlob: job.ResultsGlob, Results: job.Results, ServerName: job.ServerName, ServerGPU: int32(job.ServerGPU), PID: int64(job.PID)}
	pj.Submit = timestamppb.New(job.Submit)
	pj.Start = timestamppb.New(job.Start)
	pj.End = timestamppb.New(job.End)
	return pj
}

// JobFromPB returns a Job based on the protobuf version.
func JobFromPB(job *pb.Job) *Job {
	bj := &Job{ID: int(job.ID), Status: Status(job.Status), Source: job.Source, Path: job.Path, Script: job.Script, Files: job.Files, ResultsGlob: job.ResultsGlob, Results: job.Results, ServerName: job.ServerName, ServerGPU: int(job.ServerGPU), PID: int(job.PID)}
	bj.Submit = job.Submit.AsTime()
	bj.Start = job.Start.AsTime()
	bj.End = job.End.AsTime()
	return bj
}
