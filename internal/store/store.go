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
}

type Store struct {
	mu    sync.RWMutex
	items map[string]Item
	TTL   int64
	cache *cache.Cache
}

func (s *Store) Get(key string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if item, ok := s.items[key]; ok {
		if item.Expiration > time.Now().UnixMilli() {
			return item, ok
		}
		// if expired, delete it
		delete(s.items, key)
	}
	return Item{}, false
}

func (s *Store) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = Item{
		Value:      value,
		Expiration: time.Now().UnixMilli() + s.TTL,
	}
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, ok := s.items[key]; ok {
		if item.Expiration < time.Now().UnixMilli() {
			delete(s.items, key)
		}
	}
}

func New(ttl int64) *Store {
	s := &Store{
		items: make(map[string]Item),
		TTL:   ttl,
	}

	s.cache = cache.New(func(key string) { s.Delete(key) })
	return s
}
