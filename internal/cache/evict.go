package cache

import "time"

func (c *Cache) removeExpiredItems(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			c.EvictExpired()
		}
	}()

}
