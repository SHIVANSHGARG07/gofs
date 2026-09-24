//go:build darwin

package main

import (
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
)

// setBtime stores the birthtime (creation time) into the attr's
// platform-specific Crtime fields. These fields only exist on Darwin
// (macOS), which is why this file is built only for darwin via the
// build tag above.
func setBtime(attr *fuse.Attr, btime time.Time) {
	attr.Crtime_ = uint64(btime.Unix())
	attr.Crtimensec_ = uint32(btime.Nanosecond())
}
