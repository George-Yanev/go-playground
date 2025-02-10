package store

import (
	"sync"

	"github.com/George-Yanev/go-fun/internal/cache"
)

// Prune mechanism
// Min-Heap implementation

type Item struct {
	Value      interface{}
	Expiration int64
}

type Store struct {
	mu    sync.RWMutex
	items map[string]Item
	cache *cache.Cache
}

func (s *Store) Get(key string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[key]
	return item, ok
}

func (s *Store) Set(key string, value interface{}, expiration int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
}

func New(ttl int64) *Store {
	return &Store{
		items: make(map[string]Item),
		cache: cache.New(ttl),
	}
}
