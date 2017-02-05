package fuse

import (
	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	sfs "github.com/ovh/svfs/fs"
)

type SVFS struct {
	sfs.Fs
}

func (svfs *SVFS) Root() (fs.Node, error) {
	dir, err := svfs.Fs.Root()
	return &Directory{dir}, err
}

func (svfs *SVFS) Statfs(ctx context.Context, req *fuse.StatfsRequest,
	resp *fuse.StatfsResponse) (err error,
) {
	stats, err := svfs.Fs.StatFs()

	resp.Bavail = stats.BlocksFree
	resp.Bfree = stats.BlocksFree
	resp.Blocks = stats.Blocks
	resp.Bsize = uint32(stats.BlockSize)
	resp.Files = stats.Files
	resp.Ffree = stats.FilesFree

	return
}

var (
	_ fs.FS         = (*SVFS)(nil)
	_ fs.FSStatfser = (*SVFS)(nil)
)
