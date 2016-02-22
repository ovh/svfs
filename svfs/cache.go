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
	list        []Node
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

func (c *Cache) Delete(container, path string) {
	if !c.content[c.key(container, path)].temporary {
		c.nodeCount -= uint64(len(c.content[c.key(container, path)].list))
	}
	delete(c.content, c.key(container, path))
}

func (c *Cache) Get(container, path string) []Node {
	v, found := c.content[c.key(container, path)]

	// Not found
	if !found {
		return nil
	}

	// Increase access counter
	v.accessCount++

	// Found but expired
	if time.Now().After(v.cachingDate.Add(c.config.Timeout)) {
		defer c.Delete(container, path)
		return nil
	}

	if v.temporary ||
		(!(c.config.MaxAccess < 0) && v.accessCount == uint64(c.config.MaxAccess)) {
		defer c.Delete(container, path)
	}

	return v.list
}

func (c *Cache) Set(container, path string, list []Node) {
	entry := &CacheEntry{
		cachingDate: time.Now(),
		list:        list,
	}

	if !(c.config.MaxEntries < 0) &&
		(c.nodeCount+uint64(len(list)) >= uint64(c.config.MaxEntries)) ||
		c.config.MaxAccess == 0 {
		entry.temporary = true
	} else {
		c.nodeCount += uint64(len(list))
	}

	c.content[c.key(container, path)] = entry
}
