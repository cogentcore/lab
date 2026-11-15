// Copyright (c) 2026, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goal

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/base/logx"
	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// syncBuffer is a buffer that can be written in different goroutines, as
// by the commands that run in the background.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func (b *syncBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf.Reset()
}

// newTestGoal returns a new goal whose standard output and error go to
// the returned buffer.
func newTestGoal(t *testing.T) (*Goal, *syncBuffer) {
	t.Helper()
	gl := NewGoal()
	buf := &syncBuffer{}
	gl.Config.StdIO.Out = buf
	gl.Config.StdIO.Err = buf
	return gl, buf
}

// setHome sets the home directory to the given one for the test,
// resetting the cache of the home directory.
func setHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	homedir.Reset()
	t.Cleanup(homedir.Reset)
}

// testCmdIO returns a CmdIO of the goal that writes to the given buffer.
func testCmdIO(gl *Goal, buf *syncBuffer) *exec.CmdIO {
	cio := exec.NewCmdIO(&gl.Config)
	cio.Out = buf
	cio.Err = buf
	return cio
}

// TestCd tests the cd builtin, which changes the directory of the goal.
func TestCd(t *testing.T) {
	t.Chdir(t.TempDir())
	gl, _ := newTestGoal(t)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, gl.Cd(nil, dir))
	assert.Equal(t, dir, gl.Config.Dir)
	wd, _ := os.Getwd()
	assert.Equal(t, dir, wd)

	require.NoError(t, os.Mkdir(filepath.Join(dir, "sub"), 0o755))
	gl.Run("cd", "sub")
	assert.Equal(t, filepath.Join(dir, "sub"), gl.Config.Dir)

	err = gl.Cd(nil, "a", "b")
	assert.Error(t, err, "too many arguments")
	err = gl.Cd(nil, filepath.Join(dir, "missing"))
	assert.ErrorContains(t, err, "cd:")
	setHome(t, dir)
	require.NoError(t, gl.Cd(nil))
	assert.Equal(t, dir, gl.Config.Dir, "home without an argument")
}

// TestSetUnset tests the set and unset builtins of environment variables
// (not on macOS, where they also set the environment of launchd).
func TestSetUnset(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("set and unset change the environment of launchd on macOS")
	}
	gl, _ := newTestGoal(t)
	t.Setenv("GOAL_TEST_VAR", "")
	require.NoError(t, gl.Set(nil, "GOAL_TEST_VAR", "value"))
	assert.Equal(t, "value", os.Getenv("GOAL_TEST_VAR"))
	require.NoError(t, gl.Set(nil, "GOAL_TEST_VAR", "/a:/b:/a"))
	assert.Equal(t, "/a:/b", os.Getenv("GOAL_TEST_VAR"), "a list is made unique")
	require.NoError(t, gl.Unset(nil, "GOAL_TEST_VAR"))
	_, has := os.LookupEnv("GOAL_TEST_VAR")
	assert.False(t, has)
}

// TestBuiltinArgErrors tests the errors of builtins with the wrong
// number of arguments.
func TestBuiltinArgErrors(t *testing.T) {
	gl, _ := newTestGoal(t)
	assert.Error(t, gl.Set(nil, "a"))
	assert.Error(t, gl.Unset(nil))
	assert.Error(t, gl.Kill(nil))
	assert.Error(t, gl.Fg(nil))
	assert.Error(t, gl.Fg(nil, "notajob"))
	assert.Error(t, gl.AddPath(nil))
	assert.Error(t, gl.Which(nil))
	assert.Error(t, gl.Source(nil))
	assert.Error(t, gl.GoSSH(nil))
	assert.Error(t, gl.Scp(nil, "one"))
	assert.Error(t, gl.Scp(nil, "@host", "b"), "a host location needs a path")
}

// TestAddPath tests the add-path builtin, which adds unique paths, with
// the home directory expanded, to the PATH.
func TestAddPath(t *testing.T) {
	gl, _ := newTestGoal(t)
	home := t.TempDir()
	setHome(t, home)
	t.Setenv("PATH", "/usr/bin:/bin")
	require.NoError(t, gl.AddPath(nil, "~/bin", "/bin"))
	assert.Equal(t, "/usr/bin:/bin:"+filepath.Join(home, "bin"), os.Getenv("PATH"))
	assert.Equal(t, []string{"a", "b"}, AddHomeExpand([]string{"a"}, "b", "a"))
}

