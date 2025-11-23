package cache

import (
	"container/list"
	"sync"
	"time"
)

type CacheEntry struct {
	Key       string
	Value     any
	ExpiresAt time.Time
}

type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	lruList  *list.List
	mutex    sync.RWMutex
	hits     int64
	misses   int64
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

func (c *LRUCache) Get(key string) (any, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*CacheEntry)
		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			c.lruList.Remove(elem)
			delete(c.items, key)
			c.misses++
			return nil, false
		}
		c.lruList.MoveToFront(elem)
		c.hits++
		return entry.Value, true
	}
	c.misses++
	return nil, false
}

func (c *LRUCache) Set(key string, value any, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	if elem, ok := c.items[key]; ok {
		c.lruList.MoveToFront(elem)
		elem.Value.(*CacheEntry).Value = value
		elem.Value.(*CacheEntry).ExpiresAt = expiresAt
		return
	}

	if c.lruList.Len() >= c.capacity {
		oldest := c.lruList.Back()
		if oldest != nil {
			c.lruList.Remove(oldest)
			delete(c.items, oldest.Value.(*CacheEntry).Key)
		}
	}

	entry := &CacheEntry{Key: key, Value: value, ExpiresAt: expiresAt}
	elem := c.lruList.PushFront(entry)
	c.items[key] = elem
}

func (c *LRUCache) GetStats() (hits, misses int64, hitRate float64) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	total := c.hits + c.misses
	if total == 0 {
		return c.hits, c.misses, 0.0
	}
	return c.hits, c.misses, float64(c.hits) / float64(total) * 100
}

func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]*list.Element)
	c.lruList.Init()
	c.hits = 0
	c.misses = 0
}

// Size returns the current number of entries in the cache
func (c *LRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lruList.Len()
}

// ResetStats resets hit/miss counters without clearing cache
func (c *LRUCache) ResetStats() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.hits = 0
	c.misses = 0
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if elem, ok := c.items[key]; ok {
		c.lruList.Remove(elem)
		delete(c.items, key)
	}
}
