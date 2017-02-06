package swift

import (
	"os"
	"testing"

	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"github.com/ovh/svfs/swift"
	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	suite.Suite
	containerNode *Container
	fs            *Fs
	ts            *swift.MockedTestSet
}

func (suite *ContainerTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *ContainerTestSuite) SetupTest() {
	httpmock.Reset()
	suite.ts = swift.NewMockedTestSet()
	suite.fs = NewMockedFs()
	suite.containerNode = &Container{suite.fs, suite.ts.Container}
}

func (suite *ContainerTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *ContainerTestSuite) TestGetAttr() {
	attr, err := suite.containerNode.GetAttr()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), attr.Ctime, suite.ts.Account.CreationTime())
	assert.Equal(suite.T(), attr.Mtime, suite.ts.Account.CreationTime())
	assert.Equal(suite.T(), os.ModeDir|suite.fs.conf.Perms, attr.Mode)
	test.EqualUint32(suite.T(), suite.fs.conf.Uid, attr.Uid)
	test.EqualUint32(suite.T(), suite.fs.conf.Gid, attr.Gid)
	test.EqualUint64(suite.T(), suite.fs.conf.BlockSize, attr.Size)
}

func TestRunContainerSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}
