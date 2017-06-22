package fs

import (
	ctx "golang.org/x/net/context"
)

type FileHandle interface {
	Read(c ctx.Context, offset int64, size int) ([]byte, error)
	Write(c ctx.Context, offet int64, data []byte) error
	Close(c ctx.Context) error
}
