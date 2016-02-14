package svfs

import (
	"regexp"
	"strings"

	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/ncw/swift"
)

var SegmentRegex = regexp.MustCompile("^.+_segments$")

type Root struct {
	*Directory
}

func (r *Root) ReadDirAll(ctx context.Context) (entries []fuse.Dirent, err error) {
	var (
		baseC = make(map[string]swift.Container)
		segC  = make(map[string]swift.Container)
	)

	if len(r.children) > 0 {
		for _, c := range r.children {
			entries = append(entries, c.Export())
		}
		return entries, nil
	}

	// Retrieve base container
	c, err := r.s.ContainersAll(nil)
	if err != nil {
		return nil, err
	}

	// Find segments
	for _, container := range c {
		name := container.Name
		if !SegmentRegex.Match([]byte(name)) {
			baseC[name] = container
			continue
		}
		if SegmentRegex.Match([]byte(name)) {
			segC[strings.TrimSuffix(name, "_segments")] = container
			continue
		}
	}

	// Register children
	for name, container := range baseC {
		segment := segC[name]
		child := Container{
			Directory: &Directory{
				s:    r.s,
				c:    &container,
				path: "",
				name: name,
			},
			cs: &segment,
		}

		r.children = append(r.children, &child)
		entries = append(entries, child.Export())
	}

	return entries, nil
}

var _ Node = (*Directory)(nil)
var _ fs.Node = (*Directory)(nil)
