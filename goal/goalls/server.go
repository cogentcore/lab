// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package goalls is a language server for Goal files, for editors, which
// is a proxy for the Go language server, gopls: it gives gopls the Go
// code of each open Goal file, as the Go file of the Goal file, which
// `goal build` writes, and maps the positions of the messages between the
// editor and gopls between the Goal file and its Go code (see
// [transpile.GoCode]). Formatting a Goal file formats it and fixes its
// imports with goalfmt. The messages for Go files are passed through, so
// it can be used instead of gopls, for both Go and Goal files.
package goalls

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"

	"cogentcore.org/lab/goal/goalfmt"
	"github.com/pmezard/go-difflib/difflib"
	"cogentcore.org/lab/goal/transpile"
)

// Server is a language server for Goal files, which is a proxy for gopls
// (see the package documentation).
type Server struct {

	// Gopls is the command to run gopls, with its arguments.
	// The default is gopls, without arguments.
	Gopls []string

	// Stderr is where the standard error of gopls goes, and the errors
	// of writing Go files. The default is [os.Stderr].
	Stderr io.Writer

	// NoWrite is whether to not write the Go file of a Goal file when it
	// is saved, as `goal build` does, which is done by default, so that
	// the Go files are in sync with the Goal files.
	NoWrite bool

	// client is the connection to the client (the editor),
	// and gopls the connection to gopls.
	client, gopls *conn

	// mu protects the fields below.
	mu sync.Mutex

	// enc is the encoding of the positions.
	enc encoding

	// docs are the open Goal files, by their URIs.
	docs map[string]*doc

	// pending are the requests of the client to gopls,
	// whose responses are mapped back, by their IDs.
	pending map[string]pending
}

// pending is a request of the client to gopls.
type pending struct {
	method string

	// doc is the Goal file of the request, if any.
	doc *doc
}

