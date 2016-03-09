package svfs

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
	"golang.org/x/net/context"
)

const (
	ManifestHeader = "X-Object-Manifest"
)

var (
	SegmentSize uint64
)

// Object is a node representing a swift object.
// It belongs to a container and segmented objects
// are bound to a container of segments.
type Object struct {
	name      string
	path      string
	so        *swift.Object
	c         *swift.Container
	cs        *swift.Container
	p         *Directory
	segmented bool
}

// Attr fills the file attributes for an object node.
func (o *Object) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Size = o.size()
	a.Mode = os.FileMode(DefaultMode)
	a.Gid = uint32(DefaultGID)
	a.Uid = uint32(DefaultUID)
	a.Mtime = o.so.LastModified
	a.Ctime = o.so.LastModified
	a.Crtime = o.so.LastModified
	return nil
}

// Export converts this object node as a direntry.
func (o *Object) Export() fuse.Dirent {
	return fuse.Dirent{
		Name: o.Name(),
		Type: fuse.DT_File,
	}
}

func (o *Object) open(mode fuse.OpenFlags, flags *fuse.OpenResponseFlags) (oh *ObjectHandle, err error) {
	oh = &ObjectHandle{target: o}

	// Append mode is not supported
	if mode&fuse.OpenAppend == fuse.OpenAppend {
		return nil, fuse.ENOTSUP
	}

	// Can't seek in an open file.
	*flags |= fuse.OpenNonSeekable

	if mode.IsReadOnly() {
		oh.rd, _, err = SwiftConnection.ObjectOpen(o.c.Name, o.so.Name, false, nil)
		return oh, err
	}
	if mode.IsWriteOnly() {

		// Direct IO
		*flags |= fuse.OpenDirectIO

		// Remove segments if the previous file was a manifest
		_, h, err := SwiftConnection.Object(o.c.Name, o.so.Name)
		if err != swift.ObjectNotFound {
			if SegmentPathRegex.Match([]byte(h[ManifestHeader])) {
				deleteSegments(o.cs.Name, h[ManifestHeader])
			}
		}
		oh.wd, err = SwiftConnection.ObjectCreate(o.c.Name, o.so.Name, false, "", ObjContentType, nil)
		return oh, err
	}

	return nil, fuse.ENOTSUP
}

// Open returns the file handle associated with this object node.
func (o *Object) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return o.open(req.Flags, &resp.Flags)
}

// Name gets the name of the underlying swift object.
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
