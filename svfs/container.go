package svfs

import "bazil.org/fuse/fs"

// Container is a node representing a directory entry bound to
// a Swift container.
type Container struct {
	*Directory
}

var (
	_ Node    = (*Container)(nil)
	_ fs.Node = (*Container)(nil)
)
