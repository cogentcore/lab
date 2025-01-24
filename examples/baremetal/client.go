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

	// grpc connection
	conn *grpc.ClientConn

	client pb.BareMetalClient
}

func NewClient() *Client {
	cl := &Client{}
	cl.Host = "localhost:8585"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sub := &pb.Submission{Source: source, Path: path, Script: script, ResultsGlob: resultsGlob, Files: files}

	job, err := cl.client.Submit(ctx, sub)
	if err != nil {
		return nil, errors.Log(fmt.Errorf("could not submit: %v", err))
	}
	return JobFromPB(job), nil
}

// UpdateJobs runs any pending jobs if there are available GPUs to run on.
// returns number of jobs started, and any errors incurred in starting jobs.
func (cl *Client) UpdateJobs() (nrun, nfinished int, err error) {
	return
}

// CancelJobs cancels list of job IDs. Returns error for jobs not found.
func (cl *Client) CancelJobs(jobs ...int) error {
	return nil
}

// JobStatus gets current job data back from server for given job id(s).
func (cl *Client) JobStatus(ids ...int) ([]*Job, error) {
	return nil, nil
}

// FetchResults gets job results back from server for given job id(s).
// Results are available as job.Results as a compressed tar file.
func (cl *Client) FetchResults(ids ...int) ([]*Job, error) {
	return nil, nil
}
