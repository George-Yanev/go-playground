package store

import (
	"testing"
	"time"
)

func TestStoreSetAndGet(t *testing.T) {
	store := New(100 * time.Millisecond)
	store.Set("key", "value")
	item, ok := store.Get("key")
	if !ok {
		t.Error("expected item to be found")
	}
	if item.Value != "value" {
		t.Errorf("expected value to be %q, got %q", "value", item.Value)
	}
	time.Sleep(150 * time.Millisecond)
	_, ok = store.Get("key")
	if ok {
		t.Error("expected item to be expired")
	}
}
