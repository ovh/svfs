package fs

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"github.com/xlucas/svfs/fs/file"
	"golang.org/x/net/context"
)

var (
	RootRegex   = regexp.MustCompile("^[^/]+/?$")
	FolderRegex = regexp.MustCompile("^.+/$")
)

type Dir struct {
	s        *swift.Connection
	f        file.Node
	children []file.Node
}

func (d *Dir) container() (*file.Container, error) {
	dir, ok := d.f.(*file.Directory)
	if ok {
		return dir.Container, nil
	}
	cont, ok := d.f.(*file.Container)
	if ok {
		return cont, nil
	}
	return nil, fmt.Errorf("Unable to find relevant swift container")
}

func (d *Dir) read() (dirs []fuse.Dirent, err error) {
	// Cache hit
	if len(d.children) > 0 {
		for _, child := range d.children {
			dirs = append(dirs, child.FuseEntry())
		}
		return dirs, nil
	}

	// Get underlying container
	container, err := d.container()
	if err != nil {
		return nil, err
	}

	// Fetch objects
	objects, err := d.s.ObjectsAll(container.C.Name, &swift.ObjectsOpts{
		Delimiter: '/',
		Prefix:    d.f.Path(),
	})
	if err != nil {
		return nil, err
	}

	// Fill cache
	for _, object := range objects {
		var (
			child    file.Node
			o        = object
			fileName = strings.TrimPrefix(o.Name, d.f.Path())
		)
		// This is a directory
		if FolderRegex.Match([]byte(o.Name)) {
			child = &file.Directory{
				File:      file.NewFile(o.Name),
				Label:     fileName[:len(fileName)-1],
				Container: container,
			}
		}
		// This is a swift object
		if !FolderRegex.Match([]byte(o.Name)) {
			child = &file.Object{
				File:      file.NewFile(o.Name),
				Label:     fileName,
				Container: container,
				SO:        &o,
			}
		}

		d.children = append(d.children, child)
		dirs = append(dirs, child.FuseEntry())
	}

	return dirs, nil
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	if d.f != nil {
		return d.f.Attr(ctx, a)
	} else {
		a.Mode = os.ModeDir
	}
	return nil
}

func (d *Dir) ReadDirAll(ctx context.Context) (entries []fuse.Dirent, err error) {
	if d.f == nil {
		for _, c := range d.children {
			entries = append(entries, c.FuseEntry())
		}
		return entries, nil
	}
	return d.read()
}

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	for _, container := range d.children {
		if container.Name() == req.Name {
			return &Dir{s: d.s, f: container}, nil
		}
	}
	return nil, fuse.ENOENT
}

var _ fs.Node = (*Dir)(nil)
