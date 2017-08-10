package backend

import (
	"github.com/boltdb/bolt"
	"github.com/ovh/svfs/fs/inode"
)

type BoltDB struct {
	db         *bolt.DB
	bucketName []byte
}

func NewBoltDB(db *bolt.DB, bucketName string) (backend *BoltDB, err error) {
	bucketNameBin := []byte(bucketName)

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(bucketNameBin)
		return err
	})

	backend = &BoltDB{
		db:         db,
		bucketName: bucketNameBin,
	}

	return
}

func (b *BoltDB) Allocate() (i inode.Inode, err error) {
	err = b.db.Update(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket(b.bucketName)

		id, err := bucket.NextSequence()
		if err != nil {
			return
		}

		i = inode.Inode(id)

		return bucket.Put(i.ToBytes(), nil)
	})

	return
}

func (b *BoltDB) Reclaim(i inode.Inode) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b.bucketName)
		return bucket.Delete(i.ToBytes())
	})
}
