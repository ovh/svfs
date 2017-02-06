package swift

import (
	"os"
	"syscall"
	"time"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"
)

type Account struct {
	*Fs
	swiftAccount *swift.Account
}

func (a *Account) Create(nodeName string) (file fs.File, err error) {
	return nil, syscall.ENOTSUP
}

func (a *Account) GetAttr() (attr *fs.Attr, err error) {
	attr = &fs.Attr{
		Atime: time.Now(),
		Ctime: a.swiftAccount.CreationTime(),
		Mtime: a.swiftAccount.CreationTime(),
		Uid:   a.Fs.conf.Uid,
		Gid:   a.Fs.conf.Gid,
		Mode:  os.ModeDir | a.Fs.conf.Perms,
		Size:  a.Fs.conf.BlockSize,
	}

	return
}

func (a *Account) Hardlink(targetPath string, linkName string) error {
	return syscall.ENOTSUP
}

func (a *Account) Mkdir(dirName string) (fs.Directory, error) {
	con := a.storage.Borrow().(*swift.Connection)
	defer a.storage.Return()

	container, err := swift.NewLogicalContainer(con, dirName)

	return &Container{Fs: a.Fs, swiftContainer: container}, err
}

func (a *Account) Remove(node fs.Node) (err error) {
	if _, ok := node.(*Container); !ok {
		return syscall.ENOTSUP
	}

	con := a.storage.Borrow().(*swift.Connection)
	defer a.storage.Return()

	return con.DeleteLogicalContainer(node.(*Container).swiftContainer)
}

func (a *Account) Rename(node fs.Node, newName string, newDir fs.Directory,
) (err error,
) {
	return syscall.ENOTSUP
}

func (a *Account) Symlink(targetPath string, linkName string) error {
	return syscall.ENOTSUP
}
