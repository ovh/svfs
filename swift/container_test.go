package swift

import (
	"testing"
	"time"

	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	suite.Suite
	ts *MockedTestSet
}

func (suite *ContainerTestSuite) SetupTest() {
	suite.ts = NewMockedTestSet()
}

func (suite *ContainerTestSuite) TestCreationTime() {
	expected := time.Unix(1446048898, 882260084)
	actual := suite.ts.Container.CreationTime()
	assert.Equal(suite.T(), expected, actual)
}

func (suite *ContainerTestSuite) TestTimestampSuccess() {
	secs, nsecs, err := suite.ts.Container.MainContainer.timestamp()
	assert.NoError(suite.T(), err)
	test.EqualInt64(suite.T(), 1446048898, secs)
	test.EqualInt64(suite.T(), 882260084, nsecs)
}

func (suite *ContainerTestSuite) TestTimestampFail() {
	suite.ts.Container.MainContainer.Headers[TimestampHeader] = "invalid"
	_, _, err := suite.ts.Container.MainContainer.timestamp()
	assert.Error(suite.T(), err)
}

func (suite *ContainerTestSuite) TestFilterByStoragePolicy() {
	segment := suite.ts.ContainerList[suite.ts.Container.SegmentContainer.Name]
	segment.Headers[StoragePolicyHeader] = "Policy2"

	filtered := suite.ts.ContainerList.FilterByStoragePolicy("Policy1")

	assert.Len(suite.T(), filtered, 1)
	assert.NotNil(suite.T(), filtered[suite.ts.Container.Name()])
}

func (suite *ContainerTestSuite) TestSelectHeaders() {
	suite.ts.Container.MainContainer.Headers["X-Container-Meta-1"] = "1"
	suite.ts.Container.MainContainer.Headers["X-Container-Meta-2"] = "2"
	suite.ts.Container.MainContainer.Headers["X-Container-Foo"] = "foo"
	suite.ts.Container.MainContainer.Headers["X-Container-Bar"] = "bar"

	headers := suite.ts.Container.MainContainer.SelectHeaders("X-Container-Meta")

	assert.Len(suite.T(), headers, 2)
	assert.Equal(suite.T(), "1", headers["X-Container-Meta-1"])
	assert.Equal(suite.T(), "2", headers["X-Container-Meta-2"])
}

func TestRunContainerSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}
