package fs

type File interface {
	Open(flags uint32) (FileHandle, error)
	Fsync() error
}
