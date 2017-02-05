package swift

import (
	"testing"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"
	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	lib "github.com/xlucas/swift"
)

type FsTestSuite struct {
	suite.Suite
	account   *swift.Account
	container *swift.LogicalContainer
	fs        *Fs
	list      swift.ContainerList
	stats     *fs.FsStats
}

func (suite *FsTestSuite) SetupSuite() {
	suite.account = &swift.Account{
		&lib.Account{
			Containers: 2,
			Objects:    1500,
			BytesUsed:  65536,
		},
		lib.Headers{},
	}
	suite.container = &swift.LogicalContainer{
		MainContainer: &swift.Container{
			Container: &lib.Container{
				Name:  "container",
				Bytes: 16384,
				Count: 200,
			},
			Headers: lib.Headers{swift.StoragePolicyHeader: "PCS"},
		},
		SegmentContainer: &swift.Container{
			Container: &lib.Container{
				Name:  "container_segments",
				Bytes: 32768,
				Count: 500,
			},
			Headers: lib.Headers{swift.StoragePolicyHeader: "PCS"},
		},
	}
	suite.list = swift.ContainerList{
		"container":          suite.container.MainContainer,
		"container_segments": suite.container.SegmentContainer,
	}
}

func (suite *FsTestSuite) SetupTest() {
	suite.stats = &fs.FsStats{
		BlockSize: 4096,
	}
	suite.fs = new(Fs)
	suite.fs.conf = &FsConfiguration{
		BlockSize:     uint64(4096),
		Container:     "container",
		Connections:   uint32(1),
		StoragePolicy: "PCS",
		OsStorageURL:  swift.MockedStorageURL,
		OsAuthToken:   swift.MockedToken,
	}
	suite.fs.storage = swift.NewMockedConnectionHolder(1,
		suite.fs.conf.StoragePolicy,
	)
}

func (suite *FsTestSuite) TestGetUsage() {
	// File system with the container mount option
	suite.fs.getUsage(suite.stats, suite.account, suite.container)
	test.EqualUint64(suite.T(), 12, suite.stats.BlocksUsed)

	// File system without the container mount option
	suite.fs.conf.Container = ""
	suite.fs.getUsage(suite.stats, suite.account, suite.container)
	test.EqualUint64(suite.T(), 16, suite.stats.BlocksUsed)
}

func (suite *FsTestSuite) TestGetFreeSpace() {
	// File system with the container mount option
	suite.account.Quota = 0
	suite.fs.getUsage(suite.stats, suite.account, suite.container)
	suite.fs.getFreeSpace(suite.stats, suite.account, suite.container)
	test.EqualUint64(suite.T(), 4503599627370495, suite.stats.Blocks)
	test.EqualUint64(suite.T(), 4503599627370483, suite.stats.BlocksFree)
	test.EqualUint64(suite.T(), 18446744073709551615, suite.stats.Files)
	test.EqualUint64(suite.T(), 18446744073709551415, suite.stats.FilesFree)

	suite.account.Quota = 163840
	suite.fs.getUsage(suite.stats, suite.account, suite.container)
	suite.fs.getFreeSpace(suite.stats, suite.account, suite.container)
	test.EqualUint64(suite.T(), 36, suite.stats.Blocks)
	test.EqualUint64(suite.T(), 24, suite.stats.BlocksFree)
	test.EqualUint64(suite.T(), 18446744073709551615, suite.stats.Files)
	test.EqualUint64(suite.T(), 18446744073709551415, suite.stats.FilesFree)

	// Filesystem without the container mount option
	suite.fs.conf.Container = ""

	suite.account.Quota = 0
	suite.fs.getUsage(suite.stats, suite.account, suite.container)
	suite.fs.getFreeSpace(suite.stats, suite.account, suite.container)
	test.EqualUint64(suite.T(), 4503599627370495, suite.stats.Blocks)
	test.EqualUint64(suite.T(), 4503599627370479, suite.stats.BlocksFree)
	test.EqualUint64(suite.T(), 18446744073709551615, suite.stats.Files)
	test.EqualUint64(suite.T(), 18446744073709550115, suite.stats.FilesFree)

	suite.account.Quota = 163840
	suite.fs.getUsage(suite.stats, suite.account, suite.container)
	suite.fs.getFreeSpace(suite.stats, suite.account, suite.container)
	test.EqualUint64(suite.T(), 40, suite.stats.Blocks)
	test.EqualUint64(suite.T(), 24, suite.stats.BlocksFree)
	test.EqualUint64(suite.T(), 18446744073709551615, suite.stats.Files)
	test.EqualUint64(suite.T(), 18446744073709550115, suite.stats.FilesFree)

}

func (suite *FsTestSuite) TestGetFsRoot() {
	// Logical container exist already
	swift.MockAccount(suite.account, suite.list,
		swift.StatusMap{
			"GET":  200,
			"HEAD": 200,
		},
	)
	swift.MockContainers(suite.list, swift.StatusMap{"HEAD": 200})

	for _, option := range []string{"container", ""} {
		suite.fs.conf.Container = option
		account, container, err := suite.fs.getFsRoot()
		assert.Nil(suite.T(), err)
		assert.NotNil(suite.T(), account)
		if option != "" {
			assert.NotNil(suite.T(), container)
		}
	}

	// Logical container is missing the segment container
	suite.list["container_segments"] = nil
	suite.fs.conf.Container = "container"
	account, container, err := suite.fs.getFsRoot()
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), account)
	assert.NotNil(suite.T(), container)
	assert.Equal(suite.T(), "container_segments",
		container.SegmentContainer.Name)
	assert.Equal(suite.T(),
		container.MainContainer.Headers[swift.StoragePolicyHeader],
		container.SegmentContainer.Headers[swift.StoragePolicyHeader],
	)
}

func TestRunFsSuite(t *testing.T) {
	suite.Run(t, new(FsTestSuite))
}
