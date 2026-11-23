//go:build unix && !solaris && !aix

package kvstore

import (
	"os"
	"syscall"
)

// lockDir takes an exclusive flock on the directory dir and returns a func that releases it.
// The lock is on the directory, not on a file inside it, so no lock file is left in the store's directory.
func lockDir(dir string) (func(), error) {
	d, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(d.Fd()), syscall.LOCK_EX); err != nil {
		d.Close()
		return nil, err
	}
	return func() {
		syscall.Flock(int(d.Fd()), syscall.LOCK_UN)
		d.Close()
	}, nil
}
