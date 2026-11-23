//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
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

// readStore decodes the store file at path directly, without going through
// the package under test.
func readStore(t *testing.T, path string) map[string]string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read store %s: %v, want a store file", path, err)
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("parse store %q: %v, want a JSON object", b, err)
	}
	return m
}

func TestAC4_ConcurrentSetsKeepEveryKey(t *testing.T) {
	const n = 50
	path := setup(t)
	var argLists [][]string
	want := map[string]string{}
	for i := 1; i <= n; i++ {
		k := fmt.Sprintf("k%02d", i)
		argLists = append(argLists, []string{"set", k, "v" + k})
		want[k] = "v" + k
	}
	requireAllExitZero(t, runConcurrently(t, path, argLists))
	if got := readStore(t, path); !reflect.DeepEqual(got, want) {
		t.Errorf("store after %d concurrent kv set has %d keys, want %d\ngot:  %v\nwant: %v",
			n, len(got), len(want), got, want)
	}
}

func TestAC5_ConcurrentDelsAndSetsLoseNothing(t *testing.T) {
	const n = 25
	path := setup(t)
	seed := map[string]string{}
	for i := 1; i <= n; i++ {
		seed[fmt.Sprintf("d%02d", i)] = "old"
	}
	b, err := json.Marshal(seed)
	if err != nil {
		t.Fatalf("marshal seed: %v", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatalf("write seed store: %v", err)
	}
	var argLists [][]string
	want := map[string]string{}
	for i := 1; i <= n; i++ {
		argLists = append(argLists,
			[]string{"del", fmt.Sprintf("d%02d", i)},
			[]string{"set", fmt.Sprintf("n%02d", i), fmt.Sprintf("v%02d", i)})
		want[fmt.Sprintf("n%02d", i)] = fmt.Sprintf("v%02d", i)
	}
	requireAllExitZero(t, runConcurrently(t, path, argLists))
	if got := readStore(t, path); !reflect.DeepEqual(got, want) {
		t.Errorf("store after %d concurrent kv del and %d concurrent kv set:\ngot:  %v\nwant exactly: %v",
			n, n, got, want)
	}
}
