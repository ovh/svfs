package fs

import (
	"os"
	"time"
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

type Node interface {
	GetAttr() (*Attr, error)
	GetXAttr(attrName string) (*XAttr, error)
	ListXAttr() ([]*XAttr, error)
	RemoveXAttr(attrName string) error
	SetAttr(*Attr) error
	SetXAttr(*XAttr) error
}
