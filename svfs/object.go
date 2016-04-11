package svfs

import (
	"os"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
	"golang.org/x/net/context"
)

const (
	ManifestHeader    = "X-Object-Manifest"
	ObjectMetaHeader  = "X-Object-Meta-"
	ObjectMtimeHeader = ObjectMetaHeader + "Mtime"
)

// Object is a node representing a swift object.
// It belongs to a container and segmented objects
// are bound to a container of segments.
type Object struct {
	name      string
	path      string
	so        *swift.Object
	sh        *swift.Headers
	c         *swift.Container
	cs        *swift.Container
	p         *Directory
	m         sync.Mutex
	segmented bool
}

// Attr fills the file attributes for an object node.
func (o *Object) Attr(ctx context.Context, a *fuse.Attr) (err error) {
	a.Size = o.size()
	a.Mode = os.FileMode(DefaultMode)
	a.Gid = uint32(DefaultGID)
	a.Uid = uint32(DefaultUID)
	a.Mtime = getMtime(o.so, o.sh)
	a.Ctime = a.Mtime
	a.Crtime = a.Mtime
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
	oh = &ObjectHandle{
		target: o,
		create: mode&fuse.OpenCreate == fuse.OpenCreate,
	}

	// Append mode is not supported
	if mode&fuse.OpenAppend == fuse.OpenAppend {
		return nil, fuse.ENOTSUP
	}

	if mode.IsReadOnly() {
		return oh, nil
	}
	if mode.IsWriteOnly() {

		o.m.Lock()
		ChangeCache.Add(o.c.Name, o.path, o)

		// Can't write with an offset
		*flags |= fuse.OpenNonSeekable
		// Don't cache writes
		*flags |= fuse.OpenDirectIO

		// Remove segments
		if o.segmented && oh.create {
			err = deleteSegments(o.cs.Name, (*o.sh)[ManifestHeader])
			if err != nil {
				return oh, err
			}
			oh.target.segmented = false
		}

		// Create new object
		if oh.create {
			headers := map[string]string{AutoContent: "true"}
			oh.wd, err = SwiftConnection.ObjectCreate(o.c.Name, o.path, false, "", "", headers)
		}

		return oh, err
	}

	return nil, fuse.ENOTSUP
}

// Open returns the file handle associated with this object node.
func (o *Object) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return o.open(req.Flags, &resp.Flags)
}

func (o *Object) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// Change file size. May be used by the kernel
	// to truncate files to 0 size instead of opening
	// them with O_TRUNC flag.
	if req.Valid.Size() {
		o.so.Bytes = int64(req.Size)
		return nil
	}

	if !ExtraAttr || !req.Valid.Mtime() {
		return fuse.ENOTSUP
	}

	// Change mtime
	if !req.Mtime.Equal(getMtime(o.so, o.sh)) {
		o.m.Lock()
		defer o.m.Unlock()

		(*o.sh)[ObjectMtimeHeader] = swift.TimeToFloatString(req.Mtime)
		h := map[string]string{ObjectMtimeHeader: (*o.sh)[ObjectMtimeHeader]}
		return SwiftConnection.ObjectUpdate(o.c.Name, o.so.Name, h)
	}

	return nil
}

// Name gets the name of the underlying swift object.
func (o *Object) Name() string {
	return o.name
}

func (o *Object) size() uint64 {
	return uint64(o.so.Bytes)
}

var (
	_ Node             = (*Object)(nil)
	_ fs.Node          = (*Object)(nil)
	_ fs.NodeSetattrer = (*Object)(nil)
	_ fs.NodeOpener    = (*Object)(nil)
)
