// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"fmt"
	"go/token"
)

// bracketPairs maps each closing bracket to its opening one.
var bracketPairs = map[token.Token]token.Token{
	token.RPAREN: token.LPAREN,
	token.RBRACE: token.LBRACE,
	token.RBRACK: token.LBRACK,
}

// openBracket is an open bracket of the source code, at a 1-based line
// (of the code since [State.ResetCode]) and column.
type openBracket struct {
	tok       token.Token
	line, col int
	src       string
}

// checkClosers adds an error for the first closing bracket of the given
// source line, at the given line number, that has no opening bracket,
// given the depths of the brackets before the line, as [State.ParenDepth]
// etc, which are then reset, so that the code after it is not taken to be
// inside of brackets. It returns whether it found one.
func (st *State) checkClosers(src string, line int, paren, brace, brack int) bool {
	depths := map[token.Token]int{token.LPAREN: paren, token.LBRACE: brace, token.LBRACK: brack}
	for _, t := range TokensFromString(src) {
		if _, ok := depths[t.Tok]; ok {
			depths[t.Tok]++
			continue
		}
		open, ok := bracketPairs[t.Tok]
		if !ok {
			continue
		}
		depths[open]--
		if depths[open] < 0 {
			st.AddError(&Error{Line: line, Col: int(t.Pos), Src: src, Msg: fmt.Sprintf("%s without a matching %s", t.Tok, open)})
			st.ParenDepth, st.BraceDepth, st.BrackDepth = 0, 0, 0
			return true
		}
	}
	return false
}

// UnclosedError returns an error for the last bracket of the code since
// [State.ResetCode] that is not closed, or nil if there is none: for
// complete code, such as that of a file, as opposed to the lines that
// an interactive shell transpiles one at a time, which can continue.
func (st *State) UnclosedError() error {
	var stack []openBracket
	line := 0
	scan := func(lines []SourceLine) {
		for _, sl := range lines {
			line++
			for _, t := range TokensFromString(sl.Src) {
				switch t.Tok {
				case token.LPAREN, token.LBRACE, token.LBRACK:
					stack = append(stack, openBracket{tok: t.Tok, line: line, col: int(t.Pos), src: sl.Src})
				case token.RPAREN, token.RBRACE, token.RBRACK:
					if n := len(stack); n > 0 && stack[n-1].tok == bracketPairs[t.Tok] {
						stack = stack[:n-1]
					}
				}
			}
		}
	}
	for _, ch := range st.ChunkSources {
		scan(ch)
	}
	scan(st.SourceLines)
	if len(stack) == 0 {
		return nil
	}
	ob := stack[len(stack)-1]
	closer := map[token.Token]string{token.LPAREN: ")", token.LBRACE: "}", token.LBRACK: "]"}[ob.tok]
	return &Error{Line: ob.line, Col: ob.col, Src: ob.src, Msg: fmt.Sprintf("%s is not closed: missing %s", ob.tok, closer)}
}
