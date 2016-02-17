package svfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
)

// SVFS implements the Swift Virtual File System.
type SVFS struct {
	s     *swift.Connection
	cName string
}

func (s *SVFS) Init(sc *swift.Connection, cName string) error {
	s.s = sc
	s.cName = cName

	// Authenticate if we don't have a token
	// and storage URL
	if !s.s.Authenticated() {
		return s.s.Authenticate()
	}

	return nil
}

func (s *SVFS) Root() (fs.Node, error) {
	if s.cName != "" {
		// If a specific container is specified
		// in mount options, find it and relevant
		// segment container too if present.
		baseC, _, err := s.s.Container(s.cName)
		if err != nil {
			return nil, fuse.ENOENT
		}
		segC, _, err := s.s.Container(s.cName + "_segments")
		if err != nil && err != swift.ContainerNotFound {
			return nil, fuse.EIO
		}

		return &Container{
			Directory: &Directory{
				apex: true,
				s:    s.s,
				c:    &baseC,
			},
			cs: &segC,
		}, nil
	}
	return &Root{
		Directory: &Directory{
			apex: true,
			s:    s.s,
		},
	}, nil
}

var _ fs.FS = (*SVFS)(nil)
