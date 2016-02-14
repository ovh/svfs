package svfs

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
	return &Root{
		Directory: &Directory{
			s: s.s,
		},
	}, nil
}

var _ fs.FS = (*SVFS)(nil)
