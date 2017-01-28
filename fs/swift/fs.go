package swift

import (
	"math"
	"time"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"
	lib "github.com/xlucas/swift"
)

const (
	BlockSizeOption fs.MountOption = iota
	ContainerOption
	StoragePolicyOption
	UseFileAttributesOption
	UseFileXAttributesOption
	UseHubicTimesOption
	MaxConnections
	StorageUrlOption
	TokenOption
)

type Fs struct {
	mountTime time.Time
	options   *OptionHolder
	storage   *swift.ResourceHolder
}

func (sfs *Fs) Setup(opts fs.MountOptions) (err error) {
	sfs.mountTime = time.Now()
	sfs.options = NewOptionHolder(opts)

	sfs.storage = swift.NewResourceHolder(sfs.options.GetUint32(MaxConnections),
		&swift.Connection{
			&lib.Connection{
				StorageUrl: sfs.options.GetString(StorageUrlOption),
				AuthToken:  sfs.options.GetString(TokenOption),
			},
			sfs.options.GetString(StoragePolicyOption),
		},
	)
	return
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
		BlockSize: sfs.options.GetUint64(BlockSizeOption),
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
	if sfs.options.IsSet(ContainerOption) {
		container, err = con.LogicalContainer(
			sfs.options.GetString(ContainerOption),
		)
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

	if sfs.options.IsSet(ContainerOption) {
		// Device block count is the sum of quota free blocks plus container
		// used blocks.
		stats.Blocks = quotaFreeSpace/stats.BlockSize + stats.BlocksUsed
	} else {
		// Device block count equals quota block count.
		stats.Blocks = uint64(account.Quota) / stats.BlockSize
	}

}

func (sfs *Fs) getUsage(stats *fs.FsStats, account *swift.Account,
	container *swift.LogicalContainer,
) {
	if sfs.options.IsSet(ContainerOption) {
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
