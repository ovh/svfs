package svfs

import (
	"fmt"
	"io"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type ObjectHandle struct {
	target        *Object
	rd            io.ReadCloser
	wd            io.WriteCloser
	wroteSegment  bool
	segmentID     uint
	uploaded      uint64
	segmentPrefix string
	segmentPath   string
}

func (fh *ObjectHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	resp.Data = make([]byte, req.Size)
	io.ReadFull(fh.rd, resp.Data)
	return nil
}

func (fh *ObjectHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.rd != nil {
		fh.rd.Close()
	}
	if fh.wd != nil {
		fh.wd.Close()
		if fh.wroteSegment {
			// Create the manifest
			headers := map[string]string{ManifestHeader: fh.target.cs.Name + "/" + fh.segmentPrefix, "Content-Length": "0"}
			SwiftConnection.ObjectDelete(fh.target.c.Name, fh.target.so.Name)
			manifest, err := SwiftConnection.ObjectCreate(fh.target.c.Name, fh.target.so.Name, false, "", ObjContentType, headers)
			if err != nil {
				return err
			}
			manifest.Write(nil)
			manifest.Close()
		}
		DirectoryCache.Set(fh.target.c.Name, fh.target.path, fh.target.name, fh.target)
	}
	return nil
}

func (fh *ObjectHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) (err error) {
	if fh.uploaded+uint64(len(req.Data)) <= uint64(SegmentSize) {
		// File size is less than the size of a segment
		// or we didn't filled the current segment yet.
		if _, err := fh.wd.Write(req.Data); err != nil {
			return err
		}
		fh.uploaded += uint64(len(req.Data))
		fh.target.so.Bytes += int64(fh.uploaded)
		goto EndWrite
	}
	if fh.uploaded+uint64(len(req.Data)) > uint64(SegmentSize) {
		if !fh.wroteSegment {
			// File size is greater than the size of a segment
			// Move it to the segment directory and start writing
			// next segment.
			fh.wd.Close()
			fh.wroteSegment = true
			fh.segmentPrefix = fmt.Sprintf("%s/%d", fh.target.path, time.Now().Unix())
			fh.segmentPath = segmentPath(fh.segmentPrefix, &fh.segmentID)
			if err := SwiftConnection.ObjectMove(fh.target.c.Name, fh.target.path, fh.target.cs.Name, fh.segmentPath); err != nil {
				return err
			}
		}
		fh.wd.Close()
		fh.wd, err = createAndWriteSegment(fh.target.cs.Name, fh.segmentPrefix, &fh.segmentID, fh.target.so, req.Data, &fh.uploaded)
		if err != nil {
			return err
		}
		goto EndWrite
	}

EndWrite:
	resp.Size = len(req.Data)
	return nil
}

var (
	_ fs.Handle         = (*ObjectHandle)(nil)
	_ fs.HandleReleaser = (*ObjectHandle)(nil)
	_ fs.HandleReader   = (*ObjectHandle)(nil)
	_ fs.HandleWriter   = (*ObjectHandle)(nil)
)
