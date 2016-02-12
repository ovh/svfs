package file

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type Directory struct {
	*File
	*Container
	Label string
}

func (f *Directory) Name() string {
	return f.Label
}

// Not available with swift
func (f *Directory) Size() uint64 {
	return 0
}

func (f *Directory) Mode() os.FileMode {
	return os.ModeDir
}

func (f *Directory) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Size = f.Size()
	return f.File.Attr(ctx, a)
}

var (
	_ fs.Node = (*Directory)(nil)
	_ Node    = (*Directory)(nil)
)
