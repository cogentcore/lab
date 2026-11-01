// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goalls

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testClient is a client of a test server, as an editor.
type testClient struct {
	t     *testing.T
	conn  *conn
	mu    sync.Mutex
	id    int
	resps map[string]chan *message
	diags chan *message
}

// newTestClient runs a server for a new test client.
func newTestClient(t *testing.T) *testClient {
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Skip("gopls not found in PATH:", err)
	}
	cr, sw := io.Pipe() // server to client
	sr, cw := io.Pipe() // client to server
	s := &Server{}
	go s.Run(sr, sw)
	t.Cleanup(func() { cw.Close() })
	c := &testClient{t: t, conn: newConn(cr, cw), resps: map[string]chan *message{}, diags: make(chan *message, 100)}
	go c.read()
	return c
}

// read reads the messages from the server, answering its requests.
func (c *testClient) read() {
	for {
		m, err := c.conn.read()
		if err != nil {
			return
		}
		switch {
		case m.isResponse():
			c.mu.Lock()
			ch := c.resps[string(m.ID)]
			c.mu.Unlock()
			if ch != nil {
				ch <- m
			}
		case m.isRequest():
			res := json.RawMessage("null")
			if m.Method == "workspace/configuration" {
				res = json.RawMessage("[null]")
			}
			c.conn.write(&message{ID: m.ID, Result: res})
		case m.Method == "textDocument/publishDiagnostics":
			c.diags <- m
		}
	}
}

// call sends the given request, returning its result.
func (c *testClient) call(method string, params any, result any) {
	c.t.Helper()
	c.mu.Lock()
	c.id++
	id := json.RawMessage(fmt.Sprint(c.id))
	ch := make(chan *message, 1)
	c.resps[string(id)] = ch
	c.mu.Unlock()
	b, _ := json.Marshal(params)
	require.NoError(c.t, c.conn.write(&message{ID: id, Method: method, Params: b}))
	select {
	case m := <-ch:
		require.Empty(c.t, string(m.Error), method)
		if result != nil {
			require.NoError(c.t, json.Unmarshal(m.Result, result), method)
		}
	case <-time.After(60 * time.Second):
		c.t.Fatalf("timeout waiting for %s", method)
	}
}

