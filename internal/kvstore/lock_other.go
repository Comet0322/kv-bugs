//go:build !unix

package kvstore

// lockDir is a no-op where flock is not available, so the package still builds.
// Concurrent writers are not serialised on these platforms: the cross-process
// write guarantee is POSIX-only.
func lockDir(dir string) (func(), error) {
	return func() {}, nil
}
