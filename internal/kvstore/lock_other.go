//go:build !unix || solaris || aix

package kvstore

// lockDir is a no-op where flock is not available. Writes from several processes are then not
// serialised, so concurrent writes can still lose updates on these platforms.
func lockDir(dir string) (func(), error) {
	return func() {}, nil
}