// notify sends the given notification.
func (c *testClient) notify(method string, params any) {
	b, _ := json.Marshal(params)
	require.NoError(c.t, c.conn.write(&message{Method: method, Params: b}))
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspRangeT struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type diagnostic struct {
	Range   lspRangeT `json:"range"`
	Message string    `json:"message"`
}

// waitDiags waits for diagnostics of the given file for which the given
// function returns true.
func (c *testClient) waitDiags(uri string, ok func([]diagnostic) bool) []diagnostic {
	c.t.Helper()
	timeout := time.After(60 * time.Second)
	for {
		select {
		case m := <-c.diags:
			var p struct {
				URI         string       `json:"uri"`
				Diagnostics []diagnostic `json:"diagnostics"`
			}
			json.Unmarshal(m.Params, &p)
			if p.URI == uri && ok(p.Diagnostics) {
				return p.Diagnostics
			}
		case <-timeout:
			c.t.Fatalf("timeout waiting for the diagnostics of %s", uri)
		}
	}
}

func hasDiag(diags []diagnostic, msg string) *diagnostic {
	for i := range diags {
		if strings.Contains(diags[i].Message, msg) {
			return &diags[i]
		}
	}
	return nil
}

// capabilities are the client capabilities of the test client.
var capabilities = map[string]any{"textDocument": map[string]any{"semanticTokens": map[string]any{
	"requests":       map[string]any{"full": true},
	"tokenTypes":     []string{"namespace", "type", "function", "method", "variable", "parameter", "keyword", "string", "number", "comment", "operator"},
	"tokenModifiers": []string{"definition", "readonly", "defaultLibrary"},
	"formats":        []string{"relative"},
}}}

// TestServer tests the server with gopls, for a Goal file.
func TestServer(t *testing.T) {
	c := newTestClient(t)
	dir := t.TempDir()
	t.Setenv("GOWORK", "off")
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n\ngo 1.21\n"), 0644)
	src := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tx:=1\n\tfmt.Println(x, y)\n\tls -la\n}\n"
	fname := filepath.Join(dir, "main.goal")
	os.WriteFile(fname, []byte(src), 0644)
	uri := filenameURI(fname)
	doc := map[string]any{"uri": uri}

	c.call("initialize", map[string]any{"processId": nil, "rootUri": filenameURI(dir), "capabilities": capabilities,
		"initializationOptions": map[string]any{"semanticTokens": true, "hints": map[string]any{"assignVariableTypes": true, "parameterNames": true}}}, nil)
	c.notify("initialized", map[string]any{})
	c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "goal", "version": 1, "text": src}})

	// the errors of the Go code are in the Goal file
	diags := c.waitDiags(uri, func(ds []diagnostic) bool { return hasDiag(ds, "undefined: y") != nil })
	assert.Equal(t, lspRangeT{Start: position{6, 16}, End: position{6, 17}}, hasDiag(diags, "undefined: y").Range)
	if d := hasDiag(diags, "undefined: goalrun"); assert.NotNil(t, d) {
		assert.Equal(t, 7, d.Range.Start.Line)
		// with its Go code, which is not the same as its source
		assert.Contains(t, d.Message, "\nGo: goalrun.Run(\"ls\", \"-la\")")
	}
	assert.NotContains(t, hasDiag(diags, "undefined: y").Message, "Go:")

	var hover struct {
		Contents struct {
			Value string `json:"value"`
		} `json:"contents"`
	}
	c.call("textDocument/hover", map[string]any{"textDocument": doc, "position": position{6, 7}}, &hover)
	assert.Contains(t, hover.Contents.Value, "Println")

	var locs []struct {
		URI   string    `json:"uri"`
		Range lspRangeT `json:"range"`
	}
	c.call("textDocument/definition", map[string]any{"textDocument": doc, "position": position{6, 13}}, &locs)
	if assert.Len(t, locs, 1) {
		assert.Equal(t, uri, locs[0].URI)
		assert.Equal(t, position{5, 1}, locs[0].Range.Start)
	}

	var hints []struct {
		Position position `json:"position"`
		Label    any      `json:"label"`
	}
	c.call("textDocument/inlayHint", map[string]any{"textDocument": doc, "range": lspRangeT{End: position{9, 0}}}, &hints)
	// the type of x, the parameter of Println, and the Go code of the
	// command with an error, without the hints of the Go code of the
	// command, which is not the same as its source
	got := map[position]string{}
	for _, h := range hints {
		got[h.Position] = labelText(h.Label)
	}
	assert.Equal(t, map[position]string{{5, 2}: "int", {6, 13}: "a...:", {7, 7}: `Go: goalrun.Run("ls", "-la")`}, got)

	var toks struct {
		Data []int `json:"data"`
	}
	c.call("textDocument/semanticTokens/full", map[string]any{"textDocument": doc}, &toks)
	if assert.NotEmpty(t, toks.Data) {
		assert.Equal(t, 0, toks.Data[0], "the first token, package, is on the first line")
	}

	// renaming edits the Goal file
	var we struct {
		DocumentChanges []struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			Edits []struct {
				Range   lspRangeT `json:"range"`
				NewText string    `json:"newText"`
			} `json:"edits"`
		} `json:"documentChanges"`
		Changes map[string][]struct {
			Range lspRangeT `json:"range"`
		} `json:"changes"`
	}
	c.call("textDocument/rename", map[string]any{"textDocument": doc, "position": position{5, 1}, "newName": "z"}, &we)
	var starts []position
	for _, dc := range we.DocumentChanges {
		assert.Equal(t, uri, dc.TextDocument.URI)
		for _, e := range dc.Edits {
			assert.Equal(t, "z", e.NewText)
			starts = append(starts, e.Range.Start)
		}
	}
	for u, es := range we.Changes {
		assert.Equal(t, uri, u)
		for _, e := range es {
			starts = append(starts, e.Range.Start)
		}
	}
	assert.ElementsMatch(t, []position{{5, 1}, {6, 13}}, starts)

	// formatting formats the Goal code
	var edits []struct {
		Range   lspRangeT `json:"range"`
		NewText string    `json:"newText"`
	}
	c.call("textDocument/formatting", map[string]any{"textDocument": doc, "options": map[string]any{"tabSize": 4, "insertSpaces": false}}, &edits)
	if assert.Len(t, edits, 1) {
		assert.Contains(t, edits[0].NewText, "\tx := 1\n\tfmt.Println(x, y)\n\tls -la\n")
	}

	// completion, after a change (a line with just fmt.Pr is a command)
	fixed := strings.Replace(src, "x, y)", "x, x)\n\ts := fmt.Pr", 1)
	c.notify("textDocument/didChange", map[string]any{"textDocument": map[string]any{"uri": uri, "version": 2},
		"contentChanges": []any{map[string]any{"text": fixed}}})
	var comps struct {
		Items []struct {
			Label string `json:"label"`
		} `json:"items"`
	}
	c.call("textDocument/completion", map[string]any{"textDocument": doc, "position": position{7, 12}}, &comps)
	found := false
	for _, it := range comps.Items {
		found = found || it.Label == "Println"
	}
	assert.True(t, found, "Println in %v", comps.Items)

	// saving writes the Go file
	c.notify("textDocument/didSave", map[string]any{"textDocument": doc})
	gofile := filepath.Join(dir, "main.go")
	deadline := time.Now().Add(30 * time.Second)
	for {
		b, err := os.ReadFile(gofile)
		if err == nil && strings.Contains(string(b), `Code generated by "goal build"`) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the Go file was not written")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// an incremental change fixes the error
	c.notify("textDocument/didChange", map[string]any{"textDocument": map[string]any{"uri": uri, "version": 3},
		"contentChanges": []any{map[string]any{"range": lspRangeT{Start: position{7, 0}, End: position{8, 0}}, "text": ""}}})
	c.waitDiags(uri, func(ds []diagnostic) bool {
		return hasDiag(ds, "undefined: y") == nil && hasDiag(ds, "goalrun") != nil
	})
}

