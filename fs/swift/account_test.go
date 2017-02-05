package swift

import (
	"os"
	"testing"

	"github.com/ovh/svfs/swift"
	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	lib "github.com/xlucas/swift"
)

type AccountTestSuite struct {
	suite.Suite
	account     *swift.Account
	accountNode *Account
	container   *swift.LogicalContainer
	fs          *Fs
	list        swift.ContainerList
}

func (suite *AccountTestSuite) SetupSuite() {
	suite.account = &swift.Account{
		&lib.Account{
			Containers: 2,
			Objects:    1500,
			BytesUsed:  65536,
		},
		lib.Headers{swift.TimestampHeader: "1446048898.88226"},
	}
	suite.container = &swift.LogicalContainer{
		MainContainer: &swift.Container{
			Container: &lib.Container{
				Name:  "container",
				Bytes: 16384,
				Count: 200,
			},
			Headers: lib.Headers{
				swift.StoragePolicyHeader:  "Policy1",
				"X-Container-Bytes-Used":   "16384",
				"X-Container-Object-Count": "200",
			},
		},
		SegmentContainer: &swift.Container{
			Container: &lib.Container{
				Name:  "container_segments",
				Bytes: 32768,
				Count: 500,
			},
			Headers: lib.Headers{
				swift.StoragePolicyHeader:  "Policy1",
				"X-Container-Bytes-Used":   "32768",
				"X-Container-Object-Count": "500",
			},
		},
	}
	suite.list = swift.ContainerList{
		"container":          suite.container.MainContainer,
		"container_segments": suite.container.SegmentContainer,
	}
}

func (suite *AccountTestSuite) SetupTest() {
	suite.fs = new(Fs)
	suite.fs.conf = &FsConfiguration{
		BlockSize:     uint64(4096),
		Container:     "container",
		Connections:   uint32(1),
		StoragePolicy: "Policy1",
		Uid:           845,
		Gid:           820,
		Perms:         0700,
		OsStorageURL:  swift.MockedStorageURL,
		OsAuthToken:   swift.MockedToken,
	}
	suite.fs.storage = swift.NewMockedConnectionHolder(1,
		suite.fs.conf.StoragePolicy,
	)
	suite.accountNode = &Account{Fs: suite.fs, swiftAccount: suite.account}
}

func (suite *AccountTestSuite) TestGetAttr() {
	attr, err := suite.accountNode.GetAttr()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), attr.Ctime, suite.account.CreationTime())
	assert.Equal(suite.T(), attr.Mtime, suite.account.CreationTime())
	assert.Equal(suite.T(), attr.Mode, os.ModeDir|0700)
	test.EqualUint32(suite.T(), attr.Uid, 845)
	test.EqualUint32(suite.T(), attr.Gid, 820)
	test.EqualUint64(suite.T(), attr.Size, 4096)
}

func (suite *AccountTestSuite) TestMkdir() {
	swift.MockAccount(nil, suite.list, swift.StatusMap{"PUT": 200})
	swift.MockContainers(suite.list, swift.StatusMap{"HEAD": 200})

	dir, err := suite.accountNode.Mkdir("container")
	assert.Nil(suite.T(), err)
	assert.IsType(suite.T(), &Container{}, dir)
	container := dir.(*Container)
	assert.EqualValues(suite.T(), suite.container, container.swiftContainer)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(AccountTestSuite))
}
