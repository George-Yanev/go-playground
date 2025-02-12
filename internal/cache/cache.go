package cache

import (
	"container/heap"
	"sync"
	"time"
)

type Cache interface {
	Add(key string, expiration int64)
}

type heapCache struct {
	mu         sync.Mutex
	heap       *cacheHeap
	onEviction func(key string)
}

var _ Cache = (*heapCache)(nil)

func (c *heapCache) Add(key string, expiration int64) {
	item := CacheItem{
		Key:        key,
		Expiration: expiration,
	}
	heap.Push(c.heap, item)
}

func (c *heapCache) EvictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixMilli()
	for int64(c.heap.Len()) > 0 {
		item := c.heap.items[0]
		if item.Expiration > now {
			break
		}
		heap.Pop(c.heap)
		c.onEviction(item.Key)
	}
}

func newHeapCache(onEviction func(key string)) *heapCache {
	c := &heapCache{
		heap: &cacheHeap{
			items: make([]CacheItem, 0),
		},
		onEviction: onEviction,
	}
	heap.Init(c.heap)
	c.removeExpiredItems(100 * time.Millisecond)
	return c

}

func New(onEviction func(key string)) Cache {
	return newHeapCache(onEviction)
}
