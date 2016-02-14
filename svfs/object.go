package svfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

type Object struct {
	name      string
	path      string
	s         *swift.Connection
	so        *swift.Object
	c         *swift.Container
	p         *Directory
	segmented bool
}

func (o *Object) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Size = o.size()
	a.Mode = 0600
	a.Mtime = o.so.LastModified
	a.Ctime = o.so.LastModified
	a.Crtime = o.so.LastModified
	return nil
}

func (o *Object) Export() fuse.Dirent {
	return fuse.Dirent{
		Name: o.Name(),
		Type: fuse.DT_File,
	}
}

func (o *Object) open(mode fuse.OpenFlags) (oh *ObjectHandle, err error) {
	oh = &ObjectHandle{p: o.p}

	// Modes
	if mode.IsReadOnly() {
		oh.r, _, err = o.s.ObjectOpen(o.c.Name, o.so.Name, false, nil)
		return oh, err
	}
	if mode.IsWriteOnly() {
		oh.w, err = o.s.ObjectCreate(o.c.Name, o.so.Name, false, "", "application/octet-sream", nil)
		return oh, err
	}
	if mode.IsReadWrite() {
		return nil, fuse.ENOTSUP
	}

	return nil, fuse.ENOTSUP
}

func (o *Object) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags = fuse.OpenDirectIO
	return o.open(req.Flags)
}

func (o *Object) Name() string {
	return o.name
}

func (o *Object) size() uint64 {
	return uint64(o.so.Bytes)
}

var (
	_ Node          = (*Object)(nil)
	_ fs.Node       = (*Object)(nil)
	_ fs.NodeOpener = (*Object)(nil)
)