// TestWhich tests the which builtin, for commands, builtins, and
// executables.
func TestWhich(t *testing.T) {
	gl, buf := newTestGoal(t)
	gl.AddCommand("mycmd", func(args ...string) {})
	cio := testCmdIO(gl, buf)
	require.NoError(t, gl.Which(cio, "mycmd"))
	assert.Contains(t, buf.String(), "mycmd is a user-defined command")
	buf.Reset()
	require.NoError(t, gl.Which(cio, "cd"))
	assert.Contains(t, buf.String(), "cd is a goal builtin command")
	if runtime.GOOS != "windows" {
		buf.Reset()
		require.NoError(t, gl.Which(cio, "sh"))
		assert.Contains(t, buf.String(), "sh")
	}
}

// TestCommands tests user-defined commands, run by RunCommands as in a
// Makefile-style goal script, and the command line arguments.
func TestCommands(t *testing.T) {
	gl, _ := newTestGoal(t)
	var ran []string
	gl.AddCommand("build", func(args ...string) { ran = append(ran, "build") })
	gl.AddCommand("test", func(args ...string) { ran = append(ran, "test") })
	require.NoError(t, gl.RunCommands([]any{"build", "test"}))
	assert.Equal(t, []string{"build", "test"}, ran)
	assert.Error(t, gl.RunCommands([]any{"missing"}))

	gl.CliArgs = []any{"a", 1}
	assert.Equal(t, []any{"a", 1}, gl.Args())
}

// TestHistory tests adding, saving, opening, and printing the history.
func TestHistory(t *testing.T) {
	gl, buf := newTestGoal(t)
	for _, l := range []string{"ls", "cd /", "echo hi"} {
		gl.AddHistory(l)
	}
	fn := filepath.Join(t.TempDir(), "hist")
	require.NoError(t, gl.SaveHistory(2, fn))
	b, err := os.ReadFile(fn)
	require.NoError(t, err)
	assert.Equal(t, "cd /\necho hi", string(b), "the last 2 lines")
	require.NoError(t, gl.SaveHistory(0, fn))

	g2, _ := newTestGoal(t)
	require.NoError(t, g2.OpenHistory(fn))
	assert.Equal(t, []string{"ls", "cd /", "echo hi"}, g2.Hist)
	assert.Error(t, g2.OpenHistory(filepath.Join(t.TempDir(), "missing")))

	require.NoError(t, gl.History(testCmdIO(gl, buf)))
	assert.Contains(t, buf.String(), "echo hi")
	buf.Reset()
	require.NoError(t, gl.History(testCmdIO(gl, buf), "1"))
	assert.Contains(t, buf.String(), "echo hi")
	assert.NotContains(t, buf.String(), "cd /", "only the last line")
	assert.Error(t, gl.History(testCmdIO(gl, buf), "x"))
}

// TestTranspile tests transpiling code, files, and the config file in the
// home directory, and sourcing files.
func TestTranspile(t *testing.T) {
	gl, _ := newTestGoal(t)
	gl.TranspileCode("x := 1")
	assert.Contains(t, gl.TrState.Code(), "x := 1")

	dir := t.TempDir()
	in := filepath.Join(dir, "a.goal")
	require.NoError(t, os.WriteFile(in, []byte("echo hi\ny := 2\n"), 0o644))
	g2, _ := newTestGoal(t)
	require.NoError(t, g2.TranspileCodeFromFile(in))
	assert.Contains(t, g2.TrState.Code(), "y := 2")
	assert.Error(t, g2.TranspileCodeFromFile(filepath.Join(dir, "missing.goal")))

	g3, _ := newTestGoal(t)
	require.NoError(t, g3.Source(nil, in))
	assert.Contains(t, g3.TrState.Code(), `goalrun.Run("echo", "hi")`)

	out := filepath.Join(dir, "a.go")
	require.NoError(t, gl.TranspileFile(in, out))
	b, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(b), "package main")

	setHome(t, dir)
	g4, _ := newTestGoal(t)
	require.NoError(t, g4.TranspileConfig(), "no config file")
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".goal"), []byte("z := 3\n"), 0o644))
	require.NoError(t, g4.TranspileConfig())
	assert.Contains(t, g4.TrState.Code(), "z := 3")
}

// TestHostAndDir tests the host and directory of the goal without ssh,
// and the ssh connections by name.
func TestHostAndDir(t *testing.T) {
	gl, _ := newTestGoal(t)
	home := t.TempDir()
	setHome(t, home)
	gl.Config.Dir = filepath.Join(home, "proj")
	assert.Equal(t, "", gl.Host())
	assert.Nil(t, gl.ActiveSSH())
	assert.Equal(t, filepath.Join("~", "proj")+string(filepath.Separator), gl.HostAndDir())
	gl.Config.Dir = "/"
	assert.Equal(t, "/"+string(filepath.Separator), gl.HostAndDir(), "outside of home")

	_, err := gl.SSHByHost("none")
	assert.Error(t, err)
	assert.Error(t, gl.GoSSH(nil, "@none"), "no connection of that name")
	require.NoError(t, gl.GoSSH(nil, "@", "0"), "back to the local host")
	assert.Equal(t, "", gl.SSHActive)
	require.NoError(t, gl.GoSSH(nil, "close"))
	gl.Close()
}

