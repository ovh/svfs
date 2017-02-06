package swift

import (
	"testing"
	"time"

	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AccountTestSuite struct {
	suite.Suite
	ts *MockedTestSet
}

func (suite *AccountTestSuite) SetupTest() {
	suite.ts = NewMockedTestSet()
	suite.ts.Account.Headers[TimestampHeader] = "1446048898.88226"
}

func (suite *AccountTestSuite) TestCreationTime() {
	expected := time.Unix(1446048898, 882260084)
	actual := suite.ts.Account.CreationTime()
	assert.Equal(suite.T(), expected, actual)
}

func (suite *AccountTestSuite) TestTimestamp() {
	secs, nsecs, err := suite.ts.Account.timestamp()
	assert.Nil(suite.T(), err)
	test.EqualInt64(suite.T(), 1446048898, secs)
	test.EqualInt64(suite.T(), 882260084, nsecs)

	suite.ts.Account.Headers[TimestampHeader] = "invalid"
	_, _, err = suite.ts.Account.timestamp()
	assert.NotNil(suite.T(), err)
}

func TestRunAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
