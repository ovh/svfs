package fs

import (
	ctx "golang.org/x/net/context"
)

type FsStats struct {
	Files      uint64
	FilesFree  uint64
	Blocks     uint64
	BlockSize  uint64
	BlocksUsed uint64
	BlocksFree uint64
}

type Fs interface {
	Setup(c ctx.Context, conf interface{}) error
	StatFs(c ctx.Context) (*FsStats, error)
	Root() (Directory, error)
	Shutdown() error
}
