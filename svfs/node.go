package svfs

import "bazil.org/fuse"

type Node interface {
	Name() string
	Export() fuse.Dirent
}
