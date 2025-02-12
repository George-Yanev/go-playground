package store

import (
	"sync"
	"time"

	"github.com/George-Yanev/go-fun/internal/cache"
)

// Prune mechanism
// Min-Heap implementation

type Item struct {
	Value      interface{}
	Expiration int64
	Count      int
}

type Store struct {
	mu    sync.RWMutex
	items map[string]Item
	TTL   time.Duration
	cache cache.Cache
}

func (s *Store) Get(key string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if item, ok := s.items[key]; ok {
		s.setOrUpdate(key, item.Value)
		return item, ok
	}
	return Item{}, false
}

func (s *Store) setOrUpdate(key string, value interface{}) int64 {
	expiration := time.Now().UnixMilli() + s.TTL.Milliseconds()
	if item, ok := s.items[key]; !ok {
		s.items[key] = Item{
			Value:      value,
			Expiration: expiration,
		}
	} else {
		item.Expiration = expiration
		item.Count += 1
	}

	return expiration
}

func (s *Store) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiration := s.setOrUpdate(key, value)
	s.cache.Add(key, expiration)
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, ok := s.items[key]; ok {
		if item.Count > 0 {
			item.Count -= 1
			return
		}
		if item.Expiration < time.Now().UnixMilli() {
			delete(s.items, key)
		}
	}
}

func New(ttl time.Duration) *Store {
	s := &Store{
		items: make(map[string]Item),
		TTL:   ttl,
	}

	s.cache = cache.New(func(key string) { s.Delete(key) })
	return s
}
