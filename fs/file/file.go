package file

import (
	"os"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// File represents a generic VSFS element.
type File struct {
	path string
}

type Node interface {
	Size() uint64
	Mode() os.FileMode
	Name() string
	Path() string
}

func NewFile(path string) *File {
	return &File{
		path: path,
	}
}

func (*File) Size() uint64 {
	return uint64(0)
}

func (*File) Name() string {
	return "N/A"
}

func (*File) Mode() os.FileMode {
	return 0 << (32 - 1)
}

func (f *File) Path() string {
	return f.path
}

func (*File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Atime = time.Now()
	return nil
}

var (
	_ fs.Node = (*File)(nil)
	_ Node    = (*File)(nil)
)
