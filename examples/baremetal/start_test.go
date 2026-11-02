// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"context"
	"net"
	"testing"
	"time"

	pb "cogentcore.org/lab/examples/baremetal/baremetal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// statusServer is a server that only answers JobStatus, with no jobs.
type statusServer struct {
	pb.UnimplementedBareMetalServer
}

func (s *statusServer) JobStatus(context.Context, *pb.JobIDList) (*pb.JobList, error) {
	return &pb.JobList{}, nil
}

// serve serves a statusServer at the given address until it is stopped.
func serve(t *testing.T, addr string) *grpc.Server {
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	s := grpc.NewServer()
	pb.RegisterBareMetalServer(s, &statusServer{})
	go s.Serve(lis)
	return s
}

// TestServerRestart tests that the client works again right after the
// server is restarted, as when simmer starts it again after it has been
// stopped, instead of after the backoff of its connection.
func TestServerRestart(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0") // a free port, the same when restarted
	require.NoError(t, err)
	addr := lis.Addr().String()
	lis.Close()

	s := serve(t, addr)
	cl := NewClient()
	cl.Host = addr
	cl.Timeout = 10 * time.Second
	require.NoError(t, cl.Connect())
	assert.True(t, cl.Running())
	_, err = cl.JobStatus()
	require.NoError(t, err)

	s.Stop() // as when the server is killed
	assert.False(t, cl.Running())
	for range 3 { // fails, in backoff
		_, err = cl.JobStatus()
		assert.Error(t, err)
		time.Sleep(time.Second)
	}

	s = serve(t, addr)
	defer s.Stop()
	start := time.Now()
	cl.waitServer(5 * time.Second)
	_, err = cl.JobStatus()
	require.NoError(t, err)
	assert.Less(t, time.Since(start), time.Second)
}
