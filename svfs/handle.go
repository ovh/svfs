package svfs

import (
	"io"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

type ObjectHandle struct {
	t      *Object
	w      *swift.ObjectCreateFile
	rd     io.ReadCloser
	buffer []byte
}

func (fh *ObjectHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if len(fh.buffer) < req.Size {
		fh.buffer = make([]byte, req.Size)
	}
	buffer := make([]byte, req.Size)
	io.ReadFull(fh.rd, buffer)
	resp.Data = buffer
	return nil
}

func (fh *ObjectHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.rd != nil {
		defer fh.rd.Close()
	}
	if fh.w != nil {
		EntryCache.Set(fh.t.c.Name, fh.t.path, fh.t.name, fh.t)
		defer fh.w.Close()
	}
	return nil
}

func (fh *ObjectHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	n, err := fh.w.Write(req.Data)
	fh.t.so.Bytes += int64(n)
	resp.Size = n
	return err
}

var (
	_ fs.Handle         = (*ObjectHandle)(nil)
	_ fs.HandleReleaser = (*ObjectHandle)(nil)
	_ fs.HandleReader   = (*ObjectHandle)(nil)
	_ fs.HandleWriter   = (*ObjectHandle)(nil)
)
