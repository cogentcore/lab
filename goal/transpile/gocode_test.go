// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transpile

import (
	"path/filepath"
	"strings"
	"testing"

	"cogentcore.org/core/text/textpos"
	"github.com/stretchr/testify/assert"
)

func TestTranspileGo(t *testing.T) {
	src := "// test\nx := 10\nfor i := range x {\n\tfmt.Println(\"héllo\", i)\n}\nls -la\n"
	gc := TranspileGo(src, filepath.Join(t.TempDir(), "test.go"), GoOptions{FindImports: true})
	// a script is in a main function, with the imports it uses
	assert.False(t, gc.HasPackage)
	assert.Contains(t, gc.Imports, `"fmt"`)
	assert.NotContains(t, gc.Imports, `"cogentcore.org/lab/goal/goalib"`)
	assert.Equal(t, "func main() {", gc.Lines[gc.Offset-3])
	// each line of the source is a line of the Go code
	assert.Equal(t, "x := 10", gc.Lines[gc.Offset+1])
	assert.Equal(t, `fmt.Println("héllo", i)`, gc.Lines[gc.Offset+3])
	assert.Equal(t, `goalrun.Run("ls", "-la")`, gc.Lines[gc.Offset+5])
	assert.Equal(t, "}", gc.Lines[gc.Offset+7])
	assert.Empty(t, gc.Errors)

	// positions are mapped both ways, in runes, through the tokens,
	// including the indentation that the Go code does not have
	pos := func(ln, ch int) textpos.Pos { return textpos.Pos{Line: ln, Char: ch} }
	goal := []textpos.Pos{pos(1, 0), pos(1, 5), pos(3, 1), pos(3, 6), pos(3, 22), pos(3, 24)}
	gos := []textpos.Pos{pos(1, 0), pos(1, 5), pos(3, 0), pos(3, 5), pos(3, 21), pos(3, 23)}
	for i, gp := range goal {
		want := gos[i]
		want.Line += gc.Offset
		p, ok := gc.ToGo(gp)
		assert.True(t, ok)
		assert.Equal(t, want, p, "ToGo %v", gp)
		p, ok = gc.ToGoal(want)
		assert.True(t, ok)
		assert.Equal(t, gp, p, "ToGoal %v", want)
	}
	// a position at a space is after the token before it, as for
	// the type of x, after it
	p, ok := gc.ToGoal(pos(gc.Offset+1, 1))
	assert.True(t, ok)
	assert.Equal(t, pos(1, 1), p)
	// the code of a command is at its start
	p, ok = gc.ToGoal(pos(gc.Offset+5, 10))
	assert.True(t, ok)
	assert.Equal(t, pos(5, 0), p)
	// the header and the end are not in the source
	_, ok = gc.ToGoal(pos(gc.Offset-1, 0))
	assert.False(t, ok)
	_, ok = gc.ToGoal(pos(gc.Offset+7, 0))
	assert.False(t, ok)
}

func TestTranspileGoPackage(t *testing.T) {
	src := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n"
	gc := TranspileGo(src, filepath.Join(t.TempDir(), "test.go"), GoOptions{FindImports: true})
	// a file with a package has its own imports and no main function
	assert.True(t, gc.HasPackage)
	assert.Equal(t, 1, gc.Offset)
	assert.Equal(t, "package main", gc.Lines[1])
	assert.Equal(t, strings.Count(src, "\n")+1, gc.NumSourceLines())
	assert.Equal(t, "}", gc.Lines[gc.Offset+6])
	p, ok := gc.ToGoal(textpos.Pos{Line: 6, Char: 4})
	assert.True(t, ok)
	assert.Equal(t, textpos.Pos{Line: 5, Char: 5}, p)
}

// TestTranspileGoTokenStart tests that the tokens after a space that has
// the column of the token after it, as before a type, and the & that the
// Go code has spaces around, are at their positions in the source. They
// were one character after them, as the semantic tokens of Goal files.
func TestTranspileGoTokenStart(t *testing.T) {
	src := "package x\n\nfunc f(a *T, b uint32) {\n\tvar caP, caD float32\n\tg(a, &caP, &caD)\n}\n"
	gc := TranspileGo(src, filepath.Join(t.TempDir(), "test.go"), GoOptions{})
	toGoal := func(goLine int, tok string, n int) textpos.Pos {
		ln := gc.Offset + goLine
		i := 0
		for range n { // the nth occurrence of tok
			i += strings.Index(gc.Lines[ln][i:], tok) + 1
		}
		p, ok := gc.ToGoal(textpos.Pos{Line: ln, Char: i - 1})
		assert.True(t, ok)
		return p
	}
	srcPos := func(line int, tok string, n int) textpos.Pos {
		sl := strings.Split(src, "\n")[line]
		i := 0
		for range n {
			i += strings.Index(sl[i:], tok) + 1
		}
		return textpos.Pos{Line: line, Char: i - 1}
	}
	for _, c := range []struct {
		line int
		tok  string
		n    int
	}{{2, "*", 1}, {2, "T", 1}, {2, "uint32", 1}, {3, "float32", 1}, {3, "caD", 1},
		{4, "&", 1}, {4, "caP", 1}, {4, "&", 2}, {4, "caD", 1}} {
		assert.Equal(t, srcPos(c.line, c.tok, c.n), toGoal(c.line, c.tok, c.n), "%s %d on line %d", c.tok, c.n, c.line)
	}
}

func TestTranspileGoErrors(t *testing.T) {
	src := "x := 1\nfor {\n"
	gc := TranspileGo(src, filepath.Join(t.TempDir(), "test.go"), GoOptions{FindImports: true, Imports: []string{`"fmt"`}})
	// the imports of the code with errors are the given ones
	assert.Equal(t, []string{`"fmt"`}, gc.Imports)
	assert.Equal(t, "x := 1", gc.Lines[gc.Offset])
}
