package svfs

import (
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
)

var (
	SwiftConnection *swift.Connection
)

// SVFS implements the Swift Virtual File System.
type SVFS struct {
	conf *Config
}

type Config struct {
	Container             string
	ConnectTimeout        time.Duration
	ReadAheadSize         uint
	MaxReaddirConcurrency uint64
}

func (s *SVFS) Init(sc *swift.Connection, conf *Config, cconf *CacheConfig) error {
	s.conf = conf
	SwiftConnection = sc
	SwiftConnection.ConnectTimeout = conf.ConnectTimeout
	EntryCache = NewCache(cconf)
	DirectoryLister = &DirLister{concurrency: conf.MaxReaddirConcurrency}

	// Authenticate if we don't have a token
	// and storage URL
	if !SwiftConnection.Authenticated() {
		return SwiftConnection.Authenticate()
	}

	// Start directory lister
	DirectoryLister.Start()

	return nil
}

func (s *SVFS) Root() (fs.Node, error) {
	if s.conf.Container != "" {
		// If a specific container is specified
		// in mount options, find it and relevant
		// segment container too if present.
		baseC, _, err := SwiftConnection.Container(s.conf.Container)
		if err != nil {
			return nil, fuse.ENOENT
		}
		segC, _, err := SwiftConnection.Container(s.conf.Container + "_segments")
		if err != nil && err != swift.ContainerNotFound {
			return nil, fuse.EIO
		}

		return &Container{
			Directory: &Directory{
				apex: true,
				c:    &baseC,
			},
			cs: &segC,
		}, nil
	}
	return &Root{
		Directory: &Directory{
			apex: true,
		},
	}, nil
}

var _ fs.FS = (*SVFS)(nil)
