// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goalls

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"unicode"
)

// mapJSON returns the given JSON value (the parameters or result of a
// message) with its positions and URIs mapped from the Goal files to
// their Go code if toGo, or else back, with the given Goal file as that
// of the positions that are not in an object with a URI, and whether
// they could all be mapped (see [mapper]).
func (s *Server) mapJSON(raw json.RawMessage, d *doc, toGo bool) (json.RawMessage, bool) {
	if len(raw) == 0 {
		return raw, true
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return raw, true
	}
	s.mu.Lock()
	nv, ok := (&mapper{s: s, toGo: toGo}).value(v, d)
	s.mu.Unlock()
	if !ok {
		return raw, false
	}
	b, err := json.Marshal(nv)
	if err != nil {
		return raw, true
	}
	return b, true
}

// mapper maps the positions and URIs in the values of the messages of the
// language server protocol, between the Goal files and their Go code,
// in the direction of toGo. The positions are the objects with only a
// line and a character, which are those of the Goal file of the URI of
// the innermost object with one around them, or else of the Goal file
// of the message. An object with a position that cannot be mapped, such
// as one in the code that the transpiler adds, is removed from the array
// that it is in, or else, its parent object is, up to the value itself.
// It must be used under [Server.mu].
type mapper struct {
	s    *Server
	toGo bool
}

// doc returns the Goal file of the given URI, and its URI in the other
// direction: its Go code for a Goal file if toGo, and the Goal file of
// Go code otherwise, or nil.
func (mp *mapper) doc(uri string) (*doc, string) {
	if mp.toGo {
		if d := mp.s.docs[uri]; d != nil {
			return d, d.goURI
		}
		return nil, uri
	}
	if d := mp.s.docForGo(uri); d != nil {
		return d, d.uri
	}
	return nil, uri
}

// value returns the given value, with the positions of the given Goal
// file mapped, and whether they could all be mapped.
func (mp *mapper) value(v any, d *doc) (any, bool) {
	switch t := v.(type) {
	case []any:
		out := t[:0]
		for _, e := range t {
			if nv, ok := mp.value(e, d); ok {
				out = append(out, nv)
			}
		}
		return out, true
	case map[string]any:
		return mp.object(t, d)
	}
	return v, true
}

// object returns the given object with its positions mapped.
func (mp *mapper) object(t map[string]any, d *doc) (any, bool) {
	for _, key := range []string{"uri", "targetUri"} {
		if u, ok := t[key].(string); ok {
			d, t[key] = mp.doc(u)
		}
	}
	if changes, ok := t["changes"].(map[string]any); ok { // the edits of files, by URI
		out := make(map[string]any, len(changes))
		for u, edits := range changes {
			cd, nu := mp.doc(u)
			out[nu], _ = mp.value(edits, cd)
		}
		t["changes"] = out
	}
	if d != nil {
		if len(t) == 2 {
			ln, lok := toInt(t["line"])
			ch, cok := toInt(t["character"])
			if lok && cok {
				l, c, ok := mp.pos(d, ln, ch)
				if !ok {
					return t, false
				}
				t["line"], t["character"] = l, c
				return t, true
			}
		}
		for _, key := range []string{"startLine", "endLine"} { // of folding ranges
			if ln, ok := toInt(t[key]); ok {
				l, _, ok := mp.pos(d, ln, 0)
				if !ok {
					return t, false
				}
				t[key] = l
			}
		}
	}
	for k, v := range t {
		if k == "changes" {
			continue
		}
		nv, ok := mp.value(v, d)
		if !ok {
			return t, false
		}
		t[k] = nv
	}
	return t, true
}

// pos returns the given position of the given Goal file mapped.
func (mp *mapper) pos(d *doc, ln, ch int) (int, int, bool) {
	if mp.toGo {
		return d.toGo(mp.s.enc, ln, ch)
	}
	return d.toGoal(mp.s.enc, ln, ch)
}

