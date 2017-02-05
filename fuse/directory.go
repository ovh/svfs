package fuse

import (
	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	sfs "github.com/ovh/svfs/fs"
)

type Directory struct {
	sfs.Directory
}

func (d *Directory) Lookup(ctx context.Context, name string) (fs.Node, error) {
	return nil, fuse.ENOENT
}

func (d *Directory) Attr(ctx context.Context, a *fuse.Attr) error {
	attr, err := d.Directory.(sfs.Node).GetAttr()
	a.Atime = attr.Atime
	a.Ctime = attr.Ctime
	a.Crtime = attr.Ctime
	a.Mtime = attr.Mtime
	a.Mode = attr.Mode
	a.Uid = attr.Uid
	a.Gid = attr.Gid
	a.Size = attr.Size

	return err
}

func (d *Directory) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (
	fs.Node, error,
) {
	dir, err := d.Directory.Mkdir(req.Name)
	return &Directory{dir}, err
}

var (
	_ fs.NodeStringLookuper = (*Directory)(nil)
	_ fs.NodeMkdirer        = (*Directory)(nil)
	_ fs.Node               = (*Directory)(nil)
)
