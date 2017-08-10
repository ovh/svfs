package swift

import "github.com/ovh/svfs/swift"

func NewMockedFs() *Fs {
	fs := new(Fs)
	fs.conf = &Configuration{
		BlockSize:     uint64(4096),
		Container:     "container_1",
		MaxConn:       uint32(1),
		StoragePolicy: "Policy1",
		Uid:           845,
		Gid:           820,
		Perms:         0700,
		OsStorageURL:  swift.MockedStorageURL,
		OsAuthToken:   swift.MockedToken,
	}
	fs.storage = swift.NewMockedConnectionHolder(1,
		fs.conf.StoragePolicy,
	)
	return fs
}
