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

func (d *Directory) Attr(ctx context.Context, a *fuse.Attr) (err error) {
	attr, err := d.Directory.(sfs.Node).GetAttr()
	if err != nil {
		return
	}

	a.Atime = attr.Atime
	a.Ctime = attr.Ctime
	a.Crtime = attr.Ctime
	a.Mtime = attr.Mtime
	a.Mode = attr.Mode
	a.Uid = attr.Uid
	a.Gid = attr.Gid
	a.Size = attr.Size

	return
}

var (
	_ fs.Node = (*Directory)(nil)
)
