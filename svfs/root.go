package svfs

import (
	"regexp"
	"strings"

	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/xlucas/swift"
)

const (
	SegmentContainerSuffix = "_segments"
)

var SegmentRegex = regexp.MustCompile("^.+_segments$")

type Root struct {
	*Directory
}

func (r *Root) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	return nil, nil, fuse.ENOTSUP
}

func (r *Root) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	return nil, fuse.ENOTSUP
}

func (r *Root) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	return fuse.ENOTSUP
}

func (r *Root) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	return fuse.ENOTSUP
}

func (r *Root) ReadDirAll(ctx context.Context) (direntries []fuse.Dirent, err error) {
	var (
		baseContainers    = make(map[string]*swift.Container)
		segmentContainers = make(map[string]*swift.Container)
		children          = make(map[string]Node)
	)

	// Cache hit
	if nodes := DirectoryCache.GetAll("", r.path); nodes != nil {
		for _, node := range nodes {
			direntries = append(direntries, node.Export())
		}
		return direntries, nil
	}

	// Retrieve all containers
	cs, err := SwiftConnection.ContainersAll(nil)
	if err != nil {
		return nil, err
	}

	// Sort base and segment containers
	for _, segmentContainer := range cs {
		s := segmentContainer
		if !SegmentRegex.Match([]byte(s.Name)) {
			baseContainers[s.Name] = &s
			continue
		}
		if SegmentRegex.Match([]byte(s.Name)) {
			segmentContainers[strings.TrimSuffix(s.Name, SegmentContainerSuffix)] = &s
			continue
		}
	}

	for _, baseContainer := range baseContainers {
		c := baseContainer
		// Create segment container if missing
		if segmentContainers[c.Name] == nil {
			segmentContainers[c.Name], err = createContainer(c.Name + SegmentContainerSuffix)
			if err != nil {
				return nil, err
			}
		}

		// Register direntries and cache entries
		child := Container{
			Directory: &Directory{
				c:    c,
				cs:   segmentContainers[c.Name],
				name: c.Name,
			},
		}

		children[c.Name] = &child
		direntries = append(direntries, child.Export())
	}

	DirectoryCache.AddAll("", r.path, children)

	return direntries, nil
}

func (r *Root) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	// Fill cache if expired
	if !DirectoryCache.Peek("", r.path) {
		r.ReadDirAll(ctx)
	}

	// Find matching child
	if item := DirectoryCache.Get("", r.path, req.Name); item != nil {
		if n, ok := item.(*Container); ok {
			return n, nil
		}
		if n, ok := item.(*Directory); ok {
			return n, nil
		}
		if n, ok := item.(*Object); ok {
			return n, nil
		}
	}

	return nil, fuse.ENOENT
}

var (
	_ Node           = (*Root)(nil)
	_ fs.Node        = (*Root)(nil)
	_ fs.NodeCreater = (*Root)(nil)
	_ fs.NodeMkdirer = (*Root)(nil)
	_ fs.NodeRemover = (*Root)(nil)
	_ fs.NodeRenamer = (*Root)(nil)
)