// Run runs the server, with the given input from the client and output
// to it, until the input ends, or gopls exits.
func (s *Server) Run(in io.Reader, out io.Writer) error {
	args := s.Gopls
	if len(args) == 0 {
		args = []string{"gopls"}
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = s.Stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	gin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	gout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.client = newConn(in, out)
	s.gopls = newConn(gout, gin)
	s.docs = map[string]*doc{}
	s.pending = map[string]pending{}
	errc := make(chan error, 2)
	go func() { errc <- s.fromClient() }()
	go func() { errc <- s.fromGopls() }()
	err = <-errc
	gin.Close()
	werr := cmd.Wait()
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err == nil && werr != nil && !strings.Contains(werr.Error(), "signal") {
		err = werr
	}
	return err
}

// fromClient handles the messages from the client.
func (s *Server) fromClient() error {
	for {
		m, err := s.client.read()
		if err != nil {
			return err
		}
		if m.isResponse() { // to a request of gopls
			err = s.gopls.write(m)
		} else {
			err = s.clientMessage(m)
		}
		if err != nil {
			return err
		}
	}
}

// fromGopls handles the messages from gopls.
func (s *Server) fromGopls() error {
	for {
		m, err := s.gopls.read()
		if err != nil {
			return err
		}
		if err := s.goplsMessage(m); err != nil {
			return err
		}
	}
}

// textDocument is the document of the parameters of a message.
type textDocument struct {
	TextDocument struct {
		URI        string `json:"uri"`
		LanguageID string `json:"languageId"`
		Version    int    `json:"version"`
		Text       string `json:"text"`
	} `json:"textDocument"`
}

// clientMessage handles the given request or notification from the client.
func (s *Server) clientMessage(m *message) error {
	var td textDocument
	json.Unmarshal(m.Params, &td)
	uri := td.TextDocument.URI
	if uri == "" || !isGoal(uri) {
		if m.isRequest() {
			s.mu.Lock()
			s.pending[string(m.ID)] = pending{method: m.Method}
			s.mu.Unlock()
		}
		return s.gopls.write(m)
	}
	s.mu.Lock()
	d := s.docs[uri]
	s.mu.Unlock()
	switch m.Method {
	case "textDocument/didOpen":
		return s.didOpen(uri, td.TextDocument.Version, td.TextDocument.Text)
	case "textDocument/didChange":
		if d != nil {
			return s.didChange(d, m.Params)
		}
	case "textDocument/didClose":
		if d != nil {
			return s.didClose(d)
		}
	case "textDocument/didSave":
		if d != nil {
			if !s.NoWrite {
				s.writeGo(d)
			}
			return s.notify(s.gopls, m.Method, map[string]any{"textDocument": map[string]any{"uri": d.goURI}})
		}
	case "textDocument/formatting", "textDocument/rangeFormatting":
		if d != nil {
			return s.format(d, m)
		}
	}
	if d == nil {
		if m.isRequest() { // not open
			return s.client.write(&message{ID: m.ID, Result: json.RawMessage("null")})
		}
		return nil
	}
	if m.isRequest() {
		s.mu.Lock()
		s.pending[string(m.ID)] = pending{method: m.Method, doc: d}
		s.mu.Unlock()
	}
	params, ok := s.mapJSON(m.Params, d, true)
	if !ok { // a position that is not in the Go code
		if m.isRequest() {
			s.mu.Lock()
			delete(s.pending, string(m.ID))
			s.mu.Unlock()
			return s.client.write(&message{ID: m.ID, Result: json.RawMessage("null")})
		}
		return nil
	}
	m.Params = params
	return s.gopls.write(m)
}

// notify sends a notification with the given method and parameters.
func (s *Server) notify(c *conn, method string, params any) error {
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return c.write(&message{Method: method, Params: b})
}

// didOpen opens the given Goal file, opening its Go code with gopls.
func (s *Server) didOpen(uri string, version int, text string) error {
	d := newDoc(uri, version, text)
	s.mu.Lock()
	s.docs[uri] = d
	code := d.code.Code
	s.mu.Unlock()
	err := s.notify(s.gopls, "textDocument/didOpen", map[string]any{"textDocument": map[string]any{
		"uri": d.goURI, "languageId": "go", "version": version, "text": code}})
	if err != nil {
		return err
	}
	return s.publish(d)
}

// didChange applies the given changes to the given Goal file, and sends
// its new Go code to gopls.
func (s *Server) didChange(d *doc, params json.RawMessage) error {
	var ch struct {
		TextDocument struct {
			Version int `json:"version"`
		} `json:"textDocument"`
		ContentChanges []struct {
			Range *struct {
				Start struct{ Line, Character int } `json:"start"`
				End   struct{ Line, Character int } `json:"end"`
			} `json:"range"`
			Text string `json:"text"`
		} `json:"contentChanges"`
	}
	if err := json.Unmarshal(params, &ch); err != nil {
		return nil
	}
	s.mu.Lock()
	text := d.text
	for _, c := range ch.ContentChanges {
		if c.Range == nil {
			text = c.Text
			continue
		}
		r := c.Range
		text = applyChange(text, s.enc, r.Start.Line, r.Start.Character, r.End.Line, r.End.Character, c.Text)
	}
	d.version = ch.TextDocument.Version
	d.setText(text)
	code := d.code.Code
	s.mu.Unlock()
	err := s.notify(s.gopls, "textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": d.goURI, "version": d.version},
		"contentChanges": []any{map[string]any{"text": code}}})
	if err != nil {
		return err
	}
	return s.publish(d)
}

// didClose closes the given Goal file, and its Go code with gopls.
func (s *Server) didClose(d *doc) error {
	s.mu.Lock()
	delete(s.docs, d.uri)
	s.mu.Unlock()
	err := s.notify(s.gopls, "textDocument/didClose", map[string]any{"textDocument": map[string]any{"uri": d.goURI}})
	if err != nil {
		return err
	}
	return s.notify(s.client, "textDocument/publishDiagnostics", map[string]any{"uri": d.uri, "diagnostics": []any{}})
}

