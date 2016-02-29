package svfs

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/xlucas/swift"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type ObjectHandle struct {
	target        *Object
	segment       *swift.ObjectCreateFile
	rd            io.ReadCloser
	writing       bool
	wroteOnce     bool
	segmentBuf    *bytes.Buffer
	segmentID     uint
	uploaded      uint64
	toUpload      uint64
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
		defer fh.rd.Close()
	}
	if fh.writing {
		// Write non-segmented object
		if !fh.wroteOnce {
			err := SwiftConnection.ObjectPutBytes(fh.target.c.Name, fh.target.so.Name, fh.segmentBuf.Bytes(), ObjContentType)
			if err != nil {
				return err
			}
			fh.target.so.Bytes = int64(fh.segmentBuf.Len())
		}
		if fh.wroteOnce {
			// Close last segment
			fh.segment.Close()

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

	// Prepare first segment
	if !fh.wroteOnce && fh.segmentBuf == nil {
		fh.segmentBuf = bytes.NewBuffer(make([]byte, 0, SegmentSize))
		fh.segmentPrefix = fmt.Sprintf("%s/%d", fh.target.path, time.Now().Unix())
	}

	// We reached the segment size
	if uint64(len(req.Data))+fh.toUpload > SegmentSize || fh.uploaded+uint64(len(req.Data)) > SegmentSize {
		if !fh.wroteOnce {
			fh.segment, err = createAndWriteSegment(fh.target.cs.Name, fh.segmentPrefix, &fh.segmentID, fh.target.so, fh.segmentBuf.Bytes(), &fh.uploaded)
			if err != nil {
				return err
			}
			fh.toUpload = 0
			fh.wroteOnce = true
			fh.segment.Close()
			fh.segmentBuf.Reset()
		} else {
			fh.segment.Close()
		}

		fh.segment, err = createAndWriteSegment(fh.target.cs.Name, fh.segmentPrefix, &fh.segmentID, fh.target.so, req.Data, &fh.uploaded)
		if err != nil {
			return err
		}
	} else if fh.wroteOnce {
		writeSegmentData(fh.segment, fh.target.so, req.Data, &fh.uploaded)
	} else if !fh.wroteOnce {
		fh.segmentBuf.Write(req.Data)
		fh.toUpload += uint64(len(req.Data))
	}

	resp.Size = len(req.Data)

	return nil
}

var (
	_ fs.Handle         = (*ObjectHandle)(nil)
	_ fs.HandleReleaser = (*ObjectHandle)(nil)
	_ fs.HandleReader   = (*ObjectHandle)(nil)
	_ fs.HandleWriter   = (*ObjectHandle)(nil)
)
