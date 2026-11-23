//go:build unix

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// runAsKV is set in the environment of child processes: the test binary then
// behaves as the kv command instead of running tests.
const runAsKV = "KV_TEST_RUN_AS_KV"

func TestMain(m *testing.M) {
	if os.Getenv(runAsKV) == "1" {
		// Wait until the parent closes stdin, so every child starts its
		// command at the same moment.
		io.Copy(io.Discard, os.Stdin)
		os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

type kvProc struct {
	args   []string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stderr bytes.Buffer
}

// runConcurrently starts one separate kv process per argument list against
// the store at path, releases them all at once and waits for every one.
func runConcurrently(t *testing.T, path string, argLists [][]string) []*kvProc {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	procs := make([]*kvProc, 0, len(argLists))
	for _, args := range argLists {
		p := &kvProc{args: args, cmd: exec.Command(exe, args...)}
		p.cmd.Env = append(os.Environ(), runAsKV+"=1", "KV_FILE="+path)
		p.cmd.Stderr = &p.stderr
		if p.stdin, err = p.cmd.StdinPipe(); err != nil {
			t.Fatalf("stdin pipe: %v", err)
		}
		if err := p.cmd.Start(); err != nil {
			t.Fatalf("start kv %v: %v", args, err)
		}
		procs = append(procs, p)
	}
	for _, p := range procs {
		p.stdin.Close()
	}
	for _, p := range procs {
		p.cmd.Wait()
	}
	return procs
}

func requireAllExitZero(t *testing.T, procs []*kvProc) {
	t.Helper()
	for _, p := range procs {
		if c := p.cmd.ProcessState.ExitCode(); c != 0 {
			t.Errorf("kv %s exit = %d (stderr %q), want 0", strings.Join(p.args, " "), c, p.stderr.String())
		}
	}
}

func listStore(t *testing.T) string {
	t.Helper()
	c, o, e := do(t, "list")
	if c != 0 {
		t.Fatalf("kv list = %d %q %q, want exit 0", c, o, e)
	}
	return o
}

func TestAC4_ConcurrentSetsKeepEveryKey(t *testing.T) {
	const n = 50
	path := setup(t)
	var argLists [][]string
	var want strings.Builder
	keys := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		keys = append(keys, fmt.Sprintf("k%02d", i))
	}
	for _, k := range keys {
		argLists = append(argLists, []string{"set", k, "v" + k})
		fmt.Fprintf(&want, "%s=v%s\n", k, k)
	}
	requireAllExitZero(t, runConcurrently(t, path, argLists))
	got := listStore(t)
	if got != want.String() {
		t.Errorf("kv list after %d concurrent sets has %d lines, want %d\ngot:\n%swant:\n%s",
			n, strings.Count(got, "\n"), n, got, want.String())
	}
}
