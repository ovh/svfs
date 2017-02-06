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
)

type AccountTestSuite struct {
	suite.Suite
	accountNode *Account
	fs          *Fs
	ts          *swift.MockedTestSet
}

func (suite *AccountTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *AccountTestSuite) SetupTest() {
	httpmock.Reset()
	suite.ts = swift.NewMockedTestSet()
	suite.fs = NewMockedFs()
	suite.accountNode = &Account{suite.fs, suite.ts.Account}
}

func (suite *AccountTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *AccountTestSuite) TestCreate() {
	_, err := suite.accountNode.Create("file")
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func (suite *AccountTestSuite) TestGetAttr() {
	attr, err := suite.accountNode.GetAttr()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), attr.Ctime, suite.ts.Account.CreationTime())
	assert.Equal(suite.T(), attr.Mtime, suite.ts.Account.CreationTime())
	assert.Equal(suite.T(), os.ModeDir|suite.fs.conf.Perms, attr.Mode)
	test.EqualUint32(suite.T(), suite.fs.conf.Uid, attr.Uid)
	test.EqualUint32(suite.T(), suite.fs.conf.Gid, attr.Gid)
	test.EqualUint64(suite.T(), suite.fs.conf.BlockSize, attr.Size)
}

func (suite *AccountTestSuite) TestHardlink() {
	err := suite.accountNode.Hardlink("container", "hardlink")
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func (suite *AccountTestSuite) TestMkdirSucces() {
	suite.ts.MockAccount(swift.StatusMap{"PUT": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	dir, err := suite.accountNode.Mkdir("container")
	container := dir.(*Container)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), suite.ts.Container, container.swiftContainer)
}

func (suite *AccountTestSuite) TestMkdirFail() {
	suite.ts.MockAccount(swift.StatusMap{"PUT": 500})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	_, err := suite.accountNode.Mkdir("container")

	assert.Error(suite.T(), err)
}

func (suite *AccountTestSuite) TestRemoveSuccess() {
	suite.ts.MockContainers(swift.StatusMap{"DELETE": 200})
	containerNode := &Container{Fs: suite.fs, swiftContainer: suite.ts.Container}

	err := suite.accountNode.Remove(containerNode)

	assert.NoError(suite.T(), err)
}

func (suite *AccountTestSuite) TestRemoveFailOnContainer() {
	suite.ts.MockContainers(swift.StatusMap{"DELETE": 500})
	containerNode := &Container{Fs: suite.fs, swiftContainer: suite.ts.Container}

	err := suite.accountNode.Remove(containerNode)

	assert.Error(suite.T(), err)
}

func (suite *AccountTestSuite) TestRemoveFailOnNode() {
	suite.ts.MockContainers(swift.StatusMap{"DELETE": 500})
	accountNode := &Account{Fs: suite.fs, swiftAccount: suite.ts.Account}

	err := suite.accountNode.Remove(accountNode)

	assert.Error(suite.T(), err)
}

func (suite *AccountTestSuite) TestRename() {
	err := suite.accountNode.Rename(nil, "newName", suite.accountNode)
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func (suite *AccountTestSuite) TestSymlink() {
	err := suite.accountNode.Symlink("container", "hardlink")
	assert.Equal(suite.T(), syscall.ENOTSUP, err)
}

func TestRunAccountSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
