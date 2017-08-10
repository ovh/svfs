package fs

import (
	"os"
	"time"

	ctx "golang.org/x/net/context"
)

type Attr struct {
	Atime time.Time
	Ctime time.Time
	Mtime time.Time
	Mode  os.FileMode
	Uid   uint32
	Gid   uint32
	Size  uint64
}

type XAttr struct {
	Key   string
	Value string
}

type Inode uint64

type Node interface {
	Name(c ctx.Context) string
	GetAttr(c ctx.Context) (*Attr, error)
	GetXAttr(c ctx.Context, attrName string) (*XAttr, error)
	ListXAttr(c ctx.Context) ([]*XAttr, error)
	RemoveXAttr(c ctx.Context, attrName string) error
	SetAttr(c ctx.Context, attr *Attr) error
	SetXAttr(c ctx.Context, xattr *XAttr) error
}
