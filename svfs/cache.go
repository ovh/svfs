package svfs

import "time"

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

func (c *Cache) Delete(path string) {
	if !c.content[path].temporary {
		c.nodeCount -= uint64(len(c.content[path].list))
	}
	delete(c.content, path)
}

func (c *Cache) Get(path string) []Node {
	v, found := c.content[path]

	// Not found
	if !found {
		return nil
	}

	// Increase access counter
	v.accessCount++

	// Found but expired
	if time.Now().After(v.cachingDate.Add(c.config.Timeout)) {
		defer c.Delete(path)
		return nil
	}

	if v.temporary ||
		(!(c.config.MaxAccess < 0) && v.accessCount == uint64(c.config.MaxAccess)) {
		defer c.Delete(path)
	}

	return v.list
}

func (c *Cache) Set(path string, list []Node) {
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

	c.content[path] = entry
}
