package swift

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

type ConnectionTestSuite struct {
	suite.Suite
	names   []string
	name    string
	segname string
	ts      *MockedTestSet
}

func (suite *ConnectionTestSuite) SetupSuite() {
	httpmock.Activate()
}

func (suite *ConnectionTestSuite) TearDownSuite() {
	httpmock.DeactivateAndReset()
}

func (suite *ConnectionTestSuite) SetupTest() {
	httpmock.Reset()
	suite.ts = NewMockedTestSet()
	suite.name = suite.ts.Container.Name()
	suite.names = []string{suite.name}
	suite.segname = suite.name + SegmentContainerSuffix
}

func (suite *ConnectionTestSuite) TestAccountSuccess() {
	suite.ts.MockAccount(StatusMap{"HEAD": 200})

	account, err := suite.ts.Connection.Account()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.ts.Account, account)
}

func (suite *ConnectionTestSuite) TestAccountFail() {
	suite.ts.MockAccount(StatusMap{"HEAD": 404})
	_, err := suite.ts.Connection.Account()
	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestCreateContainersSuccess() {
	suite.ts.MockAccount(StatusMap{"PUT": 201})
	err := suite.ts.Connection.createContainers(suite.names)
	assert.NoError(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestCreateContainersFail() {
	suite.ts.MockAccount(StatusMap{"PUT": 500})
	err := suite.ts.Connection.createContainers(suite.names)
	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestDeleteContainersSuccess() {
	suite.ts.MockContainers(StatusMap{"DELETE": 200})
	err := suite.ts.Connection.deleteContainers(suite.names)
	assert.NoError(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestDeleteContainersFail() {
	suite.ts.MockContainers(StatusMap{"DELETE": 500})
	err := suite.ts.Connection.deleteContainers(suite.names)
	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestDeleteLogicalContainerSuccess() {
	suite.ts.MockContainers(StatusMap{"DELETE": 200})
	err := suite.ts.Connection.DeleteLogicalContainer(suite.ts.Container)
	assert.NoError(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestDeleteLogicalContainerFail() {
	suite.ts.MockContainers(StatusMap{"DELETE": 500})
	err := suite.ts.Connection.DeleteLogicalContainer(suite.ts.Container)
	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestGetExistingContainersSuccess() {
	suite.ts.MockAccount(StatusMap{"GET": 200})

	containers, err := suite.ts.Connection.getContainers()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), containers, 2)
	assert.EqualValues(suite.T(), suite.ts.ContainerList[suite.name].Container,
		containers[suite.name].Container,
	)
}

func (suite *ConnectionTestSuite) TestGetIncompleteContainersSuccess() {
	suite.ts.ContainerList = ContainerList{}
	suite.ts.MockAccount(StatusMap{"GET": 200})

	containers, err := suite.ts.Connection.getContainers()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), containers, 0)
}

func (suite *ConnectionTestSuite) TestGetContainersFail() {
	suite.ts.MockAccount(StatusMap{"GET": 500})
	_, err := suite.ts.Connection.getContainers()
	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestGetContainersFromNamesSuccess() {
	suite.ts.MockContainers(StatusMap{"HEAD": 200})

	list, err := suite.ts.Connection.getContainersByNames(suite.names)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 1)
	assert.EqualValues(suite.T(), suite.ts.ContainerList[suite.name],
		list[suite.name])
	assert.Equal(suite.T(), suite.ts.Connection.StoragePolicy,
		list[suite.name].Headers[StoragePolicyHeader],
	)
}

func (suite *ConnectionTestSuite) TestGetNoContainersFromNamesSuccess() {
	suite.ts.MockContainers(StatusMap{"HEAD": 404})

	list, err := suite.ts.Connection.getContainersByNames(suite.names)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), list, 0)
}

func (suite *ConnectionTestSuite) TestGetContainersFromNamesFail() {
	suite.ts.MockContainers(StatusMap{"HEAD": 500})
	_, err := suite.ts.Connection.getContainersByNames(suite.names)
	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestLogicalContainerSuccess() {
	suite.ts.MockAccount(StatusMap{"PUT": 201})
	suite.ts.MockContainers(StatusMap{"HEAD": 200})

	c, err := suite.ts.Connection.LogicalContainer(suite.name)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.name, c.MainContainer.Name)
	assert.Equal(suite.T(), suite.segname, c.SegmentContainer.Name)
}

func (suite *ConnectionTestSuite) TestLogicalContainerMissingSegmentSuccess() {
	suite.ts.ContainerList[suite.segname] = nil
	suite.ts.MockAccount(StatusMap{"PUT": 201})
	suite.ts.MockContainers(StatusMap{"HEAD": 200})

	c, err := suite.ts.Connection.LogicalContainer(suite.name)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.name, c.MainContainer.Name)
	assert.Equal(suite.T(), suite.segname, c.SegmentContainer.Name)
}

func (suite *ConnectionTestSuite) TestLogicalContainerFailOnAccount() {
	suite.ts.ContainerList[suite.segname] = nil
	suite.ts.MockAccount(StatusMap{"PUT": 500})
	suite.ts.MockContainers(StatusMap{"HEAD": 200})

	_, err := suite.ts.Connection.LogicalContainer(suite.name)

	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestLogicalContainerFailOnContainers() {
	suite.ts.MockAccount(StatusMap{"PUT": 201})
	suite.ts.MockContainers(StatusMap{"HEAD": 500})

	_, err := suite.ts.Connection.LogicalContainer(suite.name)

	assert.Error(suite.T(), err)
}

func (suite *ConnectionTestSuite) TestLogicalContainerFailOnMainContainer() {
	suite.ts.MockAccount(StatusMap{"PUT": 201})
	suite.ts.MockContainers(StatusMap{"HEAD": 404})

	_, err := suite.ts.Connection.LogicalContainer(suite.name)

	assert.Error(suite.T(), err)
}

func TestRunConnectionSuite(t *testing.T) {
	suite.Run(t, new(ConnectionTestSuite))
}
