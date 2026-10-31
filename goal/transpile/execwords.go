// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

// BackslashLiteral is whether a backslash outside of quotes and Go code is
// a literal char, unless it escapes a char that the shell gives a meaning
// (see [Escapes]), such that it is the path separator, as in C:\Users
// and cmd\goal. It is set on Windows, where backslash is the path
// separator.
var BackslashLiteral = runtime.GOOS == "windows"

// ShellSpecial are the chars, other than spaces, that the shell gives a
// meaning outside of quotes, which a backslash escapes.
const ShellSpecial = "\\\"'`{}[]|&;<>$*?#()"

// Escapes returns whether a backslash outside of quotes and Go code
// escapes the given char that follows it (0 if none), as it always does,
// except with [BackslashLiteral], when it only escapes spaces and the
// other [ShellSpecial] chars than backslash, and is otherwise literal.
func Escapes(next rune) bool {
	if !BackslashLiteral {
		return true
	}
	return next != 0 && next != '\\' && (unicode.IsSpace(next) || strings.ContainsRune(ShellSpecial, next))
}

// WordsError is an error of [ExecWords], such as an unterminated string,
// at a byte offset in the line.
type WordsError struct {
	// Pos is the byte offset of the error in the line, such as that of the
	// quote that starts an unterminated string.
	Pos int

	Msg string
}

func (e *WordsError) Error() string { return e.Msg }

func ExecWords(ln string) ([]string, error) {
	base := len(ln) - len(strings.TrimLeftFunc(ln, unicode.IsSpace)) // of ln in the line
	ln = strings.TrimSpace(ln)
	n := len(ln)
	if n == 0 {
		return nil, nil
	}

	if ln[0] == '$' {
		rest := ln[1:]
		base += 1 + len(rest) - len(strings.TrimLeftFunc(rest, unicode.IsSpace))
		ln = strings.TrimSpace(rest)
		n = len(ln)
		if n == 0 {
			return nil, nil
		}
		if ln[n-1] == '$' {
			ln = strings.TrimSpace(ln[:n-1])
			n = len(ln)
			if n == 0 {
				return nil, nil
			}
		}
	}

	word := ""
	esc := false
	dQuote := false
	bQuote := false
	sQuote := false
	brace := 0
	brack := 0
	redir := false

	var words []string
	addWord := func() {
		if brace > 0 { // always accum into one token inside brace
			return
		}
		if len(word) > 0 {
			words = append(words, word)
			word = ""
		}
	}

	atStart := true
	quoteStart, brackStart := 0, 0 // offsets of an open quote and [
	sbrack := (ln[0] == '[')
	if sbrack {
		word = "["
		addWord()
		brack++
		brackStart = base
		ln = ln[1:]
		base++
		atStart = false
	}

	for i, r := range ln {
		quote := dQuote || bQuote || sQuote
		// A // that starts a word outside of quotes and Go code starts a
		// comment, which has no words: an apostrophe in it, as in
		// // don't, is not a quote.
		if !quote && !esc && brace == 0 && word == "" && strings.HasPrefix(ln[i:], "//") {
			break
		}
		if !quote && (r == '"' || r == '`' || r == '\'' && brace == 0) {
			quoteStart = base + i
		}
		if r == '[' && brack == 0 {
			brackStart = base + i
		}

		if redir {
			redir = false
			if r == '&' {
				word += string(r)
				addWord()
				continue
			}
			if r == '>' {
				word += string(r)
				redir = true
				continue
			}
			addWord()
		}

		switch {
		case sQuote: // everything is literal until the closing quote, as in bash
			if r == '\'' {
				sQuote = false
			}
			word += string(r)
		case esc:
			// an escaped char outside of quotes and Go code is literal, as in
			// bash: the escape is removed, and the word is quoted later.
			if brace == 0 && !quote {
				word = word[:len(word)-1]
			}
			word += string(r)
			esc = false
		case r == '\\' && !quote && brace == 0 && !Escapes(nextRune(ln[i+1:])):
			word += string(r) // literal (see BackslashLiteral)
		case r == '\\':
			esc = true
			word += string(r)
		case r == '"':
			if !bQuote {
				dQuote = !dQuote
			}
			word += string(r)
		case r == '`':
			if !dQuote {
				bQuote = !bQuote
			}
			word += string(r)
		case r == '\'' && !quote && brace == 0:
			sQuote = true
			word += string(r)
		case quote: // absorbs quote -- no need to check below
			word += string(r)
		case unicode.IsSpace(r):
			addWord()
			continue // don't reset at start
		case r == '{':
			if brace == 0 {
				addWord()
				word = "{"
				addWord()
			}
			brace++
		case r == '}':
			brace--
			if brace == 0 {
				addWord()
				word = "}"
				addWord()
			}
		case r == '[':
			word += string(r)
			if atStart && brack == 0 {
				sbrack = true
				addWord()
			}
			brack++
		case r == ']':
			brack--
			if brack == 0 && sbrack { // only point of tracking brack is to get this end guy
				addWord()
				word = "]"
				addWord()
			} else {
				word += string(r)
			}
		case r == '<' || r == '>' || r == '|':
			addWord()
			word += string(r)
			redir = true
		case r == '&': // known to not be redir
			addWord()
			word += string(r)
		case r == ';':
			addWord()
			word += string(r)
			addWord()
			atStart = true
			continue // avoid reset
		default:
			word += string(r)
		}
		atStart = false
	}
	addWord()
	switch {
	case dQuote:
		return words, &WordsError{Pos: quoteStart, Msg: "unterminated \" string"}
	case bQuote:
		return words, &WordsError{Pos: quoteStart, Msg: "unterminated ` string"}
	case sQuote:
		return words, &WordsError{Pos: quoteStart, Msg: "unterminated ' string"}
	case brack > 0:
		return words, &WordsError{Pos: brackStart, Msg: "missing closing ]"}
	}
	return words, nil
}

// nextRune returns the first rune of s, or 0 if it is empty.
func nextRune(s string) rune {
	if s == "" {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

// ExecWordIsCommand returns true if given exec word is a command-like string
// (excluding any paths)
func ExecWordIsCommand(f string) bool {
	if strings.Contains(f, "(") || strings.Contains(f, "[") || strings.Contains(f, "=") {
		return false
	}
	return true
}
