package swift

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	lib "github.com/xlucas/swift"
)

type ContainerTestSuite struct {
	suite.Suite
	container     *Container
	containerList ContainerList
}

func (suite *ContainerTestSuite) SetupSuite() {
	suite.container = &Container{
		Container: &lib.Container{
			Name: "container",
		},
		Headers: lib.Headers{
			"X-Container-Meta-1": "1",
			"X-Container-Meta-2": "2",
			"X-Container-Foo":    "foo",
			"X-Container-Bar":    "bar",
		},
	}
	suite.containerList = ContainerList{
		"container1": &Container{
			&lib.Container{
				Name: "container1",
			},
			lib.Headers{
				"X-Storage-Policy": "Policy1",
			},
		},
		"container2": &Container{
			&lib.Container{
				Name: "container2",
			},
			lib.Headers{
				"X-Storage-Policy": "Policy2",
			},
		},
	}
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
