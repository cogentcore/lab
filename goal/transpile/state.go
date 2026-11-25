// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"errors"
	"fmt"
	"go/format"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"cogentcore.org/core/base/logx"
	"cogentcore.org/core/base/num"
	"cogentcore.org/core/base/stringsx"
	"golang.org/x/tools/imports"
)

// State holds the transpiling state
type State struct {
	// FuncToVar translates function definitions into variable definitions,
	// which is the default for interactive use of random code fragments
	// without the complete go formatting.
	// For pure transpiling of a complete codebase with full proper Go formatting
	// this should be turned off.
	FuncToVar bool

	// MathMode is on when math mode is turned on.
	MathMode bool

	// MathRecord is state of auto-recording of data into current directory
	// in math mode.
	MathRecord bool

	// depth of delim at the end of the current line. if 0, was complete.
	ParenDepth, BraceDepth, BrackDepth, TypeDepth, DeclDepth int

	// typeBrace is the brace depth at which the current type declaration
	// started, and declParen, declBrace the depths at which the current
	// var, const or import declaration started: they end at these depths,
	// which are not 0 inside of a function.
	typeBrace, declParen, declBrace int

	// goContinues is whether the last line of Go code ends in a token after
	// which its expression or list continues on the next line (see
	// [continues]), which is then Go code too, even if it starts with a
	// string, as a command can.
	goContinues bool

	// Chunks of code lines that are accumulated during Transpile,
	// each of which should be evaluated separately, to avoid
	// issues with contextual effects from import, package etc.
	Chunks []string

	// ChunkSources are the source lines of each of the Chunks,
	// for mapping the errors of their Go code back to the source.
	ChunkSources [][]SourceLine

	// current stack of transpiled lines, that are accumulated into
	// code Chunks.
	Lines []string

	// SourceLines are the source lines of each of the Lines.
	SourceLines []SourceLine

	// stack of runtime errors.
	Errors []error

	// if this is non-empty, it is the name of the last command defined.
	// triggers insertion of the AddCommand call to add to list of defined commands.
	lastCommand string

	// math has what math mode knows about the variables and blocks
	// of the code so far.
	math mathNames

	// lastCols is the column map of the line last transpiled
	// (see [SourceLine.Cols]).
	lastCols []token.Pos

	// curSrc is the source line being transpiled.
	curSrc string
}

// NewState returns a new transpiling state; mostly for testing
func NewState() *State {
	st := &State{FuncToVar: true}
	return st
}

// TranspileCode processes each line of given code,
// adding the results to the LineStack
func (st *State) TranspileCode(code string) {
	lns := strings.Split(code, "\n")
	n := len(lns)
	if n == 0 {
		return
	}
	for i := 0; i < n; i++ {
		ln := lns[i]
		// a top-level declaration is a chunk (not one in a function)
		hasDecl := st.DeclDepth > 0 && st.declBrace == 0 && st.declParen == 0
		// a raw string or block comment over lines is in one statement with
		// its lines
		nsrc := 1
		if !st.MathMode {
			if openBlockComment(ln) {
				for i+nsrc < n {
					next := lns[i+nsrc]
					ln += "\n" + next
					nsrc++
					if strings.Contains(next, "*/") {
						break
					}
				}
			} else {
				for i+nsrc < n && openRawString(ln) {
					ln += "\n" + lns[i+nsrc]
					nsrc++
				}
			}
		}
		st.curSrc = ln
		st.lastCols = nil
		line, paren, brace, brack := st.lineNumber(), st.ParenDepth, st.BraceDepth, st.BrackDepth
		tl := st.TranspileLine(ln)
		if st.ParenDepth < 0 || st.BraceDepth < 0 || st.BrackDepth < 0 {
			if !st.checkClosers(ln, line, paren, brace, brack) {
				st.ParenDepth, st.BraceDepth, st.BrackDepth = max(st.ParenDepth, 0), max(st.BraceDepth, 0), max(st.BrackDepth, 0)
			}
		}
		if nsrc == 1 {
			st.AddLine(tl)
			st.SourceLines[len(st.SourceLines)-1] = SourceLine{Src: ln, Cols: st.lastCols}
		} else {
			st.addLines(tl, lns[i:i+nsrc])
			i += nsrc - 1
		}
		if st.BraceDepth == 0 && st.BrackDepth == 0 && st.ParenDepth == 1 && st.lastCommand != "" {
			st.lastCommand = ""
			nl := len(st.Lines)
			st.Lines[nl-1] = st.Lines[nl-1] + ")"
			st.ParenDepth--
		}
		if hasDecl && st.DeclDepth == 0 { // break at decl
			st.AddChunk()
		}
	}
}

// TranspileFile transpiles the given input goal file to the
// given output Go file. If no existing package declaration
// is found, then package main and func main declarations are
// added. This also affects how functions are interpreted.
func (st *State) TranspileFile(in string, out string) error {
	b, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	res, err := st.GoFile(in, out, string(b))
	if err != nil {
		fmt.Println(err.Error())
	}
	werr := os.WriteFile(out, res, 0666)
	return errors.Join(err, werr)
}

// lineFilename returns the name of the input goal file in for the //line
// directive of the output Go file out: relative to the directory of out,
// as the Go tools resolve relative //line names, so that the generated
// file is the same however the paths are given (e.g., absolute paths
// from an editor, or relative ones from goal build).
func lineFilename(in, out string) string {
	ain, err1 := filepath.Abs(in)
	aout, err2 := filepath.Abs(out)
	if err1 != nil || err2 != nil {
		return in
	}
	rel, err := filepath.Rel(filepath.Dir(aout), ain)
	if err != nil {
		return in
	}
	return filepath.ToSlash(rel)
}

