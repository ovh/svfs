package file

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

type Container struct {
	*File
	C  *swift.Container
	CS *swift.Container
}

func (f *Container) Name() string {
	return f.C.Name
}

func (f *Container) Size() uint64 {
	if f.CS != nil {
		return uint64(f.C.Bytes + f.CS.Bytes)
	}
	return uint64(f.C.Bytes)
}

func (f *Container) Mode() os.FileMode {
	return os.ModeDir
}

func (f *Container) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Size = f.Size()
	return f.File.Attr(ctx, a)
}

var (
	_ fs.Node = (*Container)(nil)
	_ Node    = (*Container)(nil)
)
