package cache

import (
	"container/heap"
	"sync"
	"time"
)

type CacheItem struct {
	Key        string
	Expiration int64
}

type Cache struct {
	Cache      []CacheItem
	onEviction func(key string)
}

func (c *Cache) EvictExpired() {
	mu := sync.RWMutex{}
	defer mu.Unlock()

	f := c.Cache[0]
	for f.Expiration < time.Now().UnixMilli() {
		mu.Lock()
		_ = heap.Pop(c)
		f = c.Cache[0]
		c.onEviction(f.Key)
	}
}

func New(onEviction func(key string)) *Cache {
	c := &Cache{
		Cache:      make([]CacheItem, 2),
		onEviction: onEviction,
	}
	c.removeExpiredItems(1 * time.Second)
	return c
}
