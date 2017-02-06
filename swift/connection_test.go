package swift

import (
	"testing"

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
	ts *MockedTestSet
}

func (suite *ConnectionTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *ConnectionTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *ConnectionTestSuite) SetupTest() {
	suite.ts = NewMockedTestSet()
	httpmock.Reset()
}

func (suite *ConnectionTestSuite) TestAccount() {
	for _, code := range []int{201, 400} {
		suite.ts.MockAccount(StatusMap{"HEAD": code})
		account, err := suite.ts.Connection.Account()

		if code == 400 {
			assert.NotNil(suite.T(), err)
		}
		if code == 200 {
			assert.Nil(suite.T(), err)
			assert.Equal(suite.T(), suite.ts.Account, account)
		}
	}
}

func (suite *ConnectionTestSuite) TestCreateContainers() {
	for _, code := range []int{201, 400} {
		suite.ts.MockAccount(StatusMap{"PUT": code})

		err := suite.ts.Connection.createContainers([]string{containerName})

		if code == 404 {
			assert.NotNil(suite.T(), err)
		}
		if code == 200 {
			assert.Nil(suite.T(), err)
		}
	}

}

func (suite *ConnectionTestSuite) TestDeleteContainers() {
	for _, code := range []int{200, 404} {
		suite.ts.MockContainers(StatusMap{"DELETE": code})

		err := suite.ts.Connection.deleteContainers([]string{containerName})

		if code == 404 {
			assert.NotNil(suite.T(), err)
		}
		if code == 200 {
			assert.Nil(suite.T(), err)
		}
	}
}

func (suite *ConnectionTestSuite) TestDeleteLogicalContainer() {
	for _, code := range []int{200, 500} {
		suite.ts.MockContainers(StatusMap{"DELETE": code})

		err := suite.ts.Connection.DeleteLogicalContainer(suite.ts.Container)

		if code == 200 {
			assert.Nil(suite.T(), err)
		}
		if code == 500 {
			assert.NotNil(suite.T(), err)
		}
	}
}

func (suite *ConnectionTestSuite) TestGetContainers() {
	for _, code := range []int{200, 500} {
		for _, list := range []ContainerList{suite.ts.ContainerList, ContainerList{}} {
			suite.ts.ContainerList = list
			suite.ts.MockAccount(StatusMap{"GET": code})

			containers, err := suite.ts.Connection.getContainers()

			if code == 500 {
				assert.NotNil(suite.T(), err)
			}
			if code == 200 {
				assert.Nil(suite.T(), err)

				if len(list) == 0 {
					assert.Len(suite.T(), containers, 0)
				} else {
					assert.Nil(suite.T(), err)
					assert.Len(suite.T(), containers, 2)
					assert.EqualValues(suite.T(),
						suite.ts.ContainerList[containerName].Container,
						containers[containerName].Container)
				}
			}
		}
	}
}

func (suite *ConnectionTestSuite) TestGetContainersFromNames() {
	for _, code := range []int{200, 404, 500} {
		suite.ts.MockContainers(StatusMap{"HEAD": code})

		list, err := suite.ts.Connection.getContainersByNames(
			[]string{containerName},
		)

		if code == 404 {
			assert.Nil(suite.T(), err)
			assert.Len(suite.T(), list, 0)
		}
		if code == 500 {
			assert.NotNil(suite.T(), err)
		}
		if code == 200 {
			assert.Nil(suite.T(), err)
			assert.Len(suite.T(), list, 1)
			assert.EqualValues(suite.T(),
				suite.ts.ContainerList[containerName],
				list[containerName])
			assert.Equal(suite.T(),
				suite.ts.Connection.StoragePolicy,
				list[containerName].Headers[StoragePolicyHeader],
			)
		}
	}
}

func (suite *ConnectionTestSuite) TestLogicalContainer() {
	listOfTwo := suite.ts.ContainerList
	listOfOne := suite.ts.ContainerList
	listOfOne[containerName+"_segments"] = nil

	for _, accountCode := range []int{200, 500} {
		for _, containerCode := range []int{200, 404, 500} {
			for _, list := range []ContainerList{listOfTwo, listOfOne} {
				suite.ts.ContainerList = list
				suite.ts.MockAccount(StatusMap{"PUT": accountCode})
				suite.ts.MockContainers(StatusMap{"HEAD": containerCode})

				container, err := suite.ts.Connection.LogicalContainer(
					containerName)

				if containerCode == 404 ||
					containerCode == 500 ||
					accountCode == 500 {
					assert.NotNil(suite.T(), err)
					continue
				}
				if containerCode == 200 {
					assert.Nil(suite.T(), err)
					assert.Equal(suite.T(),
						containerName, container.MainContainer.Name)
					assert.Equal(suite.T(),
						containerName+"_segments",
						container.SegmentContainer.Name)
				}
			}
		}
	}
}

func TestRunConnectionSuite(t *testing.T) {
	suite.Run(t, new(ConnectionTestSuite))
}
