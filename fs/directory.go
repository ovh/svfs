package fs

import (
	ctx "golang.org/x/net/context"
)

type Directory interface {
	Create(c ctx.Context, nodeName string) (File, error)
	Hardlink(c ctx.Context, targetPath, linkName string) error
	Mkdir(c ctx.Context, dirName string) (Directory, error)
	Remove(c ctx.Context, node Node) error
	Rename(c ctx.Context, node Node, newName string, newDir Directory) error
	Symlink(c ctx.Context, targetPath, linkName string) error
}
