package cache

type CacheItem struct {
	Key        string
	Expiration int64
}

type Cache struct {
	Cache      []CacheItem
	onEviction func(key string)
}

func New(onEviction func(key string)) *Cache {
	return &Cache{
		Cache:      make([]CacheItem, 2),
		onEviction: onEviction,
	}
}
