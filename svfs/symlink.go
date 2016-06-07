package svfs

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/xlucas/swift"
	"golang.org/x/net/context"
)

const (
	objectSymlinkHeader = objectMetaHeader + "Symlink-Target"
)

// Symlink represents a symbolic link to an object within a container.
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
	a.Size = uint64(s.so.Bytes)
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

// Export converts this symlink node to a direntry.
func (s *Symlink) Export() fuse.Dirent {
	return fuse.Dirent{Name: s.Name(), Type: fuse.DT_Link}
}

// Name gets the name of the underlying swift object.
func (s *Symlink) Name() string {
	return s.name
}

// Readlink gets the symlink target path.
func (s *Symlink) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return s.sh[objectSymlinkHeader], nil
}

func (s *Symlink) copy(dir *Directory, name string) (*Symlink, error) {
	_, err := SwiftConnection.ObjectCopy(s.c.Name, s.path, dir.c.Name, dir.path+name, nil)
	if err != nil {
		return nil, err
	}

	link := *s
	*link.so = *s.so
	link.c = dir.c
	link.p = dir
	link.name = name
	link.path = dir.path + name

	directoryCache.Set(dir.c.Name, dir.path, name, &link)

	return &link, nil
}

func (s *Symlink) delete() error {
	directoryCache.Delete(s.c.Name, s.p.path, s.name)
	return SwiftConnection.ObjectDelete(s.c.Name, s.path)
}

func (s *Symlink) rename(dir *Directory, name string) error {
	copy, err := s.copy(dir, name)
	if err != nil {
		return err
	}

	err = s.delete()
	if err != nil {
		return err
	}

	*s = *copy

	return nil
}

var (
	_ Node              = (*Symlink)(nil)
	_ fs.Node           = (*Symlink)(nil)
	_ fs.NodeReadlinker = (*Symlink)(nil)
)
