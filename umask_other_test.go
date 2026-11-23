//go:build !unix

package main

import "testing"

// withUmask is a no-op where there is no umask.
func withUmask(t *testing.T, mask int) {}
