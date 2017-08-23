package bolt

import (
	"github.com/boltdb/bolt"
	"github.com/ovh/svfs/driver"
	"github.com/ovh/svfs/util"
)

func init() {
	driver.GetGroup("store").Register((*Bolt)(nil))
}

type Bolt struct {
	db *bolt.DB
}

func (b *Bolt) Init(path string) (err error) {
	b.db, err = bolt.Open(path, 0600, nil)
	return err
}

func (b *Bolt) Append(namespace string, v []byte) (id uint64, err error) {
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(namespace))
		id, err = bucket.NextSequence()
		if err != nil {
			return err
		}

		return bucket.Put(util.Uint64Bytes(id), v)
	})

	return id, err
}

func (b *Bolt) Close() error {
	return b.db.Close()
}

func (b *Bolt) Delete(namespace string, k []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(namespace)).Delete(k)
	})

}

func (b *Bolt) Get(namespace string, k []byte) (v []byte, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		v = tx.Bucket([]byte(namespace)).Get(k)
		return nil
	})
	return v, err
}

func (b *Bolt) Path() string {
	return b.db.Path()
}

func (b *Bolt) Prepare(namespace string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(namespace))
		return err
	})
}

func (b *Bolt) Save(namespace string, k, v []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(namespace)).Put(k, v)
	})

}
