// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"fmt"
	"go/token"
	"strings"
)

// SourceLine is a line of Goal source code, with the map from the columns
// of the Go code that it is transpiled to back to its own columns, so that
// the errors of the Go code can be shown in the Goal source.
type SourceLine struct {

	// Src is the Goal source line.
	Src string

	// Cols is the 1-based column in Src of each byte of the Go code,
	// 0 if not known (see [Tokens.CodeMap]).
	Cols []token.Pos
}

// SourceCol returns the 1-based column in the source line of the given
// 1-based column of its Go code, or 1 if it is not known.
func (sl *SourceLine) SourceCol(col int) int {
	if len(sl.Cols) == 0 {
		return 1
	}
	c := int(sl.Cols[min(max(col-1, 0), len(sl.Cols)-1)])
	return max(c, 1)
}

// Error is an error in Goal code, at a position in its source code, such
// as an error of the Go code that it is transpiled to, at the position in
// the Goal code that it comes from.
type Error struct {

	// Line and Col are the 1-based line and column (in bytes) of the
	// error in the source code, 0 if not known.
	Line, Col int

	// Src is the source line of the error.
	Src string

	// Msg is the message of the error, without its position.
	Msg string

	// Err is the underlying error, if any.
	Err error
}

// Error returns the message of the error with its position, and the
// source line with a caret under the column of the error.
func (e *Error) Error() string {
	if e.Line == 0 {
		return e.Msg
	}
	s := fmt.Sprintf("%d:%d: %s", e.Line, e.Col, e.Msg)
	if e.Src == "" {
		return s
	}
	src := strings.ReplaceAll(e.Src, "\t", " ")
	return s + "\n    " + src + "\n    " + strings.Repeat(" ", max(e.Col-1, 0)) + "^"
}

func (e *Error) Unwrap() error { return e.Err }

// CodePos returns the 1-based line and column of the error in the source
// code, which tools such as editors use to show where it is.
func (e *Error) CodePos() (line, col int) { return e.Line, e.Col }

// CodeMessage returns the message of the error, without its position.
func (e *Error) CodeMessage() string { return e.Msg }

// SourceError returns an [Error] at the given 1-based line and column of
// the given chunk of Go code (see [State.Chunks]) of the given
// SourceLines of all the chunks, at its position in the Goal source code:
// the line of the chunk, after the lines of the chunks before it.
func SourceError(sources [][]SourceLine, chunk, line, col int, msg string, err error) *Error {
	e := &Error{Msg: msg, Err: err}
	if chunk >= len(sources) || line < 1 || line > len(sources[chunk]) {
		return e
	}
	e.Line = line
	for _, ch := range sources[:chunk] {
		e.Line += len(ch)
	}
	sl := &sources[chunk][line-1]
	e.Src = sl.Src
	e.Col = sl.SourceCol(col)
	return e
}
