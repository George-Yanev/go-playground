package main

import (
	"fmt"
	"time"

	"github.com/George-Yanev/go-playground/internal/store"
)

func main() {
	store := store.New(1 * time.Second)
	store.Set("test", 64)
	store.Set("test", 64)
	store.Set("test2", 12)
	item, exists := store.Get("test")
	fmt.Printf("Before sleep - exists: %t, value: %+v\n", exists, item)
	time.Sleep(2 * time.Second)
	item, exists = store.Get("test")
	fmt.Printf("After sleep - exists: %t, value: %+v\n", exists, item)
}
