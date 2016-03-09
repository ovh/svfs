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

// Root is a fake root node used to hold a list of container nodes.
type Root struct {
	*Directory
}

// Create is not supported on a root node.
func (r *Root) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	return nil, nil, fuse.ENOTSUP
}

// Mkdir is not supported on a root node.
func (r *Root) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	return nil, fuse.ENOTSUP
}

// Remove is not supported on a root node.
func (r *Root) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	return fuse.ENOTSUP
}

// Rename is not supported on a root node.
func (r *Root) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	return fuse.ENOTSUP
}

// ReadDirAll retrieves all containers within the current Openstack tenant, as direntries.
// Segment containers are not shown and created if missing.
func (r *Root) ReadDirAll(ctx context.Context) (direntries []fuse.Dirent, err error) {
	var (
		baseContainers    = make(map[string]*swift.Container)
		segmentContainers = make(map[string]*swift.Container)
		children          = make(map[string]Node)
	)

	// Cache hit
	if _, nodes := DirectoryCache.GetAll("", r.path); nodes != nil {
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

	DirectoryCache.AddAll("", r.path, r, children)

	return direntries, nil
}

// Lookup gets a container node if its name matches the request
// name within the current context.
func (r *Root) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	// Fill cache if expired
	if _, found := DirectoryCache.Peek("", r.path); !found {
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
