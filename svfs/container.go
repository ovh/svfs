package svfs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ncw/swift"
	"golang.org/x/net/context"
)

type Container struct {
	*Directory
	cs *swift.Container
}

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

var _ Node = (*Directory)(nil)
var _ fs.Node = (*Directory)(nil)
