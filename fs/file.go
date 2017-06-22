package fs

import (
	ctx "golang.org/x/net/context"
)

type File interface {
	Open(c ctx.Context, flags uint32) (FileHandle, error)
	Fsync(c ctx.Context) error
}
