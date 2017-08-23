package swift

import (
	"os"
	"syscall"
	"testing"

	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"github.com/ovh/svfs/swift"
	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

type AccountTestSuite struct {
	suite.Suite
	accountNode *Account
	fs          *Fs
	ts          *swift.MockedTestSet
	c           context.Context
}

func (suite *AccountTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *AccountTestSuite) SetupTest() {
	httpmock.Reset()
	suite.c = context.Background()
	suite.ts = swift.NewMockedTestSet()
	suite.fs = NewMockedFs()
	suite.accountNode = NewAccount(suite.fs, suite.ts.Account)
}

func (suite *AccountTestSuite) TearDownTest() {
	suite.fs.Shutdown()
}

func (suite *AccountTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *AccountTestSuite) TestCreate() {
	_, err := suite.accountNode.Create(suite.c, "file")
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func (suite *AccountTestSuite) TestGetAttr() {
	attr, err := suite.accountNode.GetAttr(suite.c)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), attr.Ctime, suite.ts.Account.CreationTime())
	assert.Equal(suite.T(), attr.Mtime, suite.ts.Account.CreationTime())
	assert.Equal(suite.T(), os.ModeDir|suite.fs.conf.Perms, attr.Mode)
	test.EqualUint32(suite.T(), suite.fs.conf.Uid, attr.Uid)
	test.EqualUint32(suite.T(), suite.fs.conf.Gid, attr.Gid)
	test.EqualUint64(suite.T(), suite.fs.conf.BlockSize, attr.Size)
}

func (suite *AccountTestSuite) TestHardlink() {
	err := suite.accountNode.Hardlink(suite.c, "container_1", "hardlink")
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func (suite *AccountTestSuite) TestMkdirSuccess() {
	suite.ts.MockAccount(swift.StatusMap{"PUT": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	dir, err := suite.accountNode.Mkdir(suite.c, "container_1")
	container := dir.(*Container)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), suite.ts.Container, container.swiftContainer)
}

func (suite *AccountTestSuite) TestMkdirFail() {
	suite.ts.MockAccount(swift.StatusMap{"PUT": 500})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	_, err := suite.accountNode.Mkdir(suite.c, "container_1")

	assert.Error(suite.T(), err)
}

func (suite *AccountTestSuite) TestRemoveSuccess() {
	suite.ts.MockContainers(swift.StatusMap{"DELETE": 200})
	containerNode := NewContainer(suite.fs, suite.ts.Container)

	err := suite.accountNode.Remove(suite.c, containerNode)

	assert.NoError(suite.T(), err)
}

func (suite *AccountTestSuite) TestRemoveFailOnContainer() {
	suite.ts.MockContainers(swift.StatusMap{"DELETE": 500})
	containerNode := NewContainer(suite.fs, suite.ts.Container)

	err := suite.accountNode.Remove(suite.c, containerNode)

	assert.Error(suite.T(), err)
}

func (suite *AccountTestSuite) TestRemoveFailOnNode() {
	suite.ts.MockContainers(swift.StatusMap{"DELETE": 500})
	accountNode := NewAccount(suite.fs, suite.ts.Account)

	err := suite.accountNode.Remove(suite.c, accountNode)

	assert.Error(suite.T(), err)
}

func (suite *AccountTestSuite) TestRename() {
	err := suite.accountNode.Rename(suite.c, nil, "newName", suite.accountNode)
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func (suite *AccountTestSuite) TestSymlink() {
	err := suite.accountNode.Symlink(suite.c, "container_1", "hardlink")
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func TestRunAccountSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
