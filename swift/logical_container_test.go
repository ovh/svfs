package swift

import (
	"testing"
	"time"

	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LogicalContainerTestSuite struct {
	suite.Suite
	ts *MockedTestSet
}

func (suite *LogicalContainerTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *LogicalContainerTestSuite) SetupTest() {
	httpmock.Reset()
	suite.ts = NewMockedTestSet()
}

func (suite *LogicalContainerTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *LogicalContainerTestSuite) TestBytes() {
	mainBytes := suite.ts.Container.MainContainer.Bytes
	segmBytes := suite.ts.Container.SegmentContainer.Bytes
	test.EqualInt64(suite.T(), mainBytes+segmBytes, suite.ts.Container.Bytes())
}

func (suite *LogicalContainerTestSuite) TestCreationTime() {
	assert.Equal(suite.T(), time.Unix(1446048898, 882260084),
		suite.ts.Container.CreationTime())
}

func (suite *LogicalContainerTestSuite) TestName() {
	assert.Equal(suite.T(), suite.ts.Container.MainContainer.Name,
		suite.ts.Container.Name())
}

func (suite *LogicalContainerTestSuite) TestNewLogicalContainerSuccess() {
	suite.ts.MockContainers(StatusMap{"HEAD": 200})
	suite.ts.MockAccount(StatusMap{"PUT": 201})

	container, err := NewLogicalContainer(suite.ts.Connection, "container")

	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), suite.ts.Container, container)
}

func (suite *LogicalContainerTestSuite) TestNewLogicalContainerFail() {
	suite.ts.MockContainers(StatusMap{"HEAD": 200})
	suite.ts.MockAccount(StatusMap{"PUT": 500})

	_, err := NewLogicalContainer(suite.ts.Connection, "container")

	assert.Error(suite.T(), err)
}

func TestRunLogicalContainerTestSuite(t *testing.T) {
	suite.Run(t, new(LogicalContainerTestSuite))
}