// GoFile returns the Go code of the given code of the given input goal
// file, as [State.TranspileFile] writes it to the given output Go file,
// with any error, in which case the Go code is not formatted.
func (st *State) GoFile(in, out, code string) ([]byte, error) {
	lns := stringsx.SplitLines(code)
	hasPackage := false
	for _, ln := range lns {
		if strings.HasPrefix(ln, "package ") {
			hasPackage = true
			break
		}
	}
	if hasPackage {
		st.FuncToVar = false // use raw functions
	}
	st.TranspileCode(code)
	st.FuncToVar = true

	hdr := strings.Join(scriptHeader(ScriptImports), "\n") + "\n"
	src := st.Code()
	gen := fmt.Sprintf("// Code generated by \"goal build\"; DO NOT EDIT.\n//line %s:1\n", lineFilename(in, out))
	var bsrc, res []byte
	var err error
	if hasPackage {
		bsrc = []byte(gen + src)
		res, err = format.Source(bsrc)
	} else {
		bsrc = []byte(gen + hdr + src + "\n}")
		res, err = imports.Process(out, bsrc, nil)
	}
	if err != nil {
		return bsrc, err
	}
	return res, st.DepthError()
}

// TotalDepth returns the sum of any unresolved paren, brace, or bracket depths.
func (st *State) TotalDepth() int {
	return num.Abs(st.ParenDepth) + num.Abs(st.BraceDepth) + num.Abs(st.BrackDepth)
}

// ResetCode resets the stack of transpiled code
func (st *State) ResetCode() {
	st.Chunks = nil
	st.ChunkSources = nil
	st.Lines = nil
	st.SourceLines = nil
}

// ResetDepth resets the current depths to 0
func (st *State) ResetDepth() {
	st.ParenDepth, st.BraceDepth, st.BrackDepth, st.TypeDepth, st.DeclDepth = 0, 0, 0, 0, 0
	st.goContinues = false
}

// DepthError reports an error if any of the parsing depths are not zero,
// to be called at the end of transpiling a complete block of code.
func (st *State) DepthError() error {
	if st.TotalDepth() == 0 {
		return nil
	}
	str := ""
	if st.ParenDepth != 0 {
		str += fmt.Sprintf("Incomplete parentheses (), remaining depth: %d\n", st.ParenDepth)
	}
	if st.BraceDepth != 0 {
		str += fmt.Sprintf("Incomplete braces {}, remaining depth: %d\n", st.BraceDepth)
	}
	if st.BrackDepth != 0 {
		str += fmt.Sprintf("Incomplete brackets [], remaining depth: %d\n", st.BrackDepth)
	}
	if str != "" {
		slog.Error(str)
		return errors.New(str)
	}
	return nil
}

// addLines adds the Go code of the given source lines, of one statement
// with a raw string over them, as one line for each, so that the lines of
// the Go code are those of the source: the newlines of the code are those
// of its raw strings, but any others are joined.
func (st *State) addLines(code string, src []string) {
	lns := strings.Split(code, "\n")
	for len(lns) > len(src) {
		lns[len(lns)-2] += " " + lns[len(lns)-1]
		lns = lns[:len(lns)-1]
	}
	for i, s := range src {
		ln := ""
		if i < len(lns) {
			ln = lns[i]
		}
		st.AddLine(ln)
		st.SourceLines[len(st.SourceLines)-1] = SourceLine{Src: s}
	}
}

// openBlockComment returns whether the given line starts a /* block
// comment that it does not close.
func openBlockComment(ln string) bool {
	ln = strings.TrimSpace(ln)
	return strings.HasPrefix(ln, "/*") && !strings.Contains(ln[2:], "*/")
}

// openRawString returns whether the given line has a ` raw string that is
// not closed on it, outside of other strings, runes and comments, as Go
// scans it.
func openRawString(ln string) bool {
	return slices.Contains(goScanErrors(ln), "raw string literal not terminated")
}

// AddLine adds line on the stack
func (st *State) AddLine(ln string) {
	st.Lines = append(st.Lines, ln)
	st.SourceLines = append(st.SourceLines, SourceLine{Src: ln})
}

// Code returns the current transpiled lines,
// split into chunks that should be compiled separately.
func (st *State) Code() string {
	st.AddChunk()
	if len(st.Chunks) == 0 {
		return ""
	}
	return strings.Join(st.Chunks, "\n")
}

// AddChunk adds current lines into a chunk of code
// that should be compiled separately.
func (st *State) AddChunk() {
	if len(st.Lines) == 0 {
		return
	}
	st.Chunks = append(st.Chunks, strings.Join(st.Lines, "\n"))
	st.ChunkSources = append(st.ChunkSources, st.SourceLines)
	st.Lines = nil
	st.SourceLines = nil
}

// AddError adds the given error to the error stack if it is non-nil,
// and calls the Cancel function if set, to stop execution.
// This is the main way that goal errors are handled.
// It also prints the error.
func (st *State) AddError(err error) error {
	if err == nil {
		return nil
	}
	st.Errors = append(st.Errors, err)
	logx.PrintlnError(err)
	return err
}

// lineNumber returns the 1-based line number of the source line being
// transpiled, in the code since [State.ResetCode].
func (st *State) lineNumber() int {
	n := len(st.SourceLines) + 1
	for _, ch := range st.ChunkSources {
		n += len(ch)
	}
	return n
}

// addSourceError adds an [Error] at the given 1-based column of the source
// line being transpiled.
func (st *State) addSourceError(col int, msg string) {
	st.AddError(&Error{Line: st.lineNumber(), Col: col, Src: st.curSrc, Msg: msg})
}
