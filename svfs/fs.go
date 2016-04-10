package svfs

import (
	"fmt"
	"time"

	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
)

var (
	SwiftConnection *swift.Connection
	Version         string = "0.5.1"
	UserAgent       string = "svfs/" + Version
	DefaultUID      uint64 = 0
	DefaultGID      uint64 = 0
	DefaultMode     uint64 = 0700
	ExtraAttr       bool   = false
)

// SVFS implements the Swift Virtual File System.
type SVFS struct {
	conf *Config
}

// Config represents SVFS configuration settings.
type Config struct {
	Container            string
	ConnectTimeout       time.Duration
	ReadAheadSize        uint
	SegmentSizeMB        uint64
	ListConcurrency      uint64
	MaxUploadConcurrency uint64
}

// Init sets up the filesystem. It sets configuration settings, starts mandatory
// services and make sure authentication in Swift has succeeded.
func (s *SVFS) Init(sc *swift.Connection, conf *Config, cconf *CacheConfig) error {
	s.conf = conf
	SwiftConnection = sc
	DirectoryCache = NewCache(cconf)
	ChangeCache = NewSimpleCache()
	DirectoryLister = &DirLister{concurrency: conf.ListConcurrency}
	SwiftConnection.ConnectTimeout = conf.ConnectTimeout
	SegmentSize = conf.SegmentSizeMB * (1 << 20)
	if SegmentSize > 5*(1<<30) {
		return fmt.Errorf("Segment size can't exceed 5 GB")
	}
	swift.DefaultUserAgent = UserAgent

	if HubicAuthorization != "" && HubicRefreshToken != "" {
		SwiftConnection.Auth = new(HubicAuth)
	}

	// Start directory lister
	DirectoryLister.Start()

	// Authenticate if we don't have a token and storage URL
	if !SwiftConnection.Authenticated() {
		return SwiftConnection.Authenticate()
	}

	return nil
}

// Root gets the root node of the filesystem. It can either be a fake root node
// filled with all the containers found for the given Openstack tenant or a container
// node if a container name have been specified in mount options.
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
