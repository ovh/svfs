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
	fs         *SVFS         // File system
	r          *Root         // Mountpoint
	c          *Directory    // Container
	d          *Directory    // Directory
	f          *Object       // File
	h          *ObjectHandle // File handle
	s          *Symlink      // Symlink
	l          *Object       // Hardlink
	b          [4096]byte    // Data
	storageURL string
	authToken  string
	set        bool   // Context state
	it         string // Expected Lookup item name
	rc         int    // Expected ReadDirAll record count
}

func TestFs(t *testing.T) {
	t.Run("Fs_InitHubic", testInitFsHubic)
	t.Run("Fs_Stat", testFsStatfs)
	t.Run("Fs_Root", testFsRoot)

	t.Run("Fs_InitOpenrc", testInitFsOpenrc)
	t.Run("Fs_Stat", testFsStatfs)
	t.Run("Fs_Root", testFsRoot)

	t.Run("Fs_InitToken", testInitFsToken)
	t.Run("Fs_Stat", testFsStatfs)
	t.Run("Fs_Root", testFsRoot)
}

func testFsInit(t *testing.T) {
	if !SwiftConnection.Authenticated() {
		setDefaultSettings()

		switch os.ExpandEnv("$SVFS_TEST_AUTH") {
		case "HUBIC":
			testInitFsHubic(t)
		case "OPENRC":
			testInitFsOpenrc(t)
		}

		ctx.set = true
	}
}

func setDefaultSettings() {
	// Default options
	ExtraAttr = true
	TransferMode = 0
	CacheMaxEntries = -1
	CacheMaxAccess = -1
	CacheTimeout = 15 * time.Minute
	SwiftConnection.Timeout = 5 * time.Minute
	SwiftConnection.ConnectTimeout = 15 * time.Second
	SegmentSize = 256 * (1 << 20)
	ReadAheadSize = 128 * (1 << 10)
	BlockSize = 4096
	ListerConcurrency = 20
}

func setTokenAuth() {
	SwiftConnection = &swift.Connection{
		AuthToken:  ctx.authToken,
		StorageUrl: ctx.storageURL,
	}
}

func setHubicAuth() {
	SwiftConnection = &swift.Connection{}
	HubicAuthorization = os.ExpandEnv("$SVFS_TEST_HUBIC_AUTH")
	HubicRefreshToken = os.ExpandEnv("$SVFS_TEST_HUBIC_TOKEN")
}

func setOpenrcAuth() {
	SwiftConnection = &swift.Connection{
		AuthUrl:  os.ExpandEnv("$SVFS_TEST_AUTH_URL"),
		UserName: os.ExpandEnv("$SVFS_TEST_USERNAME"),
		ApiKey:   os.ExpandEnv("$SVFS_TEST_PASSWORD"),
		Tenant:   os.ExpandEnv("$SVFS_TEST_TENANT_NAME"),
		Region:   os.ExpandEnv("$SVFS_TEST_REGION_NAME"),
	}
}

func testInitFsToken(t *testing.T) {
	setDefaultSettings()
	setTokenAuth()
	assert.Nil(t, ctx.fs.Init())
}

func testInitFsHubic(t *testing.T) {
	setDefaultSettings()
	setHubicAuth()
	assert.Nil(t, ctx.fs.Init())
}

func testInitFsOpenrc(t *testing.T) {
	setDefaultSettings()
	setOpenrcAuth()
	assert.Nil(t, ctx.fs.Init())
	ctx.storageURL = SwiftConnection.StorageUrl
	ctx.authToken = SwiftConnection.AuthToken
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
