package cache

type CacheItem struct {
	Key        string
	Expiration int64
}

type Cache struct {
	Cache []CacheItem
}

func (c *Cache) Len() int { return len(c.Cache) }

func (c *Cache) Less(i, j int) bool {
	return c.Cache[i].Expiration < c.Cache[j].Expiration
}

func (c *Cache) Swap(i, j int) {
	c.Cache[i], c.Cache[j] = c.Cache[j], c.Cache[i]
}

func (c *Cache) Push(x any) {
	item := x.(CacheItem)
	c.Cache = append(c.Cache, item)
}

func (c *Cache) Pop() any {
	l := c.Cache[len(c.Cache)-1]
	c.Cache = c.Cache[:len(c.Cache)-1]

	return l
}

func New() *Cache {
	return &Cache{
		Cache: make([]CacheItem, 2),
	}
}