// TestContextErrors tests the processing context and the errors of the
// goal, which cancel the context.
func TestContextErrors(t *testing.T) {
	gl, _ := newTestGoal(t)
	ctx := gl.StartContext()
	assert.NotNil(t, gl.Cancel)
	assert.NoError(t, gl.AddError(nil))
	err := gl.AddError(errors.New("bad"))
	assert.Error(t, err)
	assert.Len(t, gl.Errors, 1)
	assert.Error(t, ctx.Err(), "canceled")
	gl.EndContext()
	assert.Nil(t, gl.Ctx)
	gl.CancelExecution() // without a context
	assert.Equal(t, "", gl.Output("echo", "not run after an error"))
}

// TestRunOutput tests running commands, with their output, errors,
// redirection to and from files, and pipes.
func TestRunOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses unix commands")
	}
	t.Chdir(t.TempDir())
	gl, buf := newTestGoal(t)
	gl.Run("echo", "hello")
	assert.Equal(t, "hello\n", buf.String())

	assert.Equal(t, "out", gl.OutputErrOK("echo", "out"))
	gl.RunErrOK("false")
	assert.Equal(t, "after", gl.Output("echo", "after"), "execution continues")

	gl.Run("echo", "to file", ">", "out.txt")
	b, err := os.ReadFile("out.txt")
	require.NoError(t, err)
	assert.Equal(t, "to file\n", string(b))
	gl.Run("echo", "more", ">>", "out.txt")
	b, _ = os.ReadFile("out.txt")
	assert.Equal(t, "to file\nmore\n", string(b), "appended")
	assert.Equal(t, "to file\nmore", gl.Output("cat", "<", "out.txt"))

	// a pipe, as the transpiler writes it
	gl.Start("cat", "out.txt", "|")
	assert.Equal(t, "2", strings.TrimSpace(gl.Output("wc", "-l")))
	assert.Empty(t, gl.Errors)
	assert.Equal(t, 0, gl.NumJobs())
}

// TestJobs tests running commands in the background as jobs, listing,
// and killing them.
func TestJobs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses unix commands")
	}
	gl, buf := newTestGoal(t)
	gl.Start("sleep", "10")
	require.Equal(t, 1, gl.NumJobs())
	require.NoError(t, gl.JobsCmd(testCmdIO(gl, buf)))
	assert.Contains(t, buf.String(), "[1]")
	assert.Contains(t, buf.String(), "sleep 10")

	args := []string{"%1", "x"}
	assert.Equal(t, 1, gl.JobIDExpand(args))
	assert.NotEqual(t, "%1", args[0], "the pid")
	assert.Equal(t, 0, gl.JobIDExpand([]string{"%9"}), "out of range")
	gl.Errors = nil
	require.NoError(t, gl.Fg(testCmdIO(gl, buf), "%1"))
	assert.Contains(t, buf.String(), "foregrounding job [1]")

	require.NoError(t, gl.Kill(testCmdIO(gl, buf), "%1"))
	gl.DeleteAllJobs()
	assert.Equal(t, 0, gl.NumJobs())
	assert.False(t, gl.DeleteJob(&Job{}))
}

// TestDebug tests turning the debug output of the transpiler on and off.
func TestDebug(t *testing.T) {
	lev := logx.UserLevel
	defer func() { logx.UserLevel = lev }()
	gl, buf := newTestGoal(t)
	require.NoError(t, gl.Debug(testCmdIO(gl, buf)))
	require.NoError(t, gl.Debug(testCmdIO(gl, buf), "on"))
	assert.Equal(t, slog.LevelDebug, logx.UserLevel)
	require.NoError(t, gl.Debug(testCmdIO(gl, buf), "off"))
	assert.Equal(t, slog.LevelInfo, logx.UserLevel)
}

// TestScpLocal tests copying a file with scp between local paths.
func TestScpLocal(t *testing.T) {
	gl, buf := newTestGoal(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0o644))
	dst := filepath.Join(dir, "b.txt")
	require.NoError(t, gl.Scp(testCmdIO(gl, buf), src, "@0:"+dst))
	b, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "data", string(b))
}
