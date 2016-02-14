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

func (o *Object) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	oh := ObjectHandle{}

	// Not supported
	if req.Flags.IsReadWrite() {
		return nil, fuse.ENOTSUP
	}

	// RO
	if req.Flags.IsReadOnly() {
		r, _, err := o.s.ObjectOpen(o.c.Name, o.so.Name, false, nil)
		if err != nil {
			return nil, fuse.EIO
		}
		oh.r = r
	}

	// WO
	if req.Flags.IsWriteOnly() {
		w, err := o.s.ObjectCreate(o.c.Name, o.so.Name, false, "", "application/octet-sream", nil)
		if err != nil {
			return nil, fuse.EIO
		}
		oh.w = w
	}

	return &oh, nil
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
