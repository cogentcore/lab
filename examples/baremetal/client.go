// Copyright (c) 2025, Cogent Lab. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"context"
	"fmt"
	"time"

	"cogentcore.org/core/base/errors"
	pb "cogentcore.org/lab/examples/baremetal/baremetal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	// The server address including port number.
	Host string `default:"localhost:8585"`

	Timeout time.Duration

	// grpc connection
	conn *grpc.ClientConn

	client pb.BareMetalClient
}

func NewClient() *Client {
	cl := &Client{}
	cl.Host = "localhost:8585"
	cl.Timeout = 120 * time.Second
	return cl
}

// Connect connects to the server
func (cl *Client) Connect() error {
	conn, err := grpc.NewClient(cl.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Log(fmt.Errorf("did not connect: %v", err))
	}
	cl.conn = conn
	cl.client = pb.NewBareMetalClient(conn)
	return nil
}

// Submit adds a new Active job with given parameters.
func (cl *Client) Submit(source, path, script, resultsGlob string, files []byte) (*Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cl.Timeout)
	defer cancel()

	sub := &pb.Submission{Source: source, Path: path, Script: script, ResultsGlob: resultsGlob, Files: files}
	job, err := cl.client.Submit(ctx, sub)
	if err != nil {
		return nil, errors.Log(fmt.Errorf("could not submit: %v", err))
	}
	return JobFromPB(job), nil
}

// JobStatus gets current job data back from server for given job id(s).
func (cl *Client) JobStatus(ids ...int) ([]*Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cl.Timeout)
	defer cancel()

	pids := &pb.JobIDList{JobID: JobIDsToPB(ids)}
	jobs, err := cl.client.JobStatus(ctx, pids)
	if err != nil {
		return nil, errors.Log(fmt.Errorf("JobStatus failed: %v", err))
	}
	return JobsFromPB(jobs), nil
}

// CancelJobs cancels list of job IDs. Returns error for jobs not found.
func (cl *Client) CancelJobs(ids ...int) error {
	ctx, cancel := context.WithTimeout(context.Background(), cl.Timeout)
	defer cancel()

	pids := &pb.JobIDList{JobID: JobIDsToPB(ids)}
	emsg, err := cl.client.CancelJobs(ctx, pids)
	if err != nil {
		return errors.Log(fmt.Errorf("CancelJobs failed: %v", err))
	}
	return errors.New(emsg.Error)
}

// FetchResults gets job results back from server for given job id(s).
// Results are available as job.Results as a compressed tar file.
func (cl *Client) FetchResults(ids ...int) ([]*Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cl.Timeout)
	defer cancel()

	pids := &pb.JobIDList{JobID: JobIDsToPB(ids)}
	jobs, err := cl.client.FetchResults(ctx, pids)
	if err != nil {
		return nil, errors.Log(fmt.Errorf("FetchResults failed: %v", err))
	}
	return JobsFromPB(jobs), nil
}

// UpdateJobs pings the server to run its updates.
// This happens automatically very 10 seconds but this is for the impatient.
func (cl *Client) UpdateJobs() {
	return
}
