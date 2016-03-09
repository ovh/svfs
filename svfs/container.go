package svfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// Container is a node representing a directory entry bound to
// a Swift container.
type Container struct {
	*Directory
}

// Attr fills the container file attributes within the current context.
func (c *Container) Attr(ctx context.Context, a *fuse.Attr) error {
	c.Directory.Attr(ctx, a)
	a.Size = c.size()
	return nil
}

func (c *Container) size() uint64 {
	if c.cs != nil {
		return uint64(c.c.Bytes + c.cs.Bytes)
	}
	return uint64(c.c.Bytes)
}

var (
	_ Node    = (*Container)(nil)
	_ fs.Node = (*Container)(nil)
)
