package swift

import (
	"math"
	"testing"

	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"github.com/ovh/svfs/fs"
	"github.com/ovh/svfs/swift"
	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	ctx "golang.org/x/net/context"
)

type FsTestSuite struct {
	suite.Suite
	fs    *Fs
	stats *fs.FsStats
	ts    *swift.MockedTestSet
	c     ctx.Context
}

func (suite *FsTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *FsTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *FsTestSuite) SetupTest() {
	httpmock.Reset()
	suite.ts = swift.NewMockedTestSet()
	suite.fs = NewMockedFs()
}

func (suite *FsTestSuite) TearDownTest() {
	suite.fs.Shutdown()
}

func (suite *FsTestSuite) TestRootAccountSuccess() {
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 200})

	suite.fs.conf.Container = ""
	root, err := suite.fs.Root()

	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), &Account{}, root)

	account := root.(*Account)
	assert.EqualValues(suite.T(), suite.ts.Account, account.swiftAccount)
}

func (suite *FsTestSuite) TestRootAccountFail() {
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 500})

	suite.fs.conf.Container = ""
	_, err := suite.fs.Root()

	assert.Error(suite.T(), err)
}

func (suite *FsTestSuite) TestRootExistingContainerSuccess() {
	suite.ts.MockAccount(swift.StatusMap{"GET": 200, "HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})
	suite.fs.conf.Container = suite.ts.Container.Name()

	root, err := suite.fs.Root()
	container := root.(*Container)

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), suite.ts.Container, container.swiftContainer)
}

func (suite *FsTestSuite) TestRootMissingSegmentContainerSuccess() {
	suite.ts.MockAccount(swift.StatusMap{"GET": 200, "HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	suite.ts.ContainerList[suite.ts.Container.SegmentContainer.Name] = nil
	suite.fs.conf.Container = suite.ts.Container.Name()
	root, err := suite.fs.Root()
	container := root.(*Container)

	assert.NoError(suite.T(), err)

	assert.EqualValues(suite.T(), suite.ts.Container, container.swiftContainer)
	assert.Equal(suite.T(), suite.ts.Container.SegmentContainer.Name,
		container.swiftContainer.SegmentContainer.Name)
	assert.Equal(suite.T(),
		suite.ts.Container.MainContainer.Headers[swift.StoragePolicyHeader],
		container.swiftContainer.SegmentContainer.Headers[swift.StoragePolicyHeader],
	)
}

func (suite *FsTestSuite) TestRootContainerFail() {
	suite.ts.MockAccount(swift.StatusMap{"GET": 200, "HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 500})

	suite.fs.conf.Container = suite.ts.Container.Name()
	_, err := suite.fs.Root()

	assert.Error(suite.T(), err)
}

func (suite *FsTestSuite) TestStatFsAccountNoQuotaSuccess() {
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	suite.fs.conf.Container = ""
	stats, err := suite.fs.StatFs(suite.c)

	assert.NoError(suite.T(), err)
	test.EqualUint64(suite.T(), math.MaxUint64, stats.Files)

	bSize := suite.fs.conf.BlockSize
	blocksUsed := uint64(suite.ts.Account.BytesUsed) / bSize
	filesFree := math.MaxUint64 - uint64(suite.ts.Account.Objects)
	assert.Equal(suite.T(), blocksUsed, stats.BlocksUsed)
	assert.Equal(suite.T(), filesFree, stats.FilesFree)

	blocks := math.MaxUint64 / bSize
	bfree := blocks - blocksUsed
	assert.Equal(suite.T(), blocks, stats.Blocks)
	assert.Equal(suite.T(), bfree, stats.BlocksFree)
}

func (suite *FsTestSuite) TestStatFsAccountQuotaSuccess() {
	suite.ts.Account.Quota = 163840
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	suite.fs.conf.Container = ""
	stats, err := suite.fs.StatFs(suite.c)
	bSize := suite.fs.conf.BlockSize

	assert.NoError(suite.T(), err)
	test.EqualUint64(suite.T(), math.MaxUint64, stats.Files)

	blocksUsed := uint64(suite.ts.Account.BytesUsed) / bSize
	filesFree := math.MaxUint64 - uint64(suite.ts.Account.Objects)
	assert.Equal(suite.T(), blocksUsed, stats.BlocksUsed)
	assert.Equal(suite.T(), filesFree, stats.FilesFree)

	bfree := uint64(suite.ts.Account.Quota-suite.ts.Account.BytesUsed) / bSize
	blocks := uint64(suite.ts.Account.Quota) / bSize
	assert.Equal(suite.T(), bfree, stats.BlocksFree)
	assert.Equal(suite.T(), blocks, stats.Blocks)

}

func (suite *FsTestSuite) TestStatFsAccountFail() {
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 500})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	_, err := suite.fs.StatFs(suite.c)

	assert.Error(suite.T(), err)
}

func (suite *FsTestSuite) TestStatFsContainerNoQuotaSuccess() {
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	stats, err := suite.fs.StatFs(suite.c)

	assert.NoError(suite.T(), err)
	test.EqualUint64(suite.T(), math.MaxUint64, stats.Files)

	bSize := suite.fs.conf.BlockSize
	blocksUsed := uint64(suite.ts.Container.Bytes()) / bSize
	filesFree := math.MaxUint64 - uint64(suite.ts.Container.MainContainer.Count)
	assert.Equal(suite.T(), blocksUsed, stats.BlocksUsed)
	assert.Equal(suite.T(), filesFree, stats.FilesFree)

	blocks := math.MaxUint64 / bSize
	bfree := blocks - blocksUsed
	assert.Equal(suite.T(), blocks, stats.Blocks)
	assert.Equal(suite.T(), bfree, stats.BlocksFree)
}

func (suite *FsTestSuite) TestStatFsContainerQuotaSuccess() {
	suite.ts.Account.Quota = 163840
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 200})

	stats, err := suite.fs.StatFs(suite.c)
	bSize := suite.fs.conf.BlockSize

	assert.NoError(suite.T(), err)
	test.EqualUint64(suite.T(), math.MaxUint64, stats.Files)

	blocksUsed := uint64(suite.ts.Container.Bytes()) / bSize
	filesFree := math.MaxUint64 - uint64(suite.ts.Container.MainContainer.Count)
	assert.Equal(suite.T(), blocksUsed, stats.BlocksUsed)
	assert.Equal(suite.T(), filesFree, stats.FilesFree)

	bfree := uint64(suite.ts.Account.Quota-suite.ts.Account.BytesUsed) / bSize
	blocks := bfree + blocksUsed
	assert.Equal(suite.T(), bfree, stats.BlocksFree)
	assert.Equal(suite.T(), blocks, stats.Blocks)
}

func (suite *FsTestSuite) TestStatFsContainerFail() {
	suite.ts.MockAccount(swift.StatusMap{"HEAD": 200})
	suite.ts.MockContainers(swift.StatusMap{"HEAD": 500})

	_, err := suite.fs.StatFs(suite.c)

	assert.Error(suite.T(), err)
}

func TestRunFsSuite(t *testing.T) {
	suite.Run(t, new(FsTestSuite))
}
