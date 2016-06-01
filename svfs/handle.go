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
	create        bool
	truncated     bool
	nonce         string
	wroteSegment  bool
	segmentID     uint
	uploaded      uint64
	segmentPrefix string
	segmentPath   string
}

// Read gets a swift object data for a request within the current context.
// The request size is always honored. We open the file on the first write.
func (fh *ObjectHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) (err error) {
	if fh.rd == nil {
		fh.rd, err = newReader(fh)
		if err != nil {
			return err
		}
	}
	fh.rd.Seek(req.Offset, 0)
	resp.Data = make([]byte, req.Size)
	io.ReadFull(fh.rd, resp.Data)
	return nil
}

// Release frees the file handle, closing all readers/writers in use.
func (fh *ObjectHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	if fh.rd != nil {
		if closer, ok := fh.rd.(io.Closer); ok {
			closer.Close()
		}
	}
	if fh.wd != nil {
		fh.wd.Close()
		if Encryption {
			if err := updateHeaders(fh.target, fh.nonce); err != nil {
				return err
			}
		}
		fh.target.writing = false
	}
	if changeCache.Exist(fh.target.c.Name, fh.target.path) {
		defer fh.target.m.Unlock()
		changeCache.Remove(fh.target.c.Name, fh.target.path)
	}
	return nil
}

// Write pushes data to a swift object.
// If we detect that we are writing more data than the configured
// segment size, then the first object we were writing to is moved
// to the segment container and named accordingly to DLO conventions.
// Remaining data will be split into segments sequentially until
// file handle release is called. If we are overwriting an object
// we handle segment deletion, and object creation.
func (fh *ObjectHandle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) (err error) {
	// Truncate the file if :
	// - this is the first write after creation
	// - this is the first write after opening an existing file
	if !fh.create && !fh.truncated ||
		fh.create && !fh.target.writing {
		if err := fh.truncate(); err != nil {
			return err
		}
	}

	// Now that we are writing to this file, make sure no lock can
	// be acquired until :
	// - all write operations have been processed
	// - this filehandle has been freed
	fh.target.writing = true

	// Write first segment or file with size smaller than a segment size.
	if fh.uploaded+uint64(len(req.Data)) <= uint64(SegmentSize) {
		if _, err := fh.wd.Write(req.Data); err != nil {
			return err
		}
		fh.uploaded += uint64(len(req.Data))
		fh.target.so.Bytes += int64(len(req.Data))
		goto EndWrite
	}

	// Data written on this writer will be larger than a segment size.
	// Close current object, move it to the segment container if this
	// is the first time this happens, then open the next segment and
	// start writing to it.
	if fh.uploaded+uint64(len(req.Data)) > uint64(SegmentSize) {
		// Close current segment
		if !fh.wroteSegment {
			if err := fh.moveToSegment(); err != nil {
				return err
			}
		}
		fh.wd.Close()

		// Open next segment
		fh.wd, err = initSegment(fh.target.cs.Name, fh.segmentPrefix, &fh.segmentID, fh.target.so, req.Data, &fh.uploaded, &fh.nonce)
		if err != nil {
			return err
		}

		goto EndWrite
	}

EndWrite:
	resp.Size = len(req.Data)
	return nil
}

func (fh *ObjectHandle) moveToSegment() error {
	// Close previous writer.
	fh.wd.Close()

	// Get the next segment name and path
	fh.segmentPrefix = fmt.Sprintf("%s/%d", fh.target.path, time.Now().Unix())
	fh.segmentPath = segmentPath(fh.segmentPrefix, &fh.segmentID)

	// Move data to segment container
	err := SwiftConnection.ObjectMove(fh.target.c.Name, fh.target.path, fh.target.cs.Name, fh.segmentPath)
	if err != nil {
		return err
	}

	// Create the manifest
	createManifest(fh.target, fh.target.c.Name, fh.target.cs.Name+"/"+fh.segmentPrefix, fh.target.path)
	fh.wroteSegment = true
	fh.target.segmented = true

	return err
}

func (fh *ObjectHandle) truncate() (err error) {
	// Remove referenced segments
	if fh.target.segmented {
		err = deleteSegments(fh.target.cs.Name, fh.target.sh[manifestHeader])
		if err != nil {
			return err
		}
		delete(fh.target.sh, manifestHeader)
		fh.target.segmented = false
	}

	// Reopen for writing
	fh.truncated = true
	fh.target.so.Bytes = 0
	fh.wd, err = newWriter(fh.target.c.Name, fh.target.so.Name, &fh.nonce)

	return err
}

var (
	_ fs.Handle         = (*ObjectHandle)(nil)
	_ fs.HandleReleaser = (*ObjectHandle)(nil)
	_ fs.HandleReader   = (*ObjectHandle)(nil)
	_ fs.HandleWriter   = (*ObjectHandle)(nil)
)
