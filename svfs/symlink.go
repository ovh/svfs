package svfs

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
	"golang.org/x/net/context"
)

const (
	SymlinkTargetHeader = "Symlink-Target"
)

// Symlink represents a symbolic link to an object within
// a container.
type Symlink struct {
	name string
	path string
	so   *swift.Object
	sh   swift.Headers
	c    *swift.Container
	p    *Directory
}

// Attr fills the file attributes for a symlink node.
func (s *Symlink) Attr(ctx context.Context, a *fuse.Attr) (err error) {
	a.Size = s.size()
	a.BlockSize = 0
	a.Blocks = 0
	a.Mode = os.ModeSymlink | os.FileMode(DefaultMode)
	a.Gid = uint32(DefaultGID)
	a.Uid = uint32(DefaultUID)
	a.Mtime = getMtime(s.so, s.sh)
	a.Ctime = a.Mtime
	a.Crtime = a.Mtime
	return nil
}

func (s *Symlink) Export() fuse.Dirent {
	return fuse.Dirent{
		Name: s.Name(),
		Type: fuse.DT_Link,
	}
}

func (s *Symlink) Name() string {
	return s.name
}

func (s *Symlink) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return s.sh[ObjectSymlinkHeader], nil
}

func (s *Symlink) size() uint64 {
	return uint64(s.so.Bytes)
}

var (
	_ Node              = (*Symlink)(nil)
	_ fs.Node           = (*Symlink)(nil)
	_ fs.NodeReadlinker = (*Symlink)(nil)
)
