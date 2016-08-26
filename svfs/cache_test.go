package svfs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/xlucas/swift"
)

type ChangeCacheTestSuite struct {
	suite.Suite
	key  string
	item *Object
}

func (suite *ChangeCacheTestSuite) SetupTest() {
	// Reset cache
	changeCache = NewSimpleCache()

	// Sample data
	suite.item = &Object{
		name: "item",
		path: "dir/item",
		c: &swift.Container{
			Name: "container",
		},
	}
	suite.key = changeCache.key(suite.item.c.Name, suite.item.path)
}

func (suite *ChangeCacheTestSuite) TestAdd() {
	changeCache.Add(suite.item.c.Name, suite.item.path, suite.item)

	require.NotNil(suite.T(), changeCache.changes[suite.key])
	assert.Equal(suite.T(), changeCache.changes[suite.key], suite.item)
}

func (suite *ChangeCacheTestSuite) TestExist() {
	suite.TestAdd()

	assert.True(suite.T(), changeCache.Exist(suite.item.c.Name, suite.item.path))
}

func (suite *ChangeCacheTestSuite) TestGet() {
	suite.TestAdd()

	node := changeCache.Get(suite.item.c.Name, suite.item.path)

	require.NotNil(suite.T(), node)
	assert.Equal(suite.T(), node, suite.item)
}

func (suite *ChangeCacheTestSuite) TestRemove() {
	suite.TestAdd()

	changeCache.Remove(suite.item.c.Name, suite.item.path)

	assert.Nil(suite.T(), changeCache.changes[suite.key])
}

func TestChangeCacheTestSuite(t *testing.T) {
	suite.Run(t, new(ChangeCacheTestSuite))
}

type CacheTestSuite struct {
	suite.Suite
	nodes  map[string]Node
	key    string
	item1  *Object
	item2  *Object
	parent *Directory
}

func (suite *CacheTestSuite) SetupTest() {
	// Cache settings
	CacheTimeout = 5 * time.Minute
	CacheMaxEntries = -1
	CacheMaxAccess = -1

	// Reset cache
	directoryCache = NewCache()

	// Sample data
	suite.nodes = make(map[string]Node)
	suite.parent = &Directory{
		name: "dir",
		path: "dir/",
		c: &swift.Container{
			Name: "container",
		},
	}
	suite.item1 = &Object{name: "item1"}
	suite.item2 = &Object{name: "item2"}
	suite.key = directoryCache.key(suite.parent.c.Name, suite.parent.path)
	suite.nodes[suite.item1.Name()] = suite.item1
}

func (suite *CacheTestSuite) TestAddAll() {
	directoryCache.AddAll(suite.parent.c.Name, suite.parent.path, suite.parent, suite.nodes)

	assert.Equal(suite.T(), directoryCache.nodeCount, uint64(1))
	assert.Len(suite.T(), directoryCache.content[suite.key].nodes, 1)
}

func (suite *CacheTestSuite) TestGetAll() {
	suite.TestAddAll()

	cachedParent, cachedNodes := directoryCache.GetAll(suite.parent.c.Name, suite.parent.path)

	assert.Len(suite.T(), cachedNodes, 1)
	assert.IsType(suite.T(), &Object{}, cachedNodes[suite.item1.Name()])
	assert.IsType(suite.T(), &Directory{}, cachedParent)
}

func (suite *CacheTestSuite) TestDeleteAll() {
	suite.TestAddAll()

	directoryCache.DeleteAll(suite.parent.c.Name, suite.parent.path)

	assert.Nil(suite.T(), directoryCache.content[suite.key])
	assert.Len(suite.T(), directoryCache.content, 0)
	assert.Equal(suite.T(), directoryCache.nodeCount, uint64(0))
}

func (suite *CacheTestSuite) TestDelete() {
	suite.TestSet()

	var (
		entries   = directoryCache.content[suite.key].nodes
		nodeCount = len(entries)
	)

	for _, node := range []Node{suite.item1, suite.item2} {
		nodeCount--
		directoryCache.Delete(suite.parent.c.Name, suite.parent.path, node.Name())
		assert.Nil(suite.T(), entries[node.Name()])
		assert.Len(suite.T(), entries, nodeCount)
	}
}

func (suite *CacheTestSuite) TestGet() {
	suite.TestSet()

	for _, node := range []Node{suite.item1, suite.item2} {
		cached := directoryCache.Get(suite.parent.c.Name, suite.parent.path, node.Name())
		require.NotNil(suite.T(), cached)
		assert.Equal(suite.T(), cached, node)
	}
}

func (suite *CacheTestSuite) TestPeek() {
	suite.TestAddAll()

	parent, found := directoryCache.Peek(suite.parent.c.Name, suite.parent.path)

	assert.True(suite.T(), found)
	require.NotNil(suite.T(), parent)
	assert.Equal(suite.T(), parent, suite.parent)
}

func (suite *CacheTestSuite) TestSet() {
	suite.TestAddAll()

	directoryCache.Set(suite.parent.c.Name, suite.parent.path, suite.item2.name, suite.item2)

	require.NotNil(suite.T(), directoryCache.content[suite.key])
	assert.NotNil(suite.T(), directoryCache.content[suite.key].nodes[suite.item2.name])
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}
