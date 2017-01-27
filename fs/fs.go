package fs

type FsStats struct {
	Files      uint64
	Blocks     uint64
	BlockSize  uint64
	BlocksUsed uint64
	BlocksFree uint64
}

type Fs interface {
	StatFs() (*FsStats, error)
	Root() (Node, error)
}