// toInt returns the given JSON number as an int.
func toInt(v any) (int, bool) {
	n, ok := v.(json.Number)
	if !ok {
		return 0, false
	}
	i, err := n.Int64()
	return int(i), err == nil
}

// semanticTokens returns the given semantic tokens of the Go code of
// the given Goal file at their positions in the Goal file, without those
// in the code that the transpiler adds. The tokens are encoded as five
// numbers each, with lines and characters relative to the previous token.
func (s *Server) semanticTokens(d *doc, result json.RawMessage) json.RawMessage {
	var res map[string]any
	dec := json.NewDecoder(bytes.NewReader(result))
	dec.UseNumber()
	if err := dec.Decode(&res); err != nil {
		return result
	}
	data, _ := res["data"].([]any)
	var toks [][5]int
	s.mu.Lock()
	ln, ch := 0, 0
	for i := 0; i+4 < len(data); i += 5 {
		var t [5]int
		for k := range 5 {
			t[k], _ = toInt(data[i+k])
		}
		if t[0] > 0 {
			ch = 0
		}
		ln += t[0]
		ch += t[1]
		l, c, ok := d.toGoal(s.enc, ln, ch)
		if !ok {
			continue
		}
		// On a line whose Go code is not the same as its source, only the
		// tokens with the same text in it are kept: not those of the code
		// that the transpiler adds, such as the calls of tensor indexing.
		if !d.sameCode(l) && !d.sameText(s.enc, ln, ch, t[2], l, c) {
			continue
		}
		toks = append(toks, [5]int{l, c, t[2], t[3], t[4]})
	}
	s.mu.Unlock()
	slices.SortStableFunc(toks, func(a, b [5]int) int {
		if a[0] != b[0] {
			return a[0] - b[0]
		}
		return a[1] - b[1]
	})
	toks = slices.CompactFunc(toks, func(a, b [5]int) bool { // at the same position
		return a[0] == b[0] && a[1] == b[1]
	})
	out := make([]int, 0, 5*len(toks))
	pl, pc := 0, 0
	for _, t := range toks {
		if t[0] != pl {
			pc = 0
		}
		out = append(out, t[0]-pl, t[1]-pc, t[2], t[3], t[4])
		pl, pc = t[0], t[1]
	}
	res["data"] = out
	b, err := json.Marshal(res)
	if err != nil {
		return result
	}
	return b
}

// inlayHints returns the given inlay hints of the Go code of the given
// Goal file at their positions in the Goal file, with the Go code of the
// lines with errors, when it is not the same as their source, after them.
// The hints of the lines whose Go code is not the same as their source,
// such as tensor indexing, are omitted, as they are about the Go code.
func (s *Server) inlayHints(d *doc, result json.RawMessage) json.RawMessage {
	res, ok := s.mapJSON(result, d, false)
	if !ok {
		res = json.RawMessage("[]")
	}
	var all []any
	dec := json.NewDecoder(bytes.NewReader(res))
	dec.UseNumber()
	dec.Decode(&all)
	s.mu.Lock()
	hints := make([]any, 0, len(all))
	for _, h := range all {
		if hm, ok := h.(map[string]any); ok {
			if p, ok := hm["position"].(map[string]any); ok {
				if ln, ok := toInt(p["line"]); ok && !d.sameCode(ln) {
					continue
				}
			}
		}
		hints = append(hints, h)
	}
	for sl, code := range d.goCode {
		src := line(d.lines, sl)
		hints = append(hints, map[string]any{
			"position":    map[string]any{"line": sl, "character": s.enc.fromRune(src, len([]rune(src)))},
			"label":       "Go: " + code,
			"paddingLeft": true})
	}
	s.mu.Unlock()
	b, err := json.Marshal(hints)
	if err != nil {
		return res
	}
	return b
}

// stripSpaces returns the given string without its spaces.
func stripSpaces(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
