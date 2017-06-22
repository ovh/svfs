package swift

import (
	"os"
	"time"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"

	ctx "golang.org/x/net/context"
)

type Container struct {
	*Fs
	swiftContainer *swift.LogicalContainer
}

func (co *Container) Create(c ctx.Context, nodeName string) (fs.File, error) {
	panic("not implemented")
}

func (co *Container) GetAttr(c ctx.Context) (attr *fs.Attr, err error) {
	attr = &fs.Attr{
		Atime: time.Now(),
		Ctime: co.swiftContainer.CreationTime(),
		Mtime: co.swiftContainer.CreationTime(),
		Uid:   co.Fs.conf.Uid,
		Gid:   co.Fs.conf.Gid,
		Mode:  os.ModeDir | co.Fs.conf.Perms,
		Size:  co.Fs.conf.BlockSize,
	}

	return
}

func (co *Container) Hardlink(c ctx.Context, targetPath string, linkName string) error {
	panic("not implemented")
}

func (co *Container) Mkdir(c ctx.Context, dirName string) (fs.Directory, error) {
	panic("not implemented")
}

func (co *Container) Remove(c ctx.Context, node fs.Node) error {
	panic("not implemented")
}

func (co *Container) Rename(c ctx.Context, node fs.Node, newName string, newDir fs.Directory) (err error) {
	panic("not implemented")
}

func (co *Container) Symlink(c ctx.Context, targetPath string, linkName string) error {
	panic("not implemented")
}
