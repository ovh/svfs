package swift

import (
	"os"
	"syscall"
	"time"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"
	ctx "golang.org/x/net/context"
)

type Account struct {
	*Fs
	swiftAccount *swift.Account
}

func NewAccount(fs *Fs, account *swift.Account) *Account {
	return &Account{Fs: fs, swiftAccount: account}
}

func (a *Account) Create(c ctx.Context, nodeName string) (file fs.File, err error) {
	return nil, syscall.ENOTSUP
}

func (a *Account) GetAttr(ctx.Context) (attr *fs.Attr, err error) {
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

func (a *Account) Hardlink(c ctx.Context, targetPath string, linkName string) error {
	return syscall.ENOTSUP
}

func (a *Account) Mkdir(c ctx.Context, dirName string) (fs.Directory, error) {
	con := a.pool.Borrow().(*swift.Connection)
	defer a.pool.Return()

	container, err := swift.NewLogicalContainer(con, dirName)
	if err != nil {
		return nil, err
	}

	return NewContainer(a.Fs, container), nil
}

func (a *Account) Name(c ctx.Context) string {
	return ""
}

func (a *Account) ReadDir(c ctx.Context) (nodes []fs.Node, err error) {
	con := a.pool.Borrow().(*swift.Connection)
	defer a.pool.Return()

	containers, err := con.LogicalContainersAll()
	if err != nil {
		return
	}

	for _, container := range containers {
		nodes = append(nodes, NewContainer(a.Fs, container))
	}

	return
}

func (a *Account) Remove(c ctx.Context, node fs.Node) (err error) {
	if _, ok := node.(*Container); !ok {
		return syscall.ENOTSUP
	}

	con := a.pool.Borrow().(*swift.Connection)
	defer a.pool.Return()

	return con.DeleteLogicalContainer(node.(*Container).swiftContainer)
}

func (a *Account) Rename(c ctx.Context, node fs.Node, newName string, newDir fs.Directory) (err error) {
	return syscall.ENOTSUP
}

func (a *Account) Symlink(c ctx.Context, targetPath string, linkName string) error {
	return syscall.ENOTSUP
}
