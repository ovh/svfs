package fs

import (
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

type File struct {
	directory   bool
	largeObject bool
	container   bool
	segments    bool
	path        string
	h           *swift.Headers
	hs          *swift.Headers
	c           *swift.Container
	cs          *swift.Container
	o           *swift.Object
}

func NewFromContainer(path string, c swift.Container, h swift.Headers) *File {
	return &File{
		c:         &c,
		h:         &h,
		path:      path,
		container: true,
		directory: true,
	}
}

func NewFromContainerWithSegments(path string, c, cs swift.Container, h, hs swift.Headers) *File {
	return &File{
		c:         &c,
		h:         &h,
		cs:        &cs,
		hs:        &hs,
		path:      path,
		container: true,
		directory: true,
		segments:  true,
	}
}

func NewFromObject(path string, directory bool, o *swift.Object, h swift.Headers, parent *File) *File {
	return &File{
		c:         parent.c,
		h:         &h,
		o:         o,
		path:      path,
		directory: directory,
	}
}

func NewFromObjectWithSegments(path string, directory bool, o *swift.Object, h swift.Headers, parent *File) *File {
	return &File{
		c:         parent.c,
		h:         &h,
		cs:        parent.cs,
		hs:        parent.hs,
		o:         o,
		path:      path,
		directory: directory,
	}

}

func (f *File) Size() uint64 {
	if f.container {
		return uint64(f.c.Bytes + f.cs.Bytes)
	}
	if f.directory {
		return uint64(0)
	}
	return uint64(f.o.Bytes)
}

func (f *File) Name() string {
	if f.container {
		return f.c.Name
	}
	return f.o.Name
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Size = f.Size()
	a.Atime = time.Now()

	if !f.container {
		a.Mtime = f.o.LastModified
		a.Ctime = f.o.LastModified
		a.Crtime = f.o.LastModified
	}
	return nil
}

// Check that we satisfy the Node interface.
var _ fs.Node = (*File)(nil)
