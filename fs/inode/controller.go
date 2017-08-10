package inode

import "github.com/ovh/svfs/fs/store"

type Controller struct {
	namespace string
	store     store.Store
}

func (ctl *Controller) Allocate() (i Inode, err error) {
	id, err := ctl.store.Append(ctl.namespace, nil)
	if err != nil {
		return
	}

	return Inode(id), err
}

func (ctl *Controller) Reclaim(i Inode) error {
	return ctl.store.Delete(ctl.namespace, i.ToBytes())
}
