// Copyright (c) 2024, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"cogentcore.org/core/base/sshclient"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/text/parse/complete"
	"cogentcore.org/lab/goal/transpile"
	"cogentcore.org/lab/tensorfs"
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

	// host is the name of the ssh host of the command, from the @name
	// that starts it, if any, which is not a word of the command.
	host string

	// cmd is the name of the command, if the word is not the first.
	cmd string
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
				v := string(value)
				switch {
				case words == 0 && w.host == "" && isHostName(v):
					w.host = v[1:]
				case words == 0:
					w.cmd = v
					words++
				default:
					words++
				}
			}
			if strings.ContainsRune(";|&", r) {
				words = 0 // a new command
				w.host, w.cmd = "", ""
			}
			value = value[:0]
			start, base = i+1, i+1
		default:
			value = append(value, r)
		}
		if r == '/' || (r == ':' && isHostName(string(value[:len(value)-1]))) { // @name:path
			base = i + 1
		}
	}
	w.value = string(value)
	w.first = words == 0
	w.baseLen = len(rs) - base
	return w
}

// isHostName returns whether the given word is the name of an ssh host,
// as @name.
func isHostName(word string) bool {
	return len(word) > 1 && word[0] == '@' && !strings.ContainsAny(word, ":/")
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
// and the commands, if it is the name of the command. The files are those
// of the host that the command runs on, as for @name at the start of the
// command or the active ssh host (see [Goal.SSHActive]), or those of the
// host of a word that is @name:path, and the names of the hosts complete
// an @name that starts the command. The text that they add continues the
// word as it is typed, escaped or quoted, and ends with a / for a
// directory, and otherwise with the closing quote, if any, and a space.
// In math mode, they are those of [mathCompletions].
func (gl *Goal) completions(line string) (w shellWord, comps []shellCompletion) {
	if ml, ok := gl.mathLine(line); ok {
		return mathCompletions(ml)
	}
	w = lastShellWord(line)
	value := w.value
	host := w.host
	if host == "" && gl.SSHActive != "" && !localBuiltins[w.cmd] {
		host = gl.SSHActive
	}
	if strings.HasPrefix(value, "@") {
		if h, p, ok := strings.Cut(value[1:], ":"); ok {
			host, value = h, p
		} else if w.first && w.host == "" {
			return w, gl.hostCompletions(value[1:])
		}
	}
	var cl *sshclient.Client
	if host != "" && host != "0" {
		cl = gl.SSHClients[host]
		if cl == nil {
			return w, nil
		}
	}
	var dir, base string
	if cl != nil {
		dir, base = path.Split(value)
	} else {
		dir, base = filepath.Split(value)
	}
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
			e = "/"
			if cl == nil {
				e = string(filepath.Separator)
			}
		}
		comps = append(comps, shellCompletion{add: escapeShell(name[len(base):], w.quote) + e, name: name, icon: icon})
	}

	var entries []dirEntry
	if cl != nil {
		entries = remoteDirEntries(cl, dir)
	} else {
		entries = gl.localDirEntries(dir)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.name, ".") && !strings.HasPrefix(base, ".") {
			continue
		}
		icon := icons.File
		if entry.isDir {
			icon = icons.Folder
		}
		add(entry.name, icon, entry.isDir)
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

// hostCompletions returns the completions of the name of an ssh host,
// as @name, that starts with the given prefix, with 0 for the local host.
func (gl *Goal) hostCompletions(prefix string) []shellCompletion {
	var comps []shellCompletion
	for _, name := range append(slices.Sorted(maps.Keys(gl.SSHClients)), "0") {
		if strings.HasPrefix(name, prefix) {
			comps = append(comps, shellCompletion{add: name[len(prefix):] + " ", name: "@" + name, icon: icons.Cloud})
		}
	}
	return comps
}

// dirEntry is a file in a directory, for completion.
type dirEntry struct {
	name  string
	isDir bool
}

// localDirEntries returns the files in the given local directory, relative
// to the current directory, with symbolic links to directories as such.
func (gl *Goal) localDirEntries(dir string) []dirEntry {
	if dir == "" {
		dir = "."
	}
	if p, err := homedir.Expand(dir); err == nil {
		dir = p
	}
	if !filepath.IsAbs(dir) && gl.Config.Dir != "" {
		dir = filepath.Join(gl.Config.Dir, dir)
	}
	entries, _ := os.ReadDir(dir)
	des := make([]dirEntry, len(entries))
	for i, entry := range entries {
		isDir := entry.IsDir()
		if !isDir && entry.Type()&os.ModeSymlink != 0 {
			if fi, err := os.Stat(filepath.Join(dir, entry.Name())); err == nil {
				isDir = fi.IsDir()
			}
		}
		des[i] = dirEntry{name: entry.Name(), isDir: isDir}
	}
	return des
}

// remoteTimeout is how long completion waits for the files on an ssh host,
// as when the network is down.
var remoteTimeout = 2 * time.Second

