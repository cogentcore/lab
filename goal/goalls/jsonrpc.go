// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goalls

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"sync"
)

// message is a JSON-RPC message of the language server protocol:
// a request (with an ID and a Method), a response (with an ID and a
// Result or an Error), or a notification (with a Method).
type message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

// isRequest returns whether the message is a request.
func (m *message) isRequest() bool { return m.Method != "" && len(m.ID) > 0 }

// isResponse returns whether the message is a response.
func (m *message) isResponse() bool { return m.Method == "" && len(m.ID) > 0 }

// conn reads and writes messages with the base protocol of the language
// server protocol: each has a Content-Length header.
type conn struct {
	r  *bufio.Reader
	w  io.Writer
	mu sync.Mutex // for writing
}

func newConn(r io.Reader, w io.Writer) *conn {
	return &conn{r: bufio.NewReader(r), w: w}
}

// read reads the next message.
func (c *conn) read() (*message, error) {
	hdr, err := textproto.NewReader(c.r).ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	n, err := strconv.Atoi(hdr.Get("Content-Length"))
	if err != nil {
		return nil, fmt.Errorf("goalls: bad Content-Length: %w", err)
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(c.r, b); err != nil {
		return nil, err
	}
	m := &message{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, fmt.Errorf("goalls: bad message: %w", err)
	}
	return m, nil
}

// write writes the given message.
func (c *conn) write(m *message) error {
	m.JSONRPC = "2.0"
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(b)); err != nil {
		return err
	}
	_, err = c.w.Write(b)
	return err
}
