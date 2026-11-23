package main

import (
	"os"
	"testing"
)

// I4: a store file that is not valid JSON makes every command exit 1 and leaves the file's bytes unchanged.
// Each row runs one command against a fresh corrupt store and checks the exit code and the file's bytes.
func TestCorruptStoreUnchangedByEveryCommand(t *testing.T) {
	corrupt := []struct {
		name string
		body string
	}{
		{"I4_unclosed_object", "{bad"},
		{"I4_plain_text", "not json at all"},
	}
	commands := []struct {
		name string
		args []string
	}{
		{"I4_set", []string{"set", "a", "1"}},
		{"I4_get", []string{"get", "a"}},
		{"I4_del", []string{"del", "a"}},
		{"I4_list", []string{"list"}},
	}
	for _, c := range corrupt {
		for _, cmd := range commands {
			t.Run(c.name+"/"+cmd.name, func(t *testing.T) {
				p := setup(t)
				if err := os.WriteFile(p, []byte(c.body), 0o644); err != nil {
					t.Fatal(err)
				}
				code, stdout, stderr := do(t, cmd.args...)
				if code != 1 {
					t.Errorf("%v: exit code = %d, want 1 (stdout %q, stderr %q)", cmd.args, code, stdout, stderr)
				}
				got, err := os.ReadFile(p)
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != c.body {
					t.Errorf("%v: store bytes = %q, want unchanged %q", cmd.args, got, c.body)
				}
			})
		}
	}
}
