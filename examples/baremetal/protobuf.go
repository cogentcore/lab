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

// JobsToPB returns the protobuf version of given Jobs list.
func JobsToPB(jobs []*Job) *pb.JobList {
	pjs := make([]*pb.Job, len(jobs))
	for i, job := range jobs {
		pjs[i] = JobToPB(job)
	}
	return &pb.JobList{Jobs: pjs}
}

// JobsFromPB returns Jobs from the protobuf version of given Jobs list.
func JobsFromPB(pjs *pb.JobList) []*Job {
	jobs := make([]*Job, len(pjs.Jobs))
	for i, pj := range pjs.Jobs {
		jobs[i] = JobFromPB(pj)
	}
	return jobs
}

// JobIDsToPB returns job id numbers as int64 for pb.JobIDList
func JobIDsToPB(ids []int) []int64 {
	i64 := make([]int64, len(ids))
	for i, id := range ids {
		i64[i] = int64(id)
	}
	return i64
}

// JobIDsFromPB returns job id numbers from int64 in pb.JobIDList
func JobIDsFromPB(ids []int64) []int {
	is := make([]int, len(ids))
	for i, id := range ids {
		is[i] = int(id)
	}
	return is
}
