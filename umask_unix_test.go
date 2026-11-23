//go:build unix

package main

import (
	"syscall"
	"testing"
)

// withUmask sets the process umask for the duration of the test.
func withUmask(t *testing.T, mask int) {
	t.Helper()
	old := syscall.Umask(mask)
	t.Cleanup(func() { syscall.Umask(old) })
}
