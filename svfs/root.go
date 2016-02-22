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

func (r *Root) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// TODO : implement container creation
	return nil, nil, fuse.ENOTSUP
}

func (r *Root) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	return nil, fuse.ENOTSUP
}

func (r *Root) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	// TODO : implement container removal
	return fuse.ENOTSUP
}

func (r *Root) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	return fuse.ENOTSUP
}

func (r *Root) ReadDirAll(ctx context.Context) (entries []fuse.Dirent, err error) {
	var (
		baseC = make(map[string]*swift.Container)
		segC  = make(map[string]*swift.Container)
		list  = make([]Node, 0)
	)

	// Cache hit
	if nodes := EntryCache.Get("", r.path); nodes != nil {
		for _, node := range nodes {
			entries = append(entries, node.Export())
		}
		return entries, nil
	}

	// Retrieve base container
	cs, err := r.s.ContainersAll(nil)
	if err != nil {
		return nil, err
	}

	// Find segments
	for _, container := range cs {
		c := container
		name := container.Name
		if !SegmentRegex.Match([]byte(name)) {
			baseC[name] = &c
			continue
		}
		if SegmentRegex.Match([]byte(name)) {
			segC[strings.TrimSuffix(name, "_segments")] = &c
			continue
		}
	}

	// Register children
	for name, container := range baseC {
		c := container
		segment := segC[name]

		child := Container{
			Directory: &Directory{
				s:    r.s,
				c:    c,
				name: name,
			},
			cs: segment,
		}

		list = append(list, &child)
		entries = append(entries, child.Export())
	}

	EntryCache.Set("", r.path, list)

	return entries, nil
}

func (r *Root) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	var nodes []Node

	// Fill cache if expired
	if nodes = EntryCache.Get("", r.path); nodes == nil {
		r.ReadDirAll(ctx)
		nodes = EntryCache.Get("", r.path)
	}

	for _, item := range nodes {
		if item.Name() == req.Name {
			if n, ok := item.(*Container); ok {
				return n, nil
			}
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
