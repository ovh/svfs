package fs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
)

// SVFS implements the Swift Virtual File System.
type SVFS struct {
	Con *swift.Connection
}

// TODO : implement it
func (SVFS) Root() (fs.Node, error) {
	return nil, fuse.ENOSYS
}

// Check that we satisfy the fs interface
var _ fs.FS = (*SVFS)(nil)
