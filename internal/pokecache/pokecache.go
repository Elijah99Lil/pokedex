package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt	time.Time
	val 		[]byte	
}

type Cache struct {
	mu			sync.Mutex
	entries		map[string]cacheEntry
	interval	time.Duration
}


func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries:  make(map[string]cacheEntry),
        interval: interval,
	}
	go c.reapLoop()
	return c
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	c.entries[key] = cacheEntry{createdAt: time.Now(), val: val}
	defer c.mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	entry, ok := c.entries[key]
	c.mu.Unlock()
	if !ok{
		return nil, false
	}
	return entry.val, true
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
		for range ticker.C {
			c.mu.Lock()
			for key, entry := range c.entries {
				if time.Since(entry.createdAt) > c.interval {
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
		}
}




