package swift

import (
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
		Mtime: a.Fs.mountTime,
	}

	return
}

func (a *Account) Hardlink(targetPath string, linkName string) error {
	return syscall.ENOTSUP
}

func (a *Account) Mkdir(dirName string) (dir fs.Directory, err error) {
	con := a.storage.Borrow().(*swift.Connection)
	defer a.storage.Return()

	container, err := swift.NewLogicalContainer(
		con,
		a.Fs.options.GetString(StoragePolicyOption),
		dirName,
	)
	if err != nil {
		return
	}

	dir = &Container{swiftContainer: container}

	return
}

func (a *Account) Remove(node fs.Node) (err error) {
	if _, ok := node.(*Container); !ok {
		return syscall.ENOTSUP
	}

	con := a.storage.Borrow().(*swift.Connection)
	defer a.storage.Return()

	err = con.DeleteLogicalContainer(node.(*Container).swiftContainer)

	return
}

func (a *Account) Rename(newName string, newDir fs.Directory) error {
	return syscall.ENOTSUP
}

func (a *Account) Symlink(targetPath string, linkName string) error {
	return syscall.ENOTSUP
}
