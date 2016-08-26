package svfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bazil.org/fuse"
)

const (
	directoryName = "directory"
	oldFileName   = "file_1"
	newFileName   = "file_2"
	symlinkName   = "symlink"
	hardlinkName  = "hardlink"
)

func TestDirectory(t *testing.T) {
	t.Run("Fs_Init", testFsInit)
	t.Run("Fs_Root", testFsRoot)
	t.Run("Root_Mkdir", testRootMkdir)
	t.Run("Root_ReadDirAll", testRootReadDirAll)

	// List container content
	ctx.rc = 0
	t.Run("Container_ReadDirAll", testContainerReadDirAll)

	// Unsupported operations
	t.Run("Container_Rename", testContainerRename)

	// Directory creation
	ctx.rc = 1
	ctx.it = directoryName
	t.Run("Container_Mkdir", testContainerMkdir)
	t.Run("Container_ReadDirAll", testContainerReadDirAll)
	t.Run("Container_Lookup", testContainerLookup)

	// File creation
	ctx.rc = 1
	ctx.it = oldFileName
	t.Run("Directory_Create", testDirectoryCreate)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_Lookup", testDirectoryLookup)

	// Hardlink creation
	ctx.rc = 2
	ctx.it = hardlinkName
	t.Run("Directory_Hardlink", testDirectoryHardlink)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_LookupHardlink", testDirectoryLookupHardlink)

	// Symlink creation
	ctx.rc = 3
	ctx.it = symlinkName
	t.Run("Directory_Symlink", testDirectorySymlink)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_LookupSymlink", testDirectoryLookupSymlink)

	// File renaming
	ctx.rc = 3
	ctx.it = newFileName
	t.Run("Directory_Rename", testDirectoryRename)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_Lookup", testDirectoryLookup)

	// File removal
	ctx.rc = 2
	ctx.it = newFileName
	t.Run("Directory_Remove", testDirectoryRemove)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_LookupMiss", testDirectoryLookupMiss)

	// Hardlink removal
	ctx.rc = 1
	ctx.it = hardlinkName
	t.Run("Directory_Remove", testDirectoryRemove)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_LookupMiss", testDirectoryLookupMiss)

	// Symlink removal
	ctx.rc = 0
	ctx.it = symlinkName
	t.Run("Directory_Remove", testDirectoryRemove)
	t.Run("Directory_ReadDirAll", testDirectoryReadDirAll)
	t.Run("Directory_LookupMiss", testDirectoryLookupMiss)

	// Directory removal
	ctx.rc = 0
	ctx.it = directoryName
	t.Run("Container_Rmdir", testContainerRmdir)
	t.Run("Container_ReadDirAll", testContainerReadDirAll)
	t.Run("Container_LookupMiss", testContainerLookupMiss)

	// Container removal
	t.Run("Root_Remove", testRootRemove)
}

func testContainerMkdir(t *testing.T) {
	req := &fuse.MkdirRequest{Name: directoryName}
	dir, err := ctx.c.Mkdir(nil, req)
	assert.Nil(t, err)
	require.IsType(t, &Directory{}, dir)
	ctx.d, _ = dir.(*Directory)
}

func testContainerLookup(t *testing.T) {
	req := &fuse.LookupRequest{Name: ctx.it}
	dir, err := ctx.c.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Nil(t, err)
	require.IsType(t, &Directory{}, dir)
	d, _ := dir.(*Directory)
	assert.Equal(t, d.Name(), req.Name)
}

func testContainerLookupMiss(t *testing.T) {
	req := &fuse.LookupRequest{Name: ctx.it}
	_, err := ctx.c.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Equal(t, err, fuse.ENOENT)
}

func testContainerReadDirAll(t *testing.T) {
	entries, err := ctx.c.ReadDirAll(nil)
	assert.Nil(t, err)
	assert.Len(t, entries, ctx.rc)
}

func testContainerRename(t *testing.T) {
	req := &fuse.RenameRequest{OldName: directoryName, NewName: "foo"}
	err := ctx.c.Rename(nil, req, ctx.c)
	assert.Equal(t, err, fuse.ENOTSUP)
}

func testContainerRmdir(t *testing.T) {
	req := &fuse.RemoveRequest{Name: ctx.d.Name(), Dir: true}
	assert.Nil(t, ctx.c.Remove(nil, req))
}

func testDirectoryCreate(t *testing.T) {
	req := &fuse.CreateRequest{Name: ctx.it}
	obj, fh, err := ctx.d.Create(nil, req, &fuse.CreateResponse{})
	assert.Nil(t, err)
	require.IsType(t, &Object{}, obj)
	require.IsType(t, &ObjectHandle{}, fh)
	ctx.f, _ = obj.(*Object)
	ctx.h, _ = fh.(*ObjectHandle)
	ctx.h.Release(nil, nil)
}

func testDirectoryHardlink(t *testing.T) {
	req := &fuse.LinkRequest{NewName: hardlinkName}
	link, err := ctx.d.Link(nil, req, ctx.f)
	assert.Nil(t, err)
	require.IsType(t, &Object{}, link)
	ctx.l, _ = link.(*Object)
}

func testDirectoryLookup(t *testing.T) {
	req := &fuse.LookupRequest{Name: ctx.it}
	file, err := ctx.d.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Nil(t, err)
	require.IsType(t, &Object{}, file)
	f, _ := file.(*Object)
	assert.Equal(t, f.Name(), req.Name)
}

func testDirectoryLookupHardlink(t *testing.T) {
	req := &fuse.LookupRequest{Name: ctx.it}
	file, err := ctx.d.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Nil(t, err)
	require.IsType(t, &Object{}, file)
	s, _ := file.(*Object)
	assert.Equal(t, s.Name(), req.Name)
}

func testDirectoryLookupSymlink(t *testing.T) {
	req := &fuse.LookupRequest{Name: ctx.it}
	file, err := ctx.d.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Nil(t, err)
	require.IsType(t, &Symlink{}, file)
	s, _ := file.(*Symlink)
	assert.Equal(t, s.Name(), req.Name)
}

func testDirectoryLookupMiss(t *testing.T) {
	req := &fuse.LookupRequest{Name: ctx.it}
	_, err := ctx.d.Lookup(nil, req, &fuse.LookupResponse{})
	assert.Equal(t, err, fuse.ENOENT)
}

func testDirectoryReadDirAll(t *testing.T) {
	entries, err := ctx.d.ReadDirAll(nil)
	assert.Nil(t, err)
	assert.Len(t, entries, ctx.rc)
}

func testDirectoryRemove(t *testing.T) {
	req := &fuse.RemoveRequest{Name: ctx.it}
	assert.Nil(t, ctx.d.Remove(nil, req))
}

func testDirectoryRename(t *testing.T) {
	req := &fuse.RenameRequest{OldName: oldFileName, NewName: newFileName}
	assert.Nil(t, ctx.d.Rename(nil, req, ctx.d))
}

func testDirectorySymlink(t *testing.T) {
	req := &fuse.SymlinkRequest{Target: oldFileName, NewName: symlinkName}
	sym, err := ctx.d.Symlink(nil, req)
	assert.Nil(t, err)
	require.IsType(t, &Symlink{}, sym)
	ctx.s, _ = sym.(*Symlink)
}