// writeGo writes the Go file of the given Goal file, as `goal build` does.
func (s *Server) writeGo(d *doc) {
	s.mu.Lock()
	filename, text := d.filename, d.text
	s.mu.Unlock()
	gofile := goalfmt.GoFilename(filename)
	res, err := transpile.NewState().GoFile(filename, gofile, text)
	err = errors.Join(err, os.WriteFile(gofile, res, 0666))
	if err != nil {
		w := s.Stderr
		if w == nil {
			w = os.Stderr
		}
		fmt.Fprintln(w, "goalls: writing", gofile+":", err)
	}
}

// format responds to the given formatting request for the given Goal
// file with the edits of formatting it with goalfmt, which also fixes its
// imports: one for each range of lines that changed, as gopls does, so
// that editors can apply the edits within a range of lines (to indent them).
func (s *Server) format(d *doc, m *message) error {
	s.mu.Lock()
	text, filename := d.text, d.filename
	s.mu.Unlock()
	res, err := goalfmt.Source(filename, []byte(text), nil)
	if err != nil {
		e, _ := json.Marshal(map[string]any{"code": -32803, "message": err.Error()})
		return s.client.write(&message{ID: m.ID, Error: e})
	}
	b, _ := json.Marshal(lineEdits(text, string(res)))
	return s.client.write(&message{ID: m.ID, Result: b})
}

// lineEdits returns the edits that change the given text to the given new
// text, replacing whole lines, one for each range of lines that changed.
func lineEdits(text, res string) []any {
	a, b := strings.SplitAfter(text, "\n"), strings.SplitAfter(res, "\n")
	edits := []any{}
	for _, op := range difflib.NewMatcher(a, b).GetOpCodes() {
		if op.Tag == 'e' {
			continue
		}
		edits = append(edits, map[string]any{
			"range":   lspRange(op.I1, 0, op.I2, 0),
			"newText": strings.Join(b[op.J1:op.J2], "")})
	}
	return edits
}

// goplsMessage handles the given message from gopls.
func (s *Server) goplsMessage(m *message) error {
	if m.isResponse() {
		s.mu.Lock()
		p, has := s.pending[string(m.ID)]
		delete(s.pending, string(m.ID))
		s.mu.Unlock()
		s.mu.Lock()
		goal := len(s.docs) > 0
		s.mu.Unlock()
		switch {
		case !has || len(m.Result) == 0 || string(m.Result) == "null":
		case p.method == "initialize":
			m.Result = s.initialized(m.Result)
		case p.doc == nil:
			if goal { // results can be in the Go code of Goal files
				m.Result, _ = s.mapJSON(m.Result, nil, false)
			}
		case p.method == "textDocument/inlayHint":
			m.Result = s.inlayHints(p.doc, m.Result)
		case strings.HasPrefix(p.method, "textDocument/semanticTokens/"):
			m.Result = s.semanticTokens(p.doc, m.Result)
		default:
			res, ok := s.mapJSON(m.Result, p.doc, false)
			if !ok {
				res = json.RawMessage("null")
			}
			m.Result = res
		}
		return s.client.write(m)
	}
	if m.Method == "textDocument/publishDiagnostics" {
		if d := s.goDoc(m.Params); d != nil {
			return s.diagnostics(d, m.Params)
		}
	}
	s.mu.Lock()
	goal := len(s.docs) > 0
	s.mu.Unlock()
	if goal {
		m.Params, _ = s.mapJSON(m.Params, nil, false)
	}
	return s.client.write(m)
}

// initialized returns the given result of the initialize request, which
// has the capabilities of gopls, after recording its position encoding.
// The semantic tokens are only for full documents, without deltas, which
// cannot be mapped.
func (s *Server) initialized(result json.RawMessage) json.RawMessage {
	var res map[string]any
	if err := json.Unmarshal(result, &res); err != nil {
		return result
	}
	caps, _ := res["capabilities"].(map[string]any)
	if caps == nil {
		return result
	}
	s.mu.Lock()
	switch caps["positionEncoding"] {
	case "utf-8":
		s.enc = utf8Encoding
	case "utf-32":
		s.enc = utf32Encoding
	}
	s.mu.Unlock()
	if st, ok := caps["semanticTokensProvider"].(map[string]any); ok {
		if _, ok := st["full"].(map[string]any); ok {
			st["full"] = true
		}
	}
	b, err := json.Marshal(res)
	if err != nil {
		return result
	}
	return b
}

