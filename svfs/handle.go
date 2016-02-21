package svfs

import (
	"io"
	"io/ioutil"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

type ObjectHandle struct {
	p *Directory
	t *Object
	r *swift.ObjectOpenFile
	w *swift.ObjectCreateFile
}

func (fh *ObjectHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	buf := make([]byte, req.Size)
	n, err := fh.r.Read(buf)
	if err != io.EOF {
		return fuse.EIO
	}
	resp.Data = buf[:n]
	return nil
}

func (fh *ObjectHandle) ReadAll(ctx context.Context) ([]byte, error) {
	data, err := ioutil.ReadAll(fh.r)
	return data, err
}

func (fh *ObjectHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.r != nil {
		fh.r.Close()
	}
	if fh.w != nil {
		fh.w.Close()
	}
	if fh.p != nil {
		fh.p.cache.Delete(fh.p.path)
	}
	return nil
}

func (fh *ObjectHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	n, err := fh.w.Write(req.Data)
	if req.Offset == 0 {
		fh.t.so.Bytes = int64(n)
	} else if req.Offset+int64(len(req.Data))+1 > fh.t.so.Bytes {
		fh.t.so.Bytes = req.Offset + int64(len(req.Data)) + 1
	}
	resp.Size = n
	return err
}

var (
	_ fs.Handle          = (*ObjectHandle)(nil)
	_ fs.HandleReleaser  = (*ObjectHandle)(nil)
	_ fs.HandleReader    = (*ObjectHandle)(nil)
	_ fs.HandleReadAller = (*ObjectHandle)(nil)
	_ fs.HandleWriter    = (*ObjectHandle)(nil)
)