func TestApplyChange(t *testing.T) {
	text := "aé😀b\nxyz\n"
	// the 😀 is two UTF-16 units, and four UTF-8 bytes
	assert.Equal(t, "aé😀Qb\nxyz\n", applyChange(text, utf16Encoding, 0, 4, 0, 4, "Q"))
	assert.Equal(t, "aé😀Qb\nxyz\n", applyChange(text, utf8Encoding, 0, 7, 0, 7, "Q"))
	assert.Equal(t, "aé😀Qb\nxyz\n", applyChange(text, utf32Encoding, 0, 3, 0, 3, "Q"))
	assert.Equal(t, "aQyz\n", applyChange(text, utf16Encoding, 0, 1, 1, 1, "Q"))
	assert.Equal(t, 4, utf16Encoding.fromRune("aé😀b", 3))
	assert.Equal(t, 3, utf16Encoding.toRune("aé😀b", 4))
}

// labelText returns the text of the given label of an inlay hint,
// which is a string or a list of parts with values.
func labelText(label any) string {
	if s, ok := label.(string); ok {
		return s
	}
	parts, _ := label.([]any)
	txt := ""
	for _, p := range parts {
		if pm, ok := p.(map[string]any); ok {
			txt += toString(pm["value"])
		}
	}
	return txt
}

// TestServerHints tests that the inlay hints of the lines of a Goal file
// whose Go code is not the same as their source are omitted.
func TestServerHints(t *testing.T) {
	c := newTestClient(t)
	dir := t.TempDir()
	t.Setenv("GOWORK", "off")
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n\ngo 1.21\n"), 0644)
	// goalrun is defined, so that the Go code of the command has hints
	src := "package main\n\ntype runner struct{}\n\nfunc (runner) Run(cmd ...any) {}\n\nvar goalrun runner\n\nfunc show(msg string) {}\n\nfunc main() {\n\tls -la\n\tshow(\"hi\")\n}\n"
	fname := filepath.Join(dir, "main.goal")
	os.WriteFile(fname, []byte(src), 0644)
	uri := filenameURI(fname)
	c.call("initialize", map[string]any{"processId": nil, "rootUri": filenameURI(dir), "capabilities": capabilities,
		"initializationOptions": map[string]any{"hints": map[string]any{"parameterNames": true}}}, nil)
	c.notify("initialized", map[string]any{})
	c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "goal", "version": 1, "text": src}})
	c.waitDiags(uri, func(ds []diagnostic) bool { return true })
	var hints []struct {
		Position position `json:"position"`
		Label    any      `json:"label"`
	}
	deadline := time.Now().Add(30 * time.Second)
	for len(hints) == 0 && time.Now().Before(deadline) {
		c.call("textDocument/inlayHint", map[string]any{"textDocument": map[string]any{"uri": uri}, "range": lspRangeT{End: position{14, 0}}}, &hints)
	}
	got := map[position]string{}
	for _, h := range hints {
		got[h.Position] = labelText(h.Label)
	}
	// the hint of show, but not that of Run, in the Go code of ls -la
	assert.Equal(t, map[position]string{{12, 6}: "msg:"}, got)
}
