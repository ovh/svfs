package swift

import (
	"testing"
	"time"

	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	lib "github.com/xlucas/swift"
)

type ContainerTestSuite struct {
	suite.Suite
	container     *Container
	containerList ContainerList
}

func (suite *ContainerTestSuite) SetupTest() {
	suite.container = &Container{
		Container: &lib.Container{
			Name: "container",
		},
		Headers: lib.Headers{
			"X-Container-Meta-1": "1",
			"X-Container-Meta-2": "2",
			"X-Container-Foo":    "foo",
			"X-Container-Bar":    "bar",
			TimestampHeader:      "1446048898.88226",
		},
	}
	suite.containerList = ContainerList{
		"container1": &Container{
			&lib.Container{
				Name: "container1",
			},
			lib.Headers{
				StoragePolicyHeader: "Policy1",
			},
		},
		"container2": &Container{
			&lib.Container{
				Name: "container2",
			},
			lib.Headers{
				StoragePolicyHeader: "Policy2",
			},
		},
	}
}

func (suite *ContainerTestSuite) TestCreationTime() {
	expected := time.Unix(1446048898, 882260084)
	actual := suite.container.CreationTime()
	assert.Equal(suite.T(), expected, actual)
}

func (suite *ContainerTestSuite) TestTimestamp() {
	// Valid timestamp
	secs, nsecs, err := suite.container.timestamp()
	assert.Nil(suite.T(), err)
	test.EqualInt64(suite.T(), 1446048898, secs)
	test.EqualInt64(suite.T(), 882260084, nsecs)

	// Invalid timestamp
	suite.container.Headers[TimestampHeader] = "invalid"
	_, _, err = suite.container.timestamp()
	assert.NotNil(suite.T(), err)
}

func (suite *ContainerTestSuite) TestFilterByStoragePolicy() {
	filtered := suite.containerList.FilterByStoragePolicy("Policy1")
	assert.Len(suite.T(), filtered, 1)
	assert.NotNil(suite.T(), filtered["container1"])
}

func (suite *ContainerTestSuite) TestSelectHeaders() {
	headers := suite.container.SelectHeaders("X-Container-Meta")
	assert.Len(suite.T(), headers, 2)
	assert.Equal(suite.T(), "1", headers["X-Container-Meta-1"])
	assert.Equal(suite.T(), "2", headers["X-Container-Meta-2"])
}

func TestRunContainerSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}
