//go:build !darwin

package main

import (
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// setBtime is a no-op on non-Darwin platforms because fuse.Attr has no
// Crtime fields there. Kept so callers don't need per-platform code.
func setBtime(attr *fuse.Attr, btime time.Time) {
	// birthtime is not supported on this platform
}