// remoteDirEntries returns the files in the given directory on the host of
// the given ssh client, relative to its current directory, with symbolic
// links to directories as such, read with SFTP. It returns nil if they
// could not be read within [remoteTimeout].
func remoteDirEntries(cl *sshclient.Client, dir string) []dirEntry {
	switch {
	case dir == "~" || strings.HasPrefix(dir, "~/"):
		dir = cl.HomeDir + dir[1:]
	case !path.IsAbs(dir):
		dir = path.Join(cl.Dir, dir)
	}
	result := make(chan []dirEntry, 1)
	go func() {
		sc, err := cl.SFTP()
		if err != nil {
			result <- nil
			return
		}
		infos, _ := sc.ReadDir(dir)
		des := make([]dirEntry, len(infos))
		for i, fi := range infos {
			isDir := fi.IsDir()
			if !isDir && fi.Mode()&os.ModeSymlink != 0 {
				if st, err := sc.Stat(path.Join(dir, fi.Name())); err == nil {
					isDir = st.IsDir()
				}
			}
			des[i] = dirEntry{name: fi.Name(), isDir: isDir}
		}
		result <- des
	}()
	select {
	case des := <-result:
		return des
	case <-time.After(remoteTimeout):
		return nil
	}
}

// mathLine returns the given line without the # that starts a line of math
// mode, and whether it is in math mode: on such a line, or when math mode
// is on (after ##).
func (gl *Goal) mathLine(line string) (string, bool) {
	if gl.TrState.MathMode {
		return line, true
	}
	t := strings.TrimLeftFunc(line, unicode.IsSpace)
	if strings.HasPrefix(t, "#") {
		return strings.TrimLeft(t, "#"), true
	}
	return line, false
}

// mathCompletions returns the completions of the word that ends at the end
// of the given line of math mode, which are the nodes of the current
// tensorfs directory ([tensorfs.CurDir]), or of the directory of the word:
// the directories in the arguments of a tensorfs command, such as ls and
// cd, which are shell words, and all of the nodes in a Go string, as in
// get("name"). The first word of a command completes the names of those
// commands.
func mathCompletions(line string) (w shellWord, comps []shellCompletion) {
	cmds := transpile.TensorfsCommands()
	w = lastShellWord(line)
	cmd := strings.Fields(line[strings.LastIndex(line, ";")+1:])
	if !w.first && len(cmd) > 0 && slices.Contains(cmds, cmd[0]) {
		end := " "
		if w.quote != 0 {
			end = string(w.quote) + end
		}
		return w, tensorfsCompletions(w.value, w.quote, end, true)
	}
	if q, s := openString(line); q != 0 {
		w = shellWord{value: s, quote: q}
		w.baseLen = utf8.RuneCountInString(s[strings.LastIndex(s, "/")+1:])
		if q == '`' {
			q = '\'' // no escapes
		}
		return w, tensorfsCompletions(w.value, q, string(w.quote), false)
	}
	if w.first {
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd, w.value) {
				comps = append(comps, shellCompletion{add: cmd[len(w.value):] + " ", name: cmd, icon: icons.Terminal})
			}
		}
	}
	return w, comps
}

// openString returns the quote of the Go string or raw string that is open
// at the end of the given line, and its text up to there, or 0 if none is
// open.
func openString(line string) (quote rune, s string) {
	start := 0
	esc := false
	for i, r := range line {
		switch {
		case quote == 0:
			if r == '"' || r == '`' || r == '\'' {
				quote, start = r, i+1
			}
		case esc:
			esc = false
		case r == '\\' && quote != '`':
			esc = true
		case r == quote:
			quote = 0
		}
	}
	if quote == 0 || quote == '\'' { // a rune
		return 0, ""
	}
	return quote, line[start:]
}

// tensorfsCompletions returns the completions of the tensorfs path p:
// the nodes of its directory whose names start with its base, or only the
// directories if dirs is set. They are escaped for the given quote (see
// [escapeShell]), and end with the given end, or a / for a directory.
func tensorfsCompletions(p string, quote rune, end string, dirs bool) []shellCompletion {
	dir, base := path.Split(p)
	cur := tensorfs.CurDir
	if cur == nil {
		cur = tensorfs.CurRoot
	}
	if dir != "" {
		d, err := cur.DirAtPath(dir)
		if err != nil {
			return nil
		}
		cur = d
	}
	entries, err := cur.ReadDir(".")
	if err != nil {
		return nil
	}
	var comps []shellCompletion
	for _, entry := range entries {
		name := entry.Name()
		isDir := entry.IsDir()
		if !strings.HasPrefix(name, base) || (dirs && !isDir) {
			continue
		}
		c := shellCompletion{add: escapeShell(name[len(base):], quote) + end, name: name, icon: icons.TableChart}
		if isDir {
			c.add = escapeShell(name[len(base):], quote) + "/"
			c.icon = icons.Folder
		}
		comps = append(comps, c)
	}
	return comps
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