// goDoc returns the open Goal file whose Go code is the document of
// the given parameters, or nil.
func (s *Server) goDoc(params json.RawMessage) *doc {
	var p struct {
		URI string `json:"uri"`
	}
	json.Unmarshal(params, &p)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.docForGo(p.URI)
}

// docForGo returns the open Goal file whose Go code has the given URI,
// or nil. Must be called under mu.
func (s *Server) docForGo(uri string) *doc {
	for _, d := range s.docs {
		if d.goURI == uri {
			return d
		}
	}
	return nil
}

// diagnostics handles the given diagnostics of gopls for the Go code of
// the given Goal file, which are published for it, at their positions in
// it, or at its start for those in the code that the transpiler adds.
func (s *Server) diagnostics(d *doc, params json.RawMessage) error {
	var p struct {
		Diagnostics []any `json:"diagnostics"`
	}
	dec := json.NewDecoder(bytes.NewReader(params))
	dec.UseNumber()
	dec.Decode(&p)
	s.mu.Lock()
	diags := make([]any, 0, len(p.Diagnostics))
	d.goCode = map[int]string{}
	mp := &mapper{s: s}
	for _, dg := range p.Diagnostics {
		dm, ok := dg.(map[string]any)
		if !ok {
			continue
		}
		var goLine int
		if r, ok := dm["range"].(map[string]any); ok {
			if st, ok := r["start"].(map[string]any); ok {
				goLine, _ = toInt(st["line"])
			}
		}
		if r, ok := mp.value(dm["range"], d); ok {
			dm["range"] = r
			// the Go code of a line that is not the same as its source
			if sl := d.code.SourceLine(goLine); sl >= 0 {
				goCode := strings.TrimSpace(d.code.Lines[goLine])
				if stripSpaces(goCode) != stripSpaces(line(d.lines, sl)) {
					dm["message"] = toString(dm["message"]) + "\nGo: " + goCode
					if fmt.Sprint(dm["severity"]) == "1" { // an error
						d.goCode[sl] = goCode
					}
				}
			}
		} else {
			dm["range"] = lspRange(0, 0, 0, 0)
			dm["message"] = "in the Go code: " + toString(dm["message"])
		}
		if ri, ok := dm["relatedInformation"]; ok {
			dm["relatedInformation"], _ = mp.value(ri, nil)
		}
		diags = append(diags, dm)
	}
	d.diags = diags
	s.mu.Unlock()
	return s.publish(d)
}

// publish publishes the diagnostics of the given Goal file: those of
// gopls for its Go code, and the errors of transpiling it.
func (s *Server) publish(d *doc) error {
	s.mu.Lock()
	diags := slices.Clone(d.diags)
	for _, e := range d.code.Errors {
		ln := max(e.Line-1, 0)
		src := line(d.lines, ln)
		ch := s.enc.fromRune(src, len([]rune(src[:min(max(e.Col-1, 0), len(src))])))
		diags = append(diags, map[string]any{"range": lspRange(ln, ch, ln, ch+1), "severity": 1, "source": "goal", "message": e.Msg})
	}
	if diags == nil {
		diags = []any{}
	}
	params := map[string]any{"uri": d.uri, "version": d.version, "diagnostics": diags}
	s.mu.Unlock()
	return s.notify(s.client, "textDocument/publishDiagnostics", params)
}

// lspRange returns a range with the given positions.
func lspRange(sl, sc, el, ec int) map[string]any {
	return map[string]any{"start": map[string]any{"line": sl, "character": sc}, "end": map[string]any{"line": el, "character": ec}}
}

// toString returns the given value as a string, if it is one.
func toString(v any) string {
	s, _ := v.(string)
	return s
}
