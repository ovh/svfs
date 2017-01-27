package fs

type FileHandle interface {
	Read(offset int64, size int) ([]byte, error)
	Write(offet int64, data []byte) error
	Close() error
}
