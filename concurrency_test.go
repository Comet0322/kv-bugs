package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

// kvChildEnv marks a re-executed test binary that must act as the kv command.
const kvChildEnv = "KV_TEST_CHILD"

// TestKVChildProcess is not a test: when the test binary is re-executed by
// startKV it behaves as `kv <args>`. It first blocks until its stdin reaches
// EOF, so the parent can release every child at the same moment.
func TestKVChildProcess(t *testing.T) {
	if os.Getenv(kvChildEnv) != "1" {
		t.Skip("helper process for the multi-process tests")
	}
	io.Copy(io.Discard, os.Stdin)
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	os.Exit(run(args, os.Stdout, os.Stderr))
}

type kvProc struct {
	args []string
	cmd  *exec.Cmd
	out  bytes.Buffer
}

// runConcurrently starts one separate kv process per argument list against
// the KV_FILE already set in the environment, holds them all at a start
// barrier, releases them together, and waits for every one to exit.
func runConcurrently(t *testing.T, argLists [][]string) []*kvProc {
	t.Helper()
	gate, release, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	procs := make([]*kvProc, 0, len(argLists))
	for _, args := range argLists {
		p := &kvProc{args: args}
		p.cmd = exec.Command(os.Args[0], append([]string{"-test.run=^TestKVChildProcess$", "--"}, args...)...)
		p.cmd.Env = append(os.Environ(), kvChildEnv+"=1")
		p.cmd.Stdin = gate
		p.cmd.Stdout = &p.out
		p.cmd.Stderr = &p.out
		if err := p.cmd.Start(); err != nil {
			release.Close()
			gate.Close()
			for _, q := range procs {
				q.cmd.Wait()
			}
			t.Fatalf("start kv %v: %v", args, err)
		}
		procs = append(procs, p)
	}
	gate.Close()
	release.Close() // EOF on every child's stdin: all start their command now
	for _, p := range procs {
		p.cmd.Wait()
	}
	return procs
}

func requireAllExitZero(t *testing.T, procs []*kvProc) {
	t.Helper()
	for _, p := range procs {
		if c := p.cmd.ProcessState.ExitCode(); c != 0 {
			t.Errorf("kv %v: exit %d, want 0; output %q", p.args, c, p.out.String())
		}
	}
}

func readStore(t *testing.T, path string) map[string]string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	m := map[string]string{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("parse store %q: %v", b, err)
	}
	return m
}

func TestAC4_ConcurrentSetsKeepEveryKey(t *testing.T) {
	const n = 50
	p := setup(t)
	var argLists [][]string
	for i := 1; i <= n; i++ {
		argLists = append(argLists, []string{"set", fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)})
	}
	requireAllExitZero(t, runConcurrently(t, argLists))
	got := readStore(t, p)
	missing := 0
	for i := 1; i <= n; i++ {
		k, want := fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)
		if v, ok := got[k]; !ok || v != want {
			missing++
			t.Errorf("after %d concurrent sets: store[%q] = %q (present %v), want %q", n, k, v, ok, want)
		}
	}
	if len(got) != n {
		t.Errorf("after %d concurrent sets: store holds %d keys (%d lost), want %d", n, len(got), missing, n)
	}
}

func TestAC5_ConcurrentDelsAndSetsLoseNothing(t *testing.T) {
	const n = 30
	p := setup(t)
	initial := map[string]string{}
	for i := 1; i <= n; i++ {
		initial[fmt.Sprintf("d%d", i)] = "x"
	}
	b, err := json.Marshal(initial)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0o600); err != nil {
		t.Fatal(err)
	}
	var argLists [][]string
	for i := 1; i <= n; i++ {
		argLists = append(argLists,
			[]string{"del", fmt.Sprintf("d%d", i)},
			[]string{"set", fmt.Sprintf("n%d", i), fmt.Sprintf("v%d", i)})
	}
	requireAllExitZero(t, runConcurrently(t, argLists))
	got := readStore(t, p)
	for i := 1; i <= n; i++ {
		if v, ok := got[fmt.Sprintf("d%d", i)]; ok {
			t.Errorf("deleted key d%d is back in the store with value %q, want it absent", i, v)
		}
		k, want := fmt.Sprintf("n%d", i), fmt.Sprintf("v%d", i)
		if v, ok := got[k]; !ok || v != want {
			t.Errorf("store[%q] = %q (present %v), want %q", k, v, ok, want)
		}
	}
	if len(got) != n {
		t.Errorf("after %d concurrent dels and %d concurrent sets: store holds %d keys %v, want exactly n1..n%d", n, n, len(got), got, n)
	}
}
