package fs

import (
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
)

// SVFS implements the Swift Virtual File System.
type SVFS struct {
	s *swift.Connection
}

func (s *SVFS) Init(sc *swift.Connection) error {
	s.s = sc
	return s.s.Authenticate()
}

func (s *SVFS) Root() (fs.Node, error) {
	return &Dir{s: s.s}, nil
}

// Check that we satisfy the fs interface.
var _ fs.FS = (*SVFS)(nil)
