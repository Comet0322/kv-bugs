package kvstore

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// I16: the module builds for Windows. lock_unix.go uses syscall.Flock, so the
// !unix file lock_other.go must supply lockDir there. This runs the same
// cross-compile the spec recorded, from the module root (two levels up from
// this package).
func TestI16_ModuleBuildsForWindows(t *testing.T) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = filepath.Join("..", "..")
	cmd.Env = append(os.Environ(), "GOOS=windows", "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("GOOS=windows go build ./... failed: %v\n%s", err, out)
	}
}
