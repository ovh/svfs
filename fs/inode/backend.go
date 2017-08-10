package inode

type Backend interface {
	Allocate() (Inode, error)
	Reclaim(Inode) error
}
