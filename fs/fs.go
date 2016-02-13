package fs

import (
	"regexp"
	"strings"

	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"github.com/xlucas/svfs/fs/file"
)

var (
	SegmentRegex = regexp.MustCompile("^.+_segments$")
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
	var (
		dir        = &Dir{s: s.s}
		containers = make(map[string]swift.Container)
		segments   = make(map[string]swift.Container)
	)

	// Retrieve containers
	c, err := s.s.ContainersAll(nil)
	if err != nil {
		return nil, err
	}

	// Find segments
	for _, container := range c {
		name := container.Name
		if !SegmentRegex.Match([]byte(name)) {
			containers[name] = container
			continue
		}
		if SegmentRegex.Match([]byte(name)) {
			segments[strings.TrimSuffix(name, "_segments")] = container
			continue
		}
	}

	// Register children
	for name, container := range containers {
		segment := segments[name]
		dir.children = append(dir.children, &file.Container{
			File: file.NewFile(""),
			C:    &container,
			CS:   &segment,
		})
	}

	return dir, nil
}

var _ fs.FS = (*SVFS)(nil)
