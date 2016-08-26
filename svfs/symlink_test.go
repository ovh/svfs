package svfs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"bazil.org/fuse"
)

func TestSymlink(t *testing.T) {
	ctx.it = oldFileName
	ctx.rc = 0

	t.Run("Fs_Init", testFsInit)
	t.Run("Fs_Root", testFsRoot)
	t.Run("Root_Mkdir", testRootMkdir)
	t.Run("Root_ReadDirAll", testRootReadDirAll)
	t.Run("Container_ReadDirAll", testContainerReadDirAll)
	t.Run("Container_Mkdir", testContainerMkdir)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_Create", testDirectoryCreate)

	ctx.it = symlinkName
	t.Run("Directory_Symlink", testDirectorySymlink)
	t.Run("Symlink_Readlink", testSymlinkReadlink)
	t.Run("Directory_Remove", testDirectoryRemove)

	ctx.it = oldFileName
	t.Run("Directory_Remove", testDirectoryRemove)
	t.Run("Container_Rmdir", testContainerRmdir)
	t.Run("RootRemove", testRootRemove)
}

func testSymlinkReadlink(t *testing.T) {
	target, err := ctx.s.Readlink(nil, &fuse.ReadlinkRequest{})
	assert.Nil(t, err)
	assert.Equal(t, target, ctx.f.Name())
}
