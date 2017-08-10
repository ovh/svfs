package store

type Store interface {
	Append(namespace string, v []byte) (id uint64, err error)
	Close() error
	Delete(namespace string, k []byte) error
	Get(namespace string, k []byte) ([]byte, error)
	Path() string
	Prepare(namespace string) error
	Save(namespace string, k, v []byte) error
}
