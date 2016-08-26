package svfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bazil.org/fuse"
)

func TestObject(t *testing.T) {
	ctx.it = newFileName
	ctx.rc = 0

	t.Run("Fs_Init", testFsInit)
	t.Run("Fs_Root", testFsRoot)
	t.Run("Root_Mkdir", testRootMkdir)
	t.Run("Root_ReadDirAll", testRootReadDirAll)
	t.Run("Container_ReadDirAll", testContainerReadDirAll)
	t.Run("Container_Mkdir", testContainerMkdir)
	t.Run("Directory_Create", testDirectoryCreate)

	ctx.rc = 1
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)

	// Unsupported operations
	t.Run("Object_OpenAppend", testObjectOpenAppend)
	t.Run("Object_OpenReadWrite", testObjectOpenReadWrite)

	// Open WO
	t.Run("Object_OpenWriteOnly", testObjectOpenWriteOnly)
	t.Run("ObjectHandle_Close", testObjectHandleClose)

	// Open RO
	t.Run("Object_OpenReadOnly", testObjectOpenReadOnly)
	t.Run("ObjectHandle_Close", testObjectHandleClose)

	t.Run("Directory_Remove", testDirectoryRemove)
	t.Run("Container_Rmdir", testContainerRmdir)
	t.Run("RootRemove", testRootRemove)
}

func testObjectOpenAppend(t *testing.T) {
	req := &fuse.OpenRequest{Flags: fuse.OpenAppend}
	_, err := ctx.f.Open(nil, req, &fuse.OpenResponse{})
	assert.Equal(t, err, fuse.ENOTSUP)
}

func testObjectOpenReadWrite(t *testing.T) {
	req := &fuse.OpenRequest{Flags: fuse.OpenReadWrite}
	_, err := ctx.f.Open(nil, req, &fuse.OpenResponse{})
	assert.Equal(t, err, fuse.ENOTSUP)
}

func testObjectOpenReadOnly(t *testing.T) {
	req := &fuse.OpenRequest{Flags: fuse.OpenReadOnly}
	rep := &fuse.OpenResponse{}

	fh, err := ctx.f.Open(nil, req, rep)

	assert.Nil(t, err)
	require.NotNil(t, rep)
	require.NotNil(t, fh)
	require.IsType(t, &ObjectHandle{}, fh)

	ctx.h, _ = fh.(*ObjectHandle)
}

func testObjectOpenWriteOnly(t *testing.T) {
	req := &fuse.OpenRequest{Flags: fuse.OpenWriteOnly}
	rep := &fuse.OpenResponse{}

	fh, err := ctx.f.Open(nil, req, rep)

	assert.Nil(t, err)
	require.NotNil(t, rep)
	require.NotNil(t, fh)
	require.IsType(t, &ObjectHandle{}, fh)
	assert.True(t, rep.Flags&fuse.OpenDirectIO != 0)
	assert.True(t, rep.Flags&fuse.OpenNonSeekable != 0)

	ctx.h, _ = fh.(*ObjectHandle)
}
