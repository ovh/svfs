package svfs

import (
	"os"
	"regexp"
	"strings"

	"github.com/ncw/swift"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

var (
	FolderRegex = regexp.MustCompile("^.+/$")
)

type Directory struct {
	name     string
	path     string
	s        *swift.Connection
	c        *swift.Container
	children []Node
}

func (d *Directory) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0600
	a.Size = uint64(4096)
	return nil
}

func (d *Directory) Export() fuse.Dirent {
	return fuse.Dirent{
		Name: d.name,
		Type: fuse.DT_Dir,
	}
}

func (d *Directory) ReadDirAll(ctx context.Context) (entries []fuse.Dirent, err error) {
	// Cache hit
	if len(d.children) > 0 {
		for _, child := range d.children {
			entries = append(entries, child.Export())
		}
		return entries, nil
	}

	// Fetch objects
	objects, err := d.s.ObjectsAll(d.c.Name, &swift.ObjectsOpts{
		Delimiter: '/',
		Prefix:    d.path,
	})
	if err != nil {
		return nil, err
	}

	// Fill cache
	for _, object := range objects {
		var (
			child    Node
			o        = object
			fileName = strings.TrimPrefix(o.Name, d.path)
		)
		// This is a directory
		if FolderRegex.Match([]byte(o.Name)) {
			child = &Directory{
				s:    d.s,
				c:    d.c,
				path: o.Name,
				name: fileName[:len(fileName)-1],
			}
		}
		// This is a swift object
		if !FolderRegex.Match([]byte(o.Name)) {
			child = &Object{
				path: o.Name,
				name: fileName,
				s:    d.s,
				c:    d.c,
				so:   &o,
			}
		}

		d.children = append(d.children, child)
		entries = append(entries, child.Export())
	}

	return entries, nil
}

func (d *Directory) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	for _, item := range d.children {
		if item.Name() == req.Name {
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
	}
	return nil, fuse.ENOENT
}

func (d *Directory) Name() string {
	return d.name
}

var _ Node = (*Directory)(nil)
var _ fs.Node = (*Directory)(nil)
