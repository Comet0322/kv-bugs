//go:build unix

package kvstore

import (
	"os"
	"syscall"
)

// lockDir takes an exclusive flock on dir and returns a func that releases it.
// The lock is on the directory, not the store file: rename replaces the store
// file's inode, and a sidecar lock file would add a directory entry.
func lockDir(dir string) (func(), error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
