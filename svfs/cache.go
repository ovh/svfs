package svfs

import (
	"fmt"
	"time"
)

type Cache struct {
	config    *CacheConfig
	content   map[string]*CacheEntry
	nodeCount uint64
}

type CacheConfig struct {
	Timeout    time.Duration
	MaxEntries int64
	MaxAccess  int64
}

type CacheEntry struct {
	cachingDate time.Time
	accessCount uint64
	temporary   bool
	nodes       map[string]Node
}

func NewCache(cconf *CacheConfig) *Cache {
	return &Cache{
		config:  cconf,
		content: make(map[string]*CacheEntry),
	}
}

func (c *Cache) key(container, path string) string {
	return fmt.Sprintf("%s:%s", container, path)
}

func (c *Cache) AddAll(container, path string, nodes map[string]Node) {
	entry := &CacheEntry{
		cachingDate: time.Now(),
		nodes:       nodes,
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

func (c *Cache) DeleteAll(container, path string) {
	k, found := c.content[c.key(container, path)]
	if found &&
		!k.temporary {
		c.nodeCount -= uint64(len(c.content[c.key(container, path)].nodes))
		delete(c.content, c.key(container, path))
	}
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
	if time.Now().After(v.cachingDate.Add(c.config.Timeout)) {
		defer c.DeleteAll(container, path)
		return nil
	}

	if v.temporary ||
		(!(c.config.MaxAccess < 0) && v.accessCount == uint64(c.config.MaxAccess)) {
		defer c.DeleteAll(container, path)
	}

	return v.nodes
}

func (c *Cache) CheckGetAll(container, path string) bool {
	v, found := c.content[c.key(container, path)]

	// Not found
	if !found {
		return false
	}

	// Found but expired
	if time.Now().After(v.cachingDate.Add(c.config.Timeout)) {
		return false
	}

	return true
}

func (c *Cache) Delete(container, path, name string) {
	h, ok := c.content[c.key(container, path)]

	if !ok {
		return
	}

	delete(h.nodes, name)
}

func (c *Cache) Get(container, path, name string) Node {
	h, ok := c.content[c.key(container, path)]

	if !ok {
		return nil
	}

	v, _ := h.nodes[name]

	return v
}

func (c *Cache) Set(container, path, name string, node Node) {
	h, ok := c.content[c.key(container, path)]

	if !ok {
		return
	}

	h.nodes[name] = node
}
