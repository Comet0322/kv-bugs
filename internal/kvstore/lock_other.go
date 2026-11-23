//go:build !unix

package kvstore

// lockDir is a no-op where flock is unavailable; writers are not serialised.
func lockDir(dir string) (func(), error) {
	return func() {}, nil
}
