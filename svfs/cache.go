package svfs

import (
	"fmt"
	"time"
)

// CacheConfig is the cache configuration.
var (
	CacheTimeout    time.Duration
	CacheMaxEntries int64
	CacheMaxAccess  int64
)

// Cache holds a map of cache entries. Its size can be configured
// as well as cache entries access limit and expiration time.
type Cache struct {
	content   map[string]*CacheValue
	nodeCount uint64
}

// CacheValue is the representation of a cache entry.
// It tracks expiration date, access count and holds
// a parent node with its children. It can be set
// as temporary, meaning that it will be stored within
// the cache but evicted on first access.
type CacheValue struct {
	date        time.Time
	accessCount uint64
	temporary   bool
	node        Node
	nodes       map[string]Node
}

// NewCache creates a new cache
func NewCache() *Cache {
	return &Cache{
		content: make(map[string]*CacheValue),
	}
}

func (c *Cache) key(container, path string) string {
	return fmt.Sprintf("%s:%s", container, path)
}

// AddAll creates a new cache entry with the key container:path and a map of nodes
// as a value. Node represents the parent node type. If the cache entry count limit is
// reached, it will be marked as temporary thus evicted after one read.
func (c *Cache) AddAll(container, path string, node Node, nodes map[string]Node) {
	entry := &CacheValue{
		date:  time.Now(),
		node:  node,
		nodes: nodes,
	}

	if !(CacheMaxEntries < 0) &&
		(c.nodeCount+uint64(len(nodes)) >= uint64(CacheMaxEntries)) ||
		CacheMaxAccess == 0 {
		entry.temporary = true
	} else {
		c.nodeCount += uint64(len(nodes))
	}

	c.content[c.key(container, path)] = entry
}

// Delete removes a node from cache.
func (c *Cache) Delete(container, path, name string) {
	v, ok := c.content[c.key(container, path)]
	if !ok {
		return
	}
	delete(v.nodes, name)
}

// DeleteAll removes all nodes for the cache key container:path.
func (c *Cache) DeleteAll(container, path string) {
	v, found := c.content[c.key(container, path)]
	if found &&
		!v.temporary {
		c.nodeCount -= uint64(len(c.content[c.key(container, path)].nodes))
		delete(c.content, c.key(container, path))
	}
}

// Get retrieves a specific node from the cache. It returns nil if
// the cache key container:path is missing.
func (c *Cache) Get(container, path, name string) Node {
	v, ok := c.content[c.key(container, path)]
	if !ok {
		return nil
	}
	return v.nodes[name]
}

// GetAll retrieves all nodes for the cache key container:path. It returns
// the parent node and its children nodes. If the cache entry is not found
// or expired or access count exceeds the limit, both values will be nil.
func (c *Cache) GetAll(container, path string) (Node, map[string]Node) {
	v, found := c.content[c.key(container, path)]

	// Not found
	if !found {
		return nil, nil
	}

	// Increase access counter
	v.accessCount++

	// Found but expired
	if time.Now().After(v.date.Add(CacheTimeout)) {
		defer c.DeleteAll(container, path)
		return nil, nil
	}

	if v.temporary ||
		(!(CacheMaxAccess < 0) && v.accessCount == uint64(CacheMaxAccess)) {
		defer c.DeleteAll(container, path)
	}

	return v.node, v.nodes
}

// Peek checks if a valid cache entry belongs to container:path
// key without changing cache access count for this entry.
// Returns the parent node with the result.
func (c *Cache) Peek(container, path string) (Node, bool) {
	v, found := c.content[c.key(container, path)]

	// Not found
	if !found {
		return nil, false
	}

	// Found but expired
	if time.Now().After(v.date.Add(CacheTimeout)) {
		return nil, false
	}

	return v.node, true
}

// Set adds a specific node in cache, given a previous peek
// operation succeeded.
func (c *Cache) Set(container, path, name string, node Node) {
	v, ok := c.content[c.key(container, path)]
	if !ok {
		return
	}
	v.nodes[name] = node
}

// SimpleCache is a simplistic caching implementation
// only relying on a hashmap with basic functions.
type SimpleCache struct {
	changes map[string]Node
}

// NewSimpleCache creates a new simplistic cache.
func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		changes: make(map[string]Node),
	}
}

func (c *SimpleCache) key(container, path string) string {
	return fmt.Sprintf("%s:%s", container, path)
}

// Add pushes a new cache entry.
func (c *SimpleCache) Add(container, path string, node Node) {
	c.changes[c.key(container, path)] = node
}

// Exist checks whether a cache key exist or not.
func (c *SimpleCache) Exist(container, path string) bool {
	return c.changes[c.key(container, path)] != nil
}

// Get retrieves a cache entry for the given key.
func (c *SimpleCache) Get(container, path string) Node {
	return c.changes[c.key(container, path)]
}

// Remove pops the cache entry at this key.
func (c *SimpleCache) Remove(container, path string) {
	delete(c.changes, c.key(container, path))
}
