// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"cogentcore.org/core/icons"
	"cogentcore.org/core/text/parse/complete"
	"github.com/mitchellh/go-homedir"
)

// shellWord is the word of a shell command line that ends at the cursor,
// as the shell reads words (see transpile.ExecWords): words are separated
// by spaces, which a backslash escapes, and can be quoted with " or '.
type shellWord struct {

	// value is the word without its escapes and quotes.
	value string

	// quote is the quote that is open at the cursor, if any.
	quote rune

	// first is whether the word is the first of its command,
	// which is the command name.
	first bool

	// baseLen is the number of runes of the word as typed after its last
	// path separator, which completion replaces.
	baseLen int
}

// lastShellWord returns the word of the given shell command line that ends
// at its end.
func lastShellWord(line string) shellWord {
	var w shellWord
	var value []rune
	esc := false
	start, base := 0, 0 // rune indexes of the start of the word and its base
	words := 0          // number of words before this one in the command
	rs := []rune(line)
	for i, r := range rs {
		switch {
		case esc:
			value = append(value, r)
			esc = false
		case w.quote == '\'':
			if r == '\'' {
				w.quote = 0
			} else {
				value = append(value, r)
			}
		case r == '\\':
			esc = true
		case w.quote == '"':
			if r == '"' {
				w.quote = 0
			} else {
				value = append(value, r)
			}
		case r == '"' || r == '\'':
			w.quote = r
		case unicode.IsSpace(r) || strings.ContainsRune(";|&<>", r):
			if len(value) > 0 || start < i {
				words++
			}
			if strings.ContainsRune(";|&", r) {
				words = 0 // a new command
			}
			value = value[:0]
			start, base = i+1, i+1
		default:
			value = append(value, r)
		}
		if r == '/' {
			base = i + 1
		}
	}
	w.value = string(value)
	w.first = words == 0
	w.baseLen = len(rs) - base
	return w
}

// escapeShell returns s as it is typed in a shell word that has the given
// quote open (0 for none), such that the shell reads it as s (see
// transpile.ExecWords): outside of quotes, spaces and the other chars that
// the shell gives a meaning are escaped with a backslash; inside " quotes,
// which are Go strings, " and \ are; inside ' quotes, nothing is.
func escapeShell(s string, quote rune) string {
	var b strings.Builder
	for _, r := range s {
		switch quote {
		case '\'':
		case '"':
			if r == '"' || r == '\\' {
				b.WriteRune('\\')
			}
		default:
			if unicode.IsSpace(r) || strings.ContainsRune("\\\"'`{}[]|&;<>$*?#()", r) {
				b.WriteRune('\\')
			}
		}
		b.WriteRune(r)
	}
	return b.String()
}

// shellCompletion is a completion of the word at the cursor: the text to
// add at the cursor, and the name that it completes.
type shellCompletion struct {
	add  string
	name string
	icon icons.Icon
}

// completions returns the completions of the word that ends at the end of
// the given shell command line: the files of its directory whose names
// start with its base (hidden files only if its base starts with a .),
// and the commands, if it is the name of the command. The text that they
// add continues the word as it is typed, escaped or quoted, and ends with
// a / for a directory, and otherwise with the closing quote, if any, and a
// space.
func (gl *Goal) completions(line string) (w shellWord, comps []shellCompletion) {
	w = lastShellWord(line)
	dir, base := filepath.Split(w.value)
	end := " "
	if w.quote != 0 {
		end = string(w.quote) + end
	}
	add := func(name string, icon icons.Icon, isDir bool) {
		if !strings.HasPrefix(name, base) {
			return
		}
		e := end
		if isDir {
			e = string(filepath.Separator)
		}
		comps = append(comps, shellCompletion{add: escapeShell(name[len(base):], w.quote) + e, name: name, icon: icon})
	}

	path := dir
	if path == "" {
		path = "."
	}
	if p, err := homedir.Expand(path); err == nil {
		path = p
	}
	if !filepath.IsAbs(path) && gl.Config.Dir != "" {
		path = filepath.Join(gl.Config.Dir, path)
	}
	entries, _ := os.ReadDir(path)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		isDir := entry.IsDir()
		if !isDir && entry.Type()&os.ModeSymlink != 0 {
			if fi, err := os.Stat(filepath.Join(path, name)); err == nil {
				isDir = fi.IsDir()
			}
		}
		icon := icons.File
		if isDir {
			icon = icons.Folder
		}
		add(name, icon, isDir)
	}
	if w.first && dir == "" {
		for cmd := range gl.Builtins {
			add(cmd, icons.Terminal, false)
		}
		for cmd := range gl.Commands {
			add(cmd, icons.Terminal, false)
		}
	}
	slices.SortFunc(comps, func(a, b shellCompletion) int { return strings.Compare(a.name, b.name) })
	comps = slices.CompactFunc(comps, func(a, b shellCompletion) bool { return a.name == b.name })
	return w, comps
}

// CompleteMatch is the [complete.MatchFunc] for the shell. The completions
// are those of the word at the cursor (see [Goal.completions]), whose Text
// replaces its base, the seed.
func (gl *Goal) CompleteMatch(data any, text string, posLine, posChar int) (md complete.Matches) {
	rs := []rune(text)
	rs = rs[:min(posChar, len(rs))]
	w, comps := gl.completions(string(rs))
	md.Seed = string(rs[len(rs)-w.baseLen:])
	for _, c := range comps {
		md.Matches = append(md.Matches, complete.Completion{
			Text: md.Seed + strings.TrimSuffix(c.add, " "),
			Icon: c.icon,
			Desc: c.name,
		})
	}
	return md
}

// CompleteEdit is the [complete.EditFunc] for the shell.
func (gl *Goal) CompleteEdit(data any, text string, cursorPos int, completion complete.Completion, seed string) (ed complete.Edit) {
	return complete.EditWord(text, cursorPos, completion.Text, seed)
}

// ReadlineCompleter implements [github.com/cogentcore/readline.AutoCompleter].
type ReadlineCompleter struct {
	Goal *Goal
}

// Do returns the completions of the word at the cursor pos of the line:
// the text that each adds at the cursor, and the number of runes before
// the cursor that they share, the base of the word (see [Goal.completions]).
func (rc *ReadlineCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	w, comps := rc.Goal.completions(string(line[:pos]))
	for _, c := range comps {
		newLine = append(newLine, []rune(c.add))
	}
	return newLine, w.baseLen
}
