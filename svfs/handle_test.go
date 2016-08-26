package svfs

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"bazil.org/fuse"
)

func TestHandle(t *testing.T) {
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

	// Write object
	t.Run("Object_OpenWriteOnly", testObjectOpenWriteOnly)
	t.Run("ObjectHandle_Write", testObjectHandleWrite)
	t.Run("ObjectHandle_Close", testObjectHandleClose)

	// Read object
	t.Run("Object_OpenReadOnly", testObjectOpenReadOnly)
	t.Run("ObjectHandle_Read", testObjectHandleRead)
	t.Run("ObjectHandle_Close", testObjectHandleClose)

	t.Run("Directory_Remove", testDirectoryRemove)
	t.Run("Container_Rmdir", testContainerRmdir)
	t.Run("RootRemove", testRootRemove)
}

func testObjectHandleClose(t *testing.T) {
	assert.Nil(t, ctx.h.Release(nil, nil))
}

func testObjectHandleRead(t *testing.T) {
	req := &fuse.ReadRequest{Size: len(ctx.b)}
	rep := &fuse.ReadResponse{Data: make([]byte, len(ctx.b))}

	err := ctx.h.Read(nil, req, rep)

	assert.Nil(t, err)
	assert.Len(t, rep.Data, len(ctx.b))

	for b := 0; b < len(ctx.b); b++ {
		assert.Equal(t, ctx.b[b], rep.Data[b])
	}
}

func testObjectHandleWrite(t *testing.T) {
	SegmentSize = uint64(len(ctx.b))

	rand.Read(ctx.b[:])
	req := &fuse.WriteRequest{Data: ctx.b[:]}
	rep := &fuse.WriteResponse{}

	err := ctx.h.Write(nil, req, rep)

	assert.Nil(t, err)
	assert.Equal(t, rep.Size, len(ctx.b))
	assert.False(t, ctx.f.segmented)
}
