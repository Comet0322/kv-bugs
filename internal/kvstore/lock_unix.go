//go:build unix

package kvstore

import (
	"fmt"
	"os"
	"syscall"
)

// lockDir takes an exclusive advisory lock on dir, blocking until it is free,
// and returns the function that releases it. Locking the directory rather
// than a file next to the store leaves no extra file behind.
func lockDir(dir string) (func(), error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
		if err != syscall.EINTR {
			break
		}
	}
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("lock %s: %w", dir, err)
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
