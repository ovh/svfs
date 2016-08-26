package svfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bazil.org/fuse"
)

const containerName = "svfs_test"

func TestRoot(t *testing.T) {
	t.Run("FsInit", testFsInit)
	t.Run("FsRoot", testFsRoot)
	t.Run("RootReadDirAll", testRootReadDirAll)

	// Unsupported operations
	t.Run("RootCreate", testRootCreate)
	t.Run("RootRename", testRootRename)
	t.Run("RootRemoveFile", testRootRemoveFile)

	// Container creation
	t.Run("RootMkdir", testRootMkdir)
	t.Run("RootLookup", testRootLookup)

	// Container removal
	t.Run("RootRemove", testRootRemove)
	t.Run("RootLookupMiss", testRootLookupMiss)
}

func testRootCreate(t *testing.T) {
	req := &fuse.CreateRequest{Name: "foo"}
	_, _, err := ctx.r.Create(nil, req, &fuse.CreateResponse{})
	assert.Equal(t, err, fuse.ENOTSUP)
}

func testRootLookup(t *testing.T) {
	req := &fuse.LookupRequest{Name: containerName}
	n, err := ctx.r.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Nil(t, err)
	require.IsType(t, &Directory{}, n)
	e, _ := n.(*Directory)
	assert.Equal(t, e.Name(), req.Name)
}

func testRootLookupMiss(t *testing.T) {
	req := &fuse.LookupRequest{Name: containerName}
	_, err := ctx.r.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Equal(t, err, fuse.ENOENT)
}

func testRootMkdir(t *testing.T) {
	req := &fuse.MkdirRequest{Name: containerName}
	c, err := ctx.r.Mkdir(nil, &fuse.MkdirRequest{Name: req.Name})
	assert.Nil(t, err)
	require.IsType(t, &Directory{}, c)
	ctx.c, _ = c.(*Directory)
	assert.Equal(t, ctx.c.Name(), req.Name)
}

func testRootReadDirAll(t *testing.T) {
	_, err := ctx.r.ReadDirAll(nil)
	assert.Nil(t, err)
}

func testRootRename(t *testing.T) {
	req := &fuse.RenameRequest{OldName: "foo", NewName: "bar"}
	assert.Equal(t, ctx.r.Rename(nil, req, ctx.r), fuse.ENOTSUP)
}

func testRootRemove(t *testing.T) {
	req := &fuse.RemoveRequest{Name: containerName, Dir: true}
	assert.Nil(t, ctx.r.Remove(nil, req))
}

func testRootRemoveFile(t *testing.T) {
	req := &fuse.RemoveRequest{Name: "foo"}
	assert.Equal(t, ctx.r.Remove(nil, req), fuse.ENOTSUP)
}
