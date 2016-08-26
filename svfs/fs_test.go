package svfs

import (
	"os"
	"testing"
	"time"

	"bazil.org/fuse"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xlucas/swift"
)

var ctx = &Ctx{
	fs: &SVFS{},
}

type Ctx struct {
	fs  *SVFS         // File system
	r   *Root         // Mountpoint
	c   *Directory    // Container
	d   *Directory    // Directory
	f   *Object       // File
	h   *ObjectHandle // File handle
	s   *Symlink      // Symlink
	l   *Object       // Hardlink
	b   [4096]byte    // Data
	set bool          // Context state
	it  string        // Expected Lookup item name
	rc  int           // Expected ReadDirAll record count
}

func TestFs(t *testing.T) {
	t.Run("FsInit", testFsInit)
	t.Run("FsStat", testFsStatfs)
	t.Run("FsRoot", testFsRoot)
}

func testFsInit(t *testing.T) {
	if !SwiftConnection.Authenticated() {

		// Default options
		ExtraAttr = true
		TransferMode = false
		CacheMaxEntries = -1
		CacheMaxAccess = -1
		CacheTimeout = 15 * time.Minute
		SwiftConnection.Timeout = 5 * time.Minute
		SwiftConnection.ConnectTimeout = 15 * time.Second
		SegmentSize = 256 * (1 << 20)
		ReadAheadSize = 128 * (1 << 10)
		BlockSize = 4096
		ListerConcurrency = 20

		switch os.ExpandEnv("$SVFS_TEST_AUTH") {
		case "HUBIC":
			SwiftConnection = &swift.Connection{}
			HubicAuthorization = os.ExpandEnv("$SVFS_TEST_HUBIC_AUTH")
			HubicRefreshToken = os.ExpandEnv("$SVFS_TEST_HUBIC_TOKEN")
		case "OPENRC":
			SwiftConnection = &swift.Connection{
				AuthUrl:  os.ExpandEnv("$SVFS_TEST_AUTH_URL"),
				UserName: os.ExpandEnv("$SVFS_TEST_USERNAME"),
				ApiKey:   os.ExpandEnv("$SVFS_TEST_PASSWORD"),
				Tenant:   os.ExpandEnv("$SVFS_TEST_TENANT_NAME"),
				Region:   os.ExpandEnv("$SVFS_TEST_REGION_NAME"),
			}
		case "TOKEN":
			SwiftConnection = &swift.Connection{
				AuthToken:  os.ExpandEnv("$SVFS_TEST_AUTH_TOKEN"),
				StorageUrl: os.ExpandEnv("$SVFS_TEST_STORAGE_URL"),
			}
		}

		ctx.set = true
		assert.Nil(t, ctx.fs.Init())
	}
}

func testFsRoot(t *testing.T) {
	n, err := ctx.fs.Root()
	assert.Nil(t, err)
	require.IsType(t, &Root{}, n)
	ctx.r, _ = n.(*Root)
}

func testFsStatfs(t *testing.T) {
	assert.Nil(t, ctx.fs.Statfs(nil, &fuse.StatfsRequest{}, &fuse.StatfsResponse{}))
}
