package fs

type MountOption int
type MountOptions map[MountOption]interface{}

type FsStats struct {
	Files      uint64
	FilesFree  uint64
	Blocks     uint64
	BlockSize  uint64
	BlocksUsed uint64
	BlocksFree uint64
}

type Fs interface {
	Setup(opts MountOptions) error
	StatFs() (*FsStats, error)
	Root() (Directory, error)
}
