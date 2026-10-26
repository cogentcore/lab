// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package interpreter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cogentcore.org/core/base/sshclient/sshtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sshTest runs the given lines of Goal code, one at a time, in an
// interpreter with an ssh connection named h to a test server, connected
// with gossh, in a new local directory. It returns the output and the
// errors of each line, and the local directory and the root directory
// of the server.
func sshTest(t *testing.T, setup func(local, remote string), lines ...string) (out string, errs []string, local, remote string) {
	t.Helper()
	s := sshtest.NewServer(t)
	local = t.TempDir()
	t.Chdir(local)
	if setup != nil {
		setup(local, s.Root)
	}
	var buf syncBuffer
	in := NewInterpreter(Options{Stdout: &buf, Stderr: &buf})
	in.Goal.SSH.User.KeyPath = s.KeyPath
	defer in.Goal.CloseSSH()
	for _, line := range append([]string{"gossh " + s.Addr + " h", "@0"}, lines...) {
		if _, _, err := in.Eval(line); err != nil {
			errs = append(errs, line+": "+err.Error())
		}
	}
	return buf.String(), errs, local, s.Root
}

// files writes the given files, as name: content, in the given directory.
func files(t *testing.T, dir string, files map[string]string) {
	for name, content := range files {
		fn := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(fn), 0755))
		require.NoError(t, os.WriteFile(fn, []byte(content), 0644))
	}
}

// hasFiles checks that the given files, as name: content, are in the
// given directory.
func hasFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if assert.NoError(t, err, name) {
			assert.Equal(t, content, strings.TrimSpace(string(b)), name)
		}
	}
}

// TestSSH tests the commands of the ssh connections of the shell
// documented in docs/content/shell.md.
func TestSSH(t *testing.T) {
	remoteFiles := map[string]string{"hostfile.tsv": "host", "b.tsv": "b", "c.txt": "c", "sub/x.tsv": "x"}
	setup := func(local, remote string) {
		files(t, remote, remoteFiles)
		files(t, local, map[string]string{"local.tsv": "local", "ldir/y.txt": "y"})
	}
	out, errs, local, remote := sshTest(t, setup,
		// commands on the host
		"@h ls",
		`server := "@h"`,
		`${server} cat c.txt$`,
		"@h cat hostfile.tsv > @0:fromhost.tsv",
		"@h cd sub",
		"@h ls",
		"@h cd",
		"echo piped | @h cat",
		"ls *.tsv | @h cat > files.txt",
		"echo hi > @0:hi.txt",
		"echo hi > @h:hi.txt",
		"echo more >> @h:hi.txt",
		"@h cat c.txt > @h:c2.txt",
		// the default host
		"@h",
		"cat b.tsv",
		"scp @h:c.txt fromdefault.txt",
		"@0",
		"gossh @h",
		"cat c.txt",
		"gossh @0",
		// scp
		"scp @h:hostfile.tsv copy.tsv",
		"scp @h:b.tsv .",
		"scp @h:*.tsv @0:data/",
		"scp @h:sub .",
		"scp local.tsv @0:data2/",
		"scp local.tsv @h:",
		"scp local.tsv @h:up/",
		"scp ldir @h:",
		"scp *.tsv @h:all",
		"scp @h:hostfile.tsv @h:sub/again.tsv",
		"gossh close",
	)
	assert.Empty(t, errs)
	t.Log(out)
	for _, want := range []string{"hostfile.tsv", "c", "x.tsv", "piped", "b"} {
		assert.Contains(t, out, want)
	}
	hasFiles(t, local, map[string]string{
		"fromhost.tsv": "host", "copy.tsv": "host", "b.tsv": "b",
		"data/hostfile.tsv": "host", "data/b.tsv": "b", "sub/x.tsv": "x",
		"data2/local.tsv": "local", "hi.txt": "hi", "fromdefault.txt": "c",
	})
	hasFiles(t, remote, map[string]string{
		"local.tsv": "local", "up/local.tsv": "local", "ldir/y.txt": "y",
		"all/local.tsv": "local", "all/fromhost.tsv": "host", "sub/again.tsv": "host",
		"files.txt": "fromhost.tsv\nlocal.tsv", "hi.txt": "hi\nmore", "c2.txt": "c",
	})
}

// TestSSHErrors tests the errors of the ssh commands.
func TestSSHErrors(t *testing.T) {
	_, errs, _, _ := sshTest(t, nil,
		"scp @h:none.tsv .",
		"scp @nohost:a.tsv .",
		"scp @h.a.tsv .",
		"scp a.tsv",
		"@nohost ls",
	)
	t.Log(strings.Join(errs, "\n"))
	assert.Len(t, errs, 5)
}
