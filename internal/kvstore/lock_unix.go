//go:build unix

package kvstore

import (
	"os"
	"syscall"
)

// lockDir takes an exclusive flock on the directory dir and returns a function
// that releases it. The lock is on the directory, not the store file, because
// save renames a new inode over the store file.
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
