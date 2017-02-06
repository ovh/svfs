package swift

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ResourceHolderTestSuite struct {
	suite.Suite
	holder *ResourceHolder
}

func (suite *ResourceHolderTestSuite) SetupTest() {
	suite.holder = NewResourceHolder(1, "resource")
}

func (suite *ResourceHolderTestSuite) TestBorrow() {
	borrows := suite.holder.borrows
	resource := suite.holder.Borrow()
	assert.Equal(suite.T(), "resource", resource)
	assert.Equal(suite.T(), borrows+1, suite.holder.borrows)
}

func (suite *ResourceHolderTestSuite) TestReturn() {
	borrows := suite.holder.borrows
	suite.holder.Return()
	assert.Equal(suite.T(), borrows-1, suite.holder.borrows)
}

func (suite *ResourceHolderTestSuite) TestOnReadLock() {
	defer func() {
		r := recover()
		assert.Equal(suite.T(), "read", r)
	}()
	suite.holder.onReadLock(func() { panic("read") })
}

func (suite *ResourceHolderTestSuite) TestOnWriteLock() {
	defer func() {
		r := recover()
		assert.Equal(suite.T(), "write", r)
	}()
	suite.holder.onWriteLock(func() { panic("write") })
}

func TestRunResourceHolderSuite(t *testing.T) {
	suite.Run(t, new(ResourceHolderTestSuite))
}
