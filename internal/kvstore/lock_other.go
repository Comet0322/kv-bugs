//go:build !unix || aix || solaris

package kvstore

// lockDir is a no-op where flock is unavailable, so the package still builds
// on Windows. Concurrent writers are not protected on these platforms.
func lockDir(dir string) (func(), error) {
	return func() {}, nil
}
