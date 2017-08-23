// +build fixtures

package swift

import (
	"github.com/ovh/svfs/swift"
)

func NewMockedFs() *Fs {
	fs := new(Fs)

	conf := &FsConfiguration{
		BlockSize:     uint64(4096),
		Container:     "container_1",
		MaxConn:       uint32(1),
		StoragePolicy: "Policy1",
		Uid:           845,
		Gid:           820,
		Perms:         0700,
		OsStorageURL:  swift.MockedStorageURL,
		OsAuthToken:   swift.MockedToken,
		StoreDriver:   "Bolt",
		StorePath:     "/dev/shm/svfs-test.store",
	}

	err := fs.Setup(nil, conf)
	if err != nil {
		panic(err)
	}

	fs.pool = swift.NewMockedConnectionHolder(1,
		conf.StoragePolicy,
	)

	return fs
}
