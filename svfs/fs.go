package svfs

import (
	"time"

	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
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
	SegmentSizeMB         uint64
	MaxReaddirConcurrency uint64
	MaxUploadConcurrency  uint64
}

func (s *SVFS) Init(sc *swift.Connection, conf *Config, cconf *CacheConfig) error {
	s.conf = conf
	SwiftConnection = sc
	DirectoryCache = NewCache(cconf)
	DirectoryLister = &DirLister{concurrency: conf.MaxReaddirConcurrency}
	SwiftConnection.ConnectTimeout = conf.ConnectTimeout
	SegmentSize = conf.SegmentSizeMB * (1 << 20)

	// Start directory lister
	DirectoryLister.Start()

	// Authenticate if we don't have a token and storage URL
	if !SwiftConnection.Authenticated() {
		return SwiftConnection.Authenticate()
	}

	return nil
}

func (s *SVFS) Root() (fs.Node, error) {
	// Mount a specific container
	if s.conf.Container != "" {
		baseContainer, _, err := SwiftConnection.Container(s.conf.Container)
		if err != nil {
			return nil, err
		}

		// Find segment container too
		segmentContainerName := s.conf.Container + SegmentContainerSuffix
		segmentContainer, _, err := SwiftConnection.Container(segmentContainerName)

		// Create it if missing
		if err == swift.ContainerNotFound {
			var container *swift.Container
			container, err = createContainer(segmentContainerName)
			segmentContainer = *container
		}
		if err != nil && err != swift.ContainerNotFound {
			return nil, err
		}

		return &Container{
			Directory: &Directory{
				apex: true,
				c:    &baseContainer,
				cs:   &segmentContainer,
			},
		}, nil
	}

	// Mount all containers within an account
	return &Root{
		Directory: &Directory{
			apex: true,
		},
	}, nil
}

var _ fs.FS = (*SVFS)(nil)
