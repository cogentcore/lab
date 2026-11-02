// Copyright (c) 2025, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package baremetal

import (
	"context"
	"fmt"
	"go/build"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"google.golang.org/grpc/connectivity"
)

// Running returns whether the server is listening at [Client.Host].
func (cl *Client) Running() bool {
	conn, err := net.DialTimeout("tcp", cl.Host, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// StartServer starts the baremetal server in a terminal in its own window
// if it is not running, as when it has been stopped, as a separate process
// that keeps running after this one ends (see the -window option of the
// baremetal command). It waits a few seconds for the server to be up, and
// then has the client connect to it right away, instead of after the
// backoff of its connection (see [Client.Connect]).
func (cl *Client) StartServer() error {
	if cl.Running() {
		cl.started = time.Time{}
		return nil
	}
	if time.Since(cl.started) < time.Minute {
		return nil // still starting, or it failed to start
	}
	exe, err := ServerExecutable()
	if err != nil {
		return err
	}
	_, port, err := net.SplitHostPort(cl.Host)
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "-window", "-port", port)
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start the baremetal server: %w", err)
	}
	cl.started = time.Now()
	go cmd.Wait()
	cl.waitServer(10 * time.Second)
	return nil
}

// waitServer waits up to the given time for the server to be up and the
// client to be connected to it, connecting right away, instead of after
// the backoff of its connection, which fails the calls while it lasts.
func (cl *Client) waitServer(wait time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	for !cl.Running() && ctx.Err() == nil {
		time.Sleep(100 * time.Millisecond)
	}
	if cl.conn == nil {
		return
	}
	cl.conn.ResetConnectBackoff()
	for st := cl.conn.GetState(); st != connectivity.Ready; st = cl.conn.GetState() {
		if st == connectivity.Idle {
			cl.conn.Connect()
		}
		if !cl.conn.WaitForStateChange(ctx, st) {
			return
		}
	}
}

// ServerExecutable returns the path of the baremetal executable: on the
// PATH, or else in the Go bin directory, which is not on the PATH of an
// app started from the Finder or Dock.
func ServerExecutable() (string, error) {
	if exe, err := exec.LookPath("baremetal"); err == nil {
		return exe, nil
	}
	bin := os.Getenv("GOBIN")
	if bin == "" {
		bin = filepath.Join(build.Default.GOPATH, "bin")
	}
	exe := filepath.Join(bin, "baremetal")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	if _, err := os.Stat(exe); err != nil {
		return "", fmt.Errorf("baremetal server executable not found on the PATH or in %s: install it with go install cogentcore.org/lab/examples/baremetal/cmd/baremetal@latest", bin)
	}
	return exe, nil
}
