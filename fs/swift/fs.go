package swift

import (
	"math"
	"os"
	"time"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"
	lib "github.com/xlucas/swift"
)

type FsConfiguration struct {
	BlockSize         uint64
	Perms             os.FileMode
	Gid               uint32
	Uid               uint32
	Size              uint64
	StoragePolicy     string
	Container         string
	Attributes        bool
	XAttributes       bool
	Connections       uint32
	OsAuthToken       string
	OsAuthURL         string
	OsUserName        string
	OsPassword        string
	OsStorageURL      string
	OsTenantName      string
	OsRegionName      string
	HubicAuthToken    string
	HubicRefreshToken string
}

type Fs struct {
	conf      *FsConfiguration
	mountTime time.Time
	storage   *swift.ResourceHolder
}

func (sfs *Fs) Setup(conf interface{}) (err error) {
	sfs.mountTime = time.Now()
	sfs.conf = conf.(*FsConfiguration)

	con := &lib.Connection{
		AuthUrl:    sfs.conf.OsAuthURL,
		ApiKey:     sfs.conf.OsPassword,
		UserName:   sfs.conf.OsUserName,
		Tenant:     sfs.conf.OsTenantName,
		StorageUrl: sfs.conf.OsStorageURL,
		AuthToken:  sfs.conf.OsAuthToken,
	}

	sfs.storage = swift.NewResourceHolder(sfs.conf.Connections,
		&swift.Connection{con, sfs.conf.StoragePolicy},
	)

	return con.Authenticate()
}

func (sfs *Fs) Root() (dir fs.Directory, err error) {
	account, container, err := sfs.getFsRoot()
	if err != nil {
		return
	}
	if container == nil {
		dir = &Account{Fs: sfs, swiftAccount: account}
	}
	if container != nil {
		dir = &Container{Fs: sfs, swiftContainer: container}
	}

	return
}

func (sfs *Fs) StatFs() (stats *fs.FsStats, err error) {
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

func (sfs *Fs) getFsRoot() (account *swift.Account,
	container *swift.LogicalContainer, err error,
) {
	con := sfs.storage.Borrow().(*swift.Connection)
	defer sfs.storage.Return()

	account, err = con.Account()
	if err != nil {
		return
	}
	if sfs.conf.Container != "" {
		container, err = con.LogicalContainer(sfs.conf.Container)
	}

	return
}

func (sfs *Fs) getFreeSpace(stats *fs.FsStats, account *swift.Account,
	container *swift.LogicalContainer,
) {
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

func (sfs *Fs) getQuotaFreeSpace(stats *fs.FsStats, account *swift.Account,
	container *swift.LogicalContainer,
) {
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

func (sfs *Fs) getUsage(stats *fs.FsStats, account *swift.Account,
	container *swift.LogicalContainer,
) {
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

func (sfs *Fs) getContainerUsage(stats *fs.FsStats,
	container *swift.LogicalContainer,
) {
	stats.BlocksUsed = uint64(container.Bytes()) / stats.BlockSize
	stats.FilesFree = math.MaxUint64 - uint64(container.MainContainer.Count)
}

var (
	_ fs.Fs = (*Fs)(nil)
)
