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
	)

	if len(r.children) > 0 {
		for _, c := range r.children {
			entries = append(entries, c.Export())
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
				path: "",
				name: name,
			},
			cs: segment,
		}

		r.children = append(r.children, &child)
		entries = append(entries, child.Export())
	}

	return entries, nil
}

var (
	_ Node           = (*Root)(nil)
	_ fs.Node        = (*Root)(nil)
	_ fs.NodeCreater = (*Root)(nil)
	_ fs.NodeRemover = (*Root)(nil)
	_ fs.NodeRenamer = (*Root)(nil)
)
