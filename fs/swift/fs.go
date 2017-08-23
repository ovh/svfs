package swift

import (
	"math"
	"os"
	"time"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"
	lib "github.com/xlucas/swift"

	ctx "golang.org/x/net/context"
)

type FsConfiguration struct {
	// Base settings
	BlockSize uint64
	Perms     os.FileMode
	Gid       uint32
	Uid       uint32
	Size      uint64

	// Swift settings
	Container     string
	StoragePolicy string
	Attributes    bool
	XAttributes   bool

	// Network settings
	MaxConn uint32

	// Openstack settings
	OsAuthToken  string
	OsAuthURL    string
	OsUserName   string
	OsPassword   string
	OsStorageURL string
	OsRegionName string
	OsTenantName string

	// Hubic settings
	HubicAuthToken    string
	HubicRefreshToken string

	// Store settings
	StorePath   string
	StoreDriver string
}

type Fs struct {
	conf      *FsConfiguration
	mountTime time.Time
	pool      *swift.ResourceHolder
}

func (sfs *Fs) Setup(c ctx.Context, conf interface{}) (err error) {
	sfs.mountTime = time.Now()
	sfs.conf = conf.(*FsConfiguration)

	con := &lib.Connection{
		AuthUrl:    sfs.conf.OsAuthURL,
		ApiKey:     sfs.conf.OsPassword,
		UserName:   sfs.conf.OsUserName,
		Tenant:     sfs.conf.OsTenantName,
		Region:     sfs.conf.OsRegionName,
		StorageUrl: sfs.conf.OsStorageURL,
		AuthToken:  sfs.conf.OsAuthToken,
	}

	sfs.pool = swift.NewResourceHolder(
		sfs.conf.MaxConn,
		&swift.Connection{con, sfs.conf.StoragePolicy},
	)

	if con.Authenticated() {
		return
	}

	return con.Authenticate()
}

func (sfs *Fs) Root() (dir fs.Directory, err error) {
	account, container, err := sfs.getFsRoot()
	if err != nil {
		return
	}
	if container == nil {
		dir = NewAccount(sfs, account)
	}
	if container != nil {
		dir = NewContainer(sfs, container)
	}

	return
}

func (sfs *Fs) Shutdown() error {
	return nil
}

func (sfs *Fs) StatFs(c ctx.Context) (stats *fs.FsStats, err error) {
	stats = &fs.FsStats{
		BlockSize: sfs.conf.BlockSize,
	}

	account, container, err := sfs.getFsRoot()
	if err != nil {
		return
	}

	sfs.getUsage(stats, account, container)
	sfs.getFreeSpace(stats, account, container)

	return
}

func (sfs *Fs) getFsRoot() (account *swift.Account, container *swift.LogicalContainer, err error) {
	con := sfs.pool.Borrow().(*swift.Connection)
	defer sfs.pool.Return()

	account, err = con.Account()
	if err != nil {
		return
	}
	if sfs.conf.Container != "" {
		container, err = con.LogicalContainer(sfs.conf.Container)
	}

	return
}

func (sfs *Fs) getFreeSpace(stats *fs.FsStats, account *swift.Account, container *swift.LogicalContainer) {
	// Device has "unlimited" inodes.
	stats.Files = math.MaxUint64

	if account.Quota > 0 {
		sfs.getQuotaFreeSpace(stats, account, container)
	} else {
		// Device has "unlimited" blocks.
		stats.Blocks = math.MaxUint64 / stats.BlockSize
		stats.BlocksFree = stats.Blocks - stats.BlocksUsed
	}
}

func (sfs *Fs) getQuotaFreeSpace(stats *fs.FsStats, account *swift.Account, container *swift.LogicalContainer) {
	quotaFreeSpace := uint64(account.Quota - account.BytesUsed)
	stats.BlocksFree = quotaFreeSpace / stats.BlockSize

	if sfs.conf.Container != "" {
		// Device block count is the sum of quota free blocks plus container
		// used blocks.
		stats.Blocks = stats.BlocksFree + stats.BlocksUsed
	} else {
		// Device block count equals quota block count.
		stats.Blocks = uint64(account.Quota) / stats.BlockSize
	}

}

func (sfs *Fs) getUsage(stats *fs.FsStats, account *swift.Account, container *swift.LogicalContainer) {
	if sfs.conf.Container != "" {
		sfs.getContainerUsage(stats, container)
	} else {
		sfs.getAccountUsage(stats, account)
	}
}

func (sfs *Fs) getAccountUsage(stats *fs.FsStats, account *swift.Account) {
	stats.BlocksUsed = uint64(account.BytesUsed) / stats.BlockSize
	stats.FilesFree = math.MaxUint64 - uint64(account.Objects)
}

func (sfs *Fs) getContainerUsage(stats *fs.FsStats, container *swift.LogicalContainer) {
	stats.BlocksUsed = uint64(container.Bytes()) / stats.BlockSize
	stats.FilesFree = math.MaxUint64 - uint64(container.MainContainer.Count)
}

var (
	_ fs.Fs = (*Fs)(nil)
)
