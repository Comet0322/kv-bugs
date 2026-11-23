//go:build !unix

package kvstore

// lockDir is a no-op on non-Unix platforms, which keep the unlocked behaviour.
func lockDir(dir string) (func(), error) {
	return func() {}, nil
}
