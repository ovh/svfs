package file

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

type Object struct {
	*File
	*Container
	Segmented bool
	SO        *swift.Object
}

func (f *Object) Name() string {
	return f.SO.Name
}

func (f *Object) Size() uint64 {
	return uint64(f.SO.Bytes)
}

func (f *Object) Mode() os.FileMode {
	return f.File.Mode()
}

func (f *Object) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Size = f.Size()
	a.Mtime = f.SO.LastModified
	a.Ctime = f.SO.LastModified
	a.Crtime = f.SO.LastModified
	return f.File.Attr(ctx, a)
}

var (
	_ fs.Node = (*Object)(nil)
	_ Node    = (*Object)(nil)
)
