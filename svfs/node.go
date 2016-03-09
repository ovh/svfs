package svfs

import "bazil.org/fuse"

// Node is the generic interface of an SVFS node.
type Node interface {
	Name() string
	Export() fuse.Dirent
}
