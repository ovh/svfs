package swift

import (
	"testing"

	"github.com/ovh/svfs/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	lib "github.com/xlucas/swift"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

const (
	containerName = "container"
)

var (
	con *lib.Connection
)

type ConnectionTestSuite struct {
	suite.Suite
	con  *Connection
	list ContainerList
}

func (suite *ConnectionTestSuite) SetupSuite() {
	httpmock.Activate()
	suite.con = &Connection{
		Connection: &lib.Connection{
			AuthToken:  MockedToken,
			StorageUrl: MockedStorageURL,
			Transport:  httpmock.DefaultTransport,
		},
		StoragePolicy: "Policy1",
	}
	suite.list = ContainerList{
		"container": &Container{
			&lib.Container{
				Name:  "container",
				Bytes: 1024,
				Count: 100,
			},
			lib.Headers{
				StoragePolicyHeader: "Policy1",
			},
		},
	}
}

func (suite *ConnectionTestSuite) TearDownSuite() {
	httpmock.Deactivate()
}

func (suite *ConnectionTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *ConnectionTestSuite) TestCreateContainers() {
	for _, code := range []int{201, 400} {
		MockAccount(nil, suite.list, StatusMap{"PUT": code})

		err := suite.con.createContainers([]string{containerName})

		if code > 300 {
			assert.NotNil(suite.T(), err)
		}
		if code >= 200 && code < 300 {
			assert.Nil(suite.T(), err)
		}
	}

}

func (suite *ConnectionTestSuite) TestGetContainers() {
	for _, list := range []ContainerList{suite.list, ContainerList{}} {
		MockAccount(nil, list, StatusMap{"GET": 200})

		containers, err := suite.con.getContainers()
		assert.Nil(suite.T(), err)

		if len(list) == 0 {
			assert.Len(suite.T(), containers, 0)
		} else {
			assert.Nil(suite.T(), err)
			assert.Len(suite.T(), containers, 1)
			assert.Equal(suite.T(), containerName, containers[containerName].Name)
			test.EqualInt64(suite.T(), 100, containers[containerName].Count)
			test.EqualInt64(suite.T(), 1024, containers[containerName].Bytes)
		}
	}
}

func (suite *ConnectionTestSuite) TestGetContainersFromNames() {
	for _, code := range []int{200, 404} {
		MockContainers(suite.list, StatusMap{"HEAD": code})

		list, err := suite.con.getContainersByNames([]string{containerName})
		assert.Nil(suite.T(), err)

		if code > 300 {
			assert.Len(suite.T(), list, 0)
		}
		if code >= 200 && code < 300 {
			assert.Len(suite.T(), list, 1)
			assert.Equal(suite.T(), containerName, list[containerName].Name)
			test.EqualInt64(suite.T(), 1024, list[containerName].Bytes)
			test.EqualInt64(suite.T(), 100, list[containerName].Count)
			assert.Equal(suite.T(),
				suite.con.StoragePolicy,
				list[containerName].Headers[StoragePolicyHeader],
			)
		}
	}
}

func (suite *ConnectionTestSuite) TestLogicalContainer() {
	suite.list[containerName+"_segments"] = &Container{
		Container: &lib.Container{
			Name:  containerName + "_segments",
			Bytes: 4096,
			Count: 500,
		},
		Headers: lib.Headers{
			StoragePolicyHeader: suite.con.StoragePolicy,
		},
	}

	for _, code := range []int{200, 404} {
		MockContainers(suite.list, StatusMap{"HEAD": code})

		container, err := suite.con.LogicalContainer(containerName)

		if code > 300 {
			assert.NotNil(suite.T(), err)
		}
		if code >= 200 && code < 300 {
			assert.Nil(suite.T(), err)

			assert.Equal(suite.T(),
				containerName, container.MainContainer.Name)
			assert.Equal(suite.T(),
				containerName+"_segments", container.SegmentContainer.Name)
		}
	}

}

func TestRunConnectionSuite(t *testing.T) {
	suite.Run(t, new(ConnectionTestSuite))
}
