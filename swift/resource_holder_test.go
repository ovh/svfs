package swift

import (
	"testing"

	"github.com/ovh/svfs/util/test"
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
	resource := suite.holder.Borrow()
	assert.Equal(suite.T(), "resource", resource)
	test.EqualUint32(suite.T(), 1, suite.holder.borrows)
}

func (suite *ResourceHolderTestSuite) TestReturn() {
	suite.holder.Borrow()
	suite.holder.Return()
	test.EqualUint32(suite.T(), 0, suite.holder.borrows)
}

func TestRunResourceHolderSuite(t *testing.T) {
	suite.Run(t, new(ResourceHolderTestSuite))
}
