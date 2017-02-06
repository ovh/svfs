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

func (suite *LogicalContainerTestSuite) TestNewLogicalContainer() {
	suite.ts.MockContainers(StatusMap{"HEAD": 200})

	for _, code := range []int{200, 400} {
		suite.ts.MockAccount(StatusMap{"PUT": code})

		container, err := NewLogicalContainer(suite.ts.Connection, "container")

		if code == 400 {
			assert.NotNil(suite.T(), err)
		}
		if code == 200 {
			assert.Nil(suite.T(), err)
			assert.EqualValues(suite.T(), suite.ts.Container, container)

		}
	}
}

func (suite *LogicalContainerTestSuite) TestBytes() {
	test.EqualInt64(suite.T(), 16384+32768, suite.ts.Container.Bytes())
}

func (suite *LogicalContainerTestSuite) TestCreationTime() {
	suite.ts.Container.MainContainer.Headers[TimestampHeader] = "1446048898.88226"
	suite.ts.Container.SegmentContainer.Headers[TimestampHeader] = "1446048897.88226"

	assert.Equal(suite.T(), time.Unix(1446048898, 882260084),
		suite.ts.Container.CreationTime())
}

func TestRunLogicalContainerTestSuite(t *testing.T) {
	suite.Run(t, new(LogicalContainerTestSuite))
}
