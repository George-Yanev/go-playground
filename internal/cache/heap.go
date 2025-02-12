package cache

type CacheItem struct {
	Key        string
	Expiration int64
}

type cacheHeap struct {
	items []CacheItem
}

func (c *cacheHeap) Len() int { return len(c.items) }

func (c *cacheHeap) Less(i, j int) bool {
	return c.items[i].Expiration < c.items[j].Expiration
}

func (c *cacheHeap) Swap(i, j int) {
	c.items[i], c.items[j] = c.items[j], c.items[i]
}

func (c *cacheHeap) Push(x any) {
	item := x.(CacheItem)
	c.items = append(c.items, item)
}

func (c *cacheHeap) Pop() any {
	l := c.items[len(c.items)-1]
	c.items = c.items[:len(c.items)-1]

	return l
}
