package svfs

import (
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
)

// SVFS implements the Swift Virtual File System.
type SVFS struct {
	s           *swift.Connection
	cache       *Cache
	lister      *DirLister
	conf        *Config
	concurrency uint64
}

type Config struct {
	Container             string
	ConnectTimeout        time.Duration
	MaxReaddirConcurrency uint64
}

func (s *SVFS) Init(sc *swift.Connection, conf *Config, cconf *CacheConfig) error {
	s.s = sc
	s.conf = conf
	s.cache = NewCache(cconf)
	s.s.ConnectTimeout = conf.ConnectTimeout
	s.lister = &DirLister{
		c:           s.s,
		concurrency: conf.MaxReaddirConcurrency,
	}

	// Authenticate if we don't have a token
	// and storage URL
	if !s.s.Authenticated() {
		return s.s.Authenticate()
	}

	// Start directory lister
	s.lister.Start()

	return nil
}

func (s *SVFS) Root() (fs.Node, error) {
	if s.conf.Container != "" {
		// If a specific container is specified
		// in mount options, find it and relevant
		// segment container too if present.
		baseC, _, err := s.s.Container(s.conf.Container)
		if err != nil {
			return nil, fuse.ENOENT
		}
		segC, _, err := s.s.Container(s.conf.Container + "_segments")
		if err != nil && err != swift.ContainerNotFound {
			return nil, fuse.EIO
		}

		return &Container{
			Directory: &Directory{
				apex:  true,
				cache: s.cache,
				s:     s.s,
				c:     &baseC,
				l:     s.lister,
			},
			cs: &segC,
		}, nil
	}
	return &Root{
		Directory: &Directory{
			apex:  true,
			cache: s.cache,
			s:     s.s,
			l:     s.lister,
		},
	}, nil
}

var _ fs.FS = (*SVFS)(nil)
