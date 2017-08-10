package backend

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/ovh/svfs/fs/inode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BoltDBTestSuite struct {
	suite.Suite
	db      *bolt.DB
	backend inode.Backend
	bucket  string
}

func (suite *BoltDBTestSuite) SetupSuite() {
	dbName := fmt.Sprintf("%d", time.Now().UnixNano())
	dbPath := os.TempDir() + "/" + dbName
	suite.db, _ = bolt.Open(dbPath, 0600, nil)
	suite.bucket = "inodes"
	suite.backend, _ = NewBoltDB(suite.db, suite.bucket)
}

func (suite *BoltDBTestSuite) TearDownSuite() {
	suite.db.Close()
	os.Remove(suite.db.Path())
}

func (suite *BoltDBTestSuite) TestAllocate() {
	for _ = range []uint64{1, 2} {
		i, err := suite.backend.Allocate()
		assert.NoError(suite.T(), err)
		assert.NotZero(suite.T(), uint64(i))
	}
}

func (suite *BoltDBTestSuite) TestReclaim() {
	i, _ := suite.backend.Allocate()

	err := suite.backend.Reclaim(i)
	assert.NoError(suite.T(), err)

	suite.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(suite.bucket))
		assert.Empty(suite.T(), b.Get(i.ToBytes()))
		return nil
	})
}

func TestRunBoltDBSuite(t *testing.T) {
	suite.Run(t, new(BoltDBTestSuite))
}
