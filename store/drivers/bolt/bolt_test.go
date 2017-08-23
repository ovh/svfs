package bolt

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ovh/svfs/store"
	"github.com/ovh/svfs/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_ "github.com/ovh/svfs/store/drivers"
)

type BoltTestSuite struct {
	suite.Suite
	store store.Store
}

func (suite *BoltTestSuite) SetupTest() {
	storeName := fmt.Sprintf("%d", time.Now().UnixNano())
	storePath := os.TempDir() + "/" + storeName
	suite.store = &Bolt{}
	suite.store.Init(storePath)
}

func (suite *BoltTestSuite) TearDownTest() {
	suite.store.Close()
	os.Remove(suite.store.Path())
}

func (suite *BoltTestSuite) TestAppend() {
	namespace := "append"
	suite.store.Prepare(namespace)
	id, err := suite.store.Append(namespace, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint64(1), id)
}

func (suite *BoltTestSuite) TestDelete() {
	namespace := "delete"
	suite.store.Prepare(namespace)
	id, _ := suite.store.Append(namespace, []byte("item"))
	err := suite.store.Delete(namespace, util.Uint64Bytes(id))
	assert.NoError(suite.T(), err)
	storedVal, _ := suite.store.Get(namespace, util.Uint64Bytes(id))
	assert.Nil(suite.T(), storedVal)
}

func (suite *BoltTestSuite) TestGet() {
	namespace := "get"
	value := []byte("item")
	suite.store.Prepare(namespace)
	id, _ := suite.store.Append(namespace, value)
	storedVal, err := suite.store.Get(namespace, util.Uint64Bytes(id))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), value, storedVal)
}

func (suite *BoltTestSuite) TestPrepare() {
	err := suite.store.Prepare("prepare")
	assert.NoError(suite.T(), err)
}

func (suite *BoltTestSuite) TestSave() {
	namespace := "save"
	k := []byte("key")
	v := []byte("value")
	suite.store.Prepare(namespace)
	err := suite.store.Save(namespace, k, v)
	assert.NoError(suite.T(), err)
	storedVal, _ := suite.store.Get(namespace, k)
	assert.Equal(suite.T(), v, storedVal)
}

func TestRunBoltSuite(t *testing.T) {
	suite.Run(t, new(BoltTestSuite))
}
