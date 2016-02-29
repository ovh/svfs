package svfs

import (
	"fmt"
	"time"
)

type Cache struct {
	config    *CacheConfig
	content   map[string]*CacheValue
	nodeCount uint64
}

type CacheConfig struct {
	Timeout    time.Duration
	MaxEntries int64
	MaxAccess  int64
}

type CacheValue struct {
	date        time.Time
	accessCount uint64
	temporary   bool
	nodes       map[string]Node
}

func NewCache(cconf *CacheConfig) *Cache {
	return &Cache{
		config:  cconf,
		content: make(map[string]*CacheValue),
	}
}

func (c *Cache) key(container, path string) string {
	return fmt.Sprintf("%s:%s", container, path)
}

func (c *Cache) AddAll(container, path string, nodes map[string]Node) {
	entry := &CacheValue{
		date:  time.Now(),
		nodes: nodes,
	}

	if !(c.config.MaxEntries < 0) &&
		(c.nodeCount+uint64(len(nodes)) >= uint64(c.config.MaxEntries)) ||
		c.config.MaxAccess == 0 {
		entry.temporary = true
	} else {
		c.nodeCount += uint64(len(nodes))
	}

	c.content[c.key(container, path)] = entry
}

func (c *Cache) Delete(container, path, name string) {
	v, ok := c.content[c.key(container, path)]
	if !ok {
		return
	}
	delete(v.nodes, name)
}

func (c *Cache) DeleteAll(container, path string) {
	v, found := c.content[c.key(container, path)]
	if found &&
		!v.temporary {
		c.nodeCount -= uint64(len(c.content[c.key(container, path)].nodes))
		delete(c.content, c.key(container, path))
	}
}

func (c *Cache) Get(container, path, name string) Node {
	v, ok := c.content[c.key(container, path)]
	if !ok {
		return nil
	}
	return v.nodes[name]
}

func (c *Cache) GetAll(container, path string) map[string]Node {
	v, found := c.content[c.key(container, path)]

	// Not found
	if !found {
		return nil
	}

	// Increase access counter
	v.accessCount++

	// Found but expired
	if time.Now().After(v.date.Add(c.config.Timeout)) {
		defer c.DeleteAll(container, path)
		return nil
	}

	if v.temporary ||
		(!(c.config.MaxAccess < 0) && v.accessCount == uint64(c.config.MaxAccess)) {
		defer c.DeleteAll(container, path)
	}

	return v.nodes
}

func (c *Cache) Peek(container, path string) bool {
	v, found := c.content[c.key(container, path)]

	// Not found
	if !found {
		return false
	}

	// Found but expired
	if time.Now().After(v.date.Add(c.config.Timeout)) {
		return false
	}

	return true
}

func (c *Cache) Set(container, path, name string, node Node) {
	v, ok := c.content[c.key(container, path)]
	if !ok {
		return
	}
	v.nodes[name] = node
}
