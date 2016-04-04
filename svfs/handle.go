package svfs

import (
	"fmt"
	"io"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// ObjectHandle represents an open object handle, similarly to
// file handles.
type ObjectHandle struct {
	target        *Object
	rd            io.ReadSeeker
	wd            io.WriteCloser
	wroteSegment  bool
	segmentID     uint
	uploaded      uint64
	segmentPrefix string
	segmentPath   string
}

// Read gets a swift object data for a request within the current context.
// The request size is always honored.
func (fh *ObjectHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	fh.rd.Seek(req.Offset, 0)
	resp.Data = make([]byte, req.Size)
	io.ReadFull(fh.rd, resp.Data)
	return nil
}

// Release frees the file handle, closing all readers/writers in use.
// In case we used this file handle to write a large object, it creates
// the manifest file. Cache is refreshed if the writer was used during
// the lifetime of this handle.
func (fh *ObjectHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.rd != nil {
		if closer, ok := fh.rd.(io.Closer); ok {
			closer.Close()
		}
	}
	if fh.wd != nil {
		fh.wd.Close()
		if fh.wroteSegment {
			// Create the manifest
			headers := map[string]string{
				ManifestHeader:   fh.target.cs.Name + "/" + fh.segmentPrefix,
				"Content-Length": "0",
				AutoContent:      "true",
			}
			SwiftConnection.ObjectDelete(fh.target.c.Name, fh.target.so.Name)
			manifest, err := SwiftConnection.ObjectCreate(fh.target.c.Name, fh.target.so.Name, false, "", "", headers)
			if err != nil {
				return err
			}
			fh.target.segmented = true
			manifest.Write(nil)
			manifest.Close()
		}
		DirectoryCache.Set(fh.target.c.Name, fh.target.path, fh.target.name, fh.target)
	}
	return nil
}

// Write pushes data to a swift object.
// If we detect that we are writing more data than the configured
// segment size, then the first object we were writing to is moved
// to the segment container and named accordingly to DLO conventions.
// Remaining data will be split into segments sequentially until
// file handle release is called.
func (fh *ObjectHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) (err error) {
	if fh.uploaded+uint64(len(req.Data)) <= uint64(SegmentSize) {
		// File size is less than the size of a segment
		// or we didn't fill the current segment yet.
		if _, err := fh.wd.Write(req.Data); err != nil {
			return err
		}
		fh.uploaded += uint64(len(req.Data))
		fh.target.so.Bytes += int64(len(req.Data))
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
