package swift

import (
	"testing"

	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	lib "github.com/xlucas/swift"
)

type AccountTestSuite struct {
	suite.Suite
	account *Account
}

func (suite *AccountTestSuite) SetupTest() {
	suite.account = &Account{
		&lib.Account{},
		lib.Headers{
			TimestampHeader: "1446048898.88226",
		},
	}
}

func (suite *AccountTestSuite) TestTimestamp() {
	secs, nsecs, err := suite.account.timestamp()
	assert.Nil(suite.T(), err)
	test.EqualInt64(suite.T(), 1446048898, secs)
	test.EqualInt64(suite.T(), 882260084, nsecs)
}

func TestRunAccountTestSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
